package compiler

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/pkg/errors"
)

func Compile(f *semantic.FunctionExpression, in semantic.Type, scope Scope) (Func, error) {
	if scope == nil {
		scope = NewScope()
	}
	if in.Nature() != semantic.Object {
		return nil, fmt.Errorf("function input must be an object @ %v", f.Location())
	}
	extern := values.BuildExternAssignments(f, scope)

	typeSol, err := semantic.InferTypes(extern, flux.StdLib())
	if err != nil {
		return nil, errors.Wrapf(err, "compile function @ %v", f.Location())
	}

	pt, err := typeSol.PolyTypeOf(f)
	if err != nil {
		return nil, errors.Wrapf(err, "reteiving compile function @ %v", f.Location())
	}
	props := in.Properties()
	parameters := make(map[string]semantic.PolyType, len(props))
	for k, p := range props {
		parameters[k] = p.PolyType()
	}
	fpt := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Parameters: parameters,
		Return:     typeSol.Fresh(),
	})
	if err := typeSol.AddConstraint(pt, fpt); err != nil {
		return nil, errors.Wrapf(err, "cannot add type constraint @ %v", f.Location())
	}
	fnType, err := typeSol.TypeOf(f)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot compile polymorphic function @ %v", f.Location())
	}

	root, err := compile(f.Block.Body, typeSol, scope, make(map[string]*semantic.FunctionExpression))
	if err != nil {
		return nil, errors.Wrapf(err, "cannot compile @ %v", f.Location())
	}
	return compiledFn{
		root:       root,
		fnType:     fnType,
		inputScope: nestScope(scope),
	}, nil
}

// monoType ignores any errors when reading the type of a node.
// This is safe becase we already validated that the function type is a mono type.
func monoType(t semantic.Type, err error) semantic.Type {
	return t
}

// polyType ignores any errors when reading the type of a node.
// This is safe becase we already validated that the function type is a poly type.
func polyType(t semantic.PolyType, err error) semantic.PolyType {
	return t
}

// compile recursively compiles semantic nodes into evaluators.
func compile(n semantic.Node, typeSol semantic.TypeSolution, scope Scope, funcExprs map[string]*semantic.FunctionExpression) (Evaluator, error) {
	switch n := n.(type) {
	case *semantic.Block:
		body := make([]Evaluator, len(n.Body))
		for i, s := range n.Body {
			node, err := compile(s, typeSol, scope, funcExprs)
			if err != nil {
				return nil, err
			}
			body[i] = node
		}
		return &blockEvaluator{
			t:    monoType(typeSol.TypeOf(n.ReturnStatement().Argument)),
			body: body,
		}, nil
	case *semantic.ExpressionStatement:
		return nil, errors.New("statement does nothing, side effects are not supported by the compiler")
	case *semantic.ReturnStatement:
		node, err := compile(n.Argument, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		return returnEvaluator{
			Evaluator: node,
		}, nil
	case *semantic.NativeVariableAssignment:
		if fe, ok := n.Init.(*semantic.FunctionExpression); ok {
			funcExprs[n.Identifier.Name] = fe
			return &noopEvaluator{}, nil
		}
		node, err := compile(n.Init, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		return &declarationEvaluator{
			t:    monoType(typeSol.TypeOf(n.Init)),
			id:   n.Identifier.Name,
			init: node,
		}, nil
	case *semantic.ObjectExpression:
		properties := make(map[string]Evaluator, len(n.Properties))
		propertyTypes := make(map[string]semantic.Type, len(n.Properties))
		obj := &objEvaluator{
			t: semantic.NewObjectType(propertyTypes),
		}

		for _, p := range n.Properties {
			node, err := compile(p.Value, typeSol, scope, funcExprs)
			if err != nil {
				return nil, err
			}
			properties[p.Key.Key()] = node
			propertyTypes[p.Key.Key()] = node.Type()
		}
		obj.properties = properties

		if n.With != nil {
			node, err := compile(n.With, typeSol, scope, funcExprs)
			if err != nil {
				return nil, err
			}
			with, ok := node.(*identifierEvaluator)
			if !ok {
				return nil, errors.New("unknown identifier in with expression")
			}
			obj.with = with

		}

		return obj, nil

	case *semantic.ArrayExpression:
		elements := make([]Evaluator, len(n.Elements))
		if len(n.Elements) == 0 {
			return &arrayEvaluator{
				t:     semantic.EmptyArrayType,
				array: nil,
			}, nil
		}
		for i, e := range n.Elements {
			node, err := compile(e, typeSol, scope, funcExprs)
			if err != nil {
				return nil, err
			}
			elements[i] = node
		}
		return &arrayEvaluator{
			t:     semantic.NewArrayType(elements[0].Type()),
			array: elements,
		}, nil
	case *semantic.IdentifierExpression:
		// Create type instance of the function
		if fe, ok := funcExprs[n.Name]; ok {
			it, err := typeSol.PolyTypeOf(n)
			if err != nil {
				return nil, err
			}
			ft, err := typeSol.PolyTypeOf(fe)
			if err != nil {
				return nil, err
			}

			typeSol := typeSol.FreshSolution()
			// Add constraint on the identifier type and the function type.
			// This way all type variables in the body of the function will know their monotype.
			err = typeSol.AddConstraint(it, ft)
			if err != nil {
				return nil, err
			}

			return compile(fe, typeSol, scope, funcExprs)
		}
		return &identifierEvaluator{
			t:    polyType(typeSol.PolyTypeOf(n)),
			name: n.Name,
		}, nil
	case *semantic.MemberExpression:
		object, err := compile(n.Object, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		return &memberEvaluator{
			t:        polyType(typeSol.PolyTypeOf(n)),
			object:   object,
			property: n.Property,
		}, nil
	case *semantic.IndexExpression:
		arr, err := compile(n.Array, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		idx, err := compile(n.Index, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		return &arrayIndexEvaluator{
			t:     monoType(typeSol.TypeOf(n)),
			array: arr,
			index: idx,
		}, nil
	case *semantic.BooleanLiteral:
		return &booleanEvaluator{
			t: monoType(typeSol.TypeOf(n)),
			b: n.Value,
		}, nil
	case *semantic.IntegerLiteral:
		return &integerEvaluator{
			t: monoType(typeSol.TypeOf(n)),
			i: n.Value,
		}, nil
	case *semantic.FloatLiteral:
		return &floatEvaluator{
			t: monoType(typeSol.TypeOf(n)),
			f: n.Value,
		}, nil
	case *semantic.StringLiteral:
		return &stringEvaluator{
			t: monoType(typeSol.TypeOf(n)),
			s: n.Value,
		}, nil
	case *semantic.RegexpLiteral:
		return &regexpEvaluator{
			t: monoType(typeSol.TypeOf(n)),
			r: n.Value,
		}, nil
	case *semantic.DateTimeLiteral:
		return &timeEvaluator{
			t:    monoType(typeSol.TypeOf(n)),
			time: values.ConvertTime(n.Value),
		}, nil
	case *semantic.DurationLiteral:
		return &durationEvaluator{
			t:        monoType(typeSol.TypeOf(n)),
			duration: values.Duration(n.Value),
		}, nil
	case *semantic.UnaryExpression:
		node, err := compile(n.Argument, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		return &unaryEvaluator{
			t:    monoType(typeSol.TypeOf(n)),
			node: node,
			op:   n.Operator,
		}, nil
	case *semantic.LogicalExpression:
		l, err := compile(n.Left, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		r, err := compile(n.Right, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		return &logicalEvaluator{
			t:        monoType(typeSol.TypeOf(n)),
			operator: n.Operator,
			left:     l,
			right:    r,
		}, nil
	case *semantic.ConditionalExpression:
		test, err := compile(n.Test, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		c, err := compile(n.Consequent, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		a, err := compile(n.Alternate, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		return &conditionalEvaluator{
			t:          monoType(typeSol.TypeOf(n.Consequent)),
			test:       test,
			consequent: c,
			alternate:  a,
		}, nil
	case *semantic.BinaryExpression:
		l, err := compile(n.Left, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		lt := l.Type()
		r, err := compile(n.Right, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		rt := r.Type()
		f, err := values.LookupBinaryFunction(values.BinaryFuncSignature{
			Operator: n.Operator,
			Left:     lt.Nature(),
			Right:    rt.Nature(),
		})
		if err != nil {
			return nil, err
		}
		return &binaryEvaluator{
			t:     monoType(typeSol.TypeOf(n)),
			left:  l,
			right: r,
			f:     f,
		}, nil
	case *semantic.CallExpression:
		args, err := compile(n.Arguments, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		callee, err := compile(n.Callee, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		return &callEvaluator{
			t:      monoType(typeSol.TypeOf(n)),
			callee: callee,
			args:   args,
		}, nil
	case *semantic.FunctionExpression:
		fnType := monoType(typeSol.TypeOf(n))
		body, err := compile(n.Block.Body, typeSol, scope, funcExprs)
		if err != nil {
			return nil, err
		}
		sig := fnType.FunctionSignature()
		params := make([]functionParam, 0, len(sig.Parameters))
		for k, pt := range sig.Parameters {
			param := functionParam{
				Key:  k,
				Type: pt,
			}
			if n.Defaults != nil {
				// Search for default value
				for _, d := range n.Defaults.Properties {
					if d.Key.Key() == k {
						d, err := compile(d.Value, typeSol, scope, funcExprs)
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
			t:      fnType,
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

// Compile returns a compiled function based on the provided type.
// The result will be cached for subsequent calls.
func (c *CompilationCache) Compile(in semantic.Type) (Func, error) {
	f, ok := c.compiled[in]
	if ok {
		return f.F, f.Err
	}
	fun, err := Compile(c.fn, in, c.scope)
	c.compiled[in] = funcErr{
		F:   fun,
		Err: err,
	}
	return fun, err
}

type funcErr struct {
	F   Func
	Err error
}

// Utility function for compiling an `fn` parameter for rename or drop/keep. In addition
// to the function expression, it takes two types to verify the result against:
// a single argument type, and a single return type.
func CompileFnParam(fn *semantic.FunctionExpression, scope Scope, paramType, returnType semantic.Type) (Func, string, error) {
	compileCache := NewCompilationCache(fn, scope)
	if fn.Block.Parameters != nil && len(fn.Block.Parameters.List) != 1 {
		return nil, "", errors.New("function should only have a single parameter")
	}
	paramName := fn.Block.Parameters.List[0].Key.Name

	compiled, err := compileCache.Compile(semantic.NewObjectType(map[string]semantic.Type{
		paramName: paramType,
	}))
	if err != nil {
		return nil, "", err
	}

	if compiled.Type() != returnType {
		return nil, "", fmt.Errorf("provided function does not evaluate to type %s", returnType.Nature())
	}

	return compiled, paramName, nil
}

func CompileReduceFn(fn *semantic.FunctionExpression, scope Scope, paramType semantic.Type) (Func, []string, error) {
	compileCache := NewCompilationCache(fn, scope)
	if len(fn.Block.Parameters.List) != 2 {
		return nil, nil, errors.New("function should only have a single parameter")
	}
	paramList := fn.Block.Parameters.List
	paramNames := []string{paramList[0].Key.Name, paramList[1].Key.Name}

	compiled, err := compileCache.Compile(semantic.NewObjectType(map[string]semantic.Type{
		paramNames[0]: paramType,
		paramNames[1]: paramType,
	}))
	if err != nil {
		return nil, nil, err
	}

	if compiled.Type() != paramType {
		return nil, nil, fmt.Errorf("provided function does not evaluate to type %s", paramType.Nature())
	}

	return compiled, paramNames, nil
}
