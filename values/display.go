package values

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/influxdata/flux/semantic"
)

// DisplayString formats the value into a string
func DisplayString(v Value) string {
	b := strings.Builder{}
	_ = Display(&b, v)
	return b.String()
}

// Display formats the value into the writer
func Display(w io.Writer, v Value) error {
	bw := bufio.NewWriter(w)
	err := display(bw, v, 0)
	if err != nil {
		return err
	}
	return bw.Flush()
}

func display(w *bufio.Writer, v Value, indent int) (err error) {
	if v.IsNull() {
		_, err = w.WriteString("<null>")
		return
	}
	switch v.Type().Nature() {
	default:
		_, err = w.WriteString("<unknown value>")
		return
	case semantic.Invalid:
		_, err = w.WriteString("<invalid>")
		return
	case semantic.String:
		_, err = w.WriteString(v.Str())
		return
	case semantic.Bytes:
		_, err = fmt.Fprint(w, v.Bytes())
		return
	case semantic.Int:
		_, err = fmt.Fprint(w, v.Int())
		return
	case semantic.UInt:
		_, err = fmt.Fprint(w, v.UInt())
		return
	case semantic.Float:
		_, err = fmt.Fprint(w, v.Float())
		return
	case semantic.Bool:
		_, err = fmt.Fprint(w, v.Bool())
		return
	case semantic.Time:
		_, err = w.WriteString(v.Time().String())
		return
	case semantic.Duration:
		_, err = w.WriteString(v.Duration().String())
		return
	case semantic.Regexp:
		_, err = w.WriteString(v.Regexp().String())
		return
	case semantic.Array:
		a := v.Array()
		multiline := a.Len() > 3
		_, err = w.WriteString("[")
		if err != nil {
			return
		}
		if multiline {
			err = newline(w, indent+1)
			if err != nil {
				return
			}
		}
		a.Range(func(i int, v Value) {
			if err != nil {
				return
			}
			if i != 0 {
				_, err = w.WriteString(", ")
				if err != nil {
					return
				}
				if multiline {
					err = newline(w, indent+1)
					if err != nil {
						return
					}
				}
			}
			err = display(w, v, indent+1)
		})
		if err != nil {
			return
		}
		if multiline {
			err = newline(w, indent)
			if err != nil {
				return
			}
		}
		_, err = w.WriteString("]")
		return
	case semantic.Object:
		o := v.Object()
		multiline := o.Len() > 3
		_, err = w.WriteString("{")
		if err != nil {
			return
		}
		if multiline {
			err = newline(w, indent+1)
			if err != nil {
				return
			}
		}
		keys := make([]string, 0, o.Len())
		o.Range(func(k string, v Value) {
			keys = append(keys, k)
		})
		sort.Strings(keys)
		for i, k := range keys {
			v, _ := o.Get(k)
			if i != 0 {
				_, err = w.WriteString(", ")
				if err != nil {
					return
				}
				if multiline {
					err = newline(w, indent+1)
					if err != nil {
						return
					}
				}
			}
			i++
			_, err = w.WriteString(k)
			if err != nil {
				return
			}
			_, err = w.WriteString(": ")
			if err != nil {
				return
			}
			err = display(w, v, indent+1)
			if err != nil {
				return
			}
		}
		if err != nil {
			return
		}
		if multiline {
			err = newline(w, indent)
			if err != nil {
				return
			}
		}
		_, err = w.WriteString("}")
		return
	case semantic.Function:
		_, err = w.WriteString(v.Type().CanonicalString())
		return
	case semantic.Dictionary:
		d := v.Dict()
		if d.Len() == 0 {
			_, err = w.WriteString("[:]")
			return
		}
		multiline := d.Len() > 3
		_, err = w.WriteString("[")
		if err != nil {
			return
		}
		if multiline {
			err = newline(w, indent+1)
			if err != nil {
				return
			}
		}
		i := 0
		d.Range(func(k, v Value) {
			if err != nil {
				return
			}
			if i != 0 {
				_, err = w.WriteString(", ")
				if err != nil {
					return
				}
				if multiline {
					err = newline(w, indent+1)
					if err != nil {
						return
					}
				}
			}
			i++
			err = display(w, k, indent+1)
			if err != nil {
				return
			}
			_, err = w.WriteString(": ")
			if err != nil {
				return
			}
			err = display(w, v, indent+1)
			if err != nil {
				return
			}
		})
		if err != nil {
			return
		}
		if multiline {
			err = newline(w, indent)
			if err != nil {
				return
			}
		}
		_, err = w.WriteString("]")
		return
	}
}

const indentStr = "    "

func writeIndent(w *bufio.Writer, indent int) (err error) {
	for i := 0; i < indent; i++ {
		_, err = w.WriteString(indentStr)
		if err != nil {
			return
		}
	}
	return
}
func newline(w *bufio.Writer, indent int) (err error) {
	_, err = w.WriteRune('\n')
	if err != nil {
		return
	}
	return writeIndent(w, indent)
}
