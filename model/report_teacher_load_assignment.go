package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

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
	mTeacherClass, err := external.GetTeacherLoadServiceProvider().BatchGetActiveClassWithStudent(ctx, op, req.TeacherIDList)
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
			}
		}

		var feedbacks entity.TeacherLoadAssignmentFeedbackSlice
		feedbacks, err = reportDA.GetTeacherLoadAssignmentFeedbackOfHomeFun(ctx, req)
		if err != nil {
			return
		}
		mFeedback := feedbacks.MapTeacherID()
		for _, item := range items {
			if sc, ok := mFeedback[item.TeacherID]; ok {
				item.Feedbacks = append(item.Feedbacks, sc.FeedbackOfAssignment)
			}
		}
	}
	if req.ClassTypeList.Contains(constant.ReportClassTypeStudy) {
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
		log.Info(
			ctx,
			"get room comments by external.GetH5PRoomCommentServiceProvider",
			log.Any("scheduleIDListForCommentOfStudy", scheduleIDListForCommentOfStudy),
			log.Any("mStudyComment", mStudyComment),
		)
		for roomID, comments := range mStudyComment {
			mAssignment := map[string]*entity.TeacherLoadAssignmentRoomAssignment{}
			for _, comment := range comments {
				if comment == nil {
					continue
				}
				for _, teacherComment := range comment.TeacherComments {
					if teacherComment == nil || teacherComment.Teacher == nil {
						continue
					}
					teacherID := teacherComment.Teacher.UserID
					roomAssignment, ok := mAssignment[teacherID]
					if !ok {
						roomAssignment = &entity.TeacherLoadAssignmentRoomAssignment{
							RoomID: roomID,
						}
						mAssignment[teacherID] = roomAssignment
					}
					roomAssignment.CountOfCompleteAssignment++
					if len(teacherComment.Comment) > 0 {
						roomAssignment.CountOfCommentAssignment++
					}

					if req.Duration.MustContain(ctx, teacherComment.Date) {
						for _, item := range items {
							if item.TeacherID != teacherID {
								continue
							}
							// 3.1 fill CountOfCompletedAssignment
							item.CountOfCompletedAssignment++
						}
					}
				}
			}

			for _, item := range items {
				if ra, ok := mAssignment[item.TeacherID]; ok {
					item.Feedbacks = append(item.Feedbacks, ra.Feedback())
				}
			}
			log.Info(ctx, "mAssignment", log.Any("mAssignment", mAssignment), log.Any("items", items))
		}
	}

	// 3.2 fill FeedbackPercentage
	for _, item := range items {
		item.FeedbackPercentage = item.Feedbacks.Avg()
	}
	log.Info(ctx, "fill FeedbackPercentage", log.Any("items", items))

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
