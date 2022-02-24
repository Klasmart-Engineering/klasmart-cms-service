package external

import (
	"context"
	"fmt"
	"testing"
)

func TestAmsTeacherLoadService_BatchGetClassWithStudent(t *testing.T) {
	ctx := context.Background()
	result, err := GetTeacherLoadServiceProvider().BatchGetActiveClassWithStudent(ctx, testOperator, []string{"8cd9f417-0812-44fb-9c50-1b78217ee76f"})
	if err != nil {
		t.Errorf("GetTeacherLoadServiceProvider().BatchGetActiveClassWithStudent() error = %v", err)
		return
	}
	fmt.Printf("%#v", result)
}
