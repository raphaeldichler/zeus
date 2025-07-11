// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package formatter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/raphaeldichler/zeus/internal/util/assert"
)

const tabWidth = 4

type Pretty struct{}

func NewPrettyFormatter() *Pretty {
	return &Pretty{}
}

func (p *Pretty) Marshal(obj any) string {
	str := strings.Builder{}
	object := newObjectGroup(0, obj, "")
	object.Marshall()
	object.write(&str)
	return str.String()
}

type group interface {
	write(out *strings.Builder)
}

type tableGroup struct {
	name   string
	tab    reflect.Value
	intend int
}

func newTableGroup(intend int, tab reflect.Value, name string) *tableGroup {
	return &tableGroup{
		name:   name,
		tab:    tab,
		intend: intend,
	}
}

func (t *tableGroup) write(out *strings.Builder) {
	assert.True(t.intend > 0, "table must be at least 2 levels deep")

	prefix := strings.Repeat(" ", tabWidth*t.intend)
	slice := t.tab
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

	if len(rows) == 0 {
		return
	}

	if t.name != "" {
		prefix := strings.Repeat(" ", (t.intend-1)*tabWidth)
		out.WriteString(prefix)
		out.WriteString(t.name)
		out.WriteRune('\n')
	}

	out.WriteString(prefix)
	for colIdx := range len(colNames) {
		pad := strings.Repeat(" ", colWidth[colIdx]-len(colNames[colIdx]))
		out.WriteString(colNames[colIdx])
		out.WriteString(pad)
	}
	out.WriteRune('\n')
	out.WriteString(prefix)

	for colIdx := range len(colNames) {
		pad := strings.Repeat(" ", colWidth[colIdx]-len(colNames[colIdx]))
		dotted := strings.Repeat("-", len(colNames[colIdx]))
		out.WriteString(dotted)
		out.WriteString(pad)
	}
	out.WriteRune('\n')
	out.WriteString(prefix)

	for idx, row := range rows {
		for colIdx := range len(colNames) {
			val := row[colIdx]
			pad := strings.Repeat(" ", colWidth[colIdx]-len(val))
			out.WriteString(val)
			out.WriteString(pad)
		}
		if idx < len(rows)-1 {
			out.WriteRune('\n')
			out.WriteString(prefix)
		}
	}
}

type primitiveGroup struct {
	intend     int
	fieldName  []string
	fieldValue []string
}

func newPrimitiveGroup(intend int) *primitiveGroup {
	return &primitiveGroup{
		intend: intend,
	}
}

func (p *primitiveGroup) add(name string, value string) {
	p.fieldName = append(p.fieldName, name)
	p.fieldValue = append(p.fieldValue, value)
}

func (p *primitiveGroup) write(sb *strings.Builder) {
	prefix := strings.Repeat(" ", tabWidth*p.intend)
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
	*sturctIterator
	name          string
	primitive     *primitiveGroup
	currentIndent int
	groups        []group
}

func newObjectGroup(intend int, obj any, name string) *objectGroup {
	return &objectGroup{
		name:           name,
		sturctIterator: newStructIterator(obj),
		currentIndent:  intend,
	}
}

func (o *objectGroup) write(out *strings.Builder) {
	if o.name != "" {
		assert.True(o.currentIndent > 0, "object must be at least 2 levels deep")
		prefix := strings.Repeat(" ", (o.currentIndent-1)*tabWidth)
		out.WriteString(prefix)
		out.WriteString(o.name)
		out.WriteRune('\n')
	}

	for _, group := range o.groups {
		group.write(out)
		out.WriteRune('\n')
	}
}

func (o *objectGroup) primary() {
	if o.primitive == nil {
		o.primitive = newPrimitiveGroup(o.currentIndent)
		o.groups = append(o.groups, o.primitive)
	}

	name, field := o.next()
	o.primitive.add(name, fmt.Sprintf("%v", field.Interface()))
}

func (o *objectGroup) table() {
	name, tab := o.next()
	table := newTableGroup(o.currentIndent+1, tab, name)
	o.groups = append(o.groups, table)
	o.primitive = nil
}

func (o *objectGroup) object() {
	name, obj := o.next()
	object := newObjectGroup(o.currentIndent+1, obj.Interface(), name)
	object.Marshall()
	o.groups = append(o.groups, object)
	o.primitive = nil
}

func (o *objectGroup) Marshall() {
	for o.hasNext() {
		switch o.currentKind() {
		case reflect.Struct:
			o.object()

		case reflect.Slice:
			o.table()

		default:
			o.primary()
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

	if obj.Kind() == reflect.Ptr {
		obj = obj.Elem()
	}

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
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}

	typField := s.typ.Field(s.currentField)
	name := typField.Name
	s.currentField += 1

	return name, field
}
