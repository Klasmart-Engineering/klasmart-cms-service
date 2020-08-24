package utils

import (
	"testing"
)

func TestSQLBuilder(t *testing.T) {
	type user struct {
		id   string
		name string
	}
	data := []user{
		{"1", "Medivh"},
		{"2", "Kwork"},
		{"3", "Unknown"},
	}
	b := NewSQLBuilder().Append("insert user(id, name) values")
	values := NewSQLBuilder()
	for _, item := range data {
		values.Append("(?,?)", item.id, item.name)
	}
	b.AppendTemplate(values.Join(","))
	t.Logf("%+v", b.Join())
}
