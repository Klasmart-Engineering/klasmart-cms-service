package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"reflect"
	"testing"
)

func Test_lockLogDA_GetByID(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		want    *entity.LockLog
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			da := &lockLogDA{}
			got, err := da.GetByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetByID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_lockLogDA_GetByRecordID(t *testing.T) {
	type args struct {
		ctx      context.Context
		recordID string
	}
	tests := []struct {
		name    string
		args    args
		want    *entity.LockLog
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			da := &lockLogDA{}
			got, err := da.GetByRecordID(tt.args.ctx, tt.args.recordID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByRecordID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetByRecordID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_lockLogDA_Insert(t *testing.T) {
	type args struct {
		ctx context.Context
		l   *entity.LockLog
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
			da := &lockLogDA{}
			if err := da.Insert(tt.args.ctx, tt.args.l); (err != nil) != tt.wantErr {
				t.Errorf("Insert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_lockLogDA_SoftDeleteByID(t *testing.T) {
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
			da := &lockLogDA{}
			if err := da.SoftDeleteByID(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("SoftDeleteByID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_lockLogDA_SoftDeleteByRecordID(t *testing.T) {
	type args struct {
		ctx      context.Context
		recordID string
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
			da := &lockLogDA{}
			if err := da.SoftDeleteByRecordID(tt.args.ctx, tt.args.recordID); (err != nil) != tt.wantErr {
				t.Errorf("SoftDeleteByRecordID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
