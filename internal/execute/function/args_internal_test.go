package function

func (t *TableObject) Equal(other *TableObject) bool {
	return t.Kind == other.Kind
}
