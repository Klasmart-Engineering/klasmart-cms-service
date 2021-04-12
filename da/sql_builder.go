package da

import (
	"fmt"
	"strings"
)

func RefInt(i int) *int {
	return &i
}

func RefString(s string) *string {
	return &s
}

func RefBool(b bool) *bool {
	return &b
}

type SQLTemplate struct {
	Format string
	Values []interface{}
}

func NewSQLTemplate(format string, values ...interface{}) *SQLTemplate {
	return &SQLTemplate{
		Format: format,
		Values: values,
	}
}

func FalseSQLTemplate() *SQLTemplate {
	return NewSQLTemplate("1 = 0")
}

func TrueSQLTemplate() *SQLTemplate {
	return NewSQLTemplate("1 = 1")
}

func (t *SQLTemplate) Append(format string, values ...interface{}) *SQLTemplate {
	t.Format += format
	t.Values = append(t.Values, values...)
	return t
}

func (t *SQLTemplate) Join(sep string, format string, values ...interface{}) *SQLTemplate {
	if !t.IsEmpty() {
		t.Append(sep)
	}
	return t.Append(format, values...)
}

func (t *SQLTemplate) JoinTemplate(sep string, other SQLTemplate) *SQLTemplate {
	return t.Join(sep, other.Format, other.Values)
}

func (t *SQLTemplate) And(format string, values ...interface{}) *SQLTemplate {
	return t.Join(" and ", format, values...)
}

func (t *SQLTemplate) AndTemplate(other SQLTemplate) *SQLTemplate {
	return t.JoinTemplate(" and ", other)
}

func (t *SQLTemplate) Or(format string, values ...interface{}) *SQLTemplate {
	return t.Join(" or ", format, values...)
}

func (t *SQLTemplate) OrTemplate(other SQLTemplate) *SQLTemplate {
	return t.JoinTemplate(" or ", other)
}

func (t *SQLTemplate) Wrap(left, right string) *SQLTemplate {
	t.Format = left + t.Format + right
	return t
}

func (t *SQLTemplate) WrapBracket() *SQLTemplate {
	return t.Wrap("(", ")")
}

func (t *SQLTemplate) IsEmpty() bool {
	return strings.TrimSpace(t.Format) == ""
}

func (t *SQLTemplate) DBOConditions() ([]string, []interface{}) {
	return []string{t.Format}, t.Values
}

type SQLBuilder struct {
	Templates []SQLTemplate
}

func NewSQLBuilder(templates ...SQLTemplate) *SQLBuilder {
	return &SQLBuilder{Templates: templates}
}

func (b *SQLBuilder) Append(format string, values ...interface{}) *SQLBuilder {
	b.Templates = append(b.Templates, SQLTemplate{Format: format, Values: values})
	return b
}

func (b *SQLBuilder) AppendTemplate(t *SQLTemplate) *SQLBuilder {
	if t == nil {
		return b
	}
	b.Templates = append(b.Templates, *t)
	return b
}

func (b *SQLBuilder) Merge(sep string) *SQLTemplate {
	var (
		formats []string
		values  []interface{}
	)
	for _, t := range b.Templates {
		formats = append(formats, t.Format)
		values = append(values, t.Values...)
	}
	return &SQLTemplate{Format: strings.Join(formats, sep), Values: values}
}

func (b *SQLBuilder) MergePlain() *SQLTemplate {
	return b.Merge("")
}

func (b *SQLBuilder) MergeWithSpace() *SQLTemplate {
	return b.Merge(" ")
}

func (b *SQLBuilder) MergeWithAnd() *SQLTemplate {
	return b.Merge(" and ")
}

func (b *SQLBuilder) MergeWithOr() *SQLTemplate {
	return b.Merge(" or ")
}

func (b *SQLBuilder) IsEmpty() bool {
	return len(b.Templates) == 0
}

func SQLBatchInsert(table string, columns []string, matrix [][]interface{}) *SQLTemplate {
	b := NewSQLBuilder().Append(fmt.Sprintf("insert into %s(%s) values", table, strings.Join(columns, ",")))
	var placeholders []string
	for i := 0; i < len(columns); i++ {
		placeholders = append(placeholders, "?")
	}
	placeholdersFormat := fmt.Sprintf("(%s)", strings.Join(placeholders, ","))
	t := NewSQLTemplate("")
	for _, values := range matrix {
		t.Join(", ", placeholdersFormat, values...)
	}
	return b.AppendTemplate(t).MergePlain()
}
