package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da/assessmentV2"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

var (
	assessmentInternalModelInstance     IAssessmentInternalModelV2
	assessmentInternalModelInstanceOnce = sync.Once{}
)

type assessmentInternalModel struct {
	AmsServices external.AmsServices
}

type IAssessmentInternalModelV2 interface {
	// live service callback
	ScheduleEndClassCallback(ctx context.Context, operator *entity.Operator, args *v2.ScheduleEndClassCallBackReq) error

	AddWhenCreateSchedules(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, req *v2.AssessmentAddWhenCreateSchedulesReq) error
	LockAssessmentContentAndOutcome(ctx context.Context, op *entity.Operator, schedule *entity.Schedule) error
	DeleteByScheduleIDsTx(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, scheduleIDs []string) error
	AnyoneAttemptedByScheduleIDs(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string]*v2.AssessmentAnyoneAttemptedReply, error)
	Query(ctx context.Context, op *entity.Operator, condition *assessmentV2.AssessmentCondition) ([]*v2.Assessment, error)
	UpdateWhenReviewScheduleSuccess(ctx context.Context, tx *dbo.DBContext, scheduleID string) error
}

func (a *assessmentInternalModel) ScheduleEndClassCallback(ctx context.Context, op *entity.Operator, req *v2.ScheduleEndClassCallBackReq) error {
	//locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixScheduleID, req.ScheduleID)
	//if err != nil {
	//	log.Error(ctx, "ScheduleEndClassCallback: lock fail",
	//		log.Err(err),
	//		log.Any("req", req),
	//	)
	//	return err
	//}
	//locker.Lock()
	//defer locker.Unlock()

	req.AttendanceIDs = utils.SliceDeduplicationExcludeEmpty(req.AttendanceIDs)
	if req.ScheduleID == "" || len(req.AttendanceIDs) <= 0 {
		log.Warn(ctx, "request data is invalid", log.Any("req", req))
		return constant.ErrInvalidArgs
	}

	assessment, err := a.getAssessmentByScheduleID(ctx, req.ScheduleID)
	if err != nil {
		return err
	}

	if assessment.Status == v2.AssessmentStatusComplete {
		log.Warn(ctx, "assessment is completed", log.Any("assessment", assessment), log.Any("req", req))
		return nil
	}

	if assessment.AssessmentType != v2.AssessmentTypeOnlineClass &&
		assessment.AssessmentType != v2.AssessmentTypeOfflineClass &&
		assessment.AssessmentType != v2.AssessmentTypeOnlineStudy &&
		assessment.AssessmentType != v2.AssessmentTypeReviewStudy {
		log.Warn(ctx, "not support this assessment type", log.Any("assessment", assessment), log.Any("req", req))
		return constant.ErrInvalidArgs
	}

	err = a.endClassCallbackUpdateAssessment(ctx, op, req, assessment)
	if err != nil {
		return err
	}

	return nil
}

func (a *assessmentInternalModel) AddWhenCreateSchedules(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, req *v2.AssessmentAddWhenCreateSchedulesReq) error {
	if !req.Valid(ctx) {
		log.Warn(ctx, "req is invalid", log.Any("req", req), log.Any("op", op))
		return constant.ErrInvalidArgs
	}

	now := time.Now().Unix()
	assessments := make([]*v2.Assessment, len(req.RepeatScheduleIDs))

	// title
	title, err := req.AssessmentType.Title(ctx, v2.GenerateAssessmentTitleInput{
		ClassName:    req.ClassRosterClassName,
		ScheduleName: req.ScheduleTitle,
	})
	if err != nil {
		return err
	}

	users := make([]*v2.AssessmentUser, 0, len(req.RepeatScheduleIDs)*len(req.Users))

	for i, scheduleID := range req.RepeatScheduleIDs {
		// assessment
		assessmentItem := &v2.Assessment{
			ID:             utils.NewID(),
			OrgID:          op.OrgID,
			ScheduleID:     scheduleID,
			AssessmentType: req.AssessmentType,
			Title:          title,
			Status:         v2.AssessmentStatusNotStarted,
			CreateAt:       now,
			MigrateFlag:    constant.AssessmentCurrentFlag,
		}

		if req.AssessmentType == v2.AssessmentTypeReviewStudy {
			assessmentItem.Status = v2.AssessmentStatusPending
		}

		assessments[i] = assessmentItem

		// assessment user
		for _, userItem := range req.Users {
			attendance := &v2.AssessmentUser{
				ID:             utils.NewID(),
				AssessmentID:   assessmentItem.ID,
				UserID:         userItem.UserID,
				UserType:       userItem.UserType,
				StatusBySystem: v2.AssessmentUserSystemStatusNotStarted,
				StatusByUser:   v2.AssessmentUserStatusParticipate,
				CreateAt:       now,
			}

			if req.AssessmentType == v2.AssessmentTypeOnlineClass {
				attendance.StatusByUser = v2.AssessmentUserStatusNotParticipate
			}

			users = append(users, attendance)
		}
	}

	_, err = assessmentV2.GetAssessmentDA().InsertInBatchesTx(ctx, tx, assessments, constant.AssessmentBatchPageSize)
	if err != nil {
		return err
	}

	_, err = assessmentV2.GetAssessmentUserDA().InsertInBatchesTx(ctx, tx, users, constant.AssessmentBatchPageSize)
	if err != nil {
		return err
	}

	return nil
}

func (a *assessmentInternalModel) LockAssessmentContentAndOutcome(ctx context.Context, op *entity.Operator, schedule *entity.Schedule) error {
	if !schedule.IsLockedLessonPlan() {
		log.Warn(ctx, "no one attempted,don't need to lock", log.Any("schedule", schedule))
		return nil
	}

	assessment, err := a.getAssessmentByScheduleID(ctx, schedule.ID)
	if err != nil {
		return err
	}

	now := time.Now().Unix()

	at, err := NewAssessmentTool(ctx, op, []*v2.Assessment{assessment})
	if err != nil {
		return err
	}
	contentsFromSchedule, err := at.firstGetLockedContentBySchedule(schedule)
	if err != nil {
		return err
	}

	// lock contents
	waitAddContents := make([]*v2.AssessmentContent, 0)
	contentMapFromSchedule := make(map[string]*v2.AssessmentContentView, len(contentsFromSchedule))
	for _, item := range contentsFromSchedule {
		if _, ok := contentMapFromSchedule[item.ID]; ok {
			log.Warn(ctx, "content repeated", log.String("repeated contentID", item.ID), log.Any("contentMapFromSchedule", contentMapFromSchedule))
			continue
		}
		contentNewItem := &v2.AssessmentContent{
			ID:           utils.NewID(),
			AssessmentID: assessment.ID,
			ContentID:    item.ID,
			ContentType:  item.ContentType,
			Status:       v2.AssessmentContentStatusCovered,
			CreateAt:     now,
		}
		waitAddContents = append(waitAddContents, contentNewItem)
		contentMapFromSchedule[item.ID] = item
	}

	// lock outcomes
	outcomeIDs := make([]string, 0)
	for _, item := range contentsFromSchedule {
		outcomeIDs = append(outcomeIDs, item.OutcomeIDs...)
	}

	// convert outcome in content to the latest outcome
	outcomeIDs = utils.SliceDeduplicationExcludeEmpty(outcomeIDs)
	waitAddUserOutcomes := make([]*v2.AssessmentUserOutcome, 0)
	assessmentUserIDs := make([]string, 0)

	if len(outcomeIDs) > 0 {
		assessmentUserMap, err := at.GetAssessmentUserMap()
		if err != nil {
			return err
		}
		assessmentUsers, ok := assessmentUserMap[assessment.ID]
		if !ok {
			log.Error(ctx, "can not found assessment users", log.String("assessmentID", assessment.ID), log.Any("assessmentUserMap", assessmentUserMap))
			return constant.ErrRecordNotFound
		}

		latestOutcomeMap, _, err := GetOutcomeModel().GetLatestOutcomes(ctx, op, dbo.MustGetDB(ctx), outcomeIDs)
		if err != nil {
			return err
		}

		log.Debug(ctx, "latestOutcomeMap data",
			log.Any("latestOutcomeMap", latestOutcomeMap),
			log.Strings("old outcomeIDs", outcomeIDs),
			log.Any("contentsFromSchedule", contentsFromSchedule))

		for _, userItem := range assessmentUsers {
			if userItem.UserType != v2.AssessmentUserTypeStudent {
				continue
			}

			assessmentUserIDs = append(assessmentUserIDs, userItem.ID)

			for _, assContentItem := range waitAddContents {
				contentItem, ok := contentMapFromSchedule[assContentItem.ContentID]
				if !ok {
					log.Warn(ctx, "not found content in contentMapFromSchedule", log.Any("contentMapFromSchedule", contentMapFromSchedule), log.String("assContentItem.ContentID", assContentItem.ContentID))
					continue
				}

				for _, oldOutcomeID := range contentItem.OutcomeIDs {
					latestOutcomeItem, ok := latestOutcomeMap[oldOutcomeID]
					if !ok {
						log.Warn(ctx, "not found outcome in latestOutcomeMap", log.Any("latestOutcomeMap", latestOutcomeMap), log.String("oldOutcomeID", oldOutcomeID))
						continue
					}

					userOutcomeItem := &v2.AssessmentUserOutcome{
						ID:                  utils.NewID(),
						AssessmentUserID:    userItem.ID,
						AssessmentContentID: assContentItem.ID,
						OutcomeID:           latestOutcomeItem.ID,
						Status:              v2.AssessmentUserOutcomeStatusUnknown,
						CreateAt:            now,
						UpdateAt:            0,
						DeleteAt:            0,
					}
					if latestOutcomeItem.Assumed {
						userOutcomeItem.Status = v2.AssessmentUserOutcomeStatusAchieved
					}

					waitAddUserOutcomes = append(waitAddUserOutcomes, userOutcomeItem)
				}
			}
		}
	}

	// The reason for deleting first and then adding is to consider the migration of old and new data
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := assessmentV2.GetAssessmentContentDA().DeleteByAssessmentIDsTx(ctx, tx, []string{assessment.ID})
		if err != nil {
			return err
		}

		_, err = assessmentV2.GetAssessmentContentDA().InsertTx(ctx, tx, waitAddContents)
		if err != nil {
			return err
		}

		if len(waitAddUserOutcomes) > 0 {
			err = assessmentV2.GetAssessmentUserOutcomeDA().DeleteByAssessmentUserIDsTx(ctx, tx, assessmentUserIDs)
			if err != nil {
				return err
			}
			_, err = assessmentV2.GetAssessmentUserOutcomeDA().InsertTx(ctx, tx, waitAddUserOutcomes)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Error(ctx, "lock assessment contents error", log.Err(err), log.Any("waitAddContents", waitAddContents), log.Any("waitAddUserOutcomes", waitAddUserOutcomes))
		return err
	}
	return nil
}

func (a *assessmentInternalModel) DeleteByScheduleIDsTx(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, scheduleIDs []string) error {
	var assessments []*v2.Assessment
	err := assessmentV2.GetAssessmentDA().Query(ctx, &assessmentV2.AssessmentCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	}, &assessments)
	if err != nil {
		log.Error(ctx, "get assessment by schedule ids error", log.Err(err), log.Strings("ScheduleIDs", scheduleIDs))
		return err
	}

	assessmentIDs := make([]string, len(assessments))
	for i, item := range assessments {
		//if item.Status != v2.AssessmentStatusNotStarted {
		//	return ErrAssessmentNotAllowDelete
		//}
		assessmentIDs[i] = item.ID
	}

	// delete assessment
	err = assessmentV2.GetAssessmentDA().DeleteByScheduleIDsTx(ctx, tx, scheduleIDs)
	if err != nil {
		log.Error(ctx, "del assessment by schedule ids error", log.Err(err), log.Strings("ScheduleIDs", scheduleIDs))
		return err
	}
	// delete assessment user
	err = assessmentV2.GetAssessmentUserDA().DeleteByAssessmentIDsTx(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "del assessment user by assessment ids error", log.Err(err), log.Strings("assessmentIDs", assessmentIDs))
		return err
	}
	// delete assessment content
	err = assessmentV2.GetAssessmentContentDA().DeleteByAssessmentIDsTx(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "del assessment content by assessment ids error", log.Err(err), log.Strings("assessmentIDs", assessmentIDs))
		return err
	}

	return err
}

// TODO:: refactor
func (a *assessmentInternalModel) UpdateAssessmentUserStatusTime(ctx context.Context, waitUpdateAssessmentUsers []*v2.AssessmentUser) error {
	now := time.Now().Unix()
	for _, item := range waitUpdateAssessmentUsers {
		switch item.StatusBySystem {
		case v2.AssessmentUserSystemStatusNotStarted:
		case v2.AssessmentUserSystemStatusInProgress:
			if item.InProgressAt == 0 {
				item.InProgressAt = now
			}
		case v2.AssessmentUserSystemStatusDone:
			if item.DoneAt == 0 {
				item.DoneAt = now
			}
		case v2.AssessmentUserSystemStatusResubmitted:
			item.ResubmittedAt = now
		case v2.AssessmentUserSystemStatusCompleted:
			if item.CompletedAt == 0 {
				item.CompletedAt = now
			}
		}
	}
	return nil
}
func (a *assessmentInternalModel) endClassCallbackUpdateAssessment(ctx context.Context, op *entity.Operator, req *v2.ScheduleEndClassCallBackReq, assessment *v2.Assessment) error {
	now := time.Now().Unix()

	attendanceCondition := &assessmentV2.AssessmentUserCondition{
		AssessmentID: sql.NullString{
			String: assessment.ID,
			Valid:  true,
		},
		UserIDs: entity.NullStrings{
			Strings: req.AttendanceIDs,
			Valid:   true,
		},
	}
	var assessmentUsers []*v2.AssessmentUser
	err := assessmentV2.GetAssessmentUserDA().Query(ctx, attendanceCondition, &assessmentUsers)
	if err != nil {
		return err
	}

	if assessment.Status == v2.AssessmentStatusNotStarted {
		// update assessment
		if assessment.AssessmentType == v2.AssessmentTypeOfflineClass ||
			assessment.AssessmentType == v2.AssessmentTypeOnlineClass {
			// update assessment title

			titleSplit := strings.SplitN(assessment.Title, "-", 2)
			if len(titleSplit) == 2 {
				var timeStr string
				if req.ClassEndAt > 0 {
					timeStr = time.Unix(req.ClassEndAt, 0).Format("20060102")
					assessment.Title = fmt.Sprintf("%s-%s", timeStr, titleSplit[1])
				}
			}
		}

		assessment.Status = v2.AssessmentStatusStarted
		assessment.UpdateAt = now
		assessment.ClassLength = req.ClassLength
		assessment.ClassEndAt = req.ClassEndAt

		for _, userItem := range assessmentUsers {
			userItem.StatusBySystem = v2.AssessmentUserSystemStatusDone
			userItem.StatusByUser = v2.AssessmentUserStatusParticipate
			userItem.UpdateAt = now
		}

		err = a.UpdateAssessmentUserStatusTime(ctx, assessmentUsers)
		if err != nil {
			return err
		}
		return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
			_, err := assessmentV2.GetAssessmentDA().UpdateTx(ctx, tx, assessment)
			if err != nil {
				return err
			}

			if len(assessmentUsers) > 0 {
				_, err := assessmentV2.GetAssessmentUserDA().UpdateTx(ctx, tx, assessmentUsers)
				if err != nil {
					return err
				}
			}

			return nil
		})
	} else if assessment.Status == v2.AssessmentStatusStarted {
		for _, userItem := range assessmentUsers {
			if userItem.StatusBySystem == v2.AssessmentUserSystemStatusDone {
				userItem.StatusBySystem = v2.AssessmentUserSystemStatusResubmitted
			} else {
				userItem.StatusBySystem = v2.AssessmentUserSystemStatusDone
			}

			userItem.StatusByUser = v2.AssessmentUserStatusParticipate
			userItem.UpdateAt = now
		}
		err = a.UpdateAssessmentUserStatusTime(ctx, assessmentUsers)
		if err != nil {
			return err
		}
		_, err := assessmentV2.GetAssessmentUserDA().Update(ctx, assessmentUsers)
		if err != nil {
			return err
		}
	} else {
		for _, userItem := range assessmentUsers {
			if userItem.StatusBySystem == v2.AssessmentUserSystemStatusDone {
				userItem.StatusBySystem = v2.AssessmentUserSystemStatusResubmitted
			} else {
				userItem.StatusBySystem = v2.AssessmentUserSystemStatusDone
			}

			userItem.UpdateAt = now
		}
		err = a.UpdateAssessmentUserStatusTime(ctx, assessmentUsers)
		if err != nil {
			return err
		}
		_, err := assessmentV2.GetAssessmentUserDA().Update(ctx, assessmentUsers)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *assessmentInternalModel) getAssessmentByScheduleID(ctx context.Context, scheduleID string) (*v2.Assessment, error) {
	var assessments []*v2.Assessment
	condition := &assessmentV2.AssessmentCondition{
		ScheduleID: sql.NullString{
			String: scheduleID,
			Valid:  true,
		},
	}
	err := assessmentV2.GetAssessmentDA().Query(ctx, condition, &assessments)
	if err != nil {
		log.Error(ctx, "query assessment error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	if len(assessments) <= 0 {
		log.Error(ctx, "assessment not found", log.Any("condition", condition))
		return nil, constant.ErrRecordNotFound
	}
	return assessments[0], nil
}

func (a *assessmentInternalModel) AnyoneAttemptedByScheduleIDs(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string]*v2.AssessmentAnyoneAttemptedReply, error) {
	var assessments []*v2.Assessment
	condition := &assessmentV2.AssessmentCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	}
	err := assessmentV2.GetAssessmentDA().Query(ctx, condition, &assessments)
	if err != nil {
		log.Error(ctx, "query assessment error", log.Err(err), log.Any("condition", condition), log.Any("op", op))
		return nil, err
	}
	assessmentIDs := make([]string, len(assessments))
	for i, item := range assessments {
		assessmentIDs[i] = item.ID
	}

	result := make(map[string]*v2.AssessmentAnyoneAttemptedReply, len(assessments))
	for _, item := range assessments {
		resultItem := &v2.AssessmentAnyoneAttemptedReply{
			AssessmentStatus: item.Status,
		}

		result[item.ScheduleID] = resultItem
	}

	return result, nil
}

func (a *assessmentInternalModel) Query(ctx context.Context, op *entity.Operator, condition *assessmentV2.AssessmentCondition) ([]*v2.Assessment, error) {
	var assessments []*v2.Assessment
	err := assessmentV2.GetAssessmentDA().Query(ctx, condition, &assessments)
	if err != nil {
		log.Error(ctx, "query assessment error", log.Err(err), log.Any("condition", condition), log.Any("op", op))
		return nil, err
	}

	return assessments, nil
}

func (a *assessmentInternalModel) UpdateWhenReviewScheduleSuccess(ctx context.Context, tx *dbo.DBContext, scheduleID string) error {
	assessment, err := a.getAssessmentByScheduleID(ctx, scheduleID)
	if err != nil {
		return err
	}

	if assessment.AssessmentType != v2.AssessmentTypeReviewStudy || assessment.Status != v2.AssessmentStatusPending {
		log.Warn(ctx, "assessment is not review study or sleep status", log.Any("assessment", assessment))
		return nil
	}

	assessment.Status = v2.AssessmentStatusNotStarted
	// update create time when schedule ready
	assessment.CreateAt = time.Now().Unix()

	_, err = assessmentV2.GetAssessmentDA().UpdateTx(ctx, tx, assessment)
	if err != nil {
		log.Error(ctx, "update assessment error", log.Any("assessment", assessment))
		return err
	}

	return nil
}

func GetAssessmentInternalModel() IAssessmentInternalModelV2 {
	assessmentInternalModelInstanceOnce.Do(func() {
		assessmentInternalModelInstance = &assessmentInternalModel{
			AmsServices: external.GetAmsServices(),
		}
	})
	return assessmentInternalModelInstance
}
