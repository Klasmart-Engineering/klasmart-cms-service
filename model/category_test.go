package model

import (
	"context"
	"fmt"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func TestCategoryModel_CreateCategory(t *testing.T) {
	type args struct {
		ctx  context.Context
		op   *entity.Operator
		data entity.CategoryObject
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx: context.Background(),
				op:  &entity.Operator{UserID: "No.1", Role: "admin"},
				data: entity.CategoryObject{
					Name: "name2",
				},
			},
			want:    "ok",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &CategoryModel{}
			got, err := cm.CreateCategory(tt.args.ctx, tt.args.op, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(got)
		})
	}
}

func TestCategoryModel_DeleteCategory(t *testing.T) {
	type args struct {
		ctx context.Context
		op  *entity.Operator
		id  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "delete category",
			args:    args{context.Background(), &entity.Operator{UserID: "No.1", Role: "admin"}, "3bdec1625fd64878"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &CategoryModel{}
			if err := cm.DeleteCategory(tt.args.ctx, tt.args.op, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteCategory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCategoryModel_GetCategoryById(t *testing.T) {
	type args struct {
		ctx context.Context
		op  *entity.Operator
		id  string
	}
	tests := []struct {
		name    string
		args    args
		want    *entity.CategoryObject
		wantErr bool
	}{
		{
			name:    "getById",
			args:    args{ctx: context.Background(), op: &entity.Operator{UserID: "No.1", Role: "admin"}, id: "id_test1"},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &CategoryModel{}
			got, err := cm.GetCategoryByID(tt.args.ctx, tt.args.op, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCategoryByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GetCategoryByID() got = %v, want %v", got, tt.want)
			//}
			fmt.Println(got)
		})
	}
}
