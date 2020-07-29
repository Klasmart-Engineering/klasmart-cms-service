package model

import (
	"calmisland/kidsloop2/entity"
	"context"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"reflect"
	"testing"
)

func TestCategoryModel_CreateCategory(t *testing.T) {
	type args struct {
		ctx  context.Context
		data entity.CategoryObject
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "",
			args: args{
				ctx: context.Background(),
				data: entity.CategoryObject{

				},
			},
			want: "ok",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &CategoryModel{}
			got, err := cm.CreateCategory(tt.args.ctx, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateCategory() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCategoryModel_DeleteCategory(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &CategoryModel{}
			if err := cm.DeleteCategory(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteCategory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCategoryModel_GetCategoryById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		want    *entity.CategoryObject
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &CategoryModel{}
			got, err := cm.GetCategoryById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCategoryById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCategoryById() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCategoryModel_SearchCategories(t *testing.T) {
	type args struct {
		ctx       context.Context
		condition *SearchCategoryCondition
	}
	tests := []struct {
		name    string
		args    args
		want    []*entity.CategoryObject
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &CategoryModel{}
			got, err := cm.SearchCategories(tt.args.ctx, tt.args.condition)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchCategories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SearchCategories() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCategoryModel_UpdateCategory(t *testing.T) {
	type args struct {
		ctx  context.Context
		data entity.CategoryObject
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &CategoryModel{}
			if err := cm.UpdateCategory(tt.args.ctx, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UpdateCategory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetCategoryModel(t *testing.T) {
	tests := []struct {
		name string
		want ICategoryModel
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCategoryModel(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCategoryModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchCategoryCondition_getConditions(t *testing.T) {
	type fields struct {
		IDs      []string
		Names    []string
		PageSize int64
		Page     int64
		OrderBy  string
	}
	tests := []struct {
		name   string
		fields fields
		want   []expression.ConditionBuilder
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SearchCategoryCondition{
				IDs:      tt.fields.IDs,
				Names:    tt.fields.Names,
				PageSize: tt.fields.PageSize,
				Page:     tt.fields.Page,
				OrderBy:  tt.fields.OrderBy,
			}
			if got := s.getConditions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}