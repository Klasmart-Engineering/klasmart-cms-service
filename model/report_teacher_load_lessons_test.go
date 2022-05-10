package model

import (
	"context"
	"fmt"
	"testing"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

var mockOperator = entity.Operator{
	OrgID:  "f27efd10-000e-4542-bef2-0ccda39b93d3",
	UserID: "6c0a0d79-86ba-461e-838e-916129db6169",
	Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjZjMGEwZDc5LTg2YmEtNDYxZS04MzhlLTkxNjEyOWRiNjE2OSIsImVtYWlsIjoidGVhY2hlcjAyXzAwMUB5b3BtYWlsLmNvbSIsImV4cCI6MTYzMjgwOTk0NCwiaXNzIjoia2lkc2xvb3AifQ.nstrx40u9TqTi18SBiJPN3zdK4YbBMAkNtUZl1Wj9vvEMyfzrDXJJhMS-9Th6q-fVcPau-HjaH3Pd_-nNUBGKJ4HRPLZVYFwftHqz7zZkYJ6eMfm0zJXlz3uhZYWZ_JdpUhCmTtHCnTCRRPSIZEOqwIyBU6vPss9eYhrd4tgCFWmUdcMZULrfJaQ8Zq1Yc8vLXmm6DVP-bvxBgyx--zKqzRMvLDKdBJujy0k1wvwh-F7GKTtFvQyqqaYekv9VQ9xv7sNMREWRYdt19KNFDj7LLaSLIIbqPSf2wNoePXllBCF7Ix7K1gZ-eZy0lYQnVgi6L0Ly88gekG62hpQzy8yTg",
}

func TestReportModel_ListTeacherLoadLessons(t *testing.T) {
	ctx := context.Background()
	args := &entity.TeacherLoadLessonArgs{
		//TeacherIDs: []string{"f05327e0-a729-52e1-a0ed-641168e37ba4", "0e6b5f9d-0383-5ac6-b13d-8af02697fa8b"},
		TeacherIDs: []string{"6c0a0d79-86ba-461e-838e-916129db6169"},
		ClassIDs:   []string{"0b3f5f4d-3540-42ab-9fda-474fbbe8b51c"},
		Duration:   entity.TimeRange("1608600600-1630419300"),
	}
	result, err := GetReportModel().ListTeacherLoadLessons(ctx, &mockOperator, args)
	if err != nil {
		t.Fatal(err)
	}
	for i := range result {
		fmt.Printf("%#v\n", result[i])
	}
}

func TestReportModel_SummaryTeacherLoadLessons(t *testing.T) {
	ctx := context.Background()
	args := &entity.TeacherLoadLessonArgs{
		TeacherIDs: []string{"f05327e0-a729-52e1-a0ed-641168e37ba4", "0e6b5f9d-0383-5ac6-b13d-8af02697fa8b"},
		ClassIDs:   []string{"0b3f5f4d-3540-42ab-9fda-474fbbe8b51c"},
		Duration:   entity.TimeRange("1608600600-1630419300"),
	}
	result, err := GetReportModel().SummaryTeacherLoadLessons(ctx, &mockOperator, args)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v", result)
}
