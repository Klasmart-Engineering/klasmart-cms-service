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
	schedules, err := l.findRelatedSchedules(ctx, tx, operator, entity.ReportLearningSummaryTypeInvalid, filter)
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
	schedules, err := l.findRelatedSchedules(ctx, tx, operator, entity.ReportLearningSummaryTypeLiveClass, filter)
	if err != nil {
		log.Error(ctx, "query live classes summary: find related schedules failed",
			log.Err(err),
			log.Any("filter", filter),
		)
		return nil, err
	}

	// find related assessments and make map by schedule id
	scheduleIDs := make([]string, 0, len(schedules))
	for _, s := range schedules {
		scheduleIDs = append(scheduleIDs, s.ID)
	}
	assessments, err := l.findRelatedAssessments(ctx, tx, operator, entity.ReportLearningSummaryTypeLiveClass, filter, scheduleIDs)
	if err != nil {
		log.Error(ctx, "query live classes summary: find related assessments failed",
			log.Err(err),
			log.Any("filter", filter),
		)
		return nil, err
	}
	assessmentMap := make(map[string]*entity.Assessment, len(assessments))
	for _, a := range assessments {
		assessmentMap[a.ScheduleID] = a
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
	assessmentIDs := make([]string, 0, len(assessments))
	for _, a := range assessments {
		assessmentIDs = append(assessmentIDs, a.ID)
	}
	assessmentIDToOutcomesMap, err := l.findRelatedAssessmentOutcomes(ctx, tx, operator, assessmentIDs)
	if err != nil {
		log.Error(ctx, "query live classes summary: find related outcomes failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("filter", filter),
		)
		return nil, err
	}

	//  assembly result
	result := &entity.QueryLiveClassesSummaryResult{}
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
			if outcomes := assessmentIDToOutcomesMap[assessment.ID]; len(outcomes) > 0 {
				for _, o := range outcomes {
					item.Outcomes = append(item.Outcomes, &entity.LearningSummaryOutcome{
						ID:   o.ID,
						Name: o.Name,
					})
				}
				l.sortOutcomesByAlphabetAsc(item.Outcomes)
			}
		}
		if comments := roomCommentMap[s.ID][filter.StudentID]; len(comments) > 0 {
			item.TeacherFeedback = comments[len(comments)-1]
		}
		result.Items = append(result.Items, &item)
	}

	// calculate student attend percent
	attend := 0
	for _, item := range result.Items {
		if !item.Absent {
			attend++
		}
	}
	if len(result.Items) > 0 {
		result.Attend = float64(attend) / float64(len(result.Items))
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

func (l *learningSummaryReportModel) findRelatedAssessmentOutcomes(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessmentIDs []string) (map[string][]*entity.Outcome, error) {
	var assessmentOutcomes []*entity.AssessmentOutcome
	if err := da.GetAssessmentOutcomeDA().Query(ctx, &da.QueryAssessmentOutcomeConditions{
		AssessmentIDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   true,
		},
		Checked: entity.NullBool{
			Bool:  true,
			Valid: true,
		},
	}, &assessmentOutcomes); err != nil {
		log.Error(ctx, "find related assessment outcomes: query assessment outcomes failed",
			log.Err(err),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}
	outcomeIDs := make([]string, 0, len(assessmentOutcomes))
	for _, o := range assessmentOutcomes {
		outcomeIDs = append(outcomeIDs, o.OutcomeID)
	}
	outcomes, err := GetOutcomeModel().GetByIDs(ctx, operator, tx, outcomeIDs)
	if err != nil {
		log.Error(ctx, "find related assessment outcomes: batch get schedule outcome failed",
			log.Err(err),
			log.Strings("outcome_ids", outcomeIDs),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}
	outcomeMap := make(map[string]*entity.Outcome, len(outcomes))
	for _, o := range outcomes {
		outcomeMap[o.ID] = o
	}
	assessmentIDToOutcomesMap := make(map[string][]*entity.Outcome, len(assessmentOutcomes))
	for _, ao := range assessmentOutcomes {
		o := outcomeMap[ao.OutcomeID]
		if o == nil {
			continue
		}
		assessmentIDToOutcomesMap[ao.AssessmentID] = append(assessmentIDToOutcomesMap[ao.AssessmentID], o)
	}
	return assessmentIDToOutcomesMap, nil
}

func (l *learningSummaryReportModel) findRelatedHomeFunStudyOutcomes(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, homeFunStudies []*entity.HomeFunStudy) (map[string][]*entity.Outcome, error) {
	homeFunStudyScheduleIDs := make([]string, 0, len(homeFunStudies))
	scheduleIDToHomeFunStudyMap := make(map[string]*entity.HomeFunStudy, len(homeFunStudies))
	for _, s := range homeFunStudies {
		homeFunStudyScheduleIDs = append(homeFunStudyScheduleIDs, s.ScheduleID)
		scheduleIDToHomeFunStudyMap[s.ScheduleID] = s
	}
	scheduleIDToOutcomeIDsMap, err := GetScheduleModel().GetLearningOutcomeIDs(ctx, operator, homeFunStudyScheduleIDs)
	if err != nil {
		log.Error(ctx, "find related home fun study outcomes: get learning outcome ids failed",
			log.Err(err),
			log.Any("home_fun_studies", homeFunStudies),
		)
		return nil, err
	}
	var allOutcomeIDs []string
	for _, outcomeIDs := range scheduleIDToOutcomeIDsMap {
		allOutcomeIDs = append(allOutcomeIDs, outcomeIDs...)
	}
	outcomes, err := GetOutcomeModel().GetByIDs(ctx, operator, tx, allOutcomeIDs)
	if err != nil {
		log.Error(ctx, "find related home fun study outcomes: get learning outcome ids failed",
			log.Err(err),
			log.Any("home_fun_studies", homeFunStudies),
			log.Strings("home_fun_study_schedule_ids", homeFunStudyScheduleIDs),
		)
		return nil, err
	}
	outcomeMap := make(map[string]*entity.Outcome, len(outcomes))
	for _, o := range outcomes {
		outcomeMap[o.ID] = o
	}
	scheduleIDToOutcomesMap := make(map[string][]*entity.Outcome, len(scheduleIDToOutcomeIDsMap))
	for scheduleID, outcomeIDs := range scheduleIDToOutcomeIDsMap {
		for _, outcomeID := range outcomeIDs {
			scheduleIDToOutcomesMap[scheduleID] = append(scheduleIDToOutcomesMap[scheduleID], outcomeMap[outcomeID])
		}
	}
	assessmentIDToOutcomesMap := make(map[string][]*entity.Outcome, len(homeFunStudies))
	for _, s := range homeFunStudies {
		assessmentIDToOutcomesMap[s.ID] = scheduleIDToOutcomesMap[s.ScheduleID]
	}
	return nil, nil
}

func (l *learningSummaryReportModel) findRelatedSchedules(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, typo entity.ReportLearningSummaryType, filter *entity.LearningSummaryFilter) ([]*entity.Schedule, error) {
	scheduleCondition := entity.ScheduleQueryCondition{
		OrgID: sql.NullString{
			String: operator.OrgID,
			Valid:  true,
		},
	}
	if typo.Valid() {
		scheduleCondition.ClassTypes.Valid = true
		switch typo {
		case entity.ReportLearningSummaryTypeLiveClass:
			scheduleCondition.ClassTypes.Strings = append(scheduleCondition.ClassTypes.Strings, string(entity.ScheduleClassTypeOnlineClass))
		case entity.ReportLearningSummaryTypeAssignment:
			scheduleCondition.ClassTypes.Strings = append(scheduleCondition.ClassTypes.Strings, string(entity.ScheduleClassTypeHomework))
		}
	}
	if typo.Valid() && typo == entity.ReportLearningSummaryTypeLiveClass {
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
	if len(filter.SubjectID) > 0 {
		// TODO: Medivh: filter subject
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

func (l *learningSummaryReportModel) findRelatedAssessments(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, typo entity.ReportLearningSummaryType, filter *entity.LearningSummaryFilter, scheduleIDs []string) ([]*entity.Assessment, error) {
	// query assessments
	var assessments []*entity.Assessment
	cond := da.QueryAssessmentConditions{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
		StudentIDs: entity.NullStrings{
			Strings: []string{filter.StudentID},
			Valid:   true,
		},
	}
	if typo == entity.ReportLearningSummaryTypeAssignment && filter.WeekStart > 0 && filter.WeekEnd > 0 {
		cond.CompleteBetween = entity.NullTimeRange{
			StartAt: filter.WeekStart,
			EndAt:   filter.WeekEnd,
			Valid:   true,
		}
	}
	if err := da.GetAssessmentDA().Query(ctx, &cond, &assessments); err != nil {
		log.Error(ctx, "find related assessments: query assessment attendance relations failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("filter", filter),
		)
		return nil, err
	}
	return assessments, nil
}

func (l *learningSummaryReportModel) QueryAssignmentsSummary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (*entity.QueryAssignmentsSummaryResult, error) {
	// find related schedules and make by schedule id
	schedules, err := l.findRelatedSchedules(ctx, tx, operator, entity.ReportLearningSummaryTypeAssignment, filter)
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
	studyAssessments, err := l.findRelatedAssessments(ctx, tx, operator, entity.ReportLearningSummaryTypeAssignment, filter, scheduleIDs)
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
	cond := da.QueryHomeFunStudyCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	}
	if filter.WeekStart > 0 && filter.WeekEnd > 0 {
		cond.CompleteBetween = entity.NullTimeRange{
			StartAt: filter.WeekStart,
			EndAt:   filter.WeekEnd,
			Valid:   true,
		}
	}
	if err := GetHomeFunStudyModel().Query(ctx, operator, &cond, &homeFunStudyAssessments); err != nil {
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
	assessmentIDs := make([]string, 0, len(studyAssessments)+len(homeFunStudyAssessments))
	for _, a := range studyAssessments {
		assessmentIDs = append(assessmentIDs, a.ID)
	}
	for _, a := range homeFunStudyAssessments {
		assessmentIDs = append(assessmentIDs, a.ID)
	}
	assessmentIDToOutcomesMap, err := l.findRelatedAssessmentOutcomes(ctx, tx, operator, assessmentIDs)
	if err != nil {
		log.Error(ctx, "query assignments summary: find related outcomes failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("filter", filter),
		)
		return nil, err
	}
	if len(homeFunStudyAssessments) > 0 {
		homeFunStudyAssessmentIDToOutcomesMap, err := l.findRelatedHomeFunStudyOutcomes(ctx, tx, operator, homeFunStudyAssessments)
		if err != nil {
			log.Error(ctx, "query assignments summary: find related home fun study outcomes failed",
				log.Err(err),
				log.Strings("schedule_ids", scheduleIDs),
				log.Any("filter", filter),
				log.Any("home_fun_studies", homeFunStudyAssessments),
			)
			return nil, err
		}
		for assessmentID, outcomes := range homeFunStudyAssessmentIDToOutcomesMap {
			assessmentIDToOutcomesMap[assessmentID] = outcomes
		}
	}

	// assembly result
	result := l.assemblyAssignmentsSummaryResult(filter, schedules, assessmentIDToOutcomesMap, lessonPlanNameMap, studyAssessmentMap, homeFunStudyAssessmentMap, roomCommentMap)

	// sort items
	l.sortAssignmentsSummaryItems(result.Items)

	log.Debug(ctx, "query assignments summary result", log.Any("result", result))

	return result, nil
}

func (l *learningSummaryReportModel) assemblyAssignmentsSummaryResult(filter *entity.LearningSummaryFilter, schedules []*entity.Schedule, assessmentIDToOutcomesMap map[string][]*entity.Outcome, lessonPlanNameMap map[string]string, studyAssessmentMap map[string]*entity.Assessment, homeFunStudyAssessmentMap map[string]*entity.HomeFunStudy, roomCommentMap map[string]map[string][]string) *entity.QueryAssignmentsSummaryResult {
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
			if outcomes := assessmentIDToOutcomesMap[assessment.ID]; len(outcomes) > 0 {
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
			if outcomes := assessmentIDToOutcomesMap[assessment.ID]; len(outcomes) > 0 {
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
