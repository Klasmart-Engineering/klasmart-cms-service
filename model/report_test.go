package model

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

func TestLessoons(t *testing.T) {
	ctx := context.Background()
	request := entity.TeacherLoadMissedLessonsRequest{TeacherId: "be381662-3f63-5bdd-8dc7-9ee0fb735d5f",
		ClassIDs: []string{"5751555a-cc18-4662-9ae5-a5ad90569f79"}, PageNumber: 2, PageSize: 2,
		Duration: "1605110400-1605715200"}
	response, err := GetReportModel().MissedLessonsList(ctx, &request)
	fmt.Println(err)
	fmt.Println(response)

}
