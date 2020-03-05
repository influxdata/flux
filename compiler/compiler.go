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

	// Retrieve the function argument types and create an object type from them.
	fnType := f.TypeOf()
	argN, err := fnType.NumArguments()
	if err != nil {
		return nil, err
	}

	// Iterate over every argument and find the equivalent
	// property inside of the input.
	// The function expression has a monotype that may have
	// tvars contained within it. We have a realized input type
	// so we can use that to construct the tvar substitutions.
	// Iterate over every argument and find the equivalent
	// property inside of the input and then generate the substitutions.
	subst := make(map[uint64]semantic.MonoType)
	for i := 0; i < argN; i++ {
		arg, err := fnType.Argument(i)
		if err != nil {
			return nil, err
		}

		name := arg.Name()
		argT, err := arg.TypeOf()
		if err != nil {
			return nil, err
		}

		prop, ok, err := findProperty(string(name), in)
		if err != nil {
			return nil, err
		} else if ok {
			mtyp, err := prop.TypeOf()
			if err != nil {
				return nil, err
			}
			if err := substituteTypes(subst, argT, mtyp); err != nil {
				return nil, err
			}
		}
	}

	root, err := compile(f.Block.Body, subst, scope)
	if err != nil {
		return nil, errors.Wrapf(err, codes.Inherit, "cannot compile @ %v", f.Location())
	}
	return compiledFn{
		root:       root,
		fnType:     fnType,
		inputScope: nestScope(scope),
	}, nil
}

// substituteTypes will generate a substitution map by recursing through
// inType and mapping any variables to the value in the other record.
// If the input type is not a type variable, it will check to ensure
// that the type in the input matches or it will return an error.
func substituteTypes(subst map[uint64]semantic.MonoType, inType, in semantic.MonoType) error {
	// If the input isn't a valid type, then don't consider it as
	// part of substituting types. We will trust type inference has
	// the correct type and that we are just handling a null value
	// which isn't represented in type inference.
	if in.Nature() == semantic.Invalid {
		return nil
	} else if inType.Kind() == semantic.Var {
		vn, err := inType.VarNum()
		if err != nil {
			return err
		}
		// If this substitution variable already exists,
		// we need to verify that it maps to the same type
		// in the input record.
		// We can do this by calling substituteTypes with the same
		// input parameter and the substituted monotype since
		// substituteTypes will verify the types.
		if t, ok := subst[vn]; ok {
			return substituteTypes(subst, t, in)
		}

		// If the input type is not invalid, mark it down
		// as the real type.
		if in.Nature() != semantic.Invalid {
			subst[vn] = in
		}
		return nil
	}

	if inType.Kind() != in.Kind() {
		return errors.Newf(codes.FailedPrecondition, "type conflict: %s != %s", inType, in)
	}

	switch inType.Kind() {
	case semantic.Basic:
		at, err := inType.Basic()
		if err != nil {
			return err
		}

		// Otherwise we have a valid type and need to ensure they match.
		bt, err := in.Basic()
		if err != nil {
			return err
		}

		if at != bt {
			return errors.Newf(codes.FailedPrecondition, "type conflict: %s != %s", inType, in)
		}
		return nil
	case semantic.Arr:
		lt, err := inType.ElemType()
		if err != nil {
			return err
		}

		rt, err := inType.ElemType()
		if err != nil {
			return err
		}
		return substituteTypes(subst, lt, rt)
	case semantic.Row:
		// We need to compare the row type that was inferred
		// and the reality. It is ok for row properties to exist
		// in the real type that aren't in the inferred type and
		// it is ok for inferred types to be missing from the actual
		// input type in the case of null values.
		// What isn't ok is that the two types conflict so we are
		// going to iterate over all of the properties in the inferred
		// type and perform substitutions on them.
		nproperties, err := inType.NumProperties()
		if err != nil {
			return err
		}

		for i := 0; i < nproperties; i++ {
			lprop, err := inType.RowProperty(i)
			if err != nil {
				return err
			}

			// Find the property in the real type if it
			// exists. If it doesn't exist, then no problem!
			name := lprop.Name()
			rprop, ok, err := findProperty(name, in)
			if err != nil {
				return err
			} else if !ok {
				// It is ok if this property doesn't exist
				// in the input type.
				continue
			}

			ltyp, err := lprop.TypeOf()
			if err != nil {
				return err
			}
			rtyp, err := rprop.TypeOf()
			if err != nil {
				return err
			}
			if err := substituteTypes(subst, ltyp, rtyp); err != nil {
				return err
			}
		}
		return nil
	case semantic.Fun:
		// TODO: https://github.com/influxdata/flux/issues/2587
		return errors.New(codes.Unimplemented)
	default:
		return errors.Newf(codes.Internal, "unknown semantic kind: %s", inType)
	}
}

func findProperty(name string, t semantic.MonoType) (*semantic.RowProperty, bool, error) {
	n, err := t.NumProperties()
	if err != nil {
		return nil, false, err
	}
	for i := 0; i < n; i++ {
		p, err := t.RowProperty(i)
		if err != nil {
			return nil, false, err
		}
		if p.Name() == name {
			return p, true, nil
		}
	}
	return nil, false, nil
}

// monoType ignores any errors when reading the type of a node.
// This is safe becase we already validated that the function type is a mono type.
func monoType(subst map[uint64]semantic.MonoType, t semantic.MonoType) semantic.MonoType {
	tv, err := t.VarNum()
	if err != nil {
		return t
	}
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
		var elements []Evaluator
		if len(n.Elements) > 0 {
			elements = make([]Evaluator, len(n.Elements))
			for i, e := range n.Elements {
				node, err := compile(e, subst, scope)
				if err != nil {
					return nil, err
				}
				elements[i] = node
			}
		}
		return &arrayEvaluator{
			t:     monoType(subst, n.TypeOf()),
			array: elements,
		}, nil
	case *semantic.IdentifierExpression:
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
		lt := l.Type().Nature()
		r, err := compile(n.Right, subst, scope)
		if err != nil {
			return nil, err
		}
		rt := r.Type().Nature()
		if lt == semantic.Invalid {
			lt = rt
		} else if rt == semantic.Invalid {
			rt = lt
		}
		f, err := values.LookupBinaryFunction(values.BinaryFuncSignature{
			Operator: n.Operator,
			Left:     lt,
			Right:    rt,
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
			fn:     n,
		}, nil
	default:
		return nil, errors.Newf(codes.Internal, "unknown semantic node of type %T", n)
	}
}
