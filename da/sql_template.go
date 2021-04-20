package da

import (
	"fmt"
	"strings"
)

type SQLTemplate struct {
	Formats []string
	Values  []interface{}
}

func NewSQLTemplate(format string, values ...interface{}) *SQLTemplate {
	return (&SQLTemplate{}).Appendf(format, values...)
}

func FalseSQLTemplate() *SQLTemplate {
	return NewSQLTemplate("1 = 0")
}

func TrueSQLTemplate() *SQLTemplate {
	return NewSQLTemplate("1 = 1")
}

func (b *SQLTemplate) IsEmpty() bool {
	return len(b.Formats) == 0
}

func (b *SQLTemplate) Appendf(format string, values ...interface{}) *SQLTemplate {
	if format == "" {
		return b
	}
	b.Formats = append(b.Formats, format)
	b.Values = append(b.Values, values...)
	return b
}

func (b *SQLTemplate) AppendResult(format string, values []interface{}) *SQLTemplate {
	return b.Appendf(format, values...)
}

func (b *SQLTemplate) AppendTemplates(others ...*SQLTemplate) *SQLTemplate {
	for _, t := range others {
		if t == nil || t.IsEmpty() {
			continue
		}
		b.Formats = append(b.Formats, t.Formats...)
		b.Values = append(b.Values, t.Values...)
	}
	return b
}

func (b *SQLTemplate) Join(sep string, left, right string) (string, []interface{}) {
	return left + strings.Join(b.Formats, sep) + right, b.Values
}

func (b *SQLTemplate) Concat() (string, []interface{}) {
	return b.Join("", "", "")
}

func (b *SQLTemplate) And() (string, []interface{}) {
	return b.Join(" and ", "(", ")")
}

func (b *SQLTemplate) Or() (string, []interface{}) {
	return b.Join(" or ", "(", ")")
}

func (b *SQLTemplate) DBOConditions() ([]string, []interface{}) {
	return b.Formats, b.Values
}

func SQLBatchInsert(table string, columns []string, matrix [][]interface{}) (string, []interface{}) {
	t := NewSQLTemplate(fmt.Sprintf("insert into %s(%s) values", table, strings.Join(columns, ",")))
	var placeholders []string
	for i := 0; i < len(columns); i++ {
		placeholders = append(placeholders, "?")
	}
	placeholdersFormat := fmt.Sprintf("(%s)", strings.Join(placeholders, ","))
	t2 := NewSQLTemplate("")
	for _, values := range matrix {
		t2.Appendf(placeholdersFormat, values...)
	}
	return t.AppendResult(t2.Join(", ", "", "")).Concat()
}

func RefInt(i int) *int {
	return &i
}

func RefString(s string) *string {
	return &s
}

func RefBool(b bool) *bool {
	return &b
}
