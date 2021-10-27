package da

import (
	"context"
	"fmt"
	"testing"
)

func TestSqlBuilder(t *testing.T) {
	ctx := context.Background()
	sb := NewSqlBuilder(ctx, `
	select * from xxx {{.condition}} 
`)
	sb.Replace(ctx, "condition", NewSqlBuilder(ctx, `
where name = ?`, "xxdkajdwa"))
	sql, args, err := sb.Build(ctx)
	fmt.Println(sql, args, err)
}
