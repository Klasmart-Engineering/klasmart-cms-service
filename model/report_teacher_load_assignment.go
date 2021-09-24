package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (m *reportModel) GetTeacherLoadReportOfAssignment(ctx context.Context, op *entity.Operator, req *entity.TeacherLoadAssignmentRequest) (items []*entity.TeacherLoadAssignmentResponseItem, err error) {
	items = make([]*entity.TeacherLoadAssignmentResponseItem, 0, len(req.TeacherIDList))
	for _, teacherID := range req.TeacherIDList {
		items = append(items, &entity.TeacherLoadAssignmentResponseItem{
			TeacherID: teacherID,
		})
	}

	reportDA := da.GetReportDA()
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

	completedCount, err := reportDA.GetTeacherLoadAssignmentCompletedCountOfHomeFun(ctx, req)
	if err != nil {
		return
	}
	mCompleteCount := completedCount.MapTeacherID()
	for _, item := range items {
		if sc, ok := mCompleteCount[item.TeacherID]; ok {
			item.CountOfScheduledAssignment = sc.CountOfCompletedAssignment
			item.CountOfCommentedAssignment = sc.CountOfCommentedAssignment
		}
	}

	mCompleteCountStudy := map[string]*entity.TeacherLoadAssignmentResponseItem{}
	scheduleIDListForCommentOfStudy, err := reportDA.GetTeacherLoadAssignmentScheduleIDListForStudyComment(ctx, req)
	if err != nil {
		return
	}
	mStudyComment, err := external.GetH5PRoomCommentServiceProvider().BatchGet(ctx, op, scheduleIDListForCommentOfStudy)
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
					return
				}
				teacherID := teacherComment.Teacher.UserID
				item, ok := mCompleteCountStudy[teacherID]
				if !ok {
					item = &entity.TeacherLoadAssignmentResponseItem{
						TeacherID: teacherID,
					}
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
	for _, item := range items {
		if item.CountOfCompletedAssignment == 0 {
			continue
		}
		item.FeedbackPercentage = float64(item.CountOfCommentedAssignment) / float64(item.CountOfCompletedAssignment)
	}

	// TODO pending count, pending days

	return
}
