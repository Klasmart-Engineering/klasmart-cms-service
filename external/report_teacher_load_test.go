package external

import (
	"context"
	"fmt"
	"testing"
)

func TestAmsTeacherLoadService_BatchGetClassWithStudent(t *testing.T) {
	ctx := context.Background()
	result, err := GetTeacherLoadServiceProvider().BatchGetActiveClassWithStudent(ctx, testOperator, []string{"fd589031-6b98-4336-8da6-a35dc52a15b3", "0d3686a6-bf6a-4777-a716-31ce4aa0f516"})
	if err != nil {
		t.Errorf("GetTeacherLoadServiceProvider().BatchGetActiveClassWithStudent() error = %v", err)
		return
	}
	fmt.Printf("%#v", result)
}
