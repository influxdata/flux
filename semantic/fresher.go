package semantic

type Fresher interface {
	Fresh() Tvar
}

func NewFresher() Fresher {
	return new(fresher)
}

type fresher Tvar

func (f *fresher) Fresh() Tvar {
	(*f)++
	return Tvar(*f)
}
