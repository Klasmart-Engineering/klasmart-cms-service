package model

import (
	"context"
	"database/sql"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sort"
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
	schedules, err := l.findRelatedSchedules(ctx, tx, operator, nil, filter)
	if err != nil {
		log.Error(ctx, "query subject filter items: query schedule failed",
			log.Err(err),
			log.Any("filter", filter),
		)
		return nil, err
	}
	scheduleIDs := make([]string, 0, len(schedules))
	for _, s := range schedules {
		scheduleIDs = append(scheduleIDs, s.ID)
	}

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
	// find related schedules and make by schedule id
	schedules, err := l.findRelatedSchedules(ctx, tx, operator, []entity.AssessmentType{entity.AssessmentTypeLive}, filter)
	if err != nil {
		return nil, err
	}

	// find related assessments and make map by schedule id
	scheduleIDs := make([]string, 0, len(schedules))
	for _, s := range schedules {
		scheduleIDs = append(scheduleIDs, s.ID)
	}
	assessments, err := l.findRelatedAssessments(ctx, tx, operator, scheduleIDs, filter.StudentID)
	if err != nil {
		return nil, err
	}
	assessmentMap := make(map[string]*entity.Assessment, len(assessments))
	for _, a := range assessments {
		assessmentMap[a.ScheduleID] = a
	}

	// calculate student attend percent
	attend := 0.0
	if len(assessments) != 0 {
		attend = float64(len(assessments)) / float64(len(schedules))
	}

	// find related comments and make map by schedule id  (live: room comments)
	roomCommentMap, err := getAssessmentH5P().batchGetRoomCommentMap(ctx, operator, scheduleIDs)
	if err != nil {
		log.Error(ctx, "query live classes summary: batch get room comment map failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("filter", filter),
		)
		return nil, err
	}

	// find related lesson plan name map
	lessonPlanIDs := make([]string, 0, len(schedules))
	for _, s := range schedules {
		lessonPlanIDs = append(lessonPlanIDs, s.LessonPlanID)
	}
	lessonPlanNames, err := GetContentModel().GetContentNameByIDList(ctx, tx, lessonPlanIDs)
	if err != nil {
		log.Error(ctx, "query live classes summary: batch get content names failed",
			log.Err(err),
			log.Strings("lesson_plan_ids", lessonPlanIDs),
			log.Any("filter", filter),
		)
		return nil, err
	}
	lessonPlanNameMap := make(map[string]string, len(lessonPlanNames))
	for _, lp := range lessonPlanNames {
		lessonPlanNameMap[lp.ID] = lp.Name
	}

	// find related outcomes and make map by schedule ids
	scheduleOutcomesMap, err := l.findRelatedOutcomes(ctx, tx, operator, scheduleIDs)
	if err != nil {
		log.Error(ctx, "query live classes summary: find related outcomes failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("filter", filter),
		)
		return nil, err
	}

	//  assembly result
	result := &entity.QueryLiveClassesSummaryResult{Attend: attend}
	for _, s := range schedules {
		assessment := assessmentMap[s.ID]
		item := entity.LiveClassSummaryItem{
			ClassStartTime: s.StartAt,
			ScheduleTitle:  s.Title,
			LessonPlanName: lessonPlanNameMap[s.LessonPlanID],
			ScheduleID:     s.ID,
		}
		if assessment == nil {
			item.Absent = true
		} else {
			item.AssessmentID = assessment.ID
			item.Status = assessment.Status
			item.CompleteAt = assessment.CompleteTime
			item.CreateAt = assessment.CreateAt
		}
		if comments := roomCommentMap[s.ID][filter.StudentID]; len(comments) > 0 {
			item.TeacherFeedback = comments[len(comments)-1]
		}
		if outcomes := scheduleOutcomesMap[s.ID]; len(outcomes) > 0 {
			for _, o := range outcomes {
				item.Outcomes = append(item.Outcomes, &entity.LearningSummaryOutcome{
					ID:   o.ID,
					Name: o.Name,
				})
			}
		}
		l.sortOutcomesByAlphabetAsc(item.Outcomes)
		result.Items = append(result.Items, &item)
	}

	// sort items
	l.sortLiveClassesSummaryItemsByStartTimeAsc(result.Items)

	log.Debug(ctx, "query live classes summary result", log.Any("result", result))

	return result, nil
}

func (l *learningSummaryReportModel) sortOutcomesByAlphabetAsc(items []*entity.LearningSummaryOutcome) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
}

func (l *learningSummaryReportModel) sortLiveClassesSummaryItemsByStartTimeAsc(items []*entity.LiveClassSummaryItem) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].ClassStartTime < items[j].ClassStartTime
	})
}

func (l *learningSummaryReportModel) findRelatedOutcomes(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) (map[string][]*entity.Outcome, error) {
	scheduleOutcomeIDsMap, err := GetScheduleModel().GetLearningOutcomeIDs(ctx, operator, scheduleIDs)
	if err != nil {
		log.Error(ctx, "find related outcomes: get schedule outcome ids failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
		)
		return nil, err
	}
	outcomeIDs := make([]string, 0, len(scheduleOutcomeIDsMap))
	for _, scheduleOutcomeIDs := range scheduleOutcomeIDsMap {
		outcomeIDs = append(outcomeIDs, scheduleOutcomeIDs...)
	}
	outcomes, err := GetOutcomeModel().GetByIDs(ctx, operator, tx, outcomeIDs)
	if err != nil {
		log.Error(ctx, "find related outcomes: batch get schedule outcome failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
		)
		return nil, err
	}
	outcomeMap := make(map[string]*entity.Outcome, len(outcomes))
	for _, o := range outcomes {
		outcomeMap[o.ID] = o
	}
	scheduleOutcomesMap := make(map[string][]*entity.Outcome, len(scheduleOutcomeIDsMap))
	for scheduleID, outcomeIDs := range scheduleOutcomeIDsMap {
		for _, outcomeID := range outcomeIDs {
			scheduleOutcomesMap[scheduleID] = append(scheduleOutcomesMap[scheduleID], outcomeMap[outcomeID])
		}
	}
	return scheduleOutcomesMap, nil
}

func (l *learningSummaryReportModel) findRelatedSchedules(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, types []entity.AssessmentType, filter *entity.LearningSummaryFilter) ([]*entity.Schedule, error) {
	scheduleCondition := entity.ScheduleQueryCondition{
		OrgID: sql.NullString{
			String: operator.OrgID,
			Valid:  true,
		},
	}
	if len(types) > 0 {
		scheduleCondition.ClassTypes.Valid = true
		for _, t := range types {
			scheduleType := t.ToScheduleClassType()
			scheduleCondition.ClassTypes.Strings = append(scheduleCondition.ClassTypes.Strings, string(scheduleType.ClassType))
			if scheduleType.IsHomeFun {
				scheduleCondition.IsHomefun = sql.NullBool{
					Bool:  true,
					Valid: true,
				}
			}
		}
	}
	if filter.WeekStart > 0 {
		scheduleCondition.StartAtGe = sql.NullInt64{
			Int64: filter.WeekStart,
			Valid: true,
		}
	}
	if filter.WeekEnd > 0 {
		scheduleCondition.StartAtLt = sql.NullInt64{
			Int64: filter.WeekEnd,
			Valid: true,
		}
	}
	if len(filter.SchoolID) > 0 {
		scheduleCondition.RelationSchoolIDs = entity.NullStrings{
			Strings: []string{filter.SchoolID},
			Valid:   true,
		}
	}
	if len(filter.ClassID) > 0 {
		scheduleCondition.RelationClassIDs = entity.NullStrings{
			Strings: []string{filter.ClassID},
			Valid:   true,
		}
	}
	if len(filter.TeacherID) > 0 {
		scheduleCondition.RelationTeacherIDs = entity.NullStrings{
			Strings: []string{filter.TeacherID},
			Valid:   true,
		}
	}
	if len(filter.StudentID) > 0 {
		scheduleCondition.RelationStudentIDs = entity.NullStrings{
			Strings: []string{filter.StudentID},
			Valid:   true,
		}
	}
	schedules, err := GetScheduleModel().QueryUnsafe(ctx, &scheduleCondition)
	if err != nil {
		log.Error(ctx, "find related schedules: query schedule failed",
			log.Err(err),
			log.Any("schedule_condition", scheduleCondition),
		)
		return nil, err
	}
	return schedules, nil
}

func (l *learningSummaryReportModel) findRelatedAssessments(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string, studentID string) ([]*entity.Assessment, error) {
	// query assessments
	var assessments []*entity.Assessment
	if err := da.GetAssessmentDA().Query(ctx, &da.QueryAssessmentConditions{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
		StudentIDs: entity.NullStrings{
			Strings: []string{studentID},
			Valid:   true,
		},
	}, &assessments); err != nil {
		log.Error(ctx, "find related assessments: query assessment attendance relations failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.String("student_id", studentID),
		)
		return nil, err
	}
	return assessments, nil
}

func (l *learningSummaryReportModel) QueryAssignmentsSummary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (*entity.QueryAssignmentsSummaryResult, error) {
	// find related schedules and make by schedule id
	schedules, err := l.findRelatedSchedules(ctx, tx, operator, []entity.AssessmentType{entity.AssessmentTypeStudy, entity.AssessmentTypeHomeFunStudy}, filter)
	if err != nil {
		log.Error(ctx, "query assignments summary: find related schedules failed",
			log.Err(err),
			log.Any("filter", filter),
		)
		return nil, err
	}

	// find related study assessments and make map by schedule id
	scheduleIDs := make([]string, 0, len(schedules))
	for _, s := range schedules {
		scheduleIDs = append(scheduleIDs, s.ID)
	}
	studyAssessments, err := l.findRelatedAssessments(ctx, tx, operator, scheduleIDs, filter.StudentID)
	if err != nil {
		log.Error(ctx, "query assignments summary: find related assessments failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("filter", filter),
		)
		return nil, err
	}
	studyAssessmentMap := make(map[string]*entity.Assessment, len(studyAssessments))
	for _, a := range studyAssessments {
		studyAssessmentMap[a.ScheduleID] = a
	}

	// find related home fun study assessments and make map by schedule id
	var homeFunStudyAssessments []*entity.HomeFunStudy
	if err := GetHomeFunStudyModel().Query(ctx, operator, &da.QueryHomeFunStudyCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	}, &homeFunStudyAssessments); err != nil {
		log.Error(ctx, "query assignments summary: query home fun study failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
		)
		return nil, err
	}
	homeFunStudyAssessmentMap := make(map[string]*entity.HomeFunStudy, len(homeFunStudyAssessments))
	for _, a := range homeFunStudyAssessments {
		homeFunStudyAssessmentMap[a.ScheduleID] = a
	}

	// find related study assessments comments and make map by schedule id (live: room comments)
	roomCommentMap, err := getAssessmentH5P().batchGetRoomCommentMap(ctx, operator, scheduleIDs)
	if err != nil {
		log.Error(ctx, "query assignments summary: batch get room comment map failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("filter", filter),
		)
		return nil, err
	}

	// find related lesson plan and make name map
	lessonPlanIDs := make([]string, 0, len(schedules))
	for _, s := range schedules {
		lessonPlanIDs = append(lessonPlanIDs, s.LessonPlanID)
	}
	lessonPlanNames, err := GetContentModel().GetContentNameByIDList(ctx, tx, lessonPlanIDs)
	if err != nil {
		log.Error(ctx, "query assignments summary: batch get lesson plans failed",
			log.Err(err),
			log.Strings("lesson_plan_ids", lessonPlanIDs),
			log.Any("filter", filter),
		)
		return nil, err
	}
	lessonPlanNameMap := make(map[string]string, len(lessonPlanNames))
	for _, lp := range lessonPlanNames {
		lessonPlanNameMap[lp.ID] = lp.Name
	}

	// find related outcomes and make map by schedule ids
	scheduleOutcomesMap, err := l.findRelatedOutcomes(ctx, tx, operator, scheduleIDs)
	if err != nil {
		log.Error(ctx, "query assignments summary: find related outcomes failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("filter", filter),
		)
		return nil, err
	}

	// assembly result
	result := l.assemblyAssignmentsSummaryResult(filter, schedules, scheduleOutcomesMap, lessonPlanNameMap, studyAssessmentMap, homeFunStudyAssessmentMap, roomCommentMap)

	// sort items
	l.sortAssignmentsSummaryItems(result.Items)

	log.Debug(ctx, "query assignments summary result", log.Any("result", result))

	return result, nil
}

func (l *learningSummaryReportModel) assemblyAssignmentsSummaryResult(filter *entity.LearningSummaryFilter, schedules []*entity.Schedule, scheduleOutcomesMap map[string][]*entity.Outcome, lessonPlanNameMap map[string]string, studyAssessmentMap map[string]*entity.Assessment, homeFunStudyAssessmentMap map[string]*entity.HomeFunStudy, roomCommentMap map[string]map[string][]string) *entity.QueryAssignmentsSummaryResult {
	result := &entity.QueryAssignmentsSummaryResult{
		StudyCount:        len(studyAssessmentMap),
		HomeFunStudyCount: len(homeFunStudyAssessmentMap),
		Items:             nil,
	}
	for _, s := range schedules {
		if s.IsHomeFun {
			assessment := homeFunStudyAssessmentMap[s.ID]
			if assessment == nil {
				continue
			}
			item := entity.AssignmentsSummaryItem{
				Type:            entity.AssessmentTypeHomeFunStudy,
				Status:          assessment.Status,
				AssessmentTitle: assessment.Title,
				TeacherFeedback: assessment.AssessComment,
				ScheduleID:      s.ID,
				AssessmentID:    assessment.ID,
				CompleteAt:      assessment.CompleteAt,
				CreateAt:        assessment.CreateAt,
			}
			if outcomes := scheduleOutcomesMap[s.ID]; len(outcomes) > 0 {
				for _, o := range outcomes {
					item.Outcomes = append(item.Outcomes, &entity.LearningSummaryOutcome{
						ID:   o.ID,
						Name: o.Name,
					})
				}
			}
			result.Items = append(result.Items, &item)
		} else {
			assessment := studyAssessmentMap[s.ID]
			if assessment == nil {
				continue
			}
			item := entity.AssignmentsSummaryItem{
				Type:            entity.AssessmentTypeHomeFunStudy,
				Status:          assessment.Status,
				AssessmentTitle: assessment.Title,
				LessonPlanName:  lessonPlanNameMap[s.LessonPlanID],
				ScheduleID:      s.ID,
				AssessmentID:    assessment.ID,
				CompleteAt:      assessment.CompleteTime,
				CreateAt:        assessment.CreateAt,
			}
			if outcomes := scheduleOutcomesMap[s.ID]; len(outcomes) > 0 {
				for _, o := range outcomes {
					item.Outcomes = append(item.Outcomes, &entity.LearningSummaryOutcome{
						ID:   o.ID,
						Name: o.Name,
					})
				}
			}
			if comments := roomCommentMap[s.ID][filter.StudentID]; len(comments) > 0 {
				item.TeacherFeedback = comments[len(comments)-1]
			}
			result.Items = append(result.Items, &item)
		}
	}
	return result
}

func (l *learningSummaryReportModel) sortAssignmentsSummaryItems(items []*entity.AssignmentsSummaryItem) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].CompleteAt < items[j].CompleteAt
	})
}
