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
	subst := &semantic.Substitution{
		TypeMap: make(map[uint64]semantic.MonoType),
	}
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
		} else if !arg.Optional() {
			return nil, errors.Newf(codes.Invalid, "missing required argument %q", string(name))
		}
	}

	root, err := compile(f.Block, subst)
	if err != nil {
		return nil, errors.Wrapf(err, codes.Inherit, "cannot compile @ %v", f.Location())
	}
	return compiledFn{
		root:        root,
		parentScope: scope,
	}, nil
}

// substituteTypes will populate a substitution map by recursing through
// inType and mapping any variables to the value in the other record.
// If the input type is not a type variable, it will check to ensure
// that the type in the input matches or it will return an error.
func substituteTypes(subst *semantic.Substitution, inferredType, actualType semantic.MonoType) error {
	// If the input isn't a valid type, then don't consider it as
	// part of substituting types. We will trust type inference has
	// the correct type and that we are just handling a null value
	// which isn't represented in type inference.
	if actualType.Nature() == semantic.Invalid {
		return nil
	} else if inferredType.Kind() == semantic.Var {
		vn, err := inferredType.VarNum()
		if err != nil {
			return err
		}
		// If this substitution variable already exists,
		// we need to verify that it maps to the same type
		// in the input record.
		// We can do this by calling substituteTypes with the same
		// input parameter and the substituted monotype since
		// substituteTypes will verify the types.
		if t, ok := subst.Apply(vn); ok {
			return substituteTypes(subst, t, actualType)
		}

		// If the input type is not invalid, mark it down
		// as the real type.
		if actualType.Nature() != semantic.Invalid {
			subst.TypeMap[vn] = actualType
		}
		return nil
	}

	if inferredType.Kind() != actualType.Kind() {
		return errors.Newf(codes.FailedPrecondition, "type conflict: %s != %s", inferredType, actualType)
	}

	switch inferredType.Kind() {
	case semantic.Basic:
		at, err := inferredType.Basic()
		if err != nil {
			return err
		}

		// Otherwise we have a valid type and need to ensure they match.
		bt, err := actualType.Basic()
		if err != nil {
			return err
		}

		if at != bt {
			return errors.Newf(codes.FailedPrecondition, "type conflict: %s != %s", inferredType, actualType)
		}
		return nil
	case semantic.Collection:
		lt, err := inferredType.ElemType()
		if err != nil {
			return err
		}

		rt, err := actualType.ElemType()
		if err != nil {
			return err
		}
		return substituteTypes(subst, lt, rt)
	case semantic.Dict:
		lk, err := inferredType.KeyType()
		if err != nil {
			return err
		}

		rk, err := actualType.KeyType()
		if err != nil {
			return err
		}

		if err := substituteTypes(subst, lk, rk); err != nil {
			return err
		}

		lv, err := inferredType.ValueType()
		if err != nil {
			return err
		}

		rv, err := actualType.ValueType()
		if err != nil {
			return err
		}

		return substituteTypes(subst, lv, rv)
	case semantic.Record:
		// We need to compare the Record type that was inferred
		// and the reality. It is ok for Record properties to exist
		// in the real type that aren't in the inferred type and
		// it is ok for inferred types to be missing from the actual
		// input type in the case of null values.
		// What isn't ok is that the two types conflict so we are
		// going to iterate over all of the properties in the inferred
		// type and perform substitutions on them.
		nproperties, err := inferredType.NumProperties()
		if err != nil {
			return err
		}

		names := make([]string, 0, nproperties)
		for i := 0; i < nproperties; i++ {
			lprop, err := inferredType.RecordProperty(i)
			if err != nil {
				return err
			}

			// Record the name of the property in the input type.
			name := lprop.Name()
			if containsStr(names, name) {
				// The input type may have the same field twice if the record was
				// extended with {r with ...}
				continue
			}
			names = append(names, name)

			// Find the property in the real type if it
			// exists. If it doesn't exist, then no problem!
			rprop, ok, err := findProperty(name, actualType)
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

		// If this object extends another, then find all of the labels
		// in the in value that were not referenced by the type.
		if withType, ok, err := inferredType.Extends(); err != nil {
			return err
		} else if ok {
			// Construct the input by filtering any of the names
			// that were referenced above. This way, extends only
			// includes the unreferenced labels.
			nproperties, err := actualType.NumProperties()
			if err != nil {
				return err
			}

			properties := make([]semantic.PropertyType, 0, nproperties)
			for i := 0; i < nproperties; i++ {
				prop, err := actualType.RecordProperty(i)
				if err != nil {
					return err
				}

				name := prop.Name()
				if containsStr(names, name) {
					// Already referenced so don't pass this
					// to the extends portion.
					continue
				}

				typ, err := prop.TypeOf()
				if err != nil {
					return err
				}
				properties = append(properties, semantic.PropertyType{
					Key:   []byte(name),
					Value: typ,
				})
			}
			with := semantic.NewObjectType(properties)
			if err := substituteTypes(subst, withType, with); err != nil {
				return err
			}
		}
		return nil
	case semantic.Fun:
		// TODO: https://github.com/influxdata/flux/issues/2587
		return errors.New(codes.Unimplemented)
	default:
		return errors.Newf(codes.Internal, "unknown semantic kind: %s", inferredType)
	}
}

func findProperty(name string, t semantic.MonoType) (*semantic.RecordProperty, bool, error) {
	n, err := t.NumProperties()
	if err != nil {
		return nil, false, err
	}
	for i := 0; i < n; i++ {
		p, err := t.RecordProperty(i)
		if err != nil {
			return nil, false, err
		}
		if p.Name() == name {
			return p, true, nil
		}
	}
	return nil, false, nil
}

// apply applies a substitution to a type.
// It will ignore any errors when reading a type.
// This is safe becase we already validated that the function type is a monotype.
func apply(sub semantic.Substitutor, props []semantic.PropertyType, t semantic.MonoType) semantic.MonoType {
	switch t.Kind() {
	case semantic.Unknown, semantic.Basic:
		// Basic types do not contain type variables.
		// As a result there is nothing to substitute.
		return t
	case semantic.Var:
		tv, err := t.VarNum()
		if err != nil {
			return t
		}
		ty, ok := sub.Apply(tv)
		if !ok {
			return t
		}
		return ty
	case semantic.Collection:
		collection, err := t.CollectionType()
		if err != nil {
			return t
		}

		element, err := t.ElemType()
		if err != nil {
			return t
		}

		return semantic.NewAppType(collection, apply(sub, props, element))
	case semantic.Dict:
		key, err := t.KeyType()
		if err != nil {
			return t
		}
		val, err := t.ValueType()
		if err != nil {
			return t
		}
		return semantic.NewDictType(
			apply(sub, props, key),
			apply(sub, props, val),
		)
	case semantic.Record:
		n, err := t.NumProperties()
		if err != nil {
			return t
		}
		for i := 0; i < n; i++ {
			pr, err := t.RecordProperty(i)
			if err != nil {
				return t
			}
			ty, err := pr.TypeOf()
			if err != nil {
				return t
			}
			props = append(props, semantic.PropertyType{
				Key:   []byte(pr.Name()),
				Value: apply(sub, nil, ty),
			})
		}
		r, extends, err := t.Extends()
		if err != nil {
			return t
		}
		if !extends {
			return semantic.NewObjectType(props)
		}
		r = apply(sub, nil, r)
		switch r.Kind() {
		case semantic.Record:
			return apply(sub, props, r)
		case semantic.Var:
			tv, err := r.VarNum()
			if err != nil {
				return t
			}
			return semantic.ExtendObjectType(props, &tv)
		default:
			panic("unknown type in record extension")
		}
	case semantic.Fun:
		n, err := t.NumArguments()
		if err != nil {
			return t
		}
		args := make([]semantic.ArgumentType, n)
		for i := 0; i < n; i++ {
			arg, err := t.Argument(i)
			if err != nil {
				return t
			}
			typ, err := arg.TypeOf()
			if err != nil {
				return t
			}
			args[i] = semantic.ArgumentType{
				Name:     arg.Name(),
				Type:     apply(sub, nil, typ),
				Pipe:     arg.Pipe(),
				Optional: arg.Optional(),
			}
		}
		retn, err := t.ReturnType()
		if err != nil {
			return t
		}
		return semantic.NewFunctionType(apply(sub, nil, retn), args)
	}
	// If none of the above cases are matched, something has gone
	// seriously wrong and we should panic.
	panic("unknown type")
}

// compile recursively compiles semantic nodes into evaluators.
func compile(n semantic.Node, subst semantic.Substitutor) (Evaluator, error) {
	switch n := n.(type) {
	case *semantic.Block:
		body := make([]Evaluator, len(n.Body))
		for i, s := range n.Body {
			node, err := compile(s, subst)
			if err != nil {
				return nil, err
			}
			body[i] = node
		}
		return &blockEvaluator{
			t:    apply(subst, nil, n.ReturnStatement().Argument.TypeOf()),
			body: body,
		}, nil
	case *semantic.ExpressionStatement:
		return nil, errors.New(codes.Internal, "statement does nothing, side effects are not supported by the compiler")
	case *semantic.ReturnStatement:
		node, err := compile(n.Argument, subst)
		if err != nil {
			return nil, err
		}
		return returnEvaluator{
			Evaluator: node,
		}, nil
	case *semantic.NativeVariableAssignment:
		subst, err := n.Typ.Instantiator(subst)
		if err != nil {
			return nil, err
		}
		node, err := compile(n.Init, subst)
		if err != nil {
			return nil, err
		}
		t := apply(subst, nil, n.Init.TypeOf())
		return &declarationEvaluator{
			t:    t,
			id:   n.Identifier.Name.Name(),
			init: node,
		}, nil
	case *semantic.ObjectExpression:
		properties := make(map[string]Evaluator, len(n.Properties))

		for _, p := range n.Properties {
			node, err := compile(p.Value, subst)
			if err != nil {
				return nil, err
			}
			properties[p.Key.Key()] = node
		}

		var extends *identifierEvaluator
		if n.With != nil {
			node, err := compile(n.With, subst)
			if err != nil {
				return nil, err
			}
			with, ok := node.(*identifierEvaluator)
			if !ok {
				return nil, errors.New(codes.Internal, "unknown identifier in with expression")
			}
			extends = with
		}

		return &objEvaluator{
			t:          apply(subst, nil, n.TypeOf()),
			properties: properties,
			with:       extends,
		}, nil

	case *semantic.ArrayExpression:
		var elements []Evaluator
		if len(n.Elements) > 0 {
			elements = make([]Evaluator, len(n.Elements))
			for i, e := range n.Elements {
				node, err := compile(e, subst)
				if err != nil {
					return nil, err
				}
				elements[i] = node
			}
		}
		return &arrayEvaluator{
			t:     apply(subst, nil, n.TypeOf()),
			array: elements,
		}, nil
	case *semantic.DictExpression:
		elements := make([]struct {
			Key Evaluator
			Val Evaluator
		}, len(n.Elements))
		for i, item := range n.Elements {
			key, err := compile(item.Key, subst)
			if err != nil {
				return nil, err
			}
			val, err := compile(item.Val, subst)
			if err != nil {
				return nil, err
			}
			elements[i] = struct {
				Key Evaluator
				Val Evaluator
			}{Key: key, Val: val}
		}
		return &dictEvaluator{
			t:        apply(subst, nil, n.TypeOf()),
			elements: elements,
		}, nil
	case *semantic.IdentifierExpression:
		return &identifierEvaluator{
			t:    apply(subst, nil, n.TypeOf()),
			name: n.Name.Name(),
		}, nil
	case *semantic.MemberExpression:
		object, err := compile(n.Object, subst)
		if err != nil {
			return nil, err
		}
		t := apply(subst, nil, n.TypeOf())
		return &memberEvaluator{
			t:        t,
			object:   object,
			property: n.Property.Name(),
			nullable: isNullable(t),
		}, nil
	case *semantic.IndexExpression:
		arr, err := compile(n.Array, subst)
		if err != nil {
			return nil, err
		}
		idx, err := compile(n.Index, subst)
		if err != nil {
			return nil, err
		}
		return &arrayIndexEvaluator{
			t:     apply(subst, nil, n.TypeOf()),
			array: arr,
			index: idx,
		}, nil
	case *semantic.StringExpression:
		parts := make([]Evaluator, len(n.Parts))
		for i, p := range n.Parts {
			e, err := compile(p, subst)
			if err != nil {
				return nil, err
			}
			parts[i] = e
		}
		return &stringExpressionEvaluator{
			parts: parts,
		}, nil
	case *semantic.TextPart:
		return &textEvaluator{
			value: n.Value,
		}, nil
	case *semantic.InterpolatedPart:
		e, err := compile(n.Expression, subst)
		if err != nil {
			return nil, err
		}
		return &interpolatedEvaluator{
			s: e,
		}, nil
	case *semantic.BooleanLiteral:
		return &booleanEvaluator{
			b: n.Value,
		}, nil
	case *semantic.IntegerLiteral:
		return &integerEvaluator{
			i: n.Value,
		}, nil
	case *semantic.UnsignedIntegerLiteral:
		return &unsignedIntegerEvaluator{
			i: n.Value,
		}, nil
	case *semantic.FloatLiteral:
		return &floatEvaluator{
			f: n.Value,
		}, nil
	case *semantic.StringLiteral:
		return &stringEvaluator{
			s: n.Value,
		}, nil
	case *semantic.RegexpLiteral:
		return &regexpEvaluator{
			r: n.Value,
		}, nil
	case *semantic.DateTimeLiteral:
		return &timeEvaluator{
			time: values.ConvertTime(n.Value),
		}, nil
	case *semantic.DurationLiteral:
		v, err := values.FromDurationValues(n.Values)
		if err != nil {
			return nil, err
		}
		return &durationEvaluator{
			duration: v,
		}, nil
	case *semantic.UnaryExpression:
		node, err := compile(n.Argument, subst)
		if err != nil {
			return nil, err
		}
		return &unaryEvaluator{
			t:    apply(subst, nil, n.TypeOf()),
			node: node,
			op:   n.Operator,
		}, nil
	case *semantic.LogicalExpression:
		l, err := compile(n.Left, subst)
		if err != nil {
			return nil, err
		}
		r, err := compile(n.Right, subst)
		if err != nil {
			return nil, err
		}

		if l.Type().Nature() == semantic.Vector && r.Type().Nature() == semantic.Vector {
			return &logicalVectorEvaluator{
				operator: n.Operator,
				left:     l,
				right:    r,
			}, nil
		}

		return &logicalEvaluator{
			operator: n.Operator,
			left:     l,
			right:    r,
		}, nil
	case *semantic.ConditionalExpression:
		test, err := compile(n.Test, subst)
		if err != nil {
			return nil, err
		}
		c, err := compile(n.Consequent, subst)
		if err != nil {
			return nil, err
		}
		a, err := compile(n.Alternate, subst)
		if err != nil {
			return nil, err
		}

		if test.Type().Nature() == semantic.Vector && c.Type().Nature() == semantic.Vector && a.Type().Nature() == semantic.Vector {
			return &conditionalVectorEvaluator{
				test:       test,
				consequent: c,
				alternate:  a,
			}, nil
		}

		return &conditionalEvaluator{
			test:       test,
			consequent: c,
			alternate:  a,
		}, nil
	case *semantic.BinaryExpression:
		l, err := compile(n.Left, subst)
		if err != nil {
			return nil, err
		}
		lt := l.Type().Nature()
		r, err := compile(n.Right, subst)
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
		if err == nil {
			return &binaryEvaluator{
				t:     apply(subst, nil, n.TypeOf()),
				left:  l,
				right: r,
				f:     f,
			}, nil
		}

		g, err := values.LookupBinaryVectorFunction(values.BinaryFuncSignature{
			Operator: n.Operator,
			Left:     lt,
			Right:    rt,
		})
		if err != nil {
			return nil, err
		}
		return &binaryVectorEvaluator{
			t:     apply(subst, nil, n.TypeOf()),
			left:  l,
			right: r,
			f:     g,
		}, nil
	case *semantic.CallExpression:
		args, err := compile(n.Arguments, subst)
		if err != nil {
			return nil, err
		}
		// Rewrite of "fake function" to vectorize constants.
		if c, ok := n.Callee.(*semantic.IdentifierExpression); ok && c.Name.LocalName == "~~vecRepeat~~" {
			v := args.(*objEvaluator).properties["v"]
			return &constVectorEvaluator{
				t: semantic.NewVectorType(v.Type()),
				v: v,
			}, nil
		}

		if n.Pipe != nil {
			pipeArg, err := n.Callee.TypeOf().PipeArgument()
			if err != nil {
				return nil, err

			}
			if pipeArg == nil {
				// This should be caught during type inference
				return nil, errors.Newf(codes.Internal, "callee lacks a pipe argument, but one was provided")
			}
			pipe, err := compile(n.Pipe, subst)
			if err != nil {
				return nil, err
			}
			args.(*objEvaluator).properties[string(pipeArg.Name())] = pipe
		}
		callee, err := compile(n.Callee, subst)
		if err != nil {
			return nil, err
		}
		return &callEvaluator{
			t:      apply(subst, nil, n.TypeOf()),
			callee: callee,
			args:   args,
		}, nil
	case *semantic.FunctionExpression:
		fnType := apply(subst, nil, n.TypeOf())
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
						d, err := compile(d.Value, subst)
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

func containsStr(strs []string, str string) bool {
	for _, s := range strs {
		if str == s {
			return true
		}
	}
	return false
}

// isNullable will report if the MonoType is capable of being nullable.
func isNullable(t semantic.MonoType) bool {
	n := t.Nature()
	return n != semantic.Array && n != semantic.Object && n != semantic.Dictionary
}
