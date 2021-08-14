package rdb

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"libs.altipla.consulting/collections"
)

type Params struct {
	values map[string]interface{}
	next   int
}

func NewParams() *Params {
	return &Params{
		values: make(map[string]interface{}),
	}
}

func (p *Params) Next(value interface{}) string {
	name := fmt.Sprintf("p%d", p.next)
	p.next++

	switch value := value.(type) {
	case time.Time:
		p.values[name] = NewDateTime(value).String()
	default:
		p.values[name] = value
	}

	return "$" + name
}

type QueryFilter interface {
	RQL(params *Params) string
}

type andFilter struct {
	children []QueryFilter
}

func And(children ...QueryFilter) QueryFilter {
	return &andFilter{children}
}

func (f *andFilter) isEmpty() bool {
	return len(f.children) == 0
}

func (f *andFilter) clone() *andFilter {
	return &andFilter{
		children: f.children,
	}
}

func (f *andFilter) RQL(params *Params) string {
	conds := make([]string, len(f.children))
	for i, child := range f.children {
		switch c := child.(type) {
		case *andFilter:
			if len(c.children) > 1 {
				conds[i] = "(" + c.RQL(params) + ")"
			} else {
				conds[i] = c.RQL(params)
			}

		case *orFilter:
			if len(c.children) > 1 {
				conds[i] = "(" + c.RQL(params) + ")"
			} else {
				conds[i] = c.RQL(params)
			}

		default:
			conds[i] = child.RQL(params)
		}
	}
	return strings.Join(conds, " and ")
}

type orFilter struct {
	children []QueryFilter
}

func Or(children ...QueryFilter) QueryFilter {
	return &orFilter{children}
}

func (f *orFilter) RQL(params *Params) string {
	conds := make([]string, len(f.children))
	for i, child := range f.children {
		switch c := child.(type) {
		case *andFilter:
			if len(c.children) > 1 {
				conds[i] = "(" + c.RQL(params) + ")"
			} else {
				conds[i] = c.RQL(params)
			}

		case *orFilter:
			if len(c.children) > 1 {
				conds[i] = "(" + c.RQL(params) + ")"
			} else {
				conds[i] = c.RQL(params)
			}

		default:
			conds[i] = child.RQL(params)
		}
	}
	return strings.Join(conds, " or ")
}

type directFilter struct {
	field string
	op    string
	value interface{}
	exact bool
}

func Filter(field string, value interface{}) QueryFilter {
	f := &directFilter{
		field: field,
		value: value,
		exact: true,
	}

	switch parts := strings.Split(field, " "); len(parts) {
	case 1:
		f.op = "=="

	case 2:
		validOperators := []string{"==", "=", "!=", "<", "<=", ">", ">="}
		if !collections.HasString(validOperators, parts[1]) {
			panic("unknown query operator: " + field + ": " + parts[1])
		}
		f.field = parts[0]
		f.op = parts[1]

	default:
		panic("unknown query expression: " + f.field)
	}

	return f
}

func FilterNotExact(field string, value interface{}) QueryFilter {
	f := Filter(field, value).(*directFilter)
	f.exact = false
	return f
}

func (f *directFilter) RQL(params *Params) string {
	if f.op == "==" {
		f.op = "="
	}
	if f.exact && (f.op == "=" || f.op == "!=") {
		if _, ok := f.value.(string); ok {
			return fmt.Sprintf("exact(%s %s %s)", f.field, f.op, params.Next(f.value))
		}
	}
	return fmt.Sprintf("%s %s %s", f.field, f.op, params.Next(f.value))
}

type inFilter struct {
	field   string
	against []interface{}
}

func FilterIn(field string, values ...interface{}) QueryFilter {
	return &inFilter{field, parseArrayFilter(values...)}
}

func (f *inFilter) RQL(params *Params) string {
	return fmt.Sprintf("%s in (%s)", f.field, params.Next(f.against))
}

type containsAllFilter struct {
	field   string
	against []interface{}
}

func FilterContainsAll(field string, values ...interface{}) QueryFilter {
	return &containsAllFilter{field, parseArrayFilter(values...)}
}

func (f *containsAllFilter) RQL(params *Params) string {
	return fmt.Sprintf("%s all in (%s)", f.field, params.Next(f.against))
}

func parseArrayFilter(values ...interface{}) []interface{} {
	var converted []interface{}

	for _, value := range values {
		if rv := reflect.ValueOf(value); rv.Kind() == reflect.Slice {
			for i := 0; i < rv.Len(); i++ {
				converted = append(converted, rv.Index(i).Interface())
			}
		} else {
			converted = append(converted, value)
		}
	}

	return converted
}

type startsWithFilter struct {
	field  string
	prefix string
}

func FilterStartsWith(field, prefix string) QueryFilter {
	return &startsWithFilter{field, prefix}
}

func (f *startsWithFilter) RQL(params *Params) string {
	return fmt.Sprintf("startsWith(%s, %s)", f.field, params.Next(f.prefix))
}

type endsWithFilter struct {
	field  string
	suffix string
}

func FilterEndsWith(field, suffix string) QueryFilter {
	return &endsWithFilter{field, suffix}
}

func (f *endsWithFilter) RQL(params *Params) string {
	return fmt.Sprintf("endsWith(%s, %s)", f.field, params.Next(f.suffix))
}

type searchFilter struct {
	field  string
	search string
	opt    SearchOption
}

func FilterSearch(field, search string, opts ...SearchOption) QueryFilter {
	if len(opts) > 1 {
		panic("cannot use more than one option")
	}

	return &searchFilter{field, search, opts[0]}
}

func (f *searchFilter) RQL(params *Params) string {
	if f.opt == "" {
		return fmt.Sprintf("search(%s, %s)", f.field, params.Next(f.search))
	}

	return fmt.Sprintf("search(%s, %s, %s)", f.field, params.Next(f.search), f.opt)
}

type hasFieldFilter struct {
	field string
}

func FilterHasField(field string) QueryFilter {
	return &hasFieldFilter{field}
}

func (f *hasFieldFilter) RQL(params *Params) string {
	return fmt.Sprintf("exists(%s)", f.field)
}

type betweenFilter struct {
	field      string
	start, end interface{}
}

func FilterBetween(field string, start, end interface{}) QueryFilter {
	return &betweenFilter{field, start, end}
}

func (f *betweenFilter) RQL(params *Params) string {
	return fmt.Sprintf("%s between %s and %s", f.field, params.Next(f.start), params.Next(f.end))
}
