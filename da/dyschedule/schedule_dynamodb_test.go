package dyschedule

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da/daschedule"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/dynamodbhelper"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}
func TestScheduleDynamoDA_Insert(t *testing.T) {
	arr := make([]*entity.Schedule, 0)

	for i := 0; i < 50; i++ {
		id := strconv.Itoa(i)
		start := time.Now().AddDate(0, 0, rand.Intn(10))
		s := &entity.Schedule{
			ID:           id,
			RepeatID:     utils.NewID(),
			Title:        fmt.Sprintf("%s_%s", id, "title"),
			ClassID:      fmt.Sprintf("%d", rand.Intn(10)),
			LessonPlanID: fmt.Sprintf("%d", rand.Intn(10)),
			TeacherIDs:   []string{"1", "2", "3"},
			OrgID:        "1",
			StartAt:      start.Unix(),
			EndAt:        start.Add(20 * time.Minute).Unix(),
			ModeType:     entity.ModeTypeAllDay,
			SubjectID:    "1",
			ProgramID:    "2",
			ClassType:    "ClassType",
			DueAt:        0,
			Description:  "深刻理解的连接方式烦死了烦死了副书记",
			Attachment:   fmt.Sprintf("%d", rand.Intn(100)),
			Version:      0,
			Repeat:       entity.RepeatOptions{},
			CreatedID:    "1",
			UpdatedID:    "",

			CreatedAt: time.Now().Unix(),
			UpdatedAt: 0,
		}
		arr = append(arr, s)
	}
	err := daschedule.GetScheduleDA().BatchInsert(context.Background(), arr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("成功")
}

func TestScheduleCondition_GetCondition(t *testing.T) {

}

func TestScheduleDynamoDA_Query(t *testing.T) {
	condition := &ScheduleCondition{
		OrgID:   "1",
		StartAt: time.Now().Unix(),
	}
	condition.Init(constant.GSI_Schedule_OrgIDAndStartAt, dynamodbhelper.SortKeyGreaterThanEqual)
	data, err := daschedule.GetScheduleDA().Query(context.Background(), condition)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, item := range data {
		fmt.Println(*item)
	}
}

func Test_scheduleDynamoDA_Delete(t *testing.T) {
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
			name:    "delete",
			args:    args{ctx: context.Background(), id: "0"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &scheduleDynamoDA{}
			if err := s.Delete(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_scheduleDynamoDA_BatchDelete(t *testing.T) {
	rangeInts := func(start, end int) []string {
		var result []string
		for num := start; num < end; num++ {
			result = append(result, strconv.Itoa(num))
		}
		return result
	}
	type args struct {
		ctx context.Context
		ids []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "batch delete",
			args: args{
				ctx: context.Background(),
				ids: rangeInts(0, 50),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &scheduleDynamoDA{}
			if err := s.BatchDelete(tt.args.ctx, tt.args.ids); (err != nil) != tt.wantErr {
				t.Errorf("BatchDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
