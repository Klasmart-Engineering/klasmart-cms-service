package da

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/dynamodbhelper"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestRand(t *testing.T) {
	for i := 0; i < 5; i++ {
		r := rand.Intn(20)
		fmt.Println(r)
	}
}

func TestTeacherScheduleDA_BatchAdd(t *testing.T) {
	start := time.Now()
	testData := make([]*entity.TeacherSchedule, 0)

	for i := 0; i < 3; i++ {
		for j := 0; j < 8; j++ {
			r := rand.Intn(20)
			ts := &entity.TeacherSchedule{
				TeacherID:  fmt.Sprintf("t_%d", i),
				ScheduleID: fmt.Sprintf("s_%d_%s", j, utils.NewID()),
				StartAt:    start.AddDate(0, 0, r).Unix(),
			}
			testData = append(testData, ts)
		}
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 8; j++ {
			r := rand.Intn(20)
			ts := &entity.TeacherSchedule{
				TeacherID:  fmt.Sprintf("t_%d", i),
				ScheduleID: fmt.Sprintf("s_%d_%s", j, utils.NewID()),
				StartAt:    start.AddDate(0, 0, 0-r).Unix(),
			}
			testData = append(testData, ts)
		}
	}
	fmt.Println("item len is ", len(testData))
	err := GetTeacherScheduleDA().BatchAdd(context.Background(), testData)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("no error")
}

func TestTeacherScheduleDA_Page(t *testing.T) {
	lastkey, result, err := GetTeacherScheduleDA().Page(context.Background(), TeacherScheduleCondition{
		Condition: dynamodbhelper.Condition{
			Pager: dynamodbhelper.Pager{
				PageSize: 5,
				LastKey:  "t_1,s_2,t_1,1599723286",
			},
			PrimaryKey: dynamodbhelper.KeyValue{
				Key:   "teacher_id",
				Value: "t_1",
			},
			SortKey: dynamodbhelper.KeyValue{
				Key:   "start_at",
				Value: time.Now().AddDate(-1, 0, 0).Unix(),
			},
			CompareType: dynamodbhelper.SortKeyGreaterThanEqual,
			IndexName:   "teacher_id_and_start_at",
		},
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(lastkey)
	for _, item := range result {
		fmt.Println(item)
	}
}

func TestTeacherScheduleCondition(t *testing.T) {
	condition := TeacherScheduleCondition{
		TeacherID: "111",
		StartAt:   time.Now().Unix(),
	}
	condition.Init(constant.GSI_TeacherSchedule_TeacherAndStartAt, dynamodbhelper.SortKeyGreaterThanEqual)
	fmt.Println(condition)
}
