package multirow

import (
	"context"
	"fmt"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/fbsemantic"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"reflect"
	"sort"
	"strings"
	"time"
	"unsafe"
)

const pkgpath = "contrib/lazarenkovegor/multirow"

type Part interface {
	AppendRow(row []values.Value) error
	AppendData(group *Group, row []values.Value) error
	End()
	Size() int
	Cols() []flux.ColMeta
	BaseKeyValue(column int) values.Value
	BaseKeyIndex(column int) int
	NewFullBaseKey() []values.Value
	DataColumnsNums() []int
	VirtualColumnsNums() []int
	KeyColumnsNums() []int
	LookupGroup(row []values.Value) *Group
}

type recordBuilder interface {
	Set(name string, value values.Value)
	Build() values.Object
}

//dynamicRecordBuilder - used for row value, skip nulls values
type dynamicRecordBuilder map[string]values.Value

//monoTypeRecordBuilder - used for table values, contains nulls values
type monoTypeRecordBuilder struct {
	values.Object
}
type Group struct {
	flux.GroupKey
	RowCount    int
	needReserve int
}

type TableBuilder struct {
	srcKey             flux.GroupKey
	mem                *memory.Allocator
	columnsBuilders    map[string]*colBuilder
	defaultValueColumn string
	currentPart        *part
	groups             map[string]*Group
	virtualColumns     []string
}

type colBuilder struct {
	groupValues map[*Group]array.Builder
	colMeta     flux.ColMeta
	index       int
	keyIndex    int
	isNewGroup  bool
	use         bool
	isVirtual   bool
}

type part struct {
	builder            *TableBuilder
	columns            []*colBuilder
	size               int
	keyColumnNums      []int
	dataColumnsNums    []int
	nullColumns        []*colBuilder
	virtualColumnsNums []int
}

var _ table.Builder = &TableBuilder{}

func (d dynamicRecordBuilder) Set(name string, value values.Value) {
	if value.IsNull() {
		return
	}
	d[name] = value
}

func (d dynamicRecordBuilder) Build() values.Object {
	return values.NewObjectWithValues(d)
}

func (d monoTypeRecordBuilder) Build() values.Object {
	return d.Object
}

func MakeRowObject(rowTp *semantic.MonoType, reader flux.ColReader, rowIndex int) values.Object {
	var builder recordBuilder
	if rowTp == nil {
		builder = &dynamicRecordBuilder{}
	} else {
		builder = &monoTypeRecordBuilder{Object: values.NewObject(*rowTp)}
	}

	var value values.Value
	for i, c := range reader.Cols() {
		switch c.Type {
		case flux.TBool:
			column := reader.Bools(i)
			if column.IsNull(rowIndex) {
				value = values.NewNull(semantic.BasicBool)
			} else {
				value = values.NewBool(column.Value(rowIndex))
			}
		case flux.TTime:
			column := reader.Times(i)
			if column.IsNull(rowIndex) {
				value = values.NewNull(semantic.BasicTime)
			} else {
				value = values.NewTime(values.Time(column.Value(rowIndex)))
			}
		case flux.TInt:
			column := reader.Ints(i)
			if column.IsNull(rowIndex) {
				value = values.NewNull(semantic.BasicInt)
			} else {
				value = values.NewInt(column.Value(rowIndex))
			}
		case flux.TUInt:
			column := reader.UInts(i)
			if column.IsNull(rowIndex) {
				value = values.NewNull(semantic.BasicUint)
			} else {
				value = values.NewUInt(column.Value(rowIndex))
			}
		case flux.TFloat:
			column := reader.Floats(i)
			if column.IsNull(rowIndex) {
				value = values.NewNull(semantic.BasicFloat)
			} else {
				value = values.NewFloat(column.Value(rowIndex))
			}
		case flux.TString:
			column := reader.Strings(i)
			if column.IsNull(rowIndex) {
				value = values.NewNull(semantic.BasicString)
			} else {
				value = values.NewString(column.Value(rowIndex))
			}

		default:
			panic(fmt.Errorf("unsupported column type %v", c.Type))
		}
		builder.Set(c.Label, value)
	}

	return builder.Build()
}

func ColsMonoType(cols []flux.ColMeta) (semantic.MonoType, error) {
	properties := make([]semantic.PropertyType, len(cols))
	for i, c := range cols {
		vtype := flux.SemanticType(c.Type)
		if vtype.Kind() == semantic.Unknown {
			return semantic.MonoType{}, errors.Newf(codes.Internal, "unknown column type: %s", c.Type)
		}
		properties[i] = semantic.PropertyType{
			Key:   []byte(c.Label),
			Value: vtype,
		}
	}
	return semantic.NewObjectType(properties), nil
}

func NewTableBuilder(GroupKey flux.GroupKey, Mem *memory.Allocator, DefaultValueColumn string, VirtualColumns []string) *TableBuilder {

	return &TableBuilder{
		groups:             map[string]*Group{"": {GroupKey: GroupKey}},
		srcKey:             GroupKey,
		mem:                Mem,
		columnsBuilders:    make(map[string]*colBuilder),
		defaultValueColumn: DefaultValueColumn,
		virtualColumns:     copyAndSort(VirtualColumns)}
}

func copyAndSort(src []string) []string {
	dst := make([]string, len(src))
	copy(dst, src)
	sort.Strings(dst)
	return dst
}

func appendBytes(ptr unsafe.Pointer, size int, builder *strings.Builder) {
	type Slice struct {
		Data unsafe.Pointer
		Len  int
		Cap  int
	}
	builder.Write(*(*[]byte)(unsafe.Pointer(&Slice{Data: ptr, Len: size, Cap: size})))
}

func MakeStringKey(values []values.Value) string {
	var sb strings.Builder
	for i, v := range values {
		if i > 0 {
			sb.WriteByte(255)
		}
		if v.IsNull() {
			continue
		}

		t, err := v.Type().Basic()
		if err != nil {
			panic(fmt.Errorf("MakeStringKey: %w", err))
		}
		switch t {
		case fbsemantic.TypeBool:
			val := v.Bool()
			const l = int(unsafe.Sizeof(val))
			appendBytes(unsafe.Pointer(&val), l, &sb)
			//sb.Write(asBytes(uintptr(unsafe.Pointer(&val)), l))
		case fbsemantic.TypeInt:
			val := v.Int()
			const l = int(unsafe.Sizeof(val))
			appendBytes(unsafe.Pointer(&val), l, &sb)
			//sb.Write(asBytes(uintptr(unsafe.Pointer(&val)), l))
		case fbsemantic.TypeUint:
			val := v.UInt()
			const l = int(unsafe.Sizeof(val))
			appendBytes(unsafe.Pointer(&val), l, &sb)
			//sb.Write(asBytes(uintptr(unsafe.Pointer(&val)), l))
		case fbsemantic.TypeFloat:
			val := v.Float()
			const l = int(unsafe.Sizeof(val))
			//sb.Write(asBytes(uintptr(unsafe.Pointer(&val)), l))
		case fbsemantic.TypeString:
			sb.WriteString(v.Str())
		case fbsemantic.TypeDuration:
			val := v.Duration()
			const l = int(unsafe.Sizeof(val))
			appendBytes(unsafe.Pointer(&val), l, &sb)
			//sb.Write(asBytes(uintptr(unsafe.Pointer(&val)), l))
		case fbsemantic.TypeTime:
			val := v.Time()
			const l = int(unsafe.Sizeof(val))
			appendBytes(unsafe.Pointer(&val), l, &sb)
			//sb.Write(asBytes(uintptr(unsafe.Pointer(&val)), l))
		case fbsemantic.TypeRegexp:
			sb.WriteString(v.Regexp().String())
		case fbsemantic.TypeBytes:
			sb.Write(v.Bytes())
		}
	}
	return sb.String()
}

func (s *part) Size() int {
	return s.size
}

func (s *part) Cols() []flux.ColMeta {
	res := make([]flux.ColMeta, len(s.columns))
	for i, v := range s.columns {
		res[i] = v.colMeta
	}
	return res
}

func (s *part) BaseKeyIndex(num int) int {
	return s.columns[num].keyIndex
}

func (s *part) BaseKeyValue(num int) values.Value {
	ki := s.columns[num].keyIndex
	return s.builder.srcKey.Value(ki)
}

func (s *part) NewFullBaseKey() []values.Value {
	src := s.builder.srcKey.Values()
	res := make([]values.Value, len(src))
	copy(res, src)
	return res
}

func (s *part) getBuilder(c *colBuilder, group *Group) array.Builder {
	if ab, found := c.groupValues[group]; found {
		return ab
	}

	ab := arrow.NewBuilder(c.colMeta.Type, s.builder.mem)
	rc := group.RowCount
	for i := 0; i < rc; i++ {
		ab.AppendNull()
	}
	c.groupValues[group] = ab
	return ab
}

func (s *part) LookupGroup(row []values.Value) *Group {
	if len(s.keyColumnNums) == 0 {
		return s.builder.groups[""]
	}
	var keyValue []values.Value
	for _, colNum := range s.keyColumnNums {
		value := row[colNum]
		if value.Equal(s.BaseKeyValue(colNum)) {
			continue
		}
		s.columns[colNum].isNewGroup = true
		if keyValue == nil {
			keyValue = s.NewFullBaseKey()
		}
		keyValue[s.BaseKeyIndex(colNum)] = value
	}

	if keyValue == nil {
		return s.builder.groups[""]
	}
	newGroupString := MakeStringKey(keyValue)

	if g, f := s.builder.groups[newGroupString]; f {
		return g
	}
	g := &Group{GroupKey: execute.NewGroupKey(s.builder.srcKey.Cols(), keyValue)}
	s.builder.groups[newGroupString] = g
	return g
}
func (s *part) VirtualColumnsNums() []int {
	return s.virtualColumnsNums
}
func (s *part) DataColumnsNums() []int {
	return s.dataColumnsNums
}

func (s *part) KeyColumnsNums() []int {
	return s.keyColumnNums
}

func (s *part) AppendData(group *Group, row []values.Value) error {
	for _, colNum := range s.dataColumnsNums {
		cb := s.columns[colNum]
		ab := s.getBuilder(cb, group)
		value := row[colNum]
		if value.IsNull() {
			ab.AppendNull()
			continue
		}
		switch b := ab.(type) {
		case *array.IntBuilder:
			if cb.colMeta.Type == flux.TTime {
				b.Append(int64(value.Time()))
			} else {
				b.Append(value.Int())
			}
		case *array.UintBuilder:
			b.Append(value.UInt())
		case *array.FloatBuilder:
			b.Append(value.Float())
		case *array.StringBuilder:
			b.Append(value.Str())
		case *array.BooleanBuilder:
			b.Append(value.Bool())
		default:
			return fmt.Errorf("unknown builder type: %v", reflect.TypeOf(ab))
		}
	}

	for _, cb := range s.nullColumns {
		s.getBuilder(cb, group).AppendNull()
	}
	return nil
}

func (s *part) AppendRow(row []values.Value) error {
	s.checkNoEnd()
	gr := s.LookupGroup(row)
	if err := s.AppendData(gr, row); err != nil {
		return err
	}
	gr.RowCount++
	return nil
}

func (s *part) checkNoEnd() {
	if s.builder == nil {
		panic(fmt.Errorf("missing or wrong BeginPart call before AppendColumnValues"))
	}
}

func (s *part) End() {
	s.checkNoEnd()

	s.builder.currentPart = nil
	s.builder = nil
}

type TableValues []array.Interface

func (s TableValues) Get(row, col int) values.Value {
	ca := s[col]
	if ca.IsNull(row) {
		return values.Null
	}
	switch b := ca.(type) {
	case *array.Int:
		return values.NewInt(b.Value(row))
	case *array.Uint:
		return values.NewUInt(b.Value(row))
	case *array.Float:
		return values.NewFloat(b.Value(row))
	case *array.String:
		return values.NewString(b.Value(row))
	case *array.Boolean:
		return values.NewBool(b.Value(row))
	default:
		panic(fmt.Errorf("unknown builder type: %v", reflect.TypeOf(ca)))
	}
}

func (s *TableBuilder) BeginPart(size int, colsMeta []flux.ColMeta) (Part, error) {
	if s.currentPart != nil {
		panic(fmt.Errorf("missing EndPart call before BeginPart"))
	}
	colCount := len(colsMeta)
	columns := make([]*colBuilder, colCount)
	keyColumnsNums := make([]int, 0, colCount)
	dataColumnsNums := make([]int, 0, colCount)
	nullColumns := make([]*colBuilder, 0, colCount)
	virtualColumnsNums := make([]int, 0, colCount)
	for colNum, colMeta := range colsMeta {
		if colMeta.Type == flux.TInvalid {
			continue
		}
		var (
			cb *colBuilder
			ok bool
		)
		if cb, ok = s.columnsBuilders[colMeta.Label]; !ok {
			keyCols := s.srcKey.Cols()
			keyColIndex := execute.ColIdx(colMeta.Label, keyCols)
			if keyColIndex > -1 && keyCols[keyColIndex].Type != colMeta.Type {
				return nil, fmt.Errorf("key column %s must be %v, but got %v", colMeta.Label, keyCols[keyColIndex].Type, colMeta.Type)
			}

			si := sort.SearchStrings(s.virtualColumns, colMeta.Label)
			isVirtual := si < len(s.virtualColumns) && s.virtualColumns[si] == colMeta.Label

			cb = &colBuilder{
				make(map[*Group]array.Builder),
				colMeta,
				len(s.columnsBuilders),
				keyColIndex,
				false,
				false,
				isVirtual,
			}
			s.columnsBuilders[colMeta.Label] = cb
		} else if cb.colMeta.Type != colMeta.Type {
			return nil, fmt.Errorf("column %s have diferent types %v and %v", colMeta.Label, cb.colMeta, colMeta.Type)
		}
		cb.use = true
		columns[colNum] = cb
		if cb.keyIndex > -1 {
			keyColumnsNums = append(keyColumnsNums, colNum)
		} else if !cb.isVirtual {
			dataColumnsNums = append(dataColumnsNums, colNum)
		} else {
			virtualColumnsNums = append(virtualColumnsNums, colNum)
		}
	}

	for _, cb := range s.columnsBuilders {
		if cb.keyIndex > -1 || cb.isVirtual || cb.use {
			cb.use = false
			continue
		}
		nullColumns = append(nullColumns, cb)
	}

	p := &part{s, columns, size, keyColumnsNums,
		dataColumnsNums, nullColumns, virtualColumnsNums}
	s.currentPart = p
	return p, nil
}

func (s *TableBuilder) Release() {
	for _, vv := range s.columnsBuilders {
		for _, v := range vv.groupValues {
			if v != nil {
				v.Release()
			}
		}
	}
}

func (s *TableBuilder) Table() (flux.Table, error) {
	if s.currentPart != nil {
		panic(fmt.Errorf("missing EndPart call before Build"))
	}
	var commonGroupKey []values.Value
	var commonGroupKeyMeta []flux.ColMeta
	for i, c := range s.srcKey.Cols() {
		if b, f := s.columnsBuilders[c.Label]; f && b.isNewGroup {
			continue
		}
		commonGroupKey = append(commonGroupKey, s.srcKey.Value(i))
		commonGroupKeyMeta = append(commonGroupKeyMeta, c)
	}

	keyCols := s.srcKey.Cols()
	keyColsCount := len(keyCols)

	columns := append(make([]flux.ColMeta, 0, len(keyCols)+len(s.columnsBuilders)), keyCols...)

	for n, v := range s.columnsBuilders {
		if v.keyIndex >= 0 || v.isVirtual {
			continue
		}
		columns = append(columns, flux.ColMeta{Label: n, Type: v.colMeta.Type})
	}
	dataColumns := columns[keyColsCount:]
	sort.Slice(dataColumns, func(i, j int) bool {
		return dataColumns[i].Label < dataColumns[j].Label
	})

	buffers := make([]flux.ColReader, 0, len(s.groups))

	for _, gr := range s.groups {
		if gr.RowCount == 0 {
			continue
		}
		buffer := arrow.TableBuffer{
			GroupKey: gr.GroupKey,
			Columns:  columns,
			Values:   make([]array.Interface, len(columns)),
		}
		buffers = append(buffers, &buffer)

		for colIndex, column := range columns {
			var a array.Interface
			if colIndex < keyColsCount {
				a = arrow.Repeat(keyCols[colIndex].Type, gr.Value(colIndex), gr.RowCount, s.mem)
			} else {
				a = s.columnsBuilders[column.Label].groupValues[gr].NewArray()
			}

			buffer.Values[colIndex] = a
		}
		if err := buffer.Validate(); err != nil {
			buffer.Release()
			return nil, err
		}

	}

	return &table.BufferedTable{
		GroupKey: execute.NewGroupKey(commonGroupKeyMeta, commonGroupKey),
		Columns:  columns,
		Buffers:  buffers,
	}, nil
}

func (s *TableBuilder) reserve() {
	for _, g := range s.groups {
		if g.needReserve == 0 {
			continue
		}
		for _, cb := range s.columnsBuilders {
			if cb.keyIndex > -1 {
				continue
			}
			if b, f := cb.groupValues[g]; f {
				b.Reserve(g.needReserve)
			}
		}
		g.needReserve = 0
	}
}

func (s *TableBuilder) AppendRows(ctx context.Context, res values.Value, needLastObject bool) (error, values.Object) {
	if to, ok := res.(*flux.TableObject); ok {
		pr, err := lang.CompileTableObject(ctx, to, time.Now())
		if err != nil {
			return err, nil
		}
		q, err := pr.Start(ctx, s.mem)
		if err != nil {
			return err, nil
		}
		r := <-q.Results()
		var obj values.Object
		err = r.Tables().Do(func(t flux.Table) error {
			return t.Do(func(reader flux.ColReader) error {
				rowCount := reader.Len()
				if rowCount == 0 {
					return nil
				}
				cols := reader.Cols()
				part, err := s.BeginPart(rowCount, cols)
				if err != nil {
					return err
				}
				defer part.End()

				colCount := len(cols)
				memTable := make(TableValues, colCount)
				for colNum := 0; colNum < colCount; colNum++ {
					memTable[colNum] = table.Values(reader, colNum)
				}

				rowGroup := make([]*Group, rowCount)
				for rowNum := 0; rowNum < rowCount; rowNum++ {
					row := make([]values.Value, colCount)
					for _, colNum := range part.KeyColumnsNums() {
						row[colNum] = memTable.Get(rowNum, colNum)
					}
					gr := part.LookupGroup(row)
					rowGroup[rowNum] = gr
					gr.needReserve++
				}
				s.reserve()

				for rowNum := 0; rowNum < rowCount; rowNum++ {
					row := make([]values.Value, colCount)
					for _, colNum := range part.DataColumnsNums() {
						row[colNum] = memTable.Get(rowNum, colNum)
					}
					gr := rowGroup[rowNum]
					if err := part.AppendData(gr, row); err != nil {
						return err
					}
					if rowNum == rowCount-1 && needLastObject {
						objKV := make(map[string]values.Value)
						cols := part.Cols()
						for _, colNum := range part.DataColumnsNums() {
							val := memTable.Get(rowNum, colNum)
							if val.IsNull() {
								continue
							}
							objKV[cols[colNum].Label] = val
						}
						for _, colNum := range part.VirtualColumnsNums() {
							val := memTable.Get(rowNum, colNum)
							if val.IsNull() {
								continue
							}
							objKV[cols[colNum].Label] = val
						}
						for i, m := range gr.GroupKey.Cols() {
							val := gr.GroupKey.Value(i)
							if val.IsNull() {
								continue
							}
							objKV[m.Label] = val
						}
						obj = values.NewObjectWithValues(objKV)
					}
					gr.RowCount++
				}

				return nil
			})
		})

		return err, obj
	}

	switch res.Type().Kind() {
	case semantic.Arr:
		var (
			err error
			obj values.Object
		)
		res.Array().Range(func(i int, v values.Value) {
			if err != nil {
				return
			}
			obj = v.Object()
			err = s.addRecord(obj)
		})
		return err, obj
	case semantic.Record:
		obj := res.Object()
		return s.addRecord(obj), obj
	default:
		obj := values.NewObjectWithValues(map[string]values.Value{s.defaultValueColumn: res})
		return s.addRecord(obj), obj
	}
}

func (s *TableBuilder) addRecord(v values.Object) error {
	l := v.Len()
	colsMeta := make([]flux.ColMeta, l)
	colsValues := make([]values.Value, l)

	i := 0
	v.Range(func(name string, v values.Value) {
		colsMeta[i] = flux.ColMeta{
			Label: name,
			Type:  flux.ColumnType(v.Type()),
		}
		colsValues[i] = v
		i++
	})

	batch, err := s.BeginPart(1, colsMeta)
	if err != nil {
		return err
	}
	defer batch.End()

	return batch.AppendRow(colsValues)
}
