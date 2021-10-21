package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var classSelectSubjectCount int
var classSelectSubjectAttendanceCount int
var studentSelectSubjectCount int
var studentSelectSubjectAttendanceCount int
var studentUnSelectSubjectCount int
var studentUnSelectSubjectAttendanceCount int

func (m *reportModel) ClassAttendanceStatistics(ctx context.Context, request *entity.ClassAttendanceRequest) (response *entity.ClassAttendanceResponse, err error) {
	response = new(entity.ClassAttendanceResponse)
	classAttendanceResponseItem := new(entity.ClassAttendanceResponseItem)
	reportDA := da.GetReportDA()
	response.RequestStudentID = request.StudentID
	response.Items = append(response.Items, classAttendanceResponseItem)
	for _, duration := range request.Durations {
		queryParameters := new(entity.ClassAttendanceQueryParameters)
		queryParameters.ClassID = request.ClassID
		queryParameters.Duration = duration
		// query condition all subject
		subjectList := make([]string, 0)
		subjectList = request.SelectedSubjectIDList
		for _, subject := range request.UnSelectedSubjectIDList {
			subjectList = append(subjectList, subject)
		}
		queryParameters.SubjectIDS = subjectList
		classAttendanceList, err := reportDA.GetClassAttendance(ctx, queryParameters)
		if err != nil {
			return nil, err
		}
		for _, classAttendance := range classAttendanceList {
			if isExistInArray(classAttendance.SubjectID, request.SelectedSubjectIDList) {
				classSelectSubjectCount++
				if classAttendance.IsAttendance {
					classSelectSubjectAttendanceCount++
				}
			}
			if classAttendance.StudentID == request.StudentID &&
				isExistInArray(classAttendance.SubjectID, request.SelectedSubjectIDList) {
				studentSelectSubjectCount++
				if classAttendance.IsAttendance {
					studentSelectSubjectAttendanceCount++
				}
			}
			if len(request.UnSelectedSubjectIDList) == 0 {
				if classAttendance.StudentID == request.StudentID &&
					isExistInArray(classAttendance.SubjectID, request.UnSelectedSubjectIDList) {
					studentUnSelectSubjectCount++
					if classAttendance.IsAttendance {
						studentUnSelectSubjectAttendanceCount++
					}
				}
			}
		}
		classAttendanceResponseItem.Duration = duration
		classAttendanceResponseItem.AttendedCount = studentSelectSubjectAttendanceCount + studentUnSelectSubjectAttendanceCount
		classAttendanceResponseItem.ScheduledCount = studentSelectSubjectCount + studentUnSelectSubjectCount
		if len(request.UnSelectedSubjectIDList) == 0 {
			if studentSelectSubjectCount > 0 && classSelectSubjectCount > 0 {
				classAttendanceResponseItem.AttendancePercentage = float64(studentSelectSubjectAttendanceCount) / float64(studentSelectSubjectCount)
				classAttendanceResponseItem.ClassAverageAttendancePercentage = float64(classSelectSubjectAttendanceCount) / float64(classSelectSubjectCount)
			}
		} else {
			if studentSelectSubjectCount > 0 && studentUnSelectSubjectCount > 0 && classSelectSubjectCount > 0 {
				classAttendanceResponseItem.AttendancePercentage = float64(studentSelectSubjectAttendanceCount) / float64(studentSelectSubjectCount)
				classAttendanceResponseItem.UnSelectedSubjectsAverageAttendancePercentage = float64(studentUnSelectSubjectAttendanceCount) / float64(studentUnSelectSubjectCount)
				classAttendanceResponseItem.ClassAverageAttendancePercentage = float64(classSelectSubjectAttendanceCount) / float64(classSelectSubjectCount)
			}
		}
		//reset
		classSelectSubjectCount = 0
		classSelectSubjectAttendanceCount = 0
		studentSelectSubjectCount = 0
		studentSelectSubjectAttendanceCount = 0
		studentUnSelectSubjectCount = 0
		studentUnSelectSubjectAttendanceCount = 0
	}
	return
}

func isExistInArray(value string, array []string) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}
