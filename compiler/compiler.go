package compiler

import (
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func Compile(scope Scope, f *semantic.FunctionExpression, in semantic.MonoType) (Func, error) {
	if scope == nil {
		scope = NewScope()
	}
	if in.Nature() != semantic.Object {
		return nil, errors.Newf(codes.Invalid, "function input must be an object @ %v", f.Location())
	}
	// TODO how do externs work now?
	//extern := values.BuildExternAssignments(f, scope)

	// The function expression may be polymorphic,
	// we can only compile and execute it if it is monomorphic.
	// Steps to convert from polymorphic to monomorphic function type:
	// 1. Create substituion for each of the free vars in the poly type
	//    Substituion should be based on the input type
	// 2. Apply substitution to poly type
	// 3. Compile monomorphic function
	//
	// forall [t0,t1] where t0 is Rowable (r: { bar: t1 | t0}) -> {foo: t1 | bar: t1 | t0}
	// (r) => ({r with foo: r.bar + 1})

	// TODO f needs to have a poly type
	ft := f.TypeOf()
	r, err := ft.Argument(0)
	if err != nil {
		return nil, err
	}
	rt, err := r.TypeOf()
	if err != nil {
		return nil, err
	}

	n, err := rt.NumProperties()
	if err != nil {
		return nil, err
	}
	// TODO allocate space for the number of poly type vars
	subst := make(map[uint64]semantic.MonoType)
	for i := 0; i < n; i++ {
		p, err := rt.RowProperty(i)
		if err != nil {
			return nil, err
		}
		pt, err := p.TypeOf()
		if err != nil {
			return nil, err
		}
		inp, err := findProperty(p.Name(), in)
		if err != nil {
			return nil, err
		}
		inpt, err := inp.TypeOf()
		if err != nil {
			return nil, err
		}
		inb, err := inpt.Basic()
		if err != nil {
			return nil, err // TODO error message about requiring basic types to compiled functions
		}
		b, err := pt.Basic()
		if err != nil {
			return nil, err // TODO error message about requiring basic types to compiled functions
		}
		if b != inb {
			return nil, errors.New(codes.Invalid, "TODO something about mismatched input types")
		}

		if pt.Kind() == semantic.Var {
			vn, err := pt.VarNum()
			if err != nil {
				return nil, err
			}
			subst[vn] = inpt
			// TODO check kind constraints
			// against the input type
		}
	}

	root, err := compile(f.Block.Body, subst, scope)
	if err != nil {
		return nil, errors.Wrapf(err, codes.Inherit, "cannot compile @ %v", f.Location())
	}
	return compiledFn{
		root:       root,
		fnType:     ft,
		inputScope: nestScope(scope),
	}, nil
}

func findProperty(name string, t semantic.MonoType) (*semantic.RowProperty, error) {
	n, err := t.NumProperties()
	if err != nil {
		return nil, err
	}
	for i := 0; i < n; i++ {
		p, err := t.RowProperty(i)
		if err != nil {
			return nil, err
		}
		if p.Name() == name {
			return p, nil
		}
	}
	// TODO use correct error here
	return nil, errors.New(codes.Internal, "not found")
}

// monoType ignores any errors when reading the type of a node.
// This is safe becase we already validated that the function type is a mono type.
func monoType(subst map[uint64]semantic.MonoType, t semantic.MonoType) semantic.MonoType {
	tv, err := t.VarNum()
	if err != nil {
		return t
	}
	// TODO do we care about the case when the tvar is not in the substitution?
	return subst[tv]
}

// compile recursively compiles semantic nodes into evaluators.
func compile(n semantic.Node, subst map[uint64]semantic.MonoType, scope Scope) (Evaluator, error) {
	switch n := n.(type) {
	case *semantic.Block:
		body := make([]Evaluator, len(n.Body))
		for i, s := range n.Body {
			node, err := compile(s, subst, scope)
			if err != nil {
				return nil, err
			}
			body[i] = node
		}
		return &blockEvaluator{
			t:    monoType(subst, n.ReturnStatement().Argument.TypeOf()),
			body: body,
		}, nil
	case *semantic.ExpressionStatement:
		return nil, errors.New(codes.Internal, "statement does nothing, side effects are not supported by the compiler")
	case *semantic.ReturnStatement:
		node, err := compile(n.Argument, subst, scope)
		if err != nil {
			return nil, err
		}
		return returnEvaluator{
			Evaluator: node,
		}, nil
	case *semantic.NativeVariableAssignment:
		node, err := compile(n.Init, subst, scope)
		if err != nil {
			return nil, err
		}
		return &declarationEvaluator{
			t:    monoType(subst, n.Init.TypeOf()),
			id:   n.Identifier.Name,
			init: node,
		}, nil
	case *semantic.ObjectExpression:
		properties := make(map[string]Evaluator, len(n.Properties))
		obj := &objEvaluator{
			t: monoType(subst, n.TypeOf()),
		}

		for _, p := range n.Properties {
			node, err := compile(p.Value, subst, scope)
			if err != nil {
				return nil, err
			}
			properties[p.Key.Key()] = node
		}
		obj.properties = properties

		if n.With != nil {
			node, err := compile(n.With, subst, scope)
			if err != nil {
				return nil, err
			}
			with, ok := node.(*identifierEvaluator)
			if !ok {
				return nil, errors.New(codes.Internal, "unknown identifier in with expression")
			}
			obj.with = with

		}

		return obj, nil

	case *semantic.ArrayExpression:
		elements := make([]Evaluator, len(n.Elements))
		if len(n.Elements) == 0 {
			return &arrayEvaluator{
				t:     monoType(subst, n.TypeOf()),
				array: nil,
			}, nil
		}
		for i, e := range n.Elements {
			node, err := compile(e, subst, scope)
			if err != nil {
				return nil, err
			}
			elements[i] = node
		}
		return &arrayEvaluator{
			t:     monoType(subst, n.TypeOf()),
			array: elements,
		}, nil
	case *semantic.IdentifierExpression:
		// TODO do we need this anymore?
		// Create type instance of the function
		//if fe, ok := funcExprs[n.Name]; ok {
		//	it, err := subst.PolyTypeOf(n)
		//	if err != nil {
		//		return nil, err
		//	}
		//	ft, err := subst.PolyTypeOf(fe)
		//	if err != nil {
		//		return nil, err
		//	}

		//	subst := subst.FreshSolution()
		//	// Add constraint on the identifier type and the function type.
		//	// This way all type variables in the body of the function will know their monotype.
		//	err = subst.AddConstraint(it, ft)
		//	if err != nil {
		//		return nil, err
		//	}

		//	return compile(fe, subst, scope, )
		//}
		return &identifierEvaluator{
			t:    monoType(subst, n.TypeOf()),
			name: n.Name,
		}, nil
	case *semantic.MemberExpression:
		object, err := compile(n.Object, subst, scope)
		if err != nil {
			return nil, err
		}
		return &memberEvaluator{
			t:        monoType(subst, n.TypeOf()),
			object:   object,
			property: n.Property,
		}, nil
	case *semantic.IndexExpression:
		arr, err := compile(n.Array, subst, scope)
		if err != nil {
			return nil, err
		}
		idx, err := compile(n.Index, subst, scope)
		if err != nil {
			return nil, err
		}
		return &arrayIndexEvaluator{
			t:     monoType(subst, n.TypeOf()),
			array: arr,
			index: idx,
		}, nil
	case *semantic.StringExpression:
		parts := make([]Evaluator, len(n.Parts))
		for i, p := range n.Parts {
			e, err := compile(p, subst, scope)
			if err != nil {
				return nil, err
			}
			parts[i] = e
		}
		return &stringExpressionEvaluator{
			t:     monoType(subst, n.TypeOf()),
			parts: parts,
		}, nil
	case *semantic.TextPart:
		return &textEvaluator{
			value: n.Value,
		}, nil
	case *semantic.InterpolatedPart:
		e, err := compile(n.Expression, subst, scope)
		if err != nil {
			return nil, err
		}
		return &interpolatedEvaluator{
			s: e,
		}, nil
	case *semantic.BooleanLiteral:
		return &booleanEvaluator{
			t: monoType(subst, n.TypeOf()),
			b: n.Value,
		}, nil
	case *semantic.IntegerLiteral:
		return &integerEvaluator{
			t: monoType(subst, n.TypeOf()),
			i: n.Value,
		}, nil
	case *semantic.UnsignedIntegerLiteral:
		return &unsignedIntegerEvaluator{
			t: monoType(subst, n.TypeOf()),
			i: n.Value,
		}, nil
	case *semantic.FloatLiteral:
		return &floatEvaluator{
			t: monoType(subst, n.TypeOf()),
			f: n.Value,
		}, nil
	case *semantic.StringLiteral:
		return &stringEvaluator{
			t: monoType(subst, n.TypeOf()),
			s: n.Value,
		}, nil
	case *semantic.RegexpLiteral:
		return &regexpEvaluator{
			t: monoType(subst, n.TypeOf()),
			r: n.Value,
		}, nil
	case *semantic.DateTimeLiteral:
		return &timeEvaluator{
			t:    monoType(subst, n.TypeOf()),
			time: values.ConvertTime(n.Value),
		}, nil
	case *semantic.DurationLiteral:
		v, err := values.FromDurationValues(n.Values)
		if err != nil {
			return nil, err
		}
		return &durationEvaluator{
			t:        monoType(subst, n.TypeOf()),
			duration: v,
		}, nil
	case *semantic.UnaryExpression:
		node, err := compile(n.Argument, subst, scope)
		if err != nil {
			return nil, err
		}
		return &unaryEvaluator{
			t:    monoType(subst, n.TypeOf()),
			node: node,
			op:   n.Operator,
		}, nil
	case *semantic.LogicalExpression:
		l, err := compile(n.Left, subst, scope)
		if err != nil {
			return nil, err
		}
		r, err := compile(n.Right, subst, scope)
		if err != nil {
			return nil, err
		}
		return &logicalEvaluator{
			t:        monoType(subst, n.TypeOf()),
			operator: n.Operator,
			left:     l,
			right:    r,
		}, nil
	case *semantic.ConditionalExpression:
		test, err := compile(n.Test, subst, scope)
		if err != nil {
			return nil, err
		}
		c, err := compile(n.Consequent, subst, scope)
		if err != nil {
			return nil, err
		}
		a, err := compile(n.Alternate, subst, scope)
		if err != nil {
			return nil, err
		}
		return &conditionalEvaluator{
			t:          monoType(subst, n.Consequent.TypeOf()),
			test:       test,
			consequent: c,
			alternate:  a,
		}, nil
	case *semantic.BinaryExpression:
		l, err := compile(n.Left, subst, scope)
		if err != nil {
			return nil, err
		}
		lt := l.Type()
		r, err := compile(n.Right, subst, scope)
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
			t:     monoType(subst, n.TypeOf()),
			left:  l,
			right: r,
			f:     f,
		}, nil
	case *semantic.CallExpression:
		args, err := compile(n.Arguments, subst, scope)
		if err != nil {
			return nil, err
		}
		callee, err := compile(n.Callee, subst, scope)
		if err != nil {
			return nil, err
		}
		return &callEvaluator{
			t:      monoType(subst, n.TypeOf()),
			callee: callee,
			args:   args,
		}, nil
	case *semantic.FunctionExpression:
		fnType := monoType(subst, n.TypeOf())
		body, err := compile(n.Block.Body, subst, scope)
		if err != nil {
			return nil, err
		}
		num, err := fnType.NumArguments()
		if err != nil {
			return nil, err
		}
		params := make([]functionParam, 0, num)
		for i := 0; i < num; i++ {
			arg, err := fnType.Argument(i)
			if err != nil {
				return nil, err
			}
			k := string(arg.Name())
			pt, err := arg.TypeOf()
			if err != nil {
				return nil, err
			}
			param := functionParam{
				Key:  k,
				Type: pt,
			}
			if n.Defaults != nil {
				// Search for default value
				for _, d := range n.Defaults.Properties {
					if d.Key.Key() == k {
						d, err := compile(d.Value, subst, scope)
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
		return nil, errors.Newf(codes.Internal, "unknown semantic node of type %T", n)
	}
}

// CompilationCache caches compilation results based on the type of the function.
type CompilationCache struct {
	fn    *semantic.FunctionExpression
	scope Scope
	//compiled map[semantic.MonoType]funcErr
}

func NewCompilationCache(fn *semantic.FunctionExpression, scope Scope) *CompilationCache {
	return &CompilationCache{
		fn:    fn,
		scope: scope,
		//compiled: make(map[semantic.MonoType]funcErr),
	}
}

// Compile returns a compiled function based on the provided type.
// The result will be cached for subsequent calls.
func (c *CompilationCache) Compile(in semantic.MonoType) (Func, error) {
	// The cache can be implemented in terms of the input type properties and basic types.
	// We do not need to cache for arbitrary mono types rather we know its a record with basic types for its properties.
	// We can special case creating a hash from property name and basic type (which is a byte).

	//f, ok := c.compiled[in]
	//if ok {
	//	return f.F, f.Err
	//}
	//fun, err := Compile(c.scope, c.fn, in)
	//c.compiled[in] = funcErr{
	//	F:   fun,
	//	Err: err,
	//}
	//return fun, err
	panic("unimpleneted")
}

type funcErr struct {
	F   Func
	Err error
}

// CompileFnParam is a utility function for compiling an `fn` parameter for rename or drop/keep. In addition
// to the function expression, it takes two types to verify the result against:
// a single argument type, and a single return type.
func CompileFnParam(fn *semantic.FunctionExpression, scope Scope, paramType, returnType semantic.MonoType) (Func, string, error) {
	//compileCache := NewCompilationCache(fn, scope)
	//if fn.Block.Parameters != nil && len(fn.Block.Parameters.List) != 1 {
	//	return nil, "", errors.New(codes.Invalid, "function should only have a single parameter")
	//}
	//paramName := fn.Block.Parameters.List[0].Key.Name

	//compiled, err := compileCache.Compile(semantic.NewObjectType(map[string]semantic.MonoType{
	//	paramName: paramType,
	//}))
	//if err != nil {
	//	return nil, "", err
	//}

	//if compiled.Type() != returnType {
	//	return nil, "", errors.Newf(codes.Invalid, "provided function does not evaluate to type %s", returnType.Nature())
	//}

	//return compiled, paramName, nil
	panic("unimplemented")
}
