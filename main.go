package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/raphaeldichler/zeus/internal/assert"
)

const tabWidth = 4

type Brithday struct {
	Day   int
	Month string
	Year  int
}

type Person struct {
	Name string
	Age  int
	City string
	Cool bool
	BDay Brithday
}

type output struct {
	builder strings.Builder
}

func (o *output) table(tab any, intend int) {
	prefix := strings.Repeat(" ", tabWidth*intend)
	slice := reflect.ValueOf(tab)
	assert.True(slice.Kind() == reflect.Slice, "we only do slices")
	var colWidth []int = nil
	var colNames []string = nil
	var rows [][]string = nil

	for eleIdx := range slice.Len() {
		obj := slice.Index(eleIdx)
		if obj.Kind() == reflect.Ptr {
			obj = obj.Elem()
		}
		numCols := obj.NumField()
		objType := obj.Type()
		var row []string = nil

		if colWidth == nil {
			colWidth = make([]int, numCols)
			colNames = make([]string, numCols)

			for idxField := range numCols {
				colNames[idxField] = objType.Field(idxField).Name
				colWidth[idxField] = len(colNames[idxField]) + tabWidth
			}
		}

		assert.True(len(colWidth) == numCols, "all structs must have the same structur")
		for idxField := range numCols {
			value := obj.Field(idxField)
			entry := fmt.Sprintf("%v", value.Interface())
			row = append(row, entry)

			colWidth[idxField] = max(len(entry)+tabWidth, colWidth[idxField])
		}

		rows = append(rows, row)
	}

	o.builder.WriteString(prefix)
	for colIdx := range len(colNames) {
		pad := strings.Repeat(" ", colWidth[colIdx]-len(colNames[colIdx]))
		o.builder.WriteString(colNames[colIdx])
		o.builder.WriteString(pad)
	}
	o.builder.WriteRune('\n')
	o.builder.WriteString(prefix)

	for colIdx := range len(colNames) {
		pad := strings.Repeat(" ", colWidth[colIdx]-len(colNames[colIdx]))
		dotted := strings.Repeat("-", len(colNames[colIdx]))
		o.builder.WriteString(dotted)
		o.builder.WriteString(pad)
	}
	o.builder.WriteRune('\n')
	o.builder.WriteString(prefix)

	for idx, row := range rows {
		for colIdx := range len(colNames) {
			val := row[colIdx]
			pad := strings.Repeat(" ", colWidth[colIdx]-len(val))
			o.builder.WriteString(val)
			o.builder.WriteString(pad)
		}
		if idx < len(rows)-1 {
			o.builder.WriteRune('\n')
			o.builder.WriteString(prefix)
		}
	}
}

type primitiveGroup struct {
	fieldName  []string
	fieldValue []string
}

func newPrimitiveGroup() *primitiveGroup {
	return &primitiveGroup{}
}

func (p *primitiveGroup) add(name string, value string) {
	p.fieldName = append(p.fieldName, name)
	p.fieldValue = append(p.fieldValue, value)
}

func (p *primitiveGroup) write(sb *strings.Builder, intend int) {
	prefix := strings.Repeat(" ", tabWidth*intend)
	assert.True(len(p.fieldName) == len(p.fieldValue), "both need to have the same length")

	width := tabWidth
	for _, name := range p.fieldName {
		width = max(len(name)+tabWidth, width)
	}

	for idx := range len(p.fieldName) {
		pad := strings.Repeat(" ", width-len(p.fieldName[idx]))
		sb.WriteString(prefix)
		sb.WriteString(p.fieldName[idx])
		sb.WriteString(pad)
		sb.WriteString(p.fieldValue[idx])
		sb.WriteRune('\n')
	}
}

type objectGroup struct {
	sturctIterator
	primitive *primitiveGroup
	object    *objectGroup

	groups int // some interface
}

func (o *objectGroup) setPrimary() *primitiveGroup {
	if o.primitive == nil {
		o.primitive = newPrimitiveGroup()
	}

	return o.primitive
}

func Marshall(o any) {
	i := newStructIterator(o)
	/*
		obj := reflect.ValueOf(o)
		assert.True(obj.Kind() == reflect.Struct, "we only do structs")
		typ := obj.Type()
	*/

	prim := &primitiveGroup{}
	for i.hasNext() {
		switch i.currentKind() {
		case reflect.Struct:
			fmt.Println("Struct")

			sb := strings.Builder{}
			prim.write(&sb, 0)
			fmt.Println(sb.String())
			return
		case reflect.Slice:
			fmt.Println("Slice")
			return
		default:
			name, field := i.next()
			prim.add(name, fmt.Sprintf("%v", field.Interface()))
		}

	}

}

type sturctIterator struct {
	val          reflect.Value
	typ          reflect.Type
	currentField int
}

func (s *sturctIterator) currentKind() reflect.Kind {
	return s.typ.Field(s.currentField).Type.Kind()
}

func newStructIterator(o any) *sturctIterator {
	obj := reflect.ValueOf(o)
	typ := obj.Type()

	return &sturctIterator{
		val:          obj,
		typ:          typ,
		currentField: 0,
	}
}

func (s *sturctIterator) hasNext() bool {
	return s.currentField < s.val.NumField()
}

func (s *sturctIterator) next() (fieldName string, fieldValue reflect.Value) {
	field := s.val.Field(s.currentField)
	typField := s.typ.Field(s.currentField)
	name := typField.Name
	s.currentField += 1

	return name, field
}

func main() {
	people := []Person{
		{Name: "Alice", Age: 30, City: "Paris"},
		{Name: "Bob", Age: 25, City: "London"},
		{Name: "Raphael Dichle", Age: 25, City: "London"},
	}

	/*
		o := &output{}
		o.table(people, 1)
		fmt.Println(o.builder.String())
	*/

	Marshall(people[0])

}
