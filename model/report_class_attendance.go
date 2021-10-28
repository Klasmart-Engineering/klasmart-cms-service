package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
		//map key is class id
		studentSelectSubjectTotalMap := make(map[string]int)
		studentSelectSubjectAttendanceTotalMap := make(map[string]int)
		studentUnSelectSubjectTotalMap := make(map[string]int)
		studentUnSelectSubjectAttendanceTotalMap := make(map[string]int)
		classSelectSubjectTotalMap := make(map[string]int)
		classSelectSubjectAttendanceTotalMap := make(map[string]int)
		studentSelectSubjectAttendanceRateMap := make(map[string]float64)
		studentUnSelectSubjectAttendanceRateMap := make(map[string]float64)
		classSelectSubjectAttendanceRateMap := make(map[string]float64)
		var studentSelectSubjectAttendanceTotalRate float64
		var studentUnSelectSubjectAttendanceTotalRate float64
		var classSelectSubjectAttendanceTotalRate float64
		var studentAttendanceCount int
		var studentScheduleCount int
		for _, classAttendance := range classAttendanceList {
			//statistics  all subject total and attendance
			classSelectSubjectTotalMap[classAttendance.SubjectID] = classSelectSubjectTotalMap[classAttendance.SubjectID] + 1
			if classAttendance.IsAttendance {
				classSelectSubjectAttendanceTotalMap[classAttendance.SubjectID] = classSelectSubjectAttendanceTotalMap[classAttendance.SubjectID] + 1
			}
			//statistics  selected subject total,attendance,
			for _, selectSubject := range request.SelectedSubjectIDList {
				if classAttendance.StudentID == request.StudentID && selectSubject == classAttendance.SubjectID {
					studentScheduleCount = studentScheduleCount + 1
					studentSelectSubjectTotalMap[classAttendance.SubjectID] = studentSelectSubjectTotalMap[classAttendance.SubjectID] + 1
					if classAttendance.IsAttendance {
						studentAttendanceCount = studentAttendanceCount + 1
						studentSelectSubjectAttendanceTotalMap[classAttendance.SubjectID] = studentSelectSubjectAttendanceTotalMap[classAttendance.SubjectID] + 1
					}
				}
			}
			//statistics  unSelected subject total and attendance
			for _, unSelectSubject := range request.UnSelectedSubjectIDList {
				if classAttendance.StudentID == request.StudentID && unSelectSubject == classAttendance.SubjectID {
					studentUnSelectSubjectTotalMap[classAttendance.SubjectID] = studentUnSelectSubjectTotalMap[classAttendance.SubjectID] + 1
					if classAttendance.IsAttendance {
						studentUnSelectSubjectAttendanceTotalMap[classAttendance.SubjectID] = studentUnSelectSubjectAttendanceTotalMap[classAttendance.SubjectID] + 1
					}
				}
			}
		}
		for k, v := range studentSelectSubjectTotalMap {
			studentSelectSubjectAttendanceRateMap[k] = float64(studentSelectSubjectAttendanceTotalMap[k]) / float64(v)
		}
		for k, v := range studentUnSelectSubjectTotalMap {
			studentUnSelectSubjectAttendanceRateMap[k] = float64(studentUnSelectSubjectAttendanceTotalMap[k]) / float64(v)
		}
		for k, v := range classSelectSubjectTotalMap {
			classSelectSubjectAttendanceRateMap[k] = float64(classSelectSubjectAttendanceTotalMap[k]) / float64(v)
		}
		for _, v := range studentSelectSubjectAttendanceRateMap {
			studentSelectSubjectAttendanceTotalRate = studentSelectSubjectAttendanceTotalRate + v
		}
		for _, v := range studentUnSelectSubjectAttendanceRateMap {
			studentUnSelectSubjectAttendanceTotalRate = studentUnSelectSubjectAttendanceTotalRate + v
		}
		for _, v := range classSelectSubjectAttendanceRateMap {
			classSelectSubjectAttendanceTotalRate = classSelectSubjectAttendanceTotalRate + v
		}
		classAttendanceResponseItem := new(entity.ClassAttendanceResponseItem)
		classAttendanceResponseItem.Duration = duration
		// when select subject all , calculate only attendancePercentage,classAverageAttendancePercentage
		if len(request.UnSelectedSubjectIDList) == 0 && len(studentSelectSubjectAttendanceRateMap) > 0 &&
			len(classSelectSubjectAttendanceRateMap) > 0 {
			classAttendanceResponseItem.AttendancePercentage = studentSelectSubjectAttendanceTotalRate / float64(len(studentSelectSubjectAttendanceRateMap))
			classAttendanceResponseItem.ClassAverageAttendancePercentage = classSelectSubjectAttendanceTotalRate / float64(len(classSelectSubjectAttendanceRateMap))
			classAttendanceResponseItem.AttendedCount = studentAttendanceCount
			classAttendanceResponseItem.ScheduledCount = studentScheduleCount
		} else if len(studentSelectSubjectAttendanceRateMap) > 0 && len(classSelectSubjectAttendanceRateMap) > 0 &&
			len(studentUnSelectSubjectAttendanceRateMap) > 0 {
			classAttendanceResponseItem.AttendedCount = studentAttendanceCount
			classAttendanceResponseItem.ScheduledCount = studentScheduleCount
			classAttendanceResponseItem.AttendancePercentage = studentSelectSubjectAttendanceTotalRate / float64(len(studentSelectSubjectAttendanceRateMap))
			classAttendanceResponseItem.ClassAverageAttendancePercentage = classSelectSubjectAttendanceTotalRate / float64(len(classSelectSubjectAttendanceRateMap))
			classAttendanceResponseItem.UnSelectedSubjectsAverageAttendancePercentage = studentUnSelectSubjectAttendanceTotalRate / float64(len(studentUnSelectSubjectAttendanceRateMap))
		}
		response.Items = append(response.Items, classAttendanceResponseItem)
	}
	return
}
