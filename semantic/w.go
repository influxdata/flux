package semantic

import "fmt"

type T interface {
	Unsolved() []TV
	Instantiate(tm map[int]TV) T
	Unify(t T) error
}

func (k Kind) Unsolved() []TV {
	return nil
}
func (k Kind) Instantiate(map[int]TV) T {
	return k
}
func (k Kind) Unify(T) error {
	return nil
}

type Env map[string]TS

func (e Env) EnvUnsolved() []TV {
	var u []TV
	for _, ts := range e {
		u = union(u, ts.Unsolved())
	}
	return u
}

func (e Env) Lookup(n string) TS {
	return e[n]
}
func (e Env) Extend(d *NativeVariableDeclaration, f *Fresher) Env {
	ee := make(Env, len(e)+1)
	for k, v := range e {
		ee[k] = v
	}
	switch d.Init.(type) {
	case *FunctionExpression:
		panic("i have no idea")
	default:
		ee[d.Identifier.Name] = schema(tcheck(e, d.Init, f), e)
	}
	return ee
}

type Fresher int

func (f *Fresher) Fresh() TV {
	v := int(*f)
	(*f)++
	return newTV(v)
}

type TV struct {
	V int
	T **T
}

func newTV(v int) TV {
	return TV{
		V: v,
		T: new(*T),
	}
}

func (tv TV) Unify(t T) error {
	return unifyVar(t, tv.T)
}

func unifyVar(t T, r **T) error {
	if *r != nil {
		return t.Unify(**r)
	}
	if tv2, ok := t.(TV); ok && *r == *tv2.T {
		// Cyclic, no need to update
		return nil
	}
	*r = &t
	return nil
}

func (tv TV) Unsolved() []TV {
	if tv.T != nil {
		return (**tv.T).Unsolved()
	}
	return []TV{tv}
}

func (tv1 TV) Instantiate(tm map[int]TV) T {
	if tv2, ok := tm[tv1.V]; ok {
		return tv2
	}
	return tv1
}

type TS struct {
	T    T
	List []TV
}

func (ts TS) Unsolved() []TV {
	return ts.T.Unsolved()
}

func union(a, b []TV) []TV {
	u := a
	for _, v := range b {
		found := false
		for _, f := range a {
			if f == v {
				found = true
				break
			}
		}
		if !found {
			u = append(u, v)
		}
	}
	return u
}

func diff(a, b []TV) []TV {
	d := make([]TV, 0, len(a))
	for _, v := range a {
		found := false
		for _, f := range b {
			if f == v {
				found = true
				break
			}
		}
		if !found {
			d = append(d, v)
		}
	}
	return d
}

func schema(t T, e Env) TS {
	uv := t.Unsolved()
	ev := e.EnvUnsolved()
	d := diff(uv, ev)
	return TS{
		T:    t,
		List: d,
	}
}

func instantiate(ts TS, f *Fresher) T {
	tm := make(map[int]TV, len(ts.List))
	for _, tv := range ts.List {
		tm[tv.V] = f.Fresh()
	}
	return ts.T.Instantiate(tm)
}

func tcheck(env Env, node Node, f *Fresher) T {
	switch n := node.(type) {
	case *IdentifierExpression:
		return instantiate(env.Lookup(n.Name), f)
	case *BooleanLiteral:
		return Bool
	case *IntegerLiteral:
		return Int
	case *NativeVariableDeclaration: // Let
		tcheck(env.Extend(n, f), n.Init, f)
		return tcheck(n.Body)
	default:
		panic(fmt.Sprintf("unsupported, %T", node))
	}
}

func Infer(n Node) Type {
	env := make(Env)
	f := new(Fresher)
	t := tcheck(env, n, f)
	return t.(Type)
}
