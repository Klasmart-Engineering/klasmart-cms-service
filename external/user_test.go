package external

import (
	"context"
	"reflect"
	"testing"
)

func TestAmsUserService_FilterByPermission(t *testing.T) {
	ids := []string{"f2626a21-3e98-517d-ac4a-ed6f33231869", "0a6091d7-1014-595d-abbf-dad456692d15"}
	want := []string{"f2626a21-3e98-517d-ac4a-ed6f33231869"}
	filtered, err := GetUserServiceProvider().FilterByPermission(context.TODO(), testOperator, ids, CreateContentPage201)
	if err != nil {
		t.Errorf("GetUserServiceProvider().FilterByPermission() error = %v", err)
		return
	}

	if !reflect.DeepEqual(filtered, want) {
		t.Errorf("GetUserServiceProvider().FilterByPermission() want %+v results got %+v", want, filtered)
		return
	}
}
