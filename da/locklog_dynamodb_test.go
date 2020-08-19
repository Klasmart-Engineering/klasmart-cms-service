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
		{
			name:    "t1",
			args:    args{ctx: context.Background(), id: "1"},
			want:    &entity.LockLog{ID: "1", RecordID: "1", OperatorID: "1", CreatedAt: 0, DeletedAt: 0},
			wantErr: false,
		},
		{
			name:    "t2",
			args:    args{ctx: context.Background(), id: "2"},
			want:    &entity.LockLog{ID: "2", RecordID: "2", OperatorID: "2", CreatedAt: 0, DeletedAt: 0},
			wantErr: false,
		},
		{
			name:    "t3",
			args:    args{ctx: context.Background(), id: "3"},
			want:    &entity.LockLog{ID: "3", RecordID: "3", OperatorID: "3", CreatedAt: 0, DeletedAt: 0},
			wantErr: false,
		},
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
		{
			name:    "t1",
			args:    args{ctx: context.Background(), recordID: "1"},
			want:    &entity.LockLog{ID: "1", RecordID: "1", OperatorID: "1", CreatedAt: 0, DeletedAt: 0},
			wantErr: false,
		},
		{
			name:    "t2",
			args:    args{ctx: context.Background(), recordID: "2"},
			want:    &entity.LockLog{ID: "2", RecordID: "2", OperatorID: "2", CreatedAt: 0, DeletedAt: 0},
			wantErr: false,
		},
		{
			name:    "t3",
			args:    args{ctx: context.Background(), recordID: "3"},
			want:    &entity.LockLog{ID: "3", RecordID: "3", OperatorID: "3", CreatedAt: 0, DeletedAt: 0},
			wantErr: false,
		},
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
	var tests = []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "t1",
			args:    args{ctx: nil, l: &entity.LockLog{ID: "1", RecordID: "1", OperatorID: "1"}},
			wantErr: false,
		},
		{
			name:    "t2",
			args:    args{ctx: nil, l: &entity.LockLog{ID: "2", RecordID: "2", OperatorID: "2"}},
			wantErr: false,
		},
		{
			name:    "t3",
			args:    args{ctx: nil, l: &entity.LockLog{ID: "3", RecordID: "3", OperatorID: "3"}},
			wantErr: false,
		},
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
		{
			name:    "t1",
			args:    args{ctx: context.Background(), id: "1"},
			wantErr: false,
		},
		{
			name:    "t2",
			args:    args{ctx: context.Background(), id: "2"},
			wantErr: false,
		},
		{
			name:    "t3",
			args:    args{ctx: context.Background(), id: "3"},
			wantErr: false,
		},
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
		{
			name:    "t1",
			args:    args{ctx: context.Background(), recordID: "1"},
			wantErr: false,
		},
		{
			name:    "t2",
			args:    args{ctx: context.Background(), recordID: "2"},
			wantErr: false,
		},
		{
			name:    "t3",
			args:    args{ctx: context.Background(), recordID: "3"},
			wantErr: false,
		},
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
