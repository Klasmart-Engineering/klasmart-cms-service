package da

import (
	"testing"
	"time"
)

func TestSQLTemplate(t *testing.T) {
	type user struct {
		id   string
		name string
	}
	data := []user{
		{"1", "Medivh"},
		{"2", "Kwork"},
		{"3", "Unknown"},
	}
	b := NewSQLTemplate("insert into user(id, name) values")
	t2 := NewSQLTemplate("")
	for _, item := range data {
		t2.Appendf("(?,?)", item.id, item.name)
	}
	b.AppendResult(t2.Join(", ", "", ""))
	format, values := b.Concat()
	t.Logf("format: %+v\n", format)
	t.Logf("values: %+v\n", values)
}

func TestSQLBatchInsert(t *testing.T) {
	data := [][]interface{}{
		{1, "Medivh", time.Now().Unix()},
		{2, "Kwork", time.Now().Unix()},
		{3, "Unknown", time.Now().Unix()},
	}
	format, values := SQLBatchInsert("user", []string{"name", "age", "created"}, data)
	t.Logf("format: %+v\n", format)
	t.Logf("values: %+v\n", values)
}
