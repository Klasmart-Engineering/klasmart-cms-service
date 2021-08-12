package model

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sort"
	"sync"
	"time"
)

type ILearningSummaryReportModel interface {
	QueryTimeFilter(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.QueryLearningSummaryTimeFilterArgs) ([]*entity.LearningSummaryFilterYear, error)
	QueryRemainingFilter(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.QueryLearningSummaryRemainingFilterArgs) ([]*entity.QueryLearningSummaryRemainingFilterResultItem, error)
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

type learningSummaryReportModel struct {
	assessmentBase
}

func (l *learningSummaryReportModel) QueryTimeFilter(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.QueryLearningSummaryTimeFilterArgs) ([]*entity.LearningSummaryFilterYear, error) {
	// try get cache
	cacheResult, err := da.GetAssessmentRedisDA().GetQueryLearningSummaryTimeFilterResult(ctx, args)
	if err != nil {
		log.Debug(ctx, "query time filter: get cache failed",
			log.Err(err),
			log.Any("args", args),
		)
	} else {
		log.Debug(ctx, "query time filter: hit cache",
			log.Err(err),
			log.Any("args", args),
			log.Any("cache_result", cacheResult),
		)
		return cacheResult, nil
	}

	fixedZone := time.FixedZone("time_filter", args.TimeOffset)
	var result []*entity.LearningSummaryFilterYear

	m := make(map[int][][2]int64)
	switch args.SummaryType {
	case entity.LearningSummaryTypeLiveClass:
		schedules, err := l.findRelatedSchedules(ctx, tx, operator, entity.LearningSummaryTypeLiveClass, &entity.LearningSummaryFilter{})
		if err != nil {
			log.Error(ctx, "query time filter: find related schedules failed",
				log.Err(err),
				log.Any("args", args),
			)
			return nil, err
		}
		for _, s := range schedules {
			year := time.Unix(s.StartAt, 0).Year()
			weekStart, weekEnd := utils.FindWeekTimeRange(s.StartAt, fixedZone)
			m[year] = append(m[year], [2]int64{weekStart, weekEnd})
		}
	case entity.LearningSummaryTypeAssignment:
		assessments, err := l.queryUnifiedAssessments(ctx, tx, operator, &entity.QueryUnifiedAssessmentArgs{
			Types: entity.NullAssessmentTypes{
				Value: []entity.AssessmentType{entity.AssessmentTypeStudy, entity.AssessmentTypeHomeFunStudy},
				Valid: true,
			},
			Status: entity.NullAssessmentStatus{
				Value: entity.AssessmentStatusComplete,
				Valid: true,
			},
			OrgID: entity.NullString{
				String: operator.OrgID,
				Valid:  true,
			},
		})
		if err != nil {
			log.Error(ctx, "query time filter: query unified assessments failed",
				log.Err(err),
				log.Any("args", args),
			)
			return nil, err
		}
		for _, a := range assessments {
			year := time.Unix(a.CompleteTime, 0).Year()
			weekStart, weekEnd := utils.FindWeekTimeRange(a.CompleteTime, fixedZone)
			m[year] = append(m[year], [2]int64{weekStart, weekEnd})
		}
	}

	// calc current week
	currentWeekStart, currentWeekEnd := utils.FindWeekTimeRange(time.Now().Unix(), fixedZone)

	// fill result
	for year, weeks := range m {
		item := entity.LearningSummaryFilterYear{Year: year}
		weeks = l.deduplicationAndSortWeeks(weeks)
		for _, w := range weeks {
			if w[0] == currentWeekStart && w[1] == currentWeekEnd {
				continue
			}
			item.Weeks = append(item.Weeks, entity.LearningSummaryFilterWeek{
				WeekStart: w[0],
				WeekEnd:   w[1],
			})
		}
		if len(item.Weeks) == 0 {
			continue
		}
		result = append(result, &item)
	}

	// sort result
	sort.Slice(result, func(i, j int) bool {
		return result[i].Year < result[j].Year
	})

	// try set cache
	if err := da.GetAssessmentRedisDA().SetQueryLearningSummaryTimeFilterResult(ctx, args, result); err != nil {
		log.Debug(ctx, "query learning summary time filter: set cache failed",
			log.Err(err),
			log.Any("args", args),
			log.Any("result", result),
		)
	}

	return result, nil
}

func (l *learningSummaryReportModel) deduplicationAndSortWeeks(weeks [][2]int64) [][2]int64 {
	deduplicationMap := make(map[[2]int64]struct{}, len(weeks))
	deduplicationItems := make([][2]int64, 0, len(weeks))
	for _, w := range weeks {
		if _, ok := deduplicationMap[w]; !ok {
			deduplicationMap[w] = struct{}{}
			deduplicationItems = append(deduplicationItems, w)
		}
	}
	sort.Slice(deduplicationItems, func(i, j int) bool {
		return deduplicationItems[i][0] < deduplicationItems[j][0]
	})
	return deduplicationItems
}

func (l *learningSummaryReportModel) QueryRemainingFilter(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.QueryLearningSummaryRemainingFilterArgs) ([]*entity.QueryLearningSummaryRemainingFilterResultItem, error) {
	schedules, err := l.findRelatedSchedules(ctx, tx, operator, args.SummaryType, &args.LearningSummaryFilter)
	if err != nil {
		log.Error(ctx, "query remaining filter school failed: find related schedules",
			log.Err(err),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}
	scheduleIDs := make([]string, 0, len(schedules))
	for _, s := range schedules {
		scheduleIDs = append(scheduleIDs, s.ID)
	}
	switch args.FilterType {
	case entity.LearningSummaryFilterTypeSchool:
		return l.queryRemainingFilterSchool(ctx, tx, operator, scheduleIDs)
	case entity.LearningSummaryFilterTypeClass:
		return l.queryRemainingFilterClass(ctx, tx, operator, scheduleIDs)
	case entity.LearningSummaryFilterTypeTeacher:
		return l.queryRemainingFilterTeacher(ctx, tx, operator, scheduleIDs)
	case entity.LearningSummaryFilterTypeStudent:
		return l.queryRemainingFilterStudent(ctx, tx, operator, scheduleIDs)
	case entity.LearningSummaryFilterTypeSubject:
		return l.queryRemainingFilterSubject(ctx, tx, operator, scheduleIDs)
	default:
		log.Error(ctx, "query remaining filter: invalid filter type")
		return nil, constant.ErrInvalidArgs
	}
}

func (l *learningSummaryReportModel) queryRemainingFilterSchool(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) ([]*entity.QueryLearningSummaryRemainingFilterResultItem, error) {
	schoolIDs, err := l.batchGetScheduleRelationIDs(ctx, operator, scheduleIDs, []entity.ScheduleRelationType{entity.ScheduleRelationTypeSchool})
	if err != nil {
		log.Error(ctx, "query remaining filter school failed: batch get school relations failed",
			log.Err(err),
			log.Any("operator", operator),
		)
		return nil, err
	}
	schoolNameMap, err := external.GetSchoolServiceProvider().BatchGetNameMap(ctx, operator, schoolIDs)
	if err != nil {
		log.Error(ctx, "query remaining filter school failed: batch get school name map failed",
			log.Err(err),
			log.Strings("school_ids", schoolIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	result := make([]*entity.QueryLearningSummaryRemainingFilterResultItem, 0, len(schoolIDs))
	for _, schoolID := range schoolIDs {
		result = append(result, &entity.QueryLearningSummaryRemainingFilterResultItem{
			SchoolID:   schoolID,
			SchoolName: schoolNameMap[schoolID],
		})
	}
	return result, nil
}

func (l *learningSummaryReportModel) queryRemainingFilterClass(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) ([]*entity.QueryLearningSummaryRemainingFilterResultItem, error) {
	classIDs, err := l.batchGetScheduleRelationIDs(ctx, operator, scheduleIDs, []entity.ScheduleRelationType{entity.ScheduleRelationTypeClassRosterClass, entity.ScheduleRelationTypeParticipantClass})
	if err != nil {
		log.Error(ctx, "query remaining filter class failed: batch get classes relations failed",
			log.Err(err),
			log.Any("operator", operator),
		)
		return nil, err
	}
	classNameMap, err := external.GetClassServiceProvider().BatchGetNameMap(ctx, operator, classIDs)
	if err != nil {
		log.Error(ctx, "query remaining filter class failed: batch get class name map failed",
			log.Err(err),
			log.Strings("class_ids", classIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	result := make([]*entity.QueryLearningSummaryRemainingFilterResultItem, 0, len(classIDs))
	for _, classID := range classIDs {
		result = append(result, &entity.QueryLearningSummaryRemainingFilterResultItem{
			ClassID:   classID,
			ClassName: classNameMap[classID],
		})
	}
	return result, nil
}

func (l *learningSummaryReportModel) queryRemainingFilterTeacher(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) ([]*entity.QueryLearningSummaryRemainingFilterResultItem, error) {
	teacherIDs, err := l.batchGetScheduleRelationIDs(ctx, operator, scheduleIDs, []entity.ScheduleRelationType{entity.ScheduleRelationTypeClassRosterTeacher, entity.ScheduleRelationTypeParticipantTeacher})
	if err != nil {
		log.Error(ctx, "query remaining filter teacher failed: batch get teachers relations failed",
			log.Err(err),
			log.Any("operator", operator),
		)
		return nil, err
	}
	teacherNameMap, err := external.GetTeacherServiceProvider().BatchGetNameMap(ctx, operator, teacherIDs)
	if err != nil {
		log.Error(ctx, "query remaining filter teacher failed: batch get teacher name map failed",
			log.Err(err),
			log.Strings("teacher_ids", teacherIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	result := make([]*entity.QueryLearningSummaryRemainingFilterResultItem, 0, len(teacherIDs))
	for _, teacherID := range teacherIDs {
		result = append(result, &entity.QueryLearningSummaryRemainingFilterResultItem{
			TeacherID:   teacherID,
			TeacherName: teacherNameMap[teacherID],
		})
	}
	return result, nil
}

func (l *learningSummaryReportModel) queryRemainingFilterStudent(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) ([]*entity.QueryLearningSummaryRemainingFilterResultItem, error) {
	studentIDs, err := l.batchGetScheduleRelationIDs(ctx, operator, scheduleIDs, []entity.ScheduleRelationType{entity.ScheduleRelationTypeClassRosterStudent, entity.ScheduleRelationTypeParticipantStudent})
	if err != nil {
		log.Error(ctx, "query remaining filter student failed: batch get students relations failed",
			log.Err(err),
			log.Any("operator", operator),
		)
		return nil, err
	}
	studentNameMap, err := external.GetStudentServiceProvider().BatchGetNameMap(ctx, operator, studentIDs)
	if err != nil {
		log.Error(ctx, "query remaining filter student failed: batch get student name map failed",
			log.Err(err),
			log.Strings("student_ids", studentIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	result := make([]*entity.QueryLearningSummaryRemainingFilterResultItem, 0, len(studentIDs))
	for _, studentID := range studentIDs {
		result = append(result, &entity.QueryLearningSummaryRemainingFilterResultItem{
			StudentID:   studentID,
			StudentName: studentNameMap[studentID],
		})
	}
	return result, nil
}

func (l *learningSummaryReportModel) queryRemainingFilterSubject(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) ([]*entity.QueryLearningSummaryRemainingFilterResultItem, error) {
	subjectIDs, err := l.batchGetScheduleRelationIDs(ctx, operator, scheduleIDs, []entity.ScheduleRelationType{entity.ScheduleRelationTypeSubject})
	if err != nil {
		log.Error(ctx, "query remaining filter student failed: batch get students relations failed",
			log.Err(err),
			log.Any("operator", operator),
		)
		return nil, err
	}
	subjectNameMap, err := external.GetSubjectServiceProvider().BatchGetNameMap(ctx, operator, subjectIDs)
	if err != nil {
		log.Error(ctx, "query remaining filter student failed: batch get student name map failed",
			log.Err(err),
			log.Strings("subject_ids", subjectIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	result := make([]*entity.QueryLearningSummaryRemainingFilterResultItem, 0, len(subjectIDs))
	for _, subjectID := range subjectIDs {
		result = append(result, &entity.QueryLearningSummaryRemainingFilterResultItem{
			SubjectID:   subjectID,
			SubjectName: subjectNameMap[subjectID],
		})
	}
	return result, nil
}

func (l *learningSummaryReportModel) batchGetScheduleRelationIDs(ctx context.Context, operator *entity.Operator, scheduleIDs []string, relationTypes []entity.ScheduleRelationType) ([]string, error) {
	cond := da.ScheduleRelationCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
		RelationTypes: entity.NullStrings{
			Strings: []string{
				string(entity.ScheduleRelationTypeSchool),
			},
			Valid: true,
		},
	}
	if len(relationTypes) > 0 {
		cond.RelationTypes.Valid = true
		for _, r := range relationTypes {
			cond.RelationTypes.Strings = append(cond.RelationTypes.Strings, string(r))
		}
	}
	relations, err := GetScheduleRelationModel().Query(ctx, operator, &cond)
	if err != nil {
		log.Error(ctx, "batch get schedule relation ids: query schedule relations failed",
			log.Err(err),
			log.Any("schedule_ids", scheduleIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	relationIDs := make([]string, 0, len(relations))
	for _, s := range relations {
		relationIDs = append(relationIDs, s.RelationID)
	}
	relationIDs = utils.SliceDeduplication(relationIDs)
	return relationIDs, nil
}

func (l *learningSummaryReportModel) QueryLiveClassesSummary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (*entity.QueryLiveClassesSummaryResult, error) {
	// find related schedules and make by schedule id
	schedules, err := l.findRelatedSchedules(ctx, tx, operator, entity.LearningSummaryTypeLiveClass, filter)
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
	assessments, err := l.findRelatedAssessments(ctx, tx, operator, entity.LearningSummaryTypeLiveClass, filter, scheduleIDs)
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

func (l *learningSummaryReportModel) findRelatedSchedules(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, typo entity.LearningSummaryType, filter *entity.LearningSummaryFilter) ([]*entity.Schedule, error) {
	scheduleCondition := entity.ScheduleQueryCondition{
		OrgID: sql.NullString{
			String: operator.OrgID,
			Valid:  true,
		},
	}
	if typo.Valid() {
		scheduleCondition.ClassTypes.Valid = true
		switch typo {
		case entity.LearningSummaryTypeLiveClass:
			scheduleCondition.ClassTypes.Strings = append(scheduleCondition.ClassTypes.Strings, string(entity.ScheduleClassTypeOnlineClass))
		case entity.LearningSummaryTypeAssignment:
			scheduleCondition.ClassTypes.Strings = append(scheduleCondition.ClassTypes.Strings, string(entity.ScheduleClassTypeHomework))
		}
	}
	if typo.Valid() && typo == entity.LearningSummaryTypeLiveClass {
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
		scheduleCondition.RelationSubjectIDs = entity.NullStrings{
			Strings: []string{filter.SubjectID},
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

	// add related assessment filter
	if typo.Valid() && typo == entity.LearningSummaryTypeAssignment &&
		filter.WeekStart > 0 && filter.WeekEnd > 0 {
		scheduleIDs := make([]string, 0, len(schedules))
		cond := entity.QueryUnifiedAssessmentArgs{
			Types: entity.NullAssessmentTypes{
				Value: []entity.AssessmentType{entity.AssessmentTypeStudy, entity.AssessmentTypeHomeFunStudy},
				Valid: true,
			},
			Status: entity.NullAssessmentStatus{
				Value: entity.AssessmentStatusComplete,
				Valid: true,
			},
			OrgID: entity.NullString{
				String: operator.OrgID,
				Valid:  true,
			},
			ScheduleIDs: entity.NullStrings{
				Strings: scheduleIDs,
				Valid:   true,
			},
			CompleteBetween: entity.NullTimeRange{
				StartAt: filter.WeekStart,
				EndAt:   filter.WeekEnd,
				Valid:   true,
			},
		}
		assessments, err := l.queryUnifiedAssessments(ctx, tx, operator, &cond)
		if err != nil {
			log.Error(ctx, "find related schedules: query unified assessments failed",
				log.Err(err),
				log.Any("cond", cond),
				log.Any("type", typo),
				log.Any("filter", filter),
			)
			return nil, err
		}
		filterScheduleIDs := make([]string, 0, len(assessments))
		for _, a := range assessments {
			filterScheduleIDs = append(filterScheduleIDs, a.ScheduleID)
		}
		filterSchedules := make([]*entity.Schedule, 0, len(schedules))
		for _, s := range filterSchedules {
			if utils.ContainsStr(filterScheduleIDs, s.ID) {
				filterSchedules = append(filterSchedules, s)
			}
		}
		schedules = filterSchedules
	}
	return schedules, nil
}

func (l *learningSummaryReportModel) findRelatedAssessments(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, typo entity.LearningSummaryType, filter *entity.LearningSummaryFilter, scheduleIDs []string) ([]*entity.Assessment, error) {
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
	if typo == entity.LearningSummaryTypeAssignment && filter.WeekStart > 0 && filter.WeekEnd > 0 {
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
	schedules, err := l.findRelatedSchedules(ctx, tx, operator, entity.LearningSummaryTypeAssignment, filter)
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
	studyAssessments, err := l.findRelatedAssessments(ctx, tx, operator, entity.LearningSummaryTypeAssignment, filter, scheduleIDs)
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
		StudentIDs: entity.NullStrings{
			Strings: []string{filter.StudentID},
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
				Type:            entity.AssessmentTypeStudy,
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
