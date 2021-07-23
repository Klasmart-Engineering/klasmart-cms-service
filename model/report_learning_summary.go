package model

import (
	"context"
	"database/sql"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ILearningSummaryReportModel interface {
	QueryFilterItems(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.QueryLearningSummaryFilterItemsArgs) ([]*entity.QueryLearningSummaryFilterResultItem, error)
	QueryLiveClassesSummary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (*entity.QueryLiveClassesSummaryResult, error)
	QueryAssignmentsSummary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (*entity.QueryAssignmentsSummaryResult, error)
}

var (
	learningSummaryReportModelInstance ILearningSummaryReportModel
	learningSummaryReportModelOnce     = sync.Once{}
)

func GetLearningSummaryReportModel() ILearningSummaryReportModel {
	learningSummaryReportModelOnce.Do(func() {
		learningSummaryReportModelInstance = &learningSummaryReportModel{}
	})
	return learningSummaryReportModelInstance
}

type learningSummaryReportModel struct{}

func (l *learningSummaryReportModel) QueryFilterItems(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.QueryLearningSummaryFilterItemsArgs) ([]*entity.QueryLearningSummaryFilterResultItem, error) {
	// check type equal subject, now only support subject
	if args.Type != entity.LearningSummaryFilterTypeSubject {
		log.Error(ctx, "query filter items: unsupported filter type in current phase", log.Any("args", args))
		return nil, errors.New("unsupported filter type in current phase")
	}
	// query subject filter items
	return l.querySubjectFilterItems(ctx, tx, operator, args.LearningSummaryFilter)
}

func (l *learningSummaryReportModel) querySubjectFilterItems(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) ([]*entity.QueryLearningSummaryFilterResultItem, error) {
	// TODO: Medivh: find related schedule ids by student id
	var scheduleIDs []string

	// batch find subject ids by schedules
	relations, err := GetScheduleRelationModel().Query(ctx, operator, &da.ScheduleRelationCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
		RelationType: sql.NullString{
			String: string(entity.ScheduleRelationTypeSubject),
			Valid:  true,
		},
	})
	subjectIDs := make([]string, 0, len(relations))
	for _, r := range relations {
		subjectIDs = append(subjectIDs, r.RelationID)
	}
	subjectIDs = utils.SliceDeduplicationExcludeEmpty(subjectIDs)

	// batch get subjects
	subjects, err := external.GetSubjectServiceProvider().BatchGet(ctx, operator, subjectIDs)
	if err != nil {
		log.Error(ctx, "query subject filter items: batch get subject failed",
			log.Err(err),
			log.Strings("subject_ids", subjectIDs),
			log.Any("filter", filter),
		)
		return nil, err
	}

	// assembly result
	result := make([]*entity.QueryLearningSummaryFilterResultItem, 0, len(subjects))
	for _, s := range subjects {
		if s.Status == external.Inactive {
			continue
		}
		result = append(result, &entity.QueryLearningSummaryFilterResultItem{
			SubjectID:   s.ID,
			SubjectName: s.Name,
		})
	}

	return result, nil
}

func (l *learningSummaryReportModel) QueryLiveClassesSummary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (*entity.QueryLiveClassesSummaryResult, error) {
	// find related schedules
	schedules, err := l.findRelatedSchedules(ctx, tx, operator, filter)
	if err != nil {
		return nil, err
	}

	// find related assessments
	assessments, err := l.findRelatedAssessments(ctx, tx, operator, filter)
	if err != nil {
		return nil, err
	}

	// find related comments (live: room comments)

	// calculate student attend percent
	attend := 0.0
	if len(assessments) != 0 {
		attend = float64(len(schedules)) / float64(len(assessments))
	}

	//  result
	result := &entity.QueryLiveClassesSummaryResult{
		Attend: attend,
		Items:  nil,
	}
	log.Debug(ctx, "query live classes summary", log.Any("result", result))

	// query schedule and assessment aggregation items
	panic("implement me")
}

func (l *learningSummaryReportModel) findRelatedSchedules(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) ([]*entity.ScheduleVariable, error) {
	// find related schedule ids by student id
	studentID := filter.StudentID
	relations, err := GetScheduleRelationModel().Query(ctx, operator, &da.ScheduleRelationCondition{
		RelationID: sql.NullString{
			String: studentID,
			Valid:  true,
		},
		RelationTypes: entity.NullStrings{
			Strings: []string{
				string(entity.ScheduleRelationTypeClassRosterStudent),
				string(entity.ScheduleRelationTypeParticipantStudent),
			},
			Valid: true,
		},
	})
	if err != nil {
		log.Error(ctx, "query subject filter items: query schedule relations failed",
			log.Err(err),
			log.Any("filter", filter),
		)
		return nil, err
	}
	scheduleIDs := make([]string, 0, len(relations))
	for _, r := range relations {
		scheduleIDs = append(scheduleIDs, r.ScheduleID)
	}
	scheduleIDs = utils.SliceDeduplicationExcludeEmpty(scheduleIDs)

	// batch get schedules
	schedules, err := GetScheduleModel().GetVariableDataByIDs(ctx, operator, scheduleIDs, nil)
	if err != nil {
		log.Error(ctx, "query subject filter items: batch get schedules failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("filter", filter),
		)
		return nil, err
	}

	// filter schedule by org
	filterSchedules := make([]*entity.ScheduleVariable, 0, len(schedules))
	for _, s := range schedules {
		if s.OrgID != operator.OrgID {
			continue
		}
		filterSchedules = append(filterSchedules, s)
	}

	return filterSchedules, nil
}

func (l *learningSummaryReportModel) findRelatedAssessments(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) ([]*entity.Assessment, error) {
	// find assessment ids by student id
	var attendances []*entity.AssessmentAttendance
	if err := da.GetAssessmentAttendanceDA().Query(ctx, &da.QueryAssessmentAttendanceConditions{
		Role: entity.NullAssessmentAttendanceRole{
			Value: entity.AssessmentAttendanceRoleStudent,
			Valid: true,
		},
		AttendanceID: entity.NullString{
			String: filter.StudentID,
			Valid:  true,
		},
	}, &attendances); err != nil {
		log.Error(ctx, "find related assessments: query assessment attendance relations failed",
			log.Err(err),
			log.Any("filter", filter),
		)
		return nil, err
	}
	assessmentIDs := make([]string, 0, len(attendances))
	for _, a := range attendances {
		assessmentIDs = append(assessmentIDs, a.ID)
	}

	// batch get assessments
	var assessments []*entity.Assessment
	if err := da.GetAssessmentDA().Query(ctx, &da.QueryAssessmentConditions{
		IDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   true,
		},
		OrgID: entity.NullString{
			String: operator.OrgID,
			Valid:  true,
		},
	}, assessments); err != nil {
		log.Error(ctx, "find related assessments: query assessment attendance relations failed",
			log.Err(err),
			log.Any("filter", filter),
		)
		return nil, err
	}

	return assessments, nil
}

func (l *learningSummaryReportModel) calcStudentAttendPercent(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (float64, error) {
	panic("implement me")
}

func (l *learningSummaryReportModel) QueryAssignmentsSummary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (*entity.QueryAssignmentsSummaryResult, error) {
	panic("implement me")
}
