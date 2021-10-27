package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

func (m *reportModel) ClassAttendanceStatistics(ctx context.Context, op *entity.Operator, request *entity.ClassAttendanceRequest) (response *entity.ClassAttendanceResponse, err error) {
	response = new(entity.ClassAttendanceResponse)
	response.RequestStudentID = request.StudentID
	for _, duration := range request.Durations {
		queryParameters := new(entity.ClassAttendanceQueryParameters)
		queryParameters.ClassID = request.ClassID
		queryParameters.Duration = duration
		// query condition all subject
		queryParameters.SubjectIDS = append(request.SelectedSubjectIDList, request.UnSelectedSubjectIDList...)
		classAttendanceList, err := da.GetReportDA().GetClassAttendance(ctx, queryParameters)
		if err != nil {
			return nil, err
		}
		var classSelectSubjectCount int
		var classSelectSubjectAttendanceCount int
		var studentSelectSubjectCount int
		var studentSelectSubjectAttendanceCount int
		var studentUnSelectSubjectCount int
		var studentUnSelectSubjectAttendanceCount int
		for _, classAttendance := range classAttendanceList {
			// statistics class average attendance selected subject
			if utils.ContainsStr(request.SelectedSubjectIDList, classAttendance.SubjectID) {
				classSelectSubjectCount++
				if classAttendance.IsAttendance {
					classSelectSubjectAttendanceCount++
				}
			}
			//statistics student attendance selected subject
			if classAttendance.StudentID == request.StudentID &&
				utils.ContainsStr(request.SelectedSubjectIDList, classAttendance.SubjectID) {
				studentSelectSubjectCount++
				if classAttendance.IsAttendance {
					studentSelectSubjectAttendanceCount++
				}
			}
			//when subject not select all,statistics student attendance other subject
			if len(request.UnSelectedSubjectIDList) != 0 && classAttendance.StudentID == request.StudentID &&
				utils.ContainsStr(request.UnSelectedSubjectIDList, classAttendance.SubjectID) {
				studentUnSelectSubjectCount++
				if classAttendance.IsAttendance {
					studentUnSelectSubjectAttendanceCount++
				}
			}
		}
		classAttendanceResponseItem := new(entity.ClassAttendanceResponseItem)
		classAttendanceResponseItem.Duration = duration
		// when select subject all , calculate only attendancePercentage,classAverageAttendancePercentage
		if len(request.UnSelectedSubjectIDList) == 0 && studentSelectSubjectCount > 0 && classSelectSubjectCount > 0 {
			classAttendanceResponseItem.AttendancePercentage = float64(studentSelectSubjectAttendanceCount) / float64(studentSelectSubjectCount)
			classAttendanceResponseItem.ClassAverageAttendancePercentage = float64(classSelectSubjectAttendanceCount) / float64(classSelectSubjectCount)
			classAttendanceResponseItem.AttendedCount = studentSelectSubjectAttendanceCount + studentUnSelectSubjectAttendanceCount
			classAttendanceResponseItem.ScheduledCount = studentSelectSubjectCount + studentUnSelectSubjectCount
		} else if studentSelectSubjectCount > 0 && studentUnSelectSubjectCount > 0 && classSelectSubjectCount > 0 {
			classAttendanceResponseItem.AttendedCount = studentSelectSubjectAttendanceCount
			classAttendanceResponseItem.ScheduledCount = studentSelectSubjectCount
			classAttendanceResponseItem.AttendancePercentage = float64(studentSelectSubjectAttendanceCount) / float64(studentSelectSubjectCount)
			classAttendanceResponseItem.UnSelectedSubjectsAverageAttendancePercentage = float64(studentUnSelectSubjectAttendanceCount) / float64(studentUnSelectSubjectCount)
			classAttendanceResponseItem.ClassAverageAttendancePercentage = float64(classSelectSubjectAttendanceCount) / float64(classSelectSubjectCount)
		}
		response.Items = append(response.Items, classAttendanceResponseItem)
	}
	return
}
