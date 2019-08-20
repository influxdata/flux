package values

type Package interface {
	Object
	SetOption(name string, v Value)
}
