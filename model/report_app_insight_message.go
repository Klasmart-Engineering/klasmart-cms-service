package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"strconv"
	"sync"
)

func (m *reportModel) GetAppInsightMessage(ctx context.Context, op *entity.Operator, req *entity.AppInsightMessageRequest) (res *entity.AppInsightMessageResponse, err error) {
	res = &entity.AppInsightMessageResponse{}
	var learningOutComeAchievementRequest *entity.LearnOutcomeAchievementRequest
	var attendanceRequest *entity.ClassAttendanceRequest
	var assignmentRequest *entity.AssignmentRequest
	var durations []entity.TimeRange
	var subjectIDList []string
	var learningOutComeAchievementResponse *entity.LearnOutcomeAchievementResponse
	var attendanceResponse *entity.ClassAttendanceResponse
	var assignmentResponse entity.AssignmentResponse
	for i := 0; i < entity.Repoet4W; i++ {
		var time entity.TimeRange
		startTime := strconv.Itoa(req.EndTime - (entity.Repoet4W-i)*7*24*60*60)
		endTime := strconv.Itoa(req.EndTime - (entity.Repoet4W-(i+1))*7*24*60*60)
		time = entity.TimeRange(startTime + "-" + endTime)
		durations = append(durations, time)
	}
	result, err := external.GetSubjectServiceProvider().GetByOrganization(ctx, op)
	if err != nil {
		return
	}
	for _, subject := range result {
		subjectIDList = append(subjectIDList, subject.ID)
	}
	learningOutComeAchievementRequest = &entity.LearnOutcomeAchievementRequest{ClassID: req.ClassID,
		StudentID: req.StudentID, SelectedSubjectIDList: subjectIDList, Durations: durations}
	attendanceRequest = &entity.ClassAttendanceRequest{ClassID: req.ClassID,
		StudentID: req.StudentID, SelectedSubjectIDList: subjectIDList, Durations: durations}
	assignmentRequest = &entity.AssignmentRequest{ClassID: req.ClassID,
		StudentID: req.StudentID, SelectedSubjectIDList: subjectIDList, Durations: durations}
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		learningOutComeAchievementResponse, err = m.GetStudentProgressLearnOutcomeAchievement(ctx, op, learningOutComeAchievementRequest)
		wg.Done()
	}()
	go func() {
		attendanceResponse, err = m.ClassAttendanceStatistics(ctx, op, attendanceRequest)
		wg.Done()
	}()
	go func() {
		assignmentResponse, err = m.GetAssignmentCompletion(ctx, op, assignmentRequest)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		return
	}
	res.LearningOutcomeAchivementLabelID = learningOutComeAchievementResponse.LabelID
	res.LearningOutcomeAchivementLabelParams = learningOutComeAchievementResponse.LabelParams
	res.AttedanceLabelID = attendanceResponse.LabelID
	res.AttedanceLabelParams = attendanceResponse.LabelParams
	res.AssignmentLabelID = assignmentResponse.LabelID
	res.AssignmentLabelParams = assignmentResponse.LabelParams
	return
}
