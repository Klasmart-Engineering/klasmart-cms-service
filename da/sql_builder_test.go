package da

import (
	"testing"
	"time"
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
	b := NewSQLBuilder().Append("insert into user(id, name) values")
	tpl := NewSQLTemplate("")
	for _, item := range data {
		tpl.Join(", ", "(?,?)", item.id, item.name)
	}
	b.AppendTemplate(tpl)
	t.Logf("%+v", b.MergePlain())
}

func TestSQLBatchInsert(t *testing.T) {
	data := [][]interface{}{
		{1, "Medivh", time.Now().Unix()},
		{2, "Kwork", time.Now().Unix()},
		{3, "Unknown", time.Now().Unix()},
	}
	template := SQLBatchInsert("user", []string{"name", "age", "created"}, data)
	t.Logf("%+v", template)
}
