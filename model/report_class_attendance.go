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
		// map key is student id
		var studentSelectSubjectTotalMap map[string]int
		var studentSelectSubjectAttendanceTotalMap map[string]int
		var studentUnSelectSubjectTotalMap map[string]int
		var studentUnSelectSubjectAttendanceTotalMap map[string]int
		var classStudentSelectSubjectAttendanceRateMap map[string]float64
		var classStudentSelectSubjectAttendanceTotalRate float64
		for _, classAttendance := range classAttendanceList {
			//statistics student attendance selected subject
			if utils.ContainsStr(request.SelectedSubjectIDList, classAttendance.SubjectID) {
				studentSelectSubjectTotalMap[classAttendance.StudentID] = studentSelectSubjectTotalMap[classAttendance.StudentID] + 1
				if classAttendance.IsAttendance {
					studentSelectSubjectAttendanceTotalMap[classAttendance.StudentID] = studentSelectSubjectAttendanceTotalMap[classAttendance.StudentID] + 1
				}
			}
			//when subject not select all,statistics student attendance other subject
			if len(request.UnSelectedSubjectIDList) != 0 &&
				utils.ContainsStr(request.UnSelectedSubjectIDList, classAttendance.SubjectID) {
				studentUnSelectSubjectTotalMap[classAttendance.StudentID] = studentUnSelectSubjectTotalMap[classAttendance.StudentID] + 1
				if classAttendance.IsAttendance {
					studentUnSelectSubjectAttendanceTotalMap[classAttendance.StudentID] = studentUnSelectSubjectAttendanceTotalMap[classAttendance.StudentID] + 1
				}
			}

		}
		// statistics the average attendance rate of students in this class under this subject
		for k, v := range studentSelectSubjectTotalMap {
			classStudentSelectSubjectAttendanceRateMap[k] = float64(studentSelectSubjectAttendanceTotalMap[k]) / float64(v)
		}

		for _, v := range classStudentSelectSubjectAttendanceRateMap {
			classStudentSelectSubjectAttendanceTotalRate = classStudentSelectSubjectAttendanceTotalRate + v
		}
		classAttendanceResponseItem := new(entity.ClassAttendanceResponseItem)
		classAttendanceResponseItem.Duration = duration
		// when select subject all , calculate only attendancePercentage,classAverageAttendancePercentage
		if len(request.UnSelectedSubjectIDList) == 0 && len(classStudentSelectSubjectAttendanceRateMap) > 0 {
			classAttendanceResponseItem.AttendancePercentage = classStudentSelectSubjectAttendanceRateMap[request.StudentID]
			classAttendanceResponseItem.ClassAverageAttendancePercentage = classStudentSelectSubjectAttendanceTotalRate / float64(len(classStudentSelectSubjectAttendanceRateMap))
			classAttendanceResponseItem.AttendedCount = studentSelectSubjectAttendanceTotalMap[request.StudentID] + studentUnSelectSubjectAttendanceTotalMap[request.StudentID]
			classAttendanceResponseItem.ScheduledCount = studentSelectSubjectTotalMap[request.StudentID] + studentUnSelectSubjectTotalMap[request.StudentID]
		} else if len(classStudentSelectSubjectAttendanceRateMap) > 0 && studentUnSelectSubjectTotalMap[request.StudentID] > 0 {
			classAttendanceResponseItem.AttendedCount = studentSelectSubjectAttendanceTotalMap[request.StudentID] + studentUnSelectSubjectAttendanceTotalMap[request.StudentID]
			classAttendanceResponseItem.ScheduledCount = studentSelectSubjectTotalMap[request.StudentID] + studentUnSelectSubjectTotalMap[request.StudentID]
			classAttendanceResponseItem.AttendancePercentage = classStudentSelectSubjectAttendanceRateMap[request.StudentID]
			classAttendanceResponseItem.UnSelectedSubjectsAverageAttendancePercentage = float64(studentUnSelectSubjectAttendanceTotalMap[request.StudentID]) / float64(studentUnSelectSubjectTotalMap[request.StudentID])
			classAttendanceResponseItem.ClassAverageAttendancePercentage = classStudentSelectSubjectAttendanceTotalRate / float64(len(classStudentSelectSubjectAttendanceRateMap))
		}
		response.Items = append(response.Items, classAttendanceResponseItem)
	}
	return
}
