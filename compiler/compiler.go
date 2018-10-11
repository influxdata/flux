package compiler

import (
	"errors"
	"fmt"
	"log"

	"github.com/influxdata/flux"

	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func Compile(f *semantic.FunctionExpression, functionType semantic.Type, builtins Scope) (Func, error) {
	if functionType.Kind() != semantic.Function {
		return nil, errors.New("type must be a function kind")
	}
	// TODO: Determine if we need to copy here:
	// If we do then we need to copy the type annotations so that copying preserves the type variable linking.
	//f = f.Copy().(*semantic.FunctionExpression)
	declarations := externDeclarations(builtins)
	extern := &semantic.Extern{
		Declarations: declarations,
		Block:        &semantic.ExternBlock{Node: f},
	}

	log.Println("Infer")
	semantic.Infer(extern)
	log.Println("Done Infer")

	pt, err := extern.PolyType()
	if err != nil {
		return nil, err
	}
	log.Println("poly type", pt)
	if err := pt.Unify(functionType.PolyType()); err != nil {
		return nil, err
	}
	fnType, mono := pt.Type()
	if !mono {
		return nil, errors.New("cannot compile polymorphic function")
	}
	log.Println("mono type", fnType)

	root, err := compile(f.Block.Body, builtins)
	if err != nil {
		return nil, err
	}
	return compiledFn{
		root:       root,
		fnType:     fnType,
		inputScope: make(Scope),
	}, nil
}

// monoType ignores any errors when reading the type of a node.
// This is safe becase we already validated that the function type is a mono type.
func monoType(t semantic.Type, _ error) semantic.Type {
	return t
}

func compile(n semantic.Node, builtIns Scope) (Evaluator, error) {
	switch n := n.(type) {
	case *semantic.BlockStatement:
		body := make([]Evaluator, len(n.Body))
		for i, s := range n.Body {
			node, err := compile(s, builtIns)
			if err != nil {
				return nil, err
			}
			body[i] = node
		}
		return &blockEvaluator{
			t:    monoType(n.ReturnStatement().Argument.Type()),
			body: body,
		}, nil
	case *semantic.ExpressionStatement:
		return nil, errors.New("statement does nothing, sideffects are not supported by the compiler")
	case *semantic.ReturnStatement:
		node, err := compile(n.Argument, builtIns)
		if err != nil {
			return nil, err
		}
		return returnEvaluator{
			Evaluator: node,
		}, nil
	case *semantic.NativeVariableDeclaration:
		node, err := compile(n.Init, builtIns)
		if err != nil {
			return nil, err
		}
		return &declarationEvaluator{
			t:    monoType(n.Init.Type()),
			id:   n.Identifier.Name,
			init: node,
		}, nil
	case *semantic.ObjectExpression:
		properties := make(map[string]Evaluator, len(n.Properties))
		propertyTypes := make(map[string]semantic.Type, len(n.Properties))
		for _, p := range n.Properties {
			node, err := compile(p.Value, builtIns)
			if err != nil {
				return nil, err
			}
			properties[p.Key.Name] = node
			propertyTypes[p.Key.Name] = node.Type()
		}
		return &objEvaluator{
			t:          semantic.NewObjectType(propertyTypes),
			properties: properties,
		}, nil
	case *semantic.IdentifierExpression:
		if v, ok := builtIns[n.Name]; ok {
			//Resolve any built in identifiers now
			return &valueEvaluator{
				value: v,
			}, nil
		}

		// TODO: How do we apply the instantiation at this stage?
		// Meaning the instatiate process decouples type variables so that each instance can have its own type.
		// Here in compliation we need to know how the type variables were linked so we can retrieve the monotype once we know the monotype of the instantiated type variables.
		t, err := n.Type()
		log.Printf("IdentifierExpression type: %v %v", t, err)
		return &identifierEvaluator{
			t:    monoType(n.Type()),
			name: n.Name,
		}, nil
	case *semantic.MemberExpression:
		object, err := compile(n.Object, builtIns)
		if err != nil {
			return nil, err
		}
		return &memberEvaluator{
			t:        monoType(n.Type()),
			object:   object,
			property: n.Property,
		}, nil
	case *semantic.BooleanLiteral:
		return &booleanEvaluator{
			t: monoType(n.Type()),
			b: n.Value,
		}, nil
	case *semantic.IntegerLiteral:
		return &integerEvaluator{
			t: monoType(n.Type()),
			i: n.Value,
		}, nil
	case *semantic.FloatLiteral:
		return &floatEvaluator{
			t: monoType(n.Type()),
			f: n.Value,
		}, nil
	case *semantic.StringLiteral:
		return &stringEvaluator{
			t: monoType(n.Type()),
			s: n.Value,
		}, nil
	case *semantic.RegexpLiteral:
		return &regexpEvaluator{
			t: monoType(n.Type()),
			r: n.Value,
		}, nil
	case *semantic.DateTimeLiteral:
		return &timeEvaluator{
			t:    monoType(n.Type()),
			time: values.ConvertTime(n.Value),
		}, nil
	case *semantic.UnaryExpression:
		node, err := compile(n.Argument, builtIns)
		if err != nil {
			return nil, err
		}
		return &unaryEvaluator{
			t:    monoType(n.Type()),
			node: node,
		}, nil
	case *semantic.LogicalExpression:
		l, err := compile(n.Left, builtIns)
		if err != nil {
			return nil, err
		}
		r, err := compile(n.Right, builtIns)
		if err != nil {
			return nil, err
		}
		return &logicalEvaluator{
			t:        monoType(n.Type()),
			operator: n.Operator,
			left:     l,
			right:    r,
		}, nil
	case *semantic.BinaryExpression:
		l, err := compile(n.Left, builtIns)
		if err != nil {
			return nil, err
		}
		log.Printf("%#v", l)
		lt := l.Type()
		r, err := compile(n.Right, builtIns)
		if err != nil {
			return nil, err
		}
		rt := r.Type()
		f, err := values.LookupBinaryFunction(values.BinaryFuncSignature{
			Operator: n.Operator,
			Left:     lt,
			Right:    rt,
		})
		if err != nil {
			return nil, err
		}
		return &binaryEvaluator{
			t:     monoType(n.Type()),
			left:  l,
			right: r,
			f:     f,
		}, nil
	case *semantic.CallExpression:
		callee, err := compile(n.Callee, builtIns)
		if err != nil {
			return nil, err
		}
		args, err := compile(n.Arguments, builtIns)
		if err != nil {
			return nil, err
		}
		return &callEvaluator{
			t:      monoType(n.Type()),
			callee: callee,
			args:   args,
		}, nil
	case *semantic.FunctionExpression:
		body, err := compile(n.Block.Body, builtIns)
		if err != nil {
			return nil, err
		}
		ft := monoType(n.Type())
		in := ft.InType()
		propertyTypes := in.Properties()
		params := make([]functionParam, 0, len(propertyTypes))
		for k, pt := range propertyTypes {
			param := functionParam{
				Key:  k,
				Type: pt,
			}
			if n.Defaults != nil {
				// Search for default value
				for _, d := range n.Defaults.List {
					if d.Key.Name == k {
						d, err := compile(d.Value, builtIns)
						if err != nil {
							return nil, err
						}
						param.Default = d
						break
					}
				}
			}
			params = append(params, param)
		}
		return &functionEvaluator{
			t:      monoType(n.Type()),
			params: params,
			body:   body,
		}, nil
	default:
		return nil, fmt.Errorf("unknown semantic node of type %T", n)
	}
}

// CompilationCache caches compilation results based on the type of the function.
type CompilationCache struct {
	fn       *semantic.FunctionExpression
	scope    Scope
	compiled map[semantic.Type]funcErr
}

func NewCompilationCache(fn *semantic.FunctionExpression, scope Scope) *CompilationCache {
	return &CompilationCache{
		fn:       fn,
		scope:    scope,
		compiled: make(map[semantic.Type]funcErr),
	}
}

// Compile returnes a compiled function bsaed on the provided type.
// The result will be cached for subsequent calls.
func (c *CompilationCache) Compile(fnType semantic.Type) (Func, error) {
	f, ok := c.compiled[fnType]
	if ok {
		return f.F, f.Err
	}
	fun, err := Compile(c.fn, fnType, c.scope)
	c.compiled[fnType] = funcErr{
		F:   fun,
		Err: err,
	}
	return fun, err
}

type funcErr struct {
	F   Func
	Err error
}

// compile recursively searches for a matching child node that has compiled the function.
// If the compilation has not been performed previously its result is cached and returned.
func (c *CompilationCache) compile(fnType semantic.Type) (Func, error) {
	Compile(c.fn, types, c.scope)
}

// Utility function for compiling an `fn` parameter for rename or drop/keep. In addition
// to the function expression, it takes two types to verify the result against:
// a single argument type, and a single return type.
func CompileFnParam(fn *semantic.FunctionExpression, paramType, returnType semantic.Type) (Func, string, error) {
	scope, decls := flux.BuiltIns()
	compileCache := NewCompilationCache(fn, scope, decls)
	if len(fn.Params) != 1 {
		return nil, "", fmt.Errorf("function should only have a single parameter, got %d", len(fn.Params))
	}
	paramName := fn.Params[0].Key.Name

	compiled, err := compileCache.Compile(map[string]semantic.Type{
		paramName: paramType,
	})
	if err != nil {
		return nil, "", err
	}

	if compiled.Type() != returnType {
		return nil, "", fmt.Errorf("provided function does not evaluate to type %s", returnType.Kind())
	}

	return compiled, paramName, nil
}

// externDeclarations produces a list of external declarations from a scope
func externDeclarations(scope Scope) []*semantic.ExternalVariableDeclaration {
	declarations := make([]*semantic.ExternalVariableDeclaration, len(scope))
	for k, v := range scope {
		declarations = append(declarations, &semantic.ExternalVariableDeclaration{
			Identifier: &semantic.Identifier{Name: k},
			ExternType: v.Type().PolyType(),
		})
	}
	return declarations
}
