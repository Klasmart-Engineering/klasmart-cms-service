package utils

import (
	"strings"
)

type SQLTemplate struct {
	Format string
	Values []interface{}
}

type SQLBuilder struct {
	items []SQLTemplate
}

func NewSQLBuilder() *SQLBuilder {
	return &SQLBuilder{}
}

func (b *SQLBuilder) Append(format string, values ...interface{}) *SQLBuilder {
	b.items = append(b.items, SQLTemplate{Format: format, Values: values})
	return b
}

func (b *SQLBuilder) AppendTemplate(template SQLTemplate) *SQLBuilder {
	b.items = append(b.items, template)
	return b
}

func (b *SQLBuilder) Join(sep ...string) SQLTemplate {
	finalSep := " "
	if len(sep) > 0 {
		finalSep = sep[0]
	}
	buffer := strings.Builder{}
	var values []interface{}
	length := len(b.items)
	for i, item := range b.items {
		buffer.WriteString(item.Format)
		if i < length-1 && finalSep != "" {
			buffer.WriteString(finalSep)
		}
		values = append(values, item.Values...)
	}
	return SQLTemplate{Format: buffer.String(), Values: values}
}
