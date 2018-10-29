package semantic

type Fresher interface {
	Fresh() Tvar
}

type fresher Tvar

func (f *fresher) Fresh() Tvar {
	fresh := *f
	(*f)++
	return Tvar(fresh)
}
