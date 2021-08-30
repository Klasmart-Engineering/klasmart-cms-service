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
		schedules, err := l.findRelatedSchedules(ctx, tx, operator, entity.LearningSummaryTypeLiveClass, &entity.LearningSummaryFilter{
			SchoolIDs: args.SchoolIDs,
			TeacherID: args.TeacherID,
			StudentID: args.StudentID,
		})
		if err != nil {
			log.Error(ctx, "query time filter: find related schedules failed",
				log.Err(err),
				log.Any("args", args),
			)
			return nil, err
		}
		for _, s := range schedules {
			year := time.Unix(s.StartAt, 0).Year()
			weekStart, weekEnd := utils.FindWeekTimeRangeFromMonday(s.StartAt, fixedZone)
			m[year] = append(m[year], [2]int64{weekStart, weekEnd})
		}
	case entity.LearningSummaryTypeAssignment:
		schedules, err := l.findRelatedSchedules(ctx, tx, operator, entity.LearningSummaryTypeAssignment, &entity.LearningSummaryFilter{
			SchoolIDs: args.SchoolIDs,
			TeacherID: args.TeacherID,
			StudentID: args.StudentID,
		})
		if err != nil {
			log.Error(ctx, "query time filter: find related schedules failed",
				log.Err(err),
				log.Any("args", args),
			)
			return nil, err
		}
		scheduleIDs := make([]string, 0, len(schedules))
		for _, s := range schedules {
			scheduleIDs = append(scheduleIDs, s.ID)
		}
		scheduleIDs = utils.SliceDeduplicationExcludeEmpty(scheduleIDs)
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
			ScheduleIDs: entity.NullStrings{
				Strings: scheduleIDs,
				Valid:   true,
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
			weekStart, weekEnd := utils.FindWeekTimeRangeFromMonday(a.CompleteTime, fixedZone)
			m[year] = append(m[year], [2]int64{weekStart, weekEnd})
		}
	}

	// calc current week
	currentWeekStart, currentWeekEnd := utils.FindWeekTimeRangeFromMonday(time.Now().Unix(), fixedZone)

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
	switch args.FilterType {
	case entity.LearningSummaryFilterTypeSchool:
		return l.queryRemainingFilterSchool(ctx, tx, operator)
	case entity.LearningSummaryFilterTypeClass:
		var teacherIDs []string
		if len(args.TeacherID) > 0 {
			teacherIDs = append(teacherIDs, args.TeacherID)
		}
		return l.queryRemainingFilterClass(ctx, tx, operator, args.SchoolIDs, teacherIDs)
	case entity.LearningSummaryFilterTypeTeacher:
		var classIDs []string
		if len(args.ClassID) > 0 {
			classIDs = append(classIDs, args.ClassID)
		}
		return l.queryRemainingFilterTeacher(ctx, tx, operator, classIDs)
	case entity.LearningSummaryFilterTypeStudent:
		var classIDs []string
		if len(args.ClassID) > 0 {
			classIDs = append(classIDs, args.ClassID)
		}
		return l.queryRemainingFilterStudent(ctx, tx, operator, classIDs)
	case entity.LearningSummaryFilterTypeSubject:
		return l.queryRemainingFilterSubject(ctx, tx, operator, args)
	default:
		log.Error(ctx, "query remaining filter: invalid filter type")
		return nil, constant.ErrInvalidArgs
	}
}

func (l *learningSummaryReportModel) queryRemainingFilterSchool(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator) ([]*entity.QueryLearningSummaryRemainingFilterResultItem, error) {
	schools, err := external.GetSchoolServiceProvider().GetByOrganizationID(ctx, operator, operator.OrgID)
	if err != nil {
		return nil, err
	}
	schoolIDs := make([]string, 0, len(schools))
	for _, s := range schools {
		schoolIDs = append(schoolIDs, s.ID)
	}
	schoolIDs = utils.SliceDeduplicationExcludeEmpty(schoolIDs)
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
		name := schoolNameMap[schoolID]
		if name != "" {
			result = append(result, &entity.QueryLearningSummaryRemainingFilterResultItem{
				SchoolID:   schoolID,
				SchoolName: name,
			})
		}
	}
	//has, err := l.hasNoneSchoolOption(ctx, tx, operator, scheduleIDs)
	//if err != nil {
	//	log.Error(ctx, "query remaining filter school failed: check has none school option failed",
	//		log.Err(err),
	//		log.Strings("schedule_ids", scheduleIDs),
	//		log.Any("operator", operator),
	//	)
	//	return nil, err
	//}
	//if has {
	//	result = append(result, &entity.QueryLearningSummaryRemainingFilterResultItem{
	//		SchoolID:   constant.LearningSummaryFilterOptionNoneID,
	//		SchoolName: constant.LearningSummaryFilterOptionNoneName,
	//	})
	//}
	return result, nil
}

func (l *learningSummaryReportModel) queryRemainingFilterClass(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, schoolIDs []string, teacherIDs []string) ([]*entity.QueryLearningSummaryRemainingFilterResultItem, error) {
	// filter schools
	if len(schoolIDs) > 0 && schoolIDs[0] == constant.LearningSummaryFilterOptionNoneID {
		log.Debug(ctx, "query remaining filter class: check 'none' option")
		return nil, nil
	}
	var classIDs []string

	if len(schoolIDs) == 0 && len(teacherIDs) > 0 {
		schoolsMap, err := external.GetSchoolServiceProvider().GetByUsers(ctx, operator, operator.OrgID, teacherIDs)
		if err != nil {
			log.Error(ctx, "query remaining filter class: get schools by user ids failed",
				log.Err(err),
				log.Strings("teacher_ids", teacherIDs),
				log.Any("operator", operator),
			)
			return nil, err
		}
		for _, schools := range schoolsMap {
			for _, s := range schools {
				schoolIDs = append(schoolIDs, s.ID)
			}
		}
		schoolIDs = utils.SliceDeduplicationExcludeEmpty(schoolIDs)
	}

	if len(schoolIDs) > 0 {
		schoolClassesMap, err := external.GetClassServiceProvider().GetBySchoolIDs(ctx, operator, schoolIDs)
		if err != nil {
			log.Error(ctx, "query remaining filter class: get classes by school ids failed",
				log.Err(err),
				log.Any("operator", operator),
			)
			return nil, err
		}
		var schoolClassIDs []string
		for _, classes := range schoolClassesMap {
			for _, c := range classes {
				schoolClassIDs = append(schoolClassIDs, c.ID)
			}
		}
		schoolClassIDs = utils.SliceDeduplicationExcludeEmpty(schoolClassIDs)
		classIDs = append(classIDs, schoolClassIDs...)
	}

	// filter teachers
	if len(teacherIDs) > 0 {
		userClassesMap, err := external.GetClassServiceProvider().GetByUserIDs(ctx, operator, teacherIDs)
		if err != nil {
			log.Error(ctx, "query remaining filter class failed: get classes by teacher ids failed",
				log.Err(err),
				log.Any("operator", operator),
			)
			return nil, err
		}
		var userClassIDs []string
		for _, classes := range userClassesMap {
			for _, c := range classes {
				userClassIDs = append(userClassIDs, c.ID)
			}
		}
		userClassIDs = utils.SliceDeduplicationExcludeEmpty(userClassIDs)
		classIDs = utils.IntersectAndDeduplicateStrSlice(classIDs, userClassIDs)
	}

	// check empty
	if len(classIDs) == 0 {
		log.Debug(ctx, "query remaining filter class: no class ids found",
			log.Strings("school_ids", schoolIDs),
			log.Strings("teacher_ids", teacherIDs),
		)
		return nil, nil
	}

	// batch get name
	classNameMap, err := external.GetClassServiceProvider().BatchGetNameMap(ctx, operator, classIDs)
	if err != nil {
		log.Error(ctx, "query remaining filter class failed: batch get class name map failed",
			log.Err(err),
			log.Strings("class_ids", classIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// assembly result
	result := make([]*entity.QueryLearningSummaryRemainingFilterResultItem, 0, len(classIDs))
	for _, classID := range classIDs {
		name := classNameMap[classID]
		if name != "" {
			result = append(result, &entity.QueryLearningSummaryRemainingFilterResultItem{
				ClassID:   classID,
				ClassName: name,
			})
		}
	}
	//has, err := l.hasNoneClassOption(ctx, tx, operator, scheduleIDs)
	//if err != nil {
	//	log.Error(ctx, "query remaining filter school failed: check has none class option failed",
	//		log.Err(err),
	//		log.Strings("schedule_ids", scheduleIDs),
	//		log.Any("operator", operator),
	//	)
	//	return nil, err
	//}
	//if has {
	//	result = append(result, &entity.QueryLearningSummaryRemainingFilterResultItem{
	//		ClassID:   constant.LearningSummaryFilterOptionNoneID,
	//		ClassName: constant.LearningSummaryFilterOptionNoneName,
	//	})
	//}
	return result, nil
}

func (l *learningSummaryReportModel) queryRemainingFilterTeacher(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, classIDs []string) ([]*entity.QueryLearningSummaryRemainingFilterResultItem, error) {
	if len(classIDs) > 0 && classIDs[0] == constant.LearningSummaryFilterOptionNoneID {
		log.Debug(ctx, "query remaining filter teacher: check 'none' option")
		return nil, nil
	}
	teachersMap, err := external.GetTeacherServiceProvider().GetByClasses(ctx, operator, classIDs)
	if err != nil {
		log.Error(ctx, "query remaining filter teacher: get teachers by class ids failed",
			log.Err(err),
			log.Any("operator", operator),
		)
		return nil, err
	}
	var teacherIDs []string
	for _, teachers := range teachersMap {
		for _, t := range teachers {
			teacherIDs = append(teacherIDs, t.ID)
		}
	}
	teacherIDs = utils.SliceDeduplicationExcludeEmpty(teacherIDs)
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
		name := teacherNameMap[teacherID]
		if name != "" {
			result = append(result, &entity.QueryLearningSummaryRemainingFilterResultItem{
				TeacherID:   teacherID,
				TeacherName: name,
			})
		}
	}
	return result, nil
}

func (l *learningSummaryReportModel) queryRemainingFilterStudent(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, classIDs []string) ([]*entity.QueryLearningSummaryRemainingFilterResultItem, error) {
	if len(classIDs) > 0 && classIDs[0] == constant.LearningSummaryFilterOptionNoneID {
		log.Debug(ctx, "query remaining filter student: check 'none' option")
		return nil, nil
	}
	studentMap, err := external.GetStudentServiceProvider().GetByClassIDs(ctx, operator, classIDs)
	var studentIDs []string
	for _, students := range studentMap {
		for _, s := range students {
			studentIDs = append(studentIDs, s.ID)
		}
	}
	studentIDs = utils.SliceDeduplicationExcludeEmpty(studentIDs)
	studentNameMap, err := external.GetStudentServiceProvider().BatchGetNameMap(ctx, operator, studentIDs)
	if err != nil {
		log.Error(ctx, "query remaining filter student: batch get student name map failed",
			log.Err(err),
			log.Strings("student_ids", studentIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	result := make([]*entity.QueryLearningSummaryRemainingFilterResultItem, 0, len(studentIDs))
	for _, studentID := range studentIDs {
		name := studentNameMap[studentID]
		if name != "" {
			result = append(result, &entity.QueryLearningSummaryRemainingFilterResultItem{
				StudentID:   studentID,
				StudentName: name,
			})
		}
	}
	return result, nil
}

func (l *learningSummaryReportModel) queryRemainingFilterSubject(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.QueryLearningSummaryRemainingFilterArgs) ([]*entity.QueryLearningSummaryRemainingFilterResultItem, error) {
	if args.WeekStart == 0 || args.WeekEnd == 0 ||
		len(args.SchoolIDs) > 0 && args.SchoolIDs[0] == constant.LearningSummaryFilterOptionNoneID ||
		args.ClassID == constant.LearningSummaryFilterOptionNoneID ||
		args.TeacherID == constant.LearningSummaryFilterOptionNoneID ||
		args.StudentID == constant.LearningSummaryFilterOptionNoneID {
		log.Debug(ctx, "query remaining filter subject: check 'none' option",
			log.Any("args", args),
		)
		return nil, nil
	}
	schedules, err := l.findRelatedSchedules(ctx, tx, operator, args.SummaryType, &args.LearningSummaryFilter)
	if err != nil {
		log.Error(ctx, "query remaining filter subject failed: find related schedules",
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
	scheduleIDs = utils.SliceDeduplication(scheduleIDs)
	subjectIDs, err := l.batchGetScheduleRelationIDs(ctx, operator, scheduleIDs, []entity.ScheduleRelationType{entity.ScheduleRelationTypeSubject})
	if err != nil {
		log.Error(ctx, "query remaining filter student failed: batch get students relations failed",
			log.Err(err),
			log.Any("operator", operator),
		)
		return nil, err
	}
	subjectIDs = utils.SliceDeduplicationExcludeEmpty(subjectIDs)
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
		name := subjectNameMap[subjectID]
		result = append(result, &entity.QueryLearningSummaryRemainingFilterResultItem{
			SubjectID:   subjectID,
			SubjectName: name,
		})
	}
	return result, nil
}

func (l *learningSummaryReportModel) hasNoneSchoolOption(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) (bool, error) {
	relatedClassIDs, err := l.batchGetScheduleRelationIDs(ctx, operator, scheduleIDs, []entity.ScheduleRelationType{entity.ScheduleRelationTypeClassRosterClass, entity.ScheduleRelationTypeParticipantClass})
	if err != nil {
		log.Error(ctx, "has none school: batch get schedule class ids failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Strings("schedule_ids", scheduleIDs),
		)
		return false, err
	}
	classes, err := external.GetClassServiceProvider().GetOnlyUnderOrgClasses(ctx, operator, operator.OrgID)
	if err != nil {
		log.Error(ctx, "has none school: get only under org classes failed",
			log.Err(err),
			log.Any("operator", operator),
		)
		return false, err
	}
	classIDs := make([]string, 0, len(classes))
	for _, c := range classes {
		classIDs = append(classIDs, c.ID)
	}

	return utils.HasIntersection(relatedClassIDs, classIDs), nil
}

func (l *learningSummaryReportModel) hasNoneClassOption(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) (bool, error) {
	relatedStudentIDs, err := l.batchGetScheduleRelationIDs(ctx, operator, scheduleIDs, []entity.ScheduleRelationType{entity.ScheduleRelationTypeClassRosterStudent, entity.ScheduleRelationTypeParticipantStudent})
	if err != nil {
		log.Error(ctx, "query remaining filter student failed: batch get students relations failed",
			log.Err(err),
			log.Any("operator", operator),
		)
		return false, err
	}
	users, err := external.GetUserServiceProvider().GetOnlyUnderOrgUsers(ctx, operator, operator.OrgID)
	if err != nil {
		log.Error(ctx, "has none school: get only under org classes failed",
			log.Err(err),
			log.Any("operator", operator),
		)
		return false, err
	}
	userIDS := make([]string, 0, len(users))
	for _, u := range users {
		userIDS = append(userIDS, u.ID)
	}
	return utils.HasIntersection(relatedStudentIDs, userIDS), nil
}

func (l *learningSummaryReportModel) batchGetScheduleRelationIDs(ctx context.Context, operator *entity.Operator, scheduleIDs []string, relationTypes []entity.ScheduleRelationType) ([]string, error) {
	cond := da.ScheduleRelationCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
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
		withAttendanceKey := entity.AssessmentOutcomeAttendanceKey{
			AssessmentID: key.AssessmentID,
			OutcomeID:    key.OutcomeID,
			AttendanceID: attendanceID,
		}
		ao := assessmentOutcomesMap[*key]
		if ao == nil {
			continue
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
		log.Any("result", result),
		log.Any("assessment_outcomes", assessmentOutcomes),
		log.Any("assessment_outcome_attendances", assessmentOutcomeAttendances),
		log.Any("assessment_content_outcome_attendances", assessmentContentOutcomeAttendances),
	)

	return result, nil
}
