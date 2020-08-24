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
func TestScheduleDynamoDA_Insert(t *testing.T) {
	arr := make([]*entity.Schedule, 0)

	for i := 0; i < 10; i++ {
		id := utils.NewID()
		start := time.Now().AddDate(0, 0, rand.Intn(10))
		s := &entity.Schedule{
			ID:           id,
			Title:        fmt.Sprintf("%s_%s", id[0:4], "title"),
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
			AttachmentID: fmt.Sprintf("%d", rand.Intn(100)),
			Version:      0,
			Repeat:       entity.RepeatOptions{},
			CreatedID:    "1",
			UpdatedID:    "",

			CreatedAt: time.Now().Unix(),
			UpdatedAt: 0,
		}
		arr = append(arr, s)
	}
	err := GetScheduleDA().BatchInsert(context.Background(), arr)
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
	data, err := GetScheduleDA().Query(context.Background(), condition)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, item := range data {
		fmt.Println(*item)
	}
}
