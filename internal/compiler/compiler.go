package compiler

import (
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func Compile(_ values.Scope, f *semantic.FunctionExpression, in semantic.MonoType) (Func, error) {
	// if scope == nil {
	// 	scope = NewScope()
	// }
	if in.Nature() != semantic.Object {
		return nil, errors.Newf(codes.Invalid, "function input must be an object @ %v", f.Location())
	}

	// Retrieve the number of input arguments which may differ from the
	// defined arguments.
	inArgNum, err := in.NumProperties()
	if err != nil {
		return nil, err
	}
	params := make([]valueMapper, inArgNum)

	// Retrieve the function argument types and create an object type from them.
	fnType := f.TypeOf()
	argN, err := fnType.NumArguments()
	if err != nil {
		return nil, err
	}

	fn := compiledFn{scope: NewScope(), params: params}

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

		propIndex, err := findPropertyIndex(string(name), in)
		if err != nil {
			return nil, err
		} else if propIndex >= 0 {
			prop, err := in.RecordProperty(propIndex)
			if err != nil {
				return nil, err
			}
			mtyp, err := prop.TypeOf()
			if err != nil {
				return nil, err
			}
			if err := substituteTypes(subst, argT, mtyp); err != nil {
				return nil, err
			}
			register := fn.scope.Declare()
			params[propIndex] = basicValueMapper(register)

			fn.scope.Define(
				prop.Name(),
				fn.staticCast(
					apply(subst, nil, argT),
					mtyp,
					register,
				),
			)
		} else if !arg.Optional() {
			return nil, errors.Newf(codes.Invalid, "missing required argument %q", string(name))
		}
	}

	ret, t, err := fn.compile(f.Block, subst)
	if err != nil {
		return nil, errors.Wrapf(err, codes.Inherit, "cannot compile @ %v", f.Location())
	}
	fn.ret, fn.t = ret, t
	return fn, nil
}

// substituteTypes will generate a substitution map by recursing through
// inType and mapping any variables to the value in the other record.
// If the input type is not a type variable, it will check to ensure
// that the type in the input matches or it will return an error.
func substituteTypes(subst map[uint64]semantic.MonoType, inferredType, actualType semantic.MonoType) error {
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
		if t, ok := subst[vn]; ok {
			return substituteTypes(subst, t, actualType)
		}

		// If the input type is not invalid, mark it down
		// as the real type.
		if actualType.Nature() != semantic.Invalid {
			subst[vn] = actualType
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
	case semantic.Arr:
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

func findPropertyIndex(name string, t semantic.MonoType) (i int, err error) {
	n, err := t.NumProperties()
	if err != nil {
		return -1, err
	}
	for i := 0; i < n; i++ {
		p, err := t.RecordProperty(i)
		if err != nil {
			return -1, err
		}
		if p.Name() == name {
			return i, nil
		}
	}
	return -1, nil
}

// apply applies a substitution to a type.
// It will ignore any errors when reading a type.
// This is safe becase we already validated that the function type is a monotype.
func apply(sub map[uint64]semantic.MonoType, props []semantic.PropertyType, t semantic.MonoType) semantic.MonoType {
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
		ty, ok := sub[tv]
		if !ok {
			return t
		}
		return ty
	case semantic.Arr:
		element, err := t.ElemType()
		if err != nil {
			return t
		}
		return semantic.NewArrayType(apply(sub, props, element))
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
func (c *compiledFn) compile(n semantic.Node, subst map[uint64]semantic.MonoType) (int, semantic.MonoType, error) {
	switch n := n.(type) {
	case *semantic.Block:
		for _, s := range n.Body[:len(n.Body)-1] {
			if _, _, err := c.compile(s, subst); err != nil {
				return -1, semantic.MonoType{}, err
			}
		}
		return c.compile(n.Body[len(n.Body)-1], subst)
	case *semantic.ExpressionStatement:
		return -1, semantic.MonoType{}, errors.New(codes.Internal, "statement does nothing, side effects are not supported by the compiler")
	case *semantic.ReturnStatement:
		return c.compile(n.Argument, subst)
	case *semantic.NativeVariableAssignment:
		reg, mt, err := c.compile(n.Init, subst)
		if err != nil {
			return -1, semantic.MonoType{}, err
		}
		c.scope.Define(n.Identifier.Name, reg)
		return reg, mt, nil
	case *semantic.ObjectExpression:
		// Note: The order of the properties in the underlying
		// type appears to have anything within the with clause
		// as after anything in the property list. Still, we
		// would generally expect that the target of with would
		// be executed before the properties that are shadowing
		// it so we execute with first.
		with := -1
		if n.With != nil {
			var err error
			with, _, err = c.compile(n.With, subst)
			if err != nil {
				return -1, semantic.MonoType{}, err
			}
		}

		// Evaluate each of the properties.
		labels := make([]int, len(n.Properties))
		for i, p := range n.Properties {
			reg, _, err := c.compile(p.Value, subst)
			if err != nil {
				return -1, semantic.MonoType{}, err
			}
			labels[i] = reg
		}

		e := &recordEvaluator{
			evaluator: evaluator{
				t:   apply(subst, nil, n.TypeOf()),
				ret: c.scope.Declare(),
			},
			labels: labels,
			with:   with,
		}
		c.scope.Set(e.ret, NewRecord(e.t))
		c.body = append(c.body, e)
		return e.ret, e.t, nil
	case *semantic.ArrayExpression:
		t := apply(subst, nil, n.TypeOf())
		elemT, err := t.ElemType()
		if err != nil {
			return -1, semantic.MonoType{}, err
		}

		var elements []int
		if len(n.Elements) > 0 {
			elements = make([]int, len(n.Elements))
			for i, e := range n.Elements {
				index, mt, err := c.compile(e, subst)
				if err != nil {
					return -1, semantic.MonoType{}, err
				}
				elements[i] = c.staticCast(elemT, mt, index)
			}
		}

		e := &arrayEvaluator{
			evaluator: evaluator{
				t:   t,
				ret: c.scope.Declare(),
			},
			array: elements,
		}
		c.body = append(c.body, e)
		return e.ret, e.t, nil
	// case *semantic.DictExpression:
	// 	elements := make([]struct {
	// 		Key Evaluator
	// 		Val Evaluator
	// 	}, len(n.Elements))
	// 	for i, item := range n.Elements {
	// 		key, err := compile(item.Key, subst)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		val, err := compile(item.Val, subst)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		elements[i] = struct {
	// 			Key Evaluator
	// 			Val Evaluator
	// 		}{Key: key, Val: val}
	// 	}
	// 	return &dictEvaluator{
	// 		t:        apply(subst, nil, n.TypeOf()),
	// 		elements: elements,
	// 	}, nil
	case *semantic.IdentifierExpression:
		reg := c.scope.Get(n.Name)
		return reg, apply(subst, nil, n.TypeOf()), nil
	case *semantic.MemberExpression:
		object, mt, err := c.compile(n.Object, subst)
		if err != nil {
			return -1, semantic.MonoType{}, err
		}
		property, err := findPropertyIndex(n.Property, mt)
		if err != nil {
			return -1, semantic.MonoType{}, err
		}

		t := apply(subst, nil, n.TypeOf())
		ret := c.scope.Declare()
		if property < 0 {
			if !isNullable(t) {
				return -1, semantic.MonoType{}, errors.Newf(codes.Invalid, "member %q with type %s is not in the record", n.Property, n.TypeOf().Nature())
			}
			return ret, t, nil
		}
		c.body = append(c.body, &memberEvaluator{
			evaluator: evaluator{
				t:   t,
				ret: ret,
			},
			object:   object,
			property: property,
		})
		return ret, t, nil
	case *semantic.IndexExpression:
		arr, _, err := c.compile(n.Array, subst)
		if err != nil {
			return -1, semantic.MonoType{}, err
		}
		idx, _, err := c.compile(n.Index, subst)
		if err != nil {
			return -1, semantic.MonoType{}, err
		}

		e := &arrayIndexEvaluator{
			evaluator: evaluator{
				t:   apply(subst, nil, n.TypeOf()),
				ret: c.scope.Declare(),
			},
			array: arr,
			index: idx,
		}
		c.body = append(c.body, e)
		return e.ret, e.t, nil
	case *semantic.StringExpression:
		parts := make([]stringPartEvaluator, len(n.Parts))
		for i, p := range n.Parts {
			switch p := p.(type) {
			case *semantic.TextPart:
				parts[i] = &textEvaluator{value: p.Value}
			case *semantic.InterpolatedPart:
				index, _, err := c.compile(p.Expression, subst)
				if err != nil {
					return -1, semantic.MonoType{}, err
				}
				parts[i] = &interpolatedEvaluator{
					index: index,
				}
			default:
				return -1, semantic.MonoType{}, errors.Newf(codes.Internal, "unknown string part node: %T", p)
			}
		}

		e := &stringExpressionEvaluator{
			evaluator: evaluator{
				t:   semantic.BasicString,
				ret: c.scope.Declare(),
			},
			parts: parts,
		}
		c.body = append(c.body, e)
		return e.ret, e.t, nil
	case *semantic.BooleanLiteral:
		v := NewBool(n.Value)
		return c.scope.Push(v), semantic.BasicBool, nil
	case *semantic.IntegerLiteral:
		v := NewInt(n.Value)
		return c.scope.Push(v), semantic.BasicInt, nil
	case *semantic.UnsignedIntegerLiteral:
		v := NewUint(n.Value)
		return c.scope.Push(v), semantic.BasicUint, nil
	case *semantic.FloatLiteral:
		v := NewFloat(n.Value)
		return c.scope.Push(v), semantic.BasicFloat, nil
	case *semantic.StringLiteral:
		v := NewString(n.Value)
		return c.scope.Push(v), semantic.BasicString, nil
	case *semantic.RegexpLiteral:
		v := NewRegexp(n.Value)
		return c.scope.Push(v), semantic.BasicRegexp, nil
	case *semantic.DateTimeLiteral:
		v := NewTime(values.ConvertTime(n.Value))
		return c.scope.Push(v), semantic.BasicTime, nil
	case *semantic.DurationLiteral:
		d, err := values.FromDurationValues(n.Values)
		if err != nil {
			return -1, semantic.MonoType{}, err
		}
		v := NewDuration(d)
		return c.scope.Push(v), semantic.BasicDuration, nil
	case *semantic.UnaryExpression:
		index, mt, err := c.compile(n.Argument, subst)
		if err != nil {
			return -1, semantic.MonoType{}, err
		}

		switch n.Operator {
		case ast.AdditionOperator:
			// Nothing is done with this operator.
			// Pass through the index and monotype as if this
			// was completely ignored.
			return index, mt, nil
		case ast.SubtractionOperator, ast.NotOperator:
			e := &unaryNotEvaluator{
				evaluator: evaluator{
					t:   apply(subst, nil, n.TypeOf()),
					ret: c.scope.Declare(),
				},
				index: index,
			}
			c.body = append(c.body, e)
			return e.ret, e.t, nil
		case ast.ExistsOperator:
			e := &unaryExistsEvaluator{
				evaluator: evaluator{
					t:   apply(subst, nil, n.TypeOf()),
					ret: c.scope.Declare(),
				},
				index: index,
			}
			c.body = append(c.body, e)
			return e.ret, e.t, nil
		default:
			return -1, semantic.MonoType{}, errors.Newf(codes.Internal, "unknown unary operator: %s", n.Operator)
		}
	case *semantic.LogicalExpression:
		// Translate this expression to a conditional expression.
		switch n.Operator {
		case ast.AndOperator:
			expr := &semantic.ConditionalExpression{
				Test:       n.Left,
				Consequent: n.Right,
				Alternate:  &semantic.BooleanLiteral{Value: false},
			}
			return c.compile(expr, subst)
		case ast.OrOperator:
			expr := &semantic.ConditionalExpression{
				Test:       n.Left,
				Consequent: &semantic.BooleanLiteral{Value: true},
				Alternate:  n.Right,
			}
			return c.compile(expr, subst)
		default:
			return -1, semantic.MonoType{}, errors.Newf(codes.Internal, "unknown logical operator %v", n.Operator)
		}
	case *semantic.ConditionalExpression:
		test, _, err := c.compile(n.Test, subst)
		if err != nil {
			return -1, semantic.MonoType{}, err
		}

		// Construct the branch for the test condition.
		// We will then compile both of the branches
		// and fill in the jump indices to this branch.
		cond := &branch{
			test: test,
			// Don't fill in the jump points yet.
		}
		c.body = append(c.body, cond)

		// Compile the consequent. We need the index and type
		// for the consequent to use in the phi node later.
		cond.t = len(c.body)
		consequent, mt, err := c.compile(n.Consequent, subst)
		if err != nil {
			return -1, semantic.MonoType{}, err
		}

		// Prepare the jump for this side of the branch.
		// We do not know where we are jumping yet.
		jmp := &jump{}
		c.body = append(c.body, jmp)
		label1 := len(c.body) - 1

		// Set the false branch from earlier to the location
		// of the alternate and then compile the alternate.
		cond.f = len(c.body)
		alternate, _, err := c.compile(n.Alternate, subst)
		if err != nil {
			return -1, semantic.MonoType{}, err
		}

		// Prepare the jump for this side of the branch.
		// It's the same jump as before but we append it
		// again in a different space.
		c.body = append(c.body, jmp)
		label2 := len(c.body) - 1

		// Construct the phi node.
		jmp.to = len(c.body)
		e := &phi{
			evaluator: evaluator{
				t:   mt,
				ret: c.scope.Declare(),
			},
			label1: label1,
			index1: consequent,
			label2: label2,
			index2: alternate,
		}
		c.body = append(c.body, e)
		return e.ret, e.t, nil
	case *semantic.BinaryExpression:
		// TODO(jsternberg): Catch null values when they are determined
		// to be null by the type system.
		l, ltyp, err := c.compile(n.Left, subst)
		if err != nil {
			return -1, semantic.MonoType{}, err
		}

		r, rtyp, err := c.compile(n.Right, subst)
		if err != nil {
			return -1, semantic.MonoType{}, err
		}

		ret := c.scope.Declare()
		switch n.Operator {
		case ast.EqualOperator:
			e, err := newEqualsEvaluator(ret, l, r, ltyp, rtyp)
			if err != nil {
				return -1, semantic.MonoType{}, err
			}
			c.body = append(c.body, e)
			return ret, semantic.BasicBool, nil
		case ast.NotEqualOperator:
			e, err := newNotEqualsEvaluator(ret, l, r, ltyp, rtyp)
			if err != nil {
				return -1, semantic.MonoType{}, err
			}
			c.body = append(c.body, e)
			return ret, semantic.BasicBool, nil
		case ast.LessThanOperator:
			e, err := newLessThanEvaluator(ret, l, r, ltyp, rtyp)
			if err != nil {
				return -1, semantic.MonoType{}, err
			}
			c.body = append(c.body, e)
			return ret, semantic.BasicBool, nil
		case ast.GreaterThanOperator:
			e, err := newGreaterThanEvaluator(ret, l, r, ltyp, rtyp)
			if err != nil {
				return -1, semantic.MonoType{}, err
			}
			c.body = append(c.body, e)
			return ret, semantic.BasicBool, nil
		case ast.LessThanEqualOperator:
			e, err := newLessThanEqualEvaluator(ret, l, r, ltyp, rtyp)
			if err != nil {
				return -1, semantic.MonoType{}, err
			}
			c.body = append(c.body, e)
			return ret, semantic.BasicBool, nil
		case ast.GreaterThanEqualOperator:
			e, err := newGreaterThanEqualEvaluator(ret, l, r, ltyp, rtyp)
			if err != nil {
				return -1, semantic.MonoType{}, err
			}
			c.body = append(c.body, e)
			return ret, semantic.BasicBool, nil
		case ast.AdditionOperator:
			e, err := newAddEvaluator(ret, l, r, ltyp)
			if err != nil {
				return -1, semantic.MonoType{}, err
			}
			c.body = append(c.body, e)
			return ret, e.Type(), nil
		case ast.SubtractionOperator:
			e, err := newSubtractEvaluator(ret, l, r, ltyp)
			if err != nil {
				return -1, semantic.MonoType{}, err
			}
			c.body = append(c.body, e)
			return ret, e.Type(), nil
		case ast.MultiplicationOperator:
			e, err := newMulEvaluator(ret, l, r, ltyp)
			if err != nil {
				return -1, semantic.MonoType{}, err
			}
			c.body = append(c.body, e)
			return ret, e.Type(), nil
		case ast.DivisionOperator:
			e, err := newDivEvaluator(ret, l, r, ltyp)
			if err != nil {
				return -1, semantic.MonoType{}, err
			}
			c.body = append(c.body, e)
			return ret, e.Type(), nil
		case ast.ModuloOperator:
			e, err := newModEvaluator(ret, l, r, ltyp)
			if err != nil {
				return -1, semantic.MonoType{}, err
			}
			c.body = append(c.body, e)
			return ret, e.Type(), nil
		case ast.RegexpMatchOperator:
			e := &regexpMatchEvaluator{
				evaluator: evaluator{
					t:   semantic.BasicBool,
					ret: ret,
				},
				left:  l,
				right: r,
			}
			c.body = append(c.body, e)
			return ret, e.t, nil
		case ast.NotRegexpMatchOperator:
			e := &regexpNotMatchEvaluator{
				evaluator: evaluator{
					t:   semantic.BasicBool,
					ret: ret,
				},
				left:  l,
				right: r,
			}
			c.body = append(c.body, e)
			return ret, e.t, nil
		default:
			return -1, semantic.MonoType{}, errors.New(codes.Unimplemented)
		}
	// 	lt := l.Type().Nature()
	// 	r, err := compile(n.Right, subst)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	rt := r.Type().Nature()
	// 	if lt == semantic.Invalid {
	// 		lt = rt
	// 	} else if rt == semantic.Invalid {
	// 		rt = lt
	// 	}
	// 	f, err := LookupBinaryFunction(BinaryFuncSignature{
	// 		Operator: n.Operator,
	// 		Left:     lt,
	// 		Right:    rt,
	// 	})
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return &binaryEvaluator{
	// 		t:     apply(subst, nil, n.TypeOf()),
	// 		left:  l,
	// 		right: r,
	// 		f:     f,
	// 	}, nil
	// case *semantic.CallExpression:
	// 	callee, err := compile(n.Callee, subst)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	fnType := callee.Type()
	// 	nargs, err := fnType.NumArguments()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	var pipeArgName string
	// 	if pipe, err := fnType.PipeArgument(); err != nil {
	// 		return nil, err
	// 	} else if pipe != nil {
	// 		pipeArgName = string(pipe.Name())
	// 		nargs++
	// 	}
	//
	// 	args, offset := make([]Evaluator, nargs), 0
	// 	if n.Pipe != nil {
	// 		// Pipe argument is always first.
	// 		if pipeArgName == "" {
	// 			// TODO(jsternberg): Investigate if this is still needed.
	// 			// It was in the original code and it doesn't seem to cause
	// 			// significant harm, but we should have test cases that hit each
	// 			// of the error conditions here and either fix the cause of those
	// 			// conditions or have an explicit rationale why it should be caught
	// 			// at this phase.
	// 			// This should be caught during type inference
	// 			return nil, errors.Newf(codes.Internal, "callee lacks a pipe argument, but one was provided")
	// 		}
	//
	// 		pipe, err := compile(n.Pipe, subst)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		args[0], offset = pipe, 1
	// 	}
	//
	// 	// We compile the object expression here in a special way.
	// 	// The object expression here is an artifact of the semantic graph
	// 	// and how it passes arguments rather than a real entity that has
	// 	// all possibilities of an object expression. We know that the
	// 	// with attribute will not be used. We know that each entry in the
	// 	// object expression will be its own key/value in the object expression.
	// 	// For that reason, we will look at each argument, find that argument
	// 	// in the object expression, then compile it in that location to map
	// 	// the expression with the parameter.
	// 	// This skips an section of indirection where we have to create a useless
	// 	// record to just pass the values to the function where it will then
	// 	// not use the record like a record.
	// 	for i := offset; i < nargs; i++ {
	// 		argInfo, err := fnType.Argument(i - offset)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	//
	// 		argName := string(argInfo.Name())
	// 		for _, prop := range n.Arguments.Properties {
	// 			if prop.Key.Key() == argName {
	// 				arg, err := compile(prop.Value, subst)
	// 				if err != nil {
	// 					return nil, err
	// 				}
	// 				args[i] = arg
	// 				break
	// 			}
	// 		}
	// 	}
	// 	return &callEvaluator{
	// 		t:      apply(subst, nil, n.TypeOf()),
	// 		callee: callee,
	// 		args:   args,
	// 	}, nil
	// case *semantic.FunctionExpression:
	// 	fnType := apply(subst, nil, n.TypeOf())
	// 	num, err := fnType.NumArguments()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	params := make([]functionParam, 0, num)
	// 	for i := 0; i < num; i++ {
	// 		arg, err := fnType.Argument(i)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		k := string(arg.Name())
	// 		pt, err := arg.TypeOf()
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		param := functionParam{
	// 			Key:  k,
	// 			Type: pt,
	// 		}
	// 		if n.Defaults != nil {
	// 			// Search for default value
	// 			for _, d := range n.Defaults.Properties {
	// 				if d.Key.Key() == k {
	// 					d, err := compile(d.Value, subst)
	// 					if err != nil {
	// 						return nil, err
	// 					}
	// 					param.Default = d
	// 					break
	// 				}
	// 			}
	// 		}
	// 		params = append(params, param)
	// 	}
	// 	return &functionEvaluator{
	// 		t:      fnType,
	// 		params: params,
	// 		fn:     n,
	// 	}, nil
	default:
		return -1, semantic.MonoType{}, errors.Newf(codes.Internal, "unknown semantic node of type %T", n)
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
