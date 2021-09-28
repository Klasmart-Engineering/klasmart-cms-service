package model

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

func TestLessoons(t *testing.T) {
	ctx := context.Background()
	request := entity.TeacherLoadMissedLessonsRequest{TeacherId: "0e6b5f9d-0383-5ac6-b13d-8af02697fa8b",
		ClassIDs: []string{"0b3f5f4d-3540-42ab-9fda-474fbbe8b51c"}, Page: 8, PageSize: -1,
		Duration: "1608600600-1630419300"}
	response, err := GetReportModel().MissedLessonsList(ctx, &request)
	fmt.Println(err)
	fmt.Println(response)

}
