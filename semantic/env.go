package semantic

type Env struct {
	parent *Env
	m      map[string]Scheme
}

func NewEnv() *Env {
	return &Env{
		m: make(map[string]Scheme),
	}
}

func (e *Env) LocalLookup(ident string) (Scheme, bool) {
	s, ok := e.m[ident]
	return s, ok
}

func (e *Env) Lookup(ident string) (Scheme, bool) {
	s, ok := e.m[ident]
	if ok {
		return s, true
	}
	if e.parent != nil {
		return e.parent.Lookup(ident)
	}
	return Scheme{}, false
}

func (e *Env) Set(ident string, s Scheme) {
	e.m[ident] = s
}

func (e *Env) Nest() *Env {
	n := NewEnv()
	n.parent = e
	return n
}

func (e *Env) FreeVars() TvarSet {
	var ftv TvarSet
	for _, s := range e.m {
		ftv = ftv.union(s.Free)
	}
	return ftv
}

func (e *Env) RangeSet(f func(k string, v Scheme) Scheme) {
	for k, v := range e.m {
		e.m[k] = f(k, v)
	}
	if e.parent != nil {
		e.parent.RangeSet(f)
	}
}
