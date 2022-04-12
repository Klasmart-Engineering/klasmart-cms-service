package model

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type ILearningSummaryReportModel interface {
	QueryTimeFilter(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.QueryLearningSummaryTimeFilterArgs) ([]*entity.LearningSummaryFilterYear, error)
	QueryLiveClassesSummary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (*entity.QueryLiveClassesSummaryResult, error)
	QueryLiveClassesSummaryV2(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (res *entity.QueryLiveClassesSummaryResultV2, err error)
	QueryAssignmentsSummary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (*entity.QueryAssignmentsSummaryResult, error)
	QueryAssignmentsSummaryV2(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (res *entity.QueryAssignmentsSummaryResultV2, err error)
	QueryOutcomesByAssessmentID(ctx context.Context, op *entity.Operator, assessmentID string, studentID string) (res []*entity.LearningSummaryOutcome, err error)
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

func (l *learningSummaryReportModel) mustGetLearningSummaryReportPermissionMap(
	ctx context.Context, operator *entity.Operator) map[external.PermissionName]bool {
	permissions := []external.PermissionName{
		external.LearningSummaryReport,
		external.ReportLearningSummaryStudent,
		external.ReportLearningSummarySchool,
		external.ReportLearningSummaryTeacher,
		external.ReportLearningSummmaryOrg,
	}
	ret, err := external.GetPermissionServiceProvider().
		HasOrganizationPermissions(ctx, operator, permissions)
	if err != nil {
		logOperatorField := log.Any("operator", operator)
		permissionMap := log.Any("permissions", ret)
		log.Panic(ctx, "failed to query permissions",
			log.Err(err), logOperatorField, permissionMap)
	}
	return ret
}

// QueryTimeFilter returns years-weeks data for frontend under proper permission
// ref: https://calmisland.atlassian.net/wiki/spaces/NKL/pages/2331050001/Sprint+13+CMS+Report+Sep+15th+-+Oct+12th
// product owner requires this date below as the beginning in the drop-down box
// block 2, point 2: `‘Year’ is single choice, values includes all years from ‘2020’, default is current year.`
// and 2019-12-30 Monday is the beginning record
func (l *learningSummaryReportModel) QueryTimeFilter(
	ctx context.Context, tx *dbo.DBContext,
	operator *entity.Operator,
	args *entity.QueryLearningSummaryTimeFilterArgs) (ret []*entity.LearningSummaryFilterYear, err error) {

	// check permission
	permissionsShouldHave := []external.PermissionName{
		external.LearningSummaryReport,
		external.ReportLearningSummaryStudent,
		external.ReportLearningSummarySchool,
		external.ReportLearningSummaryTeacher,
		external.ReportLearningSummmaryOrg,
	}
	permissionMap := l.mustGetLearningSummaryReportPermissionMap(ctx, operator)
	logOperatorField := log.Any("operator", operator)
	logPermissionsShouldHaveField := log.Any("permissions", permissionsShouldHave)
	hitOne := false
	for _, p := range permissionMap {
		hitOne = hitOne || p
	}
	if !hitOne {
		log.Debug(ctx, "all permissions check failed",
			logOperatorField, logPermissionsShouldHaveField)
		return nil, constant.ErrForbidden
	}

	// make data
	// benchmark info
	// BenchmarkWeeks100y
	//	cpu: Intel(R) Core(TM) i5-8500 CPU @ 3.00GHz
	//	BenchmarkWeeks100y
	//  BenchmarkWeeks100y-6   	    1797	   2758555 ns/op
	// BenchmarkWeeks10y
	//	cpu: Intel(R) Core(TM) i5-8500 CPU @ 3.00GHz
	//	BenchmarkWeeks10y
	//	BenchmarkWeeks10y-6   	   10000	    304150 ns/op
	fixedZone := time.FixedZone("fixed-zone:"+strconv.Itoa(args.TimeOffset), args.TimeOffset)
	nowWithZone := time.Now().In(fixedZone)
	ret = l.getYearsWeeksData(nowWithZone)
	return
}

func (l *learningSummaryReportModel) getYearsWeeksData(nowWithZone time.Time) (ret []*entity.LearningSummaryFilterYear) {
	result := make(map[int][]entity.LearningSummaryFilterWeek)
	cursor := time.Date(2019, 12, 30, 0, 0, 0, 0, nowWithZone.Location())
	for {
		endCursor := cursor.Add(time.Hour * 24 * 7)
		if endCursor.After(nowWithZone) {
			break
		}
		endYear, _, _ := endCursor.Date()
		if _, ok := result[endYear]; !ok {
			result[endYear] = make([]entity.LearningSummaryFilterWeek, 0, 54)
		}
		result[endYear] = append(result[endYear], entity.LearningSummaryFilterWeek{
			WeekStart: cursor.Unix(),
			WeekEnd:   endCursor.Unix(),
		})
		cursor = endCursor
	}
	for year := nowWithZone.Year(); year >= 2020; year-- {
		// the first year in slice is the current year,
		// but the week may cross two years,
		// so the weekList is nil and should skip
		weekList, exist := result[year]
		if !exist {
			continue
		}
		utils.ReverseSliceInPlace(weekList)
		item := entity.LearningSummaryFilterYear{
			Year:  year,
			Weeks: weekList,
		}
		ret = append(ret, &item)
	}
	return
}

func (l *learningSummaryReportModel) QueryLiveClassesSummaryV2(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (res *entity.QueryLiveClassesSummaryResultV2, err error) {
	items, err := da.GetReportDA().QueryLiveClassesSummaryV2(ctx, tx, operator, filter)
	if err != nil {
		return
	}
	var scheduleIDs []string
	for _, item := range items {
		scheduleIDs = append(scheduleIDs, item.ScheduleID)
	}

	// find related comments and make map by schedule id  (live: room comments)
	roomCommentMap, err := getAssessmentH5P().batchGetRoomCommentMap(ctx, operator, scheduleIDs)
	if err != nil {
		log.Error(ctx, "query live classes summary: batch get room comment map failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("filter", filter),
		)
		return
	}
	for _, item := range items {
		comments := roomCommentMap[item.ScheduleID][filter.StudentID]
		if len(comments) > 0 {
			item.TeacherFeedback = comments[len(comments)-1]
		}
	}

	res = &entity.QueryLiveClassesSummaryResultV2{
		Attend: 0,
		Items:  items,
	}
	if len(items) > 0 {
		absentCount := 0
		for _, item := range items {
			if item.Absent {
				absentCount++
			}
		}
		res.Attend = (float64(len(items)) - float64(absentCount)) / float64(len(items))
	}

	return
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

	// batch get assessment outcome status
	var assessmentOutcomeKeys []*entity.AssessmentOutcomeKey
	for assessmentID, outcomes := range assessmentIDToOutcomesMap {
		for _, o := range outcomes {
			assessmentOutcomeKeys = append(assessmentOutcomeKeys, &entity.AssessmentOutcomeKey{
				AssessmentID: assessmentID,
				OutcomeID:    o.ID,
			})
		}
	}
	outcomeStatusMap, err := l.batchGetAssessmentOutcomeStatus(ctx, filter.StudentID, assessmentOutcomeKeys)
	if err != nil {
		log.Error(ctx, "query live classes summary: find related outcomes failed",
			log.Err(err),
			log.String("student_id", filter.StudentID),
			log.Any("keys", assessmentOutcomeKeys),
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
					status, ok := outcomeStatusMap[entity.AssessmentOutcomeKey{
						AssessmentID: assessment.ID,
						OutcomeID:    o.ID,
					}]
					if !ok {
						continue
					}
					item.Outcomes = append(item.Outcomes, &entity.LearningSummaryOutcome{
						ID:     o.ID,
						Name:   o.Name,
						Status: status,
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
	outcomeIDs = utils.SliceDeduplicationExcludeEmpty(outcomeIDs)
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
		// deduplication
		exists := false
		for _, o := range assessmentIDToOutcomesMap[ao.AssessmentID] {
			if ao.OutcomeID == o.ID {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
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
			scheduleCondition.ClassTypes.Strings = append(scheduleCondition.ClassTypes.Strings, entity.ScheduleClassTypeOnlineClass.String())
		case entity.LearningSummaryTypeAssignment:
			scheduleCondition.ClassTypes.Strings = append(scheduleCondition.ClassTypes.Strings, entity.ScheduleClassTypeHomework.String())
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
	if len(filter.SchoolIDs) > 0 {
		if filter.SchoolIDs[0] == constant.LearningSummaryFilterOptionNoneID {
			classes, err := external.GetClassServiceProvider().GetOnlyUnderOrgClasses(ctx, operator, operator.OrgID)
			if err != nil {
				log.Error(ctx, "find related schedules: get only under org classes failed",
					log.Err(err),
					log.Any("schedule_condition", scheduleCondition),
				)
				return nil, err
			}
			classIDs := make([]string, 0, len(classes))
			for _, c := range classes {
				classIDs = append(classIDs, c.ID)
			}
			scheduleCondition.RelationClassIDs = entity.NullStrings{
				Strings: classIDs,
				Valid:   true,
			}
		}
		scheduleCondition.RelationSchoolIDs = entity.NullStrings{
			Strings: filter.SchoolIDs,
			Valid:   true,
		}
	}
	if len(filter.ClassID) > 0 {
		if filter.ClassID == constant.LearningSummaryFilterOptionNoneID {
			users, err := external.GetUserServiceProvider().GetOnlyUnderOrgUsers(ctx, operator, operator.OrgID)
			if err != nil {
				log.Error(ctx, "find related schedules: get only under org users failed",
					log.Err(err),
					log.Any("schedule_condition", scheduleCondition),
				)
				return nil, err
			}
			userIDs := make([]string, 0, len(users))
			for _, u := range users {
				userIDs = append(userIDs, u.ID)
			}
			scheduleCondition.RelationTeacherIDs = entity.NullStrings{
				Strings: userIDs,
				Valid:   true,
			}
			scheduleCondition.RelationStudentIDs = entity.NullStrings{
				Strings: userIDs,
				Valid:   true,
			}
		}
		if scheduleCondition.RelationClassIDs.Valid {
			scheduleCondition.RelationClassIDs.Strings = append(scheduleCondition.RelationClassIDs.Strings, filter.ClassID)
		} else {
			scheduleCondition.RelationClassIDs = entity.NullStrings{
				Strings: []string{filter.ClassID},
				Valid:   true,
			}
		}
	}
	if len(filter.TeacherID) > 0 {
		if scheduleCondition.RelationTeacherIDs.Valid {
			scheduleCondition.RelationTeacherIDs.Strings = append(scheduleCondition.RelationTeacherIDs.Strings, filter.TeacherID)
		} else {
			scheduleCondition.RelationTeacherIDs = entity.NullStrings{
				Strings: []string{filter.TeacherID},
				Valid:   true,
			}
		}
	}
	if len(filter.StudentID) > 0 {
		if scheduleCondition.RelationStudentIDs.Valid {
			scheduleCondition.RelationStudentIDs.Strings = append(scheduleCondition.RelationStudentIDs.Strings, filter.StudentID)
		} else {
			scheduleCondition.RelationStudentIDs = entity.NullStrings{
				Strings: []string{filter.StudentID},
				Valid:   true,
			}
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
		for _, s := range schedules {
			scheduleIDs = append(scheduleIDs, s.ID)
		}
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
		for _, s := range schedules {
			if utils.ContainsString(filterScheduleIDs, s.ID) {
				filterSchedules = append(filterSchedules, s)
			}
		}
		schedules = filterSchedules
	}
	return schedules, nil
}

func (l *learningSummaryReportModel) findRelatedAssessments(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, typo entity.LearningSummaryType, filter *entity.LearningSummaryFilter, scheduleIDs []string) ([]*entity.Assessment, error) {
	// query assessments

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

	assessments, err := da.GetAssessmentDA().Query(ctx, &cond)
	if err != nil {
		log.Error(ctx, "find related assessments: query assessment attendance relations failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("filter", filter),
		)
		return nil, err
	}

	return assessments, nil
}

func (l *learningSummaryReportModel) QueryOutcomesByAssessmentID(ctx context.Context, op *entity.Operator, assessmentID string, studentID string) (res []*entity.LearningSummaryOutcome, err error) {
	if assessmentID == "" {
		log.Warn(ctx, "assessment_id is required")
		err = constant.ErrInvalidArgs
		return
	}
	if studentID == "" {
		log.Warn(ctx, "student_id is required")
		err = constant.ErrInvalidArgs
		return
	}
	res = []*entity.LearningSummaryOutcome{}
	items, err := da.GetReportDA().QueryOutcomesByAssessmentID(ctx, op, assessmentID, studentID)
	if err != nil {
		return
	}
	for _, item := range items {
		if item.CountOfAll == item.CountOfUnknown {
			continue
		}
		o := &entity.LearningSummaryOutcome{
			ID:   item.OutcomeID,
			Name: item.OutcomeName,
		}
		if item.CountOfNotCovered == item.CountOfAll {
			continue
		}
		if item.CountOfAchieved == item.CountOfAll {
			o.Status = entity.AssessmentOutcomeStatusAchieved
		} else if item.CountOfAchieved > 0 && item.CountOfNotAchieved > 0 {
			o.Status = entity.AssessmentOutcomeStatusPartially
		} else {
			o.Status = entity.AssessmentOutcomeStatusNotAchieved
		}

		res = append(res, o)
	}

	return
}

func (l *learningSummaryReportModel) QueryAssignmentsSummaryV2(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (res *entity.QueryAssignmentsSummaryResultV2, err error) {
	items, err := da.GetReportDA().QueryAssignmentsSummaryV2(ctx, tx, operator, filter)
	if err != nil {
		return
	}
	res = &entity.QueryAssignmentsSummaryResultV2{
		Items: items,
	}
	for _, item := range items {
		switch item.Type {
		case entity.AssessmentTypeHomeFunStudy:
			res.HomeFunStudyCount++
		case entity.AssessmentTypeStudy:
			res.StudyCount++
		}
	}

	var scheduleIDs []string
	for _, item := range items {
		scheduleIDs = append(scheduleIDs, item.ScheduleID)
	}

	// find related study assessments comments and make map by schedule id (live: room comments)
	roomCommentMap, err := getAssessmentH5P().batchGetRoomCommentMap(ctx, operator, scheduleIDs)
	if err != nil {
		log.Error(ctx, "query assignments summary: batch get room comment map failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("filter", filter),
		)
		return
	}
	for _, item := range items {
		comments := roomCommentMap[item.ScheduleID][filter.StudentID]
		if len(comments) > 0 {
			item.TeacherFeedback = comments[len(comments)-1]
		}
	}
	return
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

	// batch get assessment outcome status
	var assessmentOutcomeKeys []*entity.AssessmentOutcomeKey
	for assessmentID, outcomes := range assessmentIDToOutcomesMap {
		for _, o := range outcomes {
			assessmentOutcomeKeys = append(assessmentOutcomeKeys, &entity.AssessmentOutcomeKey{
				AssessmentID: assessmentID,
				OutcomeID:    o.ID,
			})
		}
	}
	outcomeStatusMap, err := l.batchGetAssessmentOutcomeStatus(ctx, filter.StudentID, assessmentOutcomeKeys)
	if err != nil {
		log.Error(ctx, "query assignments summary: batch get assessment outcome status failed",
			log.Err(err),
			log.String("student_id", filter.StudentID),
			log.Any("keys", assessmentOutcomeKeys),
			log.Any("filter", filter),
			log.Any("home_fun_studies", homeFunStudyAssessments),
		)
		return nil, err
	}

	// assembly result
	result := l.assemblyAssignmentsSummaryResult(filter, schedules, assessmentIDToOutcomesMap, lessonPlanNameMap, studyAssessmentMap, homeFunStudyAssessmentMap, roomCommentMap, outcomeStatusMap)

	// sort items
	l.sortAssignmentsSummaryItems(result.Items)

	log.Debug(ctx, "query assignments summary result", log.Any("result", result))

	return result, nil
}

func (l *learningSummaryReportModel) assemblyAssignmentsSummaryResult(
	filter *entity.LearningSummaryFilter,
	schedules []*entity.Schedule,
	assessmentIDToOutcomesMap map[string][]*entity.Outcome,
	lessonPlanNameMap map[string]string,
	studyAssessmentMap map[string]*entity.Assessment,
	homeFunStudyAssessmentMap map[string]*entity.HomeFunStudy,
	roomCommentMap map[string]map[string][]string,
	outcomeStatusMap map[entity.AssessmentOutcomeKey]entity.AssessmentOutcomeStatus,
) *entity.QueryAssignmentsSummaryResult {
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
					status, ok := outcomeStatusMap[entity.AssessmentOutcomeKey{
						AssessmentID: assessment.ID,
						OutcomeID:    o.ID,
					}]
					if !ok {
						continue
					}
					item.Outcomes = append(item.Outcomes, &entity.LearningSummaryOutcome{
						ID:     o.ID,
						Name:   o.Name,
						Status: status,
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
					status, ok := outcomeStatusMap[entity.AssessmentOutcomeKey{
						AssessmentID: assessment.ID,
						OutcomeID:    o.ID,
					}]
					if !ok {
						continue
					}
					item.Outcomes = append(item.Outcomes, &entity.LearningSummaryOutcome{
						ID:     o.ID,
						Name:   o.Name,
						Status: status,
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

func (l *learningSummaryReportModel) batchGetAssessmentOutcomeStatus(ctx context.Context, attendanceID string, keys []*entity.AssessmentOutcomeKey) (map[entity.AssessmentOutcomeKey]entity.AssessmentOutcomeStatus, error) {
	if len(keys) == 0 {
		log.Debug(ctx, "batch get assessment outcome status: empty keys")
		return map[entity.AssessmentOutcomeKey]entity.AssessmentOutcomeStatus{}, nil
	}

	// query assessment outcome overall status
	assessmentOutcomeCond := da.QueryAssessmentOutcomeConditions{
		Keys: entity.NullAssessmentOutcomeKeys{
			Value: keys,
			Valid: true,
		},
	}
	var assessmentOutcomes []*entity.AssessmentOutcome
	if err := da.GetAssessmentOutcomeDA().Query(ctx, &assessmentOutcomeCond, &assessmentOutcomes); err != nil {
		log.Error(ctx, "batch get assessment outcome status: query assessment outcome failed",
			log.Err(err),
			log.String("student_id", attendanceID),
			log.Any("keys", keys),
		)
		return nil, err
	}
	assessmentOutcomesMap := make(map[entity.AssessmentOutcomeKey]*entity.AssessmentOutcome, len(assessmentOutcomes))
	for _, ao := range assessmentOutcomes {
		assessmentOutcomesMap[entity.AssessmentOutcomeKey{AssessmentID: ao.AssessmentID, OutcomeID: ao.OutcomeID}] = ao
	}

	// query assessment outcome attendances
	assessmentOutcomeAttendanceCond := da.QueryAssessmentOutcomeAttendanceCondition{
		AttendanceIDs: entity.NullStrings{
			Strings: []string{attendanceID},
			Valid:   true,
		},
		AssessmentIDAndOutcomeIDPairs: entity.NullAssessmentOutcomeKeys{
			Value: keys,
			Valid: true,
		},
	}
	var assessmentOutcomeAttendances []*entity.OutcomeAttendance
	if err := da.GetOutcomeAttendanceDA().Query(ctx, &assessmentOutcomeAttendanceCond, &assessmentOutcomeAttendances); err != nil {
		log.Error(ctx, "batch get assessment outcome status: query assessment outcome attendance failed",
			log.Err(err),
			log.Any("cond", assessmentOutcomeAttendanceCond),
			log.String("attendance_id", attendanceID),
			log.Any("keys", keys),
		)
		return nil, err
	}
	assessmentOutcomeAttendMap := make(map[entity.AssessmentOutcomeAttendanceKey]bool, len(keys))
	for _, a := range assessmentOutcomeAttendances {
		assessmentOutcomeAttendMap[entity.AssessmentOutcomeAttendanceKey{
			AssessmentID: a.AssessmentID,
			OutcomeID:    a.OutcomeID,
			AttendanceID: a.AttendanceID,
		}] = true
	}

	// query assessment content outcome attend map
	assessmentContentOutcomeAttendanceCond := da.QueryAssessmentContentOutcomeAttendanceCondition{
		AttendanceIDs: entity.NullStrings{},
		AssessmentIDAndOutcomeIDPairs: entity.NullAssessmentOutcomeKeys{
			Value: keys,
			Valid: true,
		},
	}
	var assessmentContentOutcomeAttendances []*entity.AssessmentContentOutcomeAttendance
	if err := da.GetAssessmentContentOutcomeAttendanceDA().Query(ctx, &assessmentContentOutcomeAttendanceCond, &assessmentContentOutcomeAttendances); err != nil {
		log.Error(ctx, "batch get assessment outcome status: query assessment content outcome attendance failed",
			log.Err(err),
			log.String("student_id", attendanceID),
			log.Any("keys", keys),
		)
		return nil, err
	}
	assessmentContentOutcomeAttendMap := map[entity.AssessmentOutcomeAttendanceKey]bool{}
	for _, item := range assessmentContentOutcomeAttendances {
		assessmentContentOutcomeAttendMap[entity.AssessmentOutcomeAttendanceKey{
			AssessmentID: item.AssessmentID,
			OutcomeID:    item.OutcomeID,
			AttendanceID: item.AttendanceID,
		}] = true
	}

	// construct partial status map
	assessmentOutcomePartiallyAttendMap := make(map[entity.AssessmentOutcomeAttendanceKey]bool, len(keys))
	for _, key := range keys {
		if key == nil {
			continue
		}
		withAttendanceKey := entity.AssessmentOutcomeAttendanceKey{
			AssessmentID: key.AssessmentID,
			OutcomeID:    key.OutcomeID,
			AttendanceID: attendanceID,
		}
		if assessmentOutcomeAttendMap[withAttendanceKey] {
			continue
		}
		if assessmentContentOutcomeAttendMap[withAttendanceKey] {
			assessmentOutcomePartiallyAttendMap[withAttendanceKey] = true
		}
	}

	// aggregate result
	result := make(map[entity.AssessmentOutcomeKey]entity.AssessmentOutcomeStatus, len(keys))
	for _, key := range keys {
		if key == nil {
			continue
		}
		ao := assessmentOutcomesMap[*key]
		if ao == nil {
			continue
		}
		withAttendanceKey := entity.AssessmentOutcomeAttendanceKey{
			AssessmentID: key.AssessmentID,
			OutcomeID:    key.OutcomeID,
			AttendanceID: attendanceID,
		}
		if ao.Skip {
			continue
		} else if ao.NoneAchieved {
			result[*key] = entity.AssessmentOutcomeStatusNotAchieved
		} else if assessmentOutcomeAttendMap[withAttendanceKey] {
			result[*key] = entity.AssessmentOutcomeStatusAchieved
		} else if assessmentOutcomePartiallyAttendMap[withAttendanceKey] {
			result[*key] = entity.AssessmentOutcomeStatusPartially
		} else {
			result[*key] = entity.AssessmentOutcomeStatusNotAchieved
		}
	}

	log.Debug(ctx, "batch get assessment outcome status: print args",
		log.String("attendance_id", attendanceID),
		log.Any("keys", keys),
		log.String("result", fmt.Sprintf("%+v", result)),
		log.Any("assessment_outcomes", assessmentOutcomes),
		log.Any("assessment_outcome_attendances", assessmentOutcomeAttendances),
		log.Any("assessment_content_outcome_attendances", assessmentContentOutcomeAttendances),
	)

	return result, nil
}
