package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (m *reportModel) GetTeacherLoadReportOfAssignment(ctx context.Context, op *entity.Operator, req *entity.TeacherLoadAssignmentRequest) (items []*entity.TeacherLoadAssignmentResponseItem, err error) {
	// 0. init response
	items = make([]*entity.TeacherLoadAssignmentResponseItem, 0, len(req.TeacherIDList))
	for _, teacherID := range req.TeacherIDList {
		items = append(items, &entity.TeacherLoadAssignmentResponseItem{
			TeacherID: teacherID,
		})
	}

	// 1. fill CountOfClasses and CountOfStudents
	mTeacherClass, err := external.GetTeacherLoadServiceProvider().BatchGetClassWithStudent(ctx, op, req.TeacherIDList)
	if err != nil {
		return
	}
	for _, item := range items {
		if classWithStudent, ok := mTeacherClass[item.TeacherID]; ok {
			classWithStudentCounter := classWithStudent.CountClassAndStudent(ctx, req.ClassIDList)
			item.CountOfClasses = int64(classWithStudentCounter.Class)
			item.CountOfStudents = int64(classWithStudentCounter.Student)
		}
	}

	reportDA := da.GetReportDA()
	// 2. fill CountOfScheduledAssignment
	scheCounts, err := reportDA.GetTeacherLoadAssignmentScheduledCount(ctx, req)
	if err != nil {
		return
	}
	mScheCount := scheCounts.MapTeacherID()
	for _, item := range items {
		if sc, ok := mScheCount[item.TeacherID]; ok {
			item.CountOfScheduledAssignment = sc.CountOfScheduledAssignment
		}
	}

	// 3. fill CountOfCompletedAssignment and FeedbackPercentage
	if req.ClassTypeList.Contains(constant.ReportClassTypeHomeFun) {
		var completedCount entity.TeacherLoadAssignmentResponseItemSlice
		completedCount, err = reportDA.GetTeacherLoadAssignmentCompletedCountOfHomeFun(ctx, req)
		if err != nil {
			return
		}
		mCompleteCount := completedCount.MapTeacherID()
		for _, item := range items {
			if sc, ok := mCompleteCount[item.TeacherID]; ok {
				item.CountOfCompletedAssignment = sc.CountOfCompletedAssignment
				item.CountOfCommentedAssignment = sc.CountOfCommentedAssignment
			}
		}
	}
	if req.ClassTypeList.Contains(constant.ReportClassTypeStudy) {
		mCompleteCountStudy := map[string]*entity.TeacherLoadAssignmentResponseItem{}
		var scheduleIDListForCommentOfStudy []string
		scheduleIDListForCommentOfStudy, err = reportDA.GetTeacherLoadAssignmentScheduleIDListForStudyComment(ctx, req)
		if err != nil {
			return
		}
		var mStudyComment map[string][]*external.H5PTeacherCommentsByStudent
		mStudyComment, err = external.GetH5PRoomCommentServiceProvider().BatchGet(ctx, op, scheduleIDListForCommentOfStudy)
		if err != nil {
			return
		}
		for _, comments := range mStudyComment {
			for _, comment := range comments {
				if comment == nil {
					continue
				}
				for _, teacherComment := range comment.TeacherComments {
					if teacherComment == nil {
						continue
					}
					if !req.Duration.MustContain(ctx, teacherComment.Date) {
						continue
					}
					if teacherComment.Teacher == nil {
						continue
					}
					teacherID := teacherComment.Teacher.UserID
					item, ok := mCompleteCountStudy[teacherID]
					if !ok {
						item = &entity.TeacherLoadAssignmentResponseItem{
							TeacherID: teacherID,
						}
						mCompleteCountStudy[teacherID] = item
					}
					item.CountOfCompletedAssignment++
					if len(teacherComment.Comment) > 0 {
						item.CountOfCommentedAssignment++
					}
				}
			}
		}
		for _, item := range items {
			if sc, ok := mCompleteCountStudy[item.TeacherID]; ok {
				item.CountOfScheduledAssignment += sc.CountOfCompletedAssignment
				item.CountOfCommentedAssignment += sc.CountOfCommentedAssignment
			}
		}
	}
	for _, item := range items {
		if item.CountOfCompletedAssignment == 0 {
			continue
		}
		item.FeedbackPercentage = float64(item.CountOfCommentedAssignment) / float64(item.CountOfCompletedAssignment)
	}

	// 4. fill CountOfPendingAssignment and AvgDaysOfPendingAssignment
	countPending, err := reportDA.GetTeacherLoadAssignmentPendingAssessment(ctx, req)
	if err != nil {
		return
	}
	mPending := countPending.MapTeacherID()
	for _, item := range items {
		if sc, ok := mPending[item.TeacherID]; ok {
			item.CountOfPendingAssignment = sc.CountOfPendingAssignment
			item.AvgDaysOfPendingAssignment = sc.AvgDaysOfPendingAssignment
		}
	}

	return
}
