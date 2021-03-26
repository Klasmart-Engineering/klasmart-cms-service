package model

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/tidwall/gjson"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type IReportModel interface {
	ListStudentsReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.ListStudentsAchievementReportRequest) (*entity.StudentsAchievementReportResponse, error)
	GetStudentReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.GetStudentAchievementReportRequest) (*entity.StudentAchievementReportResponse, error)
	GetTeacherReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, teacherID string) (*entity.TeacherReport, error)

	ListStudentsPerformanceReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.ListStudentsPerformanceReportRequest) (*entity.ListStudentsPerformanceReportResponse, error)
	GetStudentPerformanceReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.GetStudentPerformanceReportRequest) (*entity.GetStudentPerformanceReportResponse, error)

	ListStudentsPerformanceH5PReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.ListStudentsPerformanceH5PReportRequest) (*entity.ListStudentsPerformanceH5PReportResponse, error)
	GetStudentPerformanceH5PReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.GetStudentPerformanceH5PReportRequest) (*entity.GetStudentPerformanceH5PReportResponse, error)
}

var (
	reportModelInstance IReportModel
	reportModelOnce     = sync.Once{}
)

func GetReportModel() IReportModel {
	reportModelOnce.Do(func() {
		reportModelInstance = &reportModel{}
	})
	return reportModelInstance
}

type reportModel struct{}

// region assessment

func (rm *reportModel) ListStudentsReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.ListStudentsAchievementReportRequest) (*entity.StudentsAchievementReportResponse, error) {
	{
		if !req.Status.Valid() {
			log.Error(ctx, "list students report: invalid status", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		if !req.SortBy.Valid() {
			log.Error(ctx, "list students report: invalid sort by", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		if req.ClassID == "" {
			log.Error(ctx, "list students report: require class id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		if req.TeacherID == "" {
			log.Error(ctx, "list students report: require teacher id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		if req.LessonPlanID == "" {
			log.Error(ctx, "list students report: require lesson plan id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		allowed, err := rm.hasReportPermission(ctx, operator, req.TeacherID)
		if err != nil {
			log.Error(ctx, "list students report: check report report permission failed",
				log.Any("req", req),
				log.Any("operator", operator),
			)
			return nil, err
		}
		if !allowed {
			log.Error(ctx, "list students report: no permission",
				log.Any("req", req),
				log.Any("operator", operator),
			)
			return nil, constant.ErrForbidden
		}
	}

	students, err := rm.getStudentsInClass(ctx, operator, req.ClassID)
	if err != nil {
		log.Error(ctx, "list students report: get students",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}

	assessmentIDs, err := rm.getCompletedAssessmentIDs(ctx, tx, operator, req.ClassID, req.LessonPlanID)
	if err != nil {
		log.Error(ctx, "list student report: get assessment ids failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}

	assessmentAttendances, err := rm.getCheckedAssessmentAttendance(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "list student report: get checked assessment attendance failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	checked := true
	assessmentOutcomes, err := da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs, &checked)
	if err != nil {
		log.Error(ctx, "ListStudentsReport: da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDs: get assessment outcomes failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	outcomeAttendances, err := rm.getOutcomeAttendances(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "list student report: call getOutcomeAttendances failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	tr, err := rm.makeLatestOutcomeIDsTranslator(ctx, tx, operator, rm.getOutcomeIDs(assessmentOutcomes))
	if err != nil {
		log.Error(ctx, "list student report: make latest outcome ids translator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Any("assessment_outcomes", assessmentOutcomes),
		)
		return nil, err
	}

	attendanceIDExistsMap := rm.getAttendanceIDsExistMap(assessmentAttendances)
	attendanceID2OutcomeIDsMap := rm.getAttendanceID2OutcomeIDsMap(assessmentAttendances, assessmentOutcomes)
	achievedAttendanceID2OutcomeIDsMap := rm.getAchievedAttendanceID2OutcomeIDsMap(outcomeAttendances)
	skipAttendanceID2OutcomeIDsMap := rm.getSkipAttendanceID2OutcomeIDsMap(assessmentAttendances, assessmentOutcomes)
	notAchievedAttendanceID2OutcomeIDsMap := rm.getNotAchievedAttendanceID2OutcomeIDsMap(attendanceID2OutcomeIDsMap, achievedAttendanceID2OutcomeIDsMap, skipAttendanceID2OutcomeIDsMap)
	log.Debug(ctx, "ListStudentsReport: print all map",
		log.Any("attendance_id_exists_map", attendanceIDExistsMap),
		log.Any("attendance_id_2_outcome_ids_map", attendanceID2OutcomeIDsMap),
		log.Any("achieved_attendance_id_2_outcome_ids_map", achievedAttendanceID2OutcomeIDsMap),
		log.Any("skip_attendance_id_2_outcome_ids_map", skipAttendanceID2OutcomeIDsMap),
		log.Any("not_achieved_attendance_id_2_outcome_ids_map", notAchievedAttendanceID2OutcomeIDsMap),
	)

	var result = entity.StudentsAchievementReportResponse{AssessmentIDs: assessmentIDs}
	for _, student := range students {
		newItem := entity.StudentAchievementReportItem{StudentID: student.ID, StudentName: student.Name}
		if !attendanceIDExistsMap[student.ID] {
			newItem.Attend = false
			result.Items = append(result.Items, &newItem)
			continue
		}
		newItem.Attend = true
		newItem.AchievedCount = len(tr(achievedAttendanceID2OutcomeIDsMap[student.ID]))
		newItem.NotAttemptedCount = len(tr(skipAttendanceID2OutcomeIDsMap[student.ID]))
		newItem.NotAchievedCount = len(tr(notAchievedAttendanceID2OutcomeIDsMap[student.ID]))
		result.Items = append(result.Items, &newItem)
	}

	sortInterface := entity.NewSortingStudentReportItems(result.Items, req.Status)
	switch req.SortBy {
	case entity.ReportSortByDesc:
		sort.Sort(sort.Reverse(sortInterface))
	case entity.ReportSortByAsc:
		fallthrough
	default:
		sort.Sort(sortInterface)
	}

	return &result, nil
}

func (rm *reportModel) GetStudentReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.GetStudentAchievementReportRequest) (*entity.StudentAchievementReportResponse, error) {
	{
		if req.ClassID == "" {
			log.Error(ctx, "get student detail report: require class id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		if req.TeacherID == "" {
			log.Error(ctx, "get student detail report: require teacher id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		if req.LessonPlanID == "" {
			log.Error(ctx, "get student detail report: require lesson plan id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		if req.StudentID == "" {
			log.Error(ctx, "get student detail report: require student id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		allowed, err := rm.hasReportPermission(ctx, operator, req.TeacherID)
		if err != nil {
			log.Error(ctx, "get student detail report: check report report permission failed",
				log.Any("operator", operator),
				log.Any("req", req),
			)
			return nil, err
		}
		if !allowed {
			log.Error(ctx, "get student detail report: no permission",
				log.Any("operator", operator),
				log.Any("req", req),
			)
			return nil, constant.ErrUnAuthorized
		}
	}

	student, err := rm.getStudentInClass(ctx, operator, req.ClassID, req.StudentID)
	if err != nil {
		log.Error(ctx, "list students report: call getStudentInClass failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}

	assessmentIDs, err := rm.getCompletedAssessmentIDs(ctx, tx, operator, req.ClassID, req.LessonPlanID)
	if err != nil {
		log.Error(ctx, "get student detail report: get assessment ids failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}

	checked := true
	assessmentOutcomes, err := da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs, &checked)
	if err != nil {
		log.Error(ctx, "GetStudentDetailReport: da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDs: get assessment outcomes failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
			log.Bool("checked", checked),
		)
		return nil, err
	}

	assessmentAttendances, err := rm.getCheckedAssessmentAttendance(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "get student detail report: get checked assessment attendances failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	outcomeAttendances, err := rm.getOutcomeAttendances(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "get student detail report: call getOutcomeAttendances failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	outcomeIDs := rm.getOutcomeIDs(assessmentOutcomes)

	outcomesMap, err := rm.getOutcomesMap(ctx, tx, operator, outcomeIDs)
	if err != nil {
		log.Error(ctx, "get student detail report: get outcomes map failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("outcome_ids", outcomeIDs),
		)
		return nil, err
	}

	tr, err := rm.makeLatestOutcomeIDsTranslator(ctx, tx, operator, outcomeIDs)
	if err != nil {
		log.Error(ctx, "get student detail report: make latest outcome ids translator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("outcome_ids", outcomeIDs),
		)
		return nil, err
	}

	developmentals, err := external.GetCategoryServiceProvider().GetByOrganization(ctx, operator)
	if err != nil {
		log.Error(ctx, "get student detail report: query all developmental failed",
			log.Err(err),
			log.Any("req", req),
			log.Any("operator", operator),
		)
		return nil, err
	}

	attendanceIDExistsMap := rm.getAttendanceIDsExistMap(assessmentAttendances)
	attendanceID2OutcomeIDsMap := rm.getAttendanceID2OutcomeIDsMap(assessmentAttendances, assessmentOutcomes)
	achievedAttendanceID2OutcomeIDsMap := rm.getAchievedAttendanceID2OutcomeIDsMap(outcomeAttendances)
	skipAttendanceID2OutcomeIDsMap := rm.getSkipAttendanceID2OutcomeIDsMap(assessmentAttendances, assessmentOutcomes)
	notAchievedAttendanceID2OutcomeIDsMap := rm.getNotAchievedAttendanceID2OutcomeIDsMap(attendanceID2OutcomeIDsMap, achievedAttendanceID2OutcomeIDsMap, skipAttendanceID2OutcomeIDsMap)
	log.Debug(ctx, "GetStudentReport: print all map",
		log.Any("attendance_id_exists_map", attendanceIDExistsMap),
		log.Any("attendance_id_2_outcome_ids_map", attendanceID2OutcomeIDsMap),
		log.Any("achieved_attendance_id_2_outcome_ids_map", achievedAttendanceID2OutcomeIDsMap),
		log.Any("skip_attendance_id_2_outcome_ids_map", skipAttendanceID2OutcomeIDsMap),
		log.Any("not_achieved_attendance_id_2_outcome_ids_map", notAchievedAttendanceID2OutcomeIDsMap),
	)

	var result = entity.StudentAchievementReportResponse{StudentName: student.Name, AssessmentIDs: assessmentIDs}
	if !attendanceIDExistsMap[req.StudentID] {
		result.Attend = false
		return &result, nil
	}
	result.Attend = true
	for _, developmental := range developmentals {
		c := entity.StudentAchievementReportCategoryItem{Name: developmental.Name}
		achievedOIDs := tr(achievedAttendanceID2OutcomeIDsMap[req.StudentID])
		for _, oid := range achievedOIDs {
			o := outcomesMap[oid]
			if o == nil {
				continue
			}
			if o.Developmental == developmental.ID {
				c.AchievedItems = append(c.AchievedItems, o.Name)
			}
		}
		notAchievedOIDs := tr(notAchievedAttendanceID2OutcomeIDsMap[req.StudentID])
		for _, oid := range notAchievedOIDs {
			o := outcomesMap[oid]
			if o == nil {
				continue
			}
			if o.Developmental == developmental.ID {
				c.NotAchievedItems = append(c.NotAchievedItems, o.Name)
			}
		}
		skipOIDs := tr(skipAttendanceID2OutcomeIDsMap[req.StudentID])
		for _, oid := range skipOIDs {
			o := outcomesMap[oid]
			if o == nil {
				continue
			}
			if o.Developmental == developmental.ID {
				c.NotAttemptedItems = append(c.NotAttemptedItems, o.Name)
			}
		}
		result.Categories = append(result.Categories, &c)
	}

	return &result, nil
}

func (rm *reportModel) GetTeacherReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, teacherID string) (*entity.TeacherReport, error) {
	var assessmentIDs []string
	{
		var assessments []*entity.Assessment
		if err := da.GetAssessmentDA().Query(ctx, &da.QueryAssessmentsCondition{
			OrgID:      &operator.OrgID,
			TeacherIDs: []string{teacherID},
		}, &assessments); err != nil {
			log.Error(ctx, "get teacher report: query failed",
				log.Err(err),
				log.Any("operator", operator),
				log.String("teacher_id", teacherID),
			)
			return nil, err
		}
		for _, item := range assessments {
			assessmentIDs = append(assessmentIDs, item.ID)
		}
	}
	var outcomes []*entity.Outcome
	{
		checked := true
		assessmentOutcomes, err := da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs, &checked)
		if err != nil {
			log.Error(ctx, "GetTeacherReport: da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDs: get assessment outcomes failed",
				log.Err(err),
				log.Any("operator", operator),
				log.Strings("assessment_ids", assessmentIDs),
				log.Bool("checked", checked),
			)
			return nil, err
		}

		outcomeIDs := rm.getOutcomeIDs(assessmentOutcomes)
		oidTr, err := rm.makeLatestOutcomeIDsTranslator(ctx, tx, operator, outcomeIDs)
		if err != nil {
			log.Error(ctx, "GetTeacherReport: make latest outcome ids translator failed",
				log.Err(err),
				log.Any("outcome_ids", outcomeIDs),
				log.Any("operator", operator),
			)
		}
		outcomeIDs = oidTr(outcomeIDs)
		outcomes, err = GetOutcomeModel().GetLearningOutcomesByIDs(ctx, tx, outcomeIDs, operator)
		if err != nil {
			log.Error(ctx, "get teacher report: get learning outcome failed by ids",
				log.Err(err),
				log.Any("operator", operator),
				log.String("teacher_id", teacherID),
			)
			return nil, err
		}
	}
	developmentalID2NameMap := map[string]string{}
	{
		developmentalList, err := external.GetCategoryServiceProvider().GetByOrganization(ctx, operator)
		if err != nil {
			log.Error(ctx, "get teacher report: query all developmental failed",
				log.Err(err),
				log.Any("teacher_id", teacherID),
				log.Any("operator", operator),
			)
			return nil, err
		}
		for _, item := range developmentalList {
			developmentalID2NameMap[item.ID] = item.Name
		}
	}
	result := &entity.TeacherReport{}
	{
		developmentalID2OutcomeCountMap := map[string][]*entity.Outcome{}
		for _, outcome := range outcomes {
			developmentalID2OutcomeCountMap[outcome.Developmental] = append(developmentalID2OutcomeCountMap[outcome.Developmental], outcome)
		}
		for developmentalID, outcomes := range developmentalID2OutcomeCountMap {
			newItem := &entity.TeacherReportCategory{
				Name: developmentalID2NameMap[developmentalID],
			}
			for _, outcome := range outcomes {
				newItem.Items = append(newItem.Items, outcome.Name)
			}
			result.Categories = append(result.Categories, newItem)
		}
		sort.Sort((*entity.TeacherReportSortByCount)(result))
	}
	return result, nil
}

func (rm *reportModel) ListStudentsPerformanceReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.ListStudentsPerformanceReportRequest) (*entity.ListStudentsPerformanceReportResponse, error) {
	{
		if req.ClassID == "" {
			log.Error(ctx, "ListStudentsPerformanceReport: require class id",
				log.Any("operator", operator),
				log.Any("req", req),
			)
			return nil, constant.ErrInvalidArgs
		}
		if req.TeacherID == "" {
			log.Error(ctx, "ListStudentsPerformanceReport: require teacher id",
				log.Any("operator", operator),
				log.Any("req", req),
			)
			return nil, constant.ErrInvalidArgs
		}
		if req.LessonPlanID == "" {
			log.Error(ctx, "ListStudentsPerformanceReport: require lesson plan id",
				log.Any("operator", operator),
				log.Any("req", req),
			)
			return nil, constant.ErrInvalidArgs
		}
		allowed, err := rm.hasReportPermission(ctx, operator, req.TeacherID)
		if err != nil {
			log.Error(ctx, "ListStudentsPerformanceReport: check report report permission failed",
				log.Any("operator", operator),
				log.Any("req", req),
			)
			return nil, err
		}
		if !allowed {
			log.Error(ctx, "ListStudentsPerformanceReport: no permission",
				log.Any("operator", operator),
				log.Any("req", req),
			)
			return nil, constant.ErrForbidden
		}
	}

	students, err := rm.getStudentsInClass(ctx, operator, req.ClassID)
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceReport: call getStudentsInClass failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}

	assessmentIDs, err := rm.getCompletedAssessmentIDs(ctx, tx, operator, req.ClassID, req.LessonPlanID)
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceReport: call getCompletedAssessmentIDs failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}

	assessmentAttendances, err := rm.getCheckedAssessmentAttendance(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceReport: call getCheckedAssessmentAttendance failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	checked := true
	assessmentOutcomes, err := da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs, &checked)
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceReport: da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDs: get assessment outcomes failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
			log.Bool("checked", checked),
		)
		return nil, err
	}

	outcomeAttendances, err := rm.getOutcomeAttendances(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceReport: call getOutcomeAttendances failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	outcomeNamesMap, err := rm.getOutcomeNamesMap(ctx, tx, operator, rm.getOutcomeIDs(assessmentOutcomes))
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceReport: call getOutcomeNamesMap failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}
	trOutcomeNames := func(ids []string) []string {
		var names []string
		for _, id := range ids {
			names = append(names, outcomeNamesMap[id])
		}
		return names
	}

	trLatestOutcomeIDs, err := rm.makeLatestOutcomeIDsTranslator(ctx, tx, operator, rm.getOutcomeIDs(assessmentOutcomes))
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceReport: call makeLatestOutcomeIDsTranslator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Any("assessment_outcomes", assessmentOutcomes),
		)
		return nil, err
	}

	attendanceIDExistsMap := rm.getAttendanceIDsExistMap(assessmentAttendances)
	attendanceID2OutcomeIDsMap := rm.getAttendanceID2OutcomeIDsMap(assessmentAttendances, assessmentOutcomes)
	achievedAttendanceID2OutcomeIDsMap := rm.getAchievedAttendanceID2OutcomeIDsMap(outcomeAttendances)
	skipAttendanceID2OutcomeIDsMap := rm.getSkipAttendanceID2OutcomeIDsMap(assessmentAttendances, assessmentOutcomes)
	notAchievedAttendanceID2OutcomeIDsMap := rm.getNotAchievedAttendanceID2OutcomeIDsMap(attendanceID2OutcomeIDsMap, achievedAttendanceID2OutcomeIDsMap, skipAttendanceID2OutcomeIDsMap)
	log.Debug(ctx, "ListStudentsPerformanceReport: print all map",
		log.Any("attendance_id_exists_map", attendanceIDExistsMap),
		log.Any("attendance_id_2_outcome_ids_map", attendanceID2OutcomeIDsMap),
		log.Any("achieved_attendance_id_2_outcome_ids_map", achievedAttendanceID2OutcomeIDsMap),
		log.Any("skip_attendance_id_2_outcome_ids_map", skipAttendanceID2OutcomeIDsMap),
		log.Any("not_achieved_attendance_id_2_outcome_ids_map", notAchievedAttendanceID2OutcomeIDsMap),
	)

	var result = entity.ListStudentsPerformanceReportResponse{AssessmentIDs: assessmentIDs}
	for _, student := range students {
		newItem := entity.StudentsPerformanceReportItem{StudentID: student.ID, StudentName: student.Name}
		if !attendanceIDExistsMap[student.ID] {
			newItem.Attend = false
			result.Items = append(result.Items, &newItem)
			continue
		}
		newItem.Attend = true
		newItem.AchievedNames = trOutcomeNames(trLatestOutcomeIDs(achievedAttendanceID2OutcomeIDsMap[student.ID]))
		newItem.NotAchievedNames = trOutcomeNames(trLatestOutcomeIDs(notAchievedAttendanceID2OutcomeIDsMap[student.ID]))
		newItem.NotAttemptedNames = trOutcomeNames(trLatestOutcomeIDs(skipAttendanceID2OutcomeIDsMap[student.ID]))
		result.Items = append(result.Items, &newItem)
	}

	return &result, nil
}

func (rm *reportModel) GetStudentPerformanceReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.GetStudentPerformanceReportRequest) (*entity.GetStudentPerformanceReportResponse, error) {
	{
		if req.ClassID == "" {
			log.Error(ctx, "GetStudentPerformanceReport: require class id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		if req.TeacherID == "" {
			log.Error(ctx, "GetStudentPerformanceReport: require teacher id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		if req.LessonPlanID == "" {
			log.Error(ctx, "GetStudentPerformanceReport: require lesson plan id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		if req.StudentID == "" {
			log.Error(ctx, "GetStudentPerformanceReport: require student id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		allowed, err := rm.hasReportPermission(ctx, operator, req.TeacherID)
		if err != nil {
			log.Error(ctx, "GetStudentPerformanceReport: check report report permission failed",
				log.Any("operator", operator),
				log.Any("req", req),
			)
			return nil, err
		}
		if !allowed {
			log.Error(ctx, "GetStudentPerformanceReport: no permission",
				log.Any("operator", operator),
				log.Any("req", req),
			)
			return nil, constant.ErrUnAuthorized
		}
	}

	student, err := rm.getStudentInClass(ctx, operator, req.ClassID, req.StudentID)
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceReport: call getStudentInClass failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}

	assessments, err := rm.getCompletedAssessments(ctx, tx, operator, req.ClassID, req.LessonPlanID)
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceReport: call getCompletedAssessments failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}
	assessmentIDs := make([]string, 0, len(assessments))
	scheduleIDs := make([]string, 0, len(assessments))
	for _, a := range assessments {
		assessmentIDs = append(assessmentIDs, a.ID)
		scheduleIDs = append(scheduleIDs, a.ScheduleID)
	}

	schedules, err := GetScheduleModel().GetByIDs(ctx, operator, scheduleIDs)
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceReport: call GetScheduleDA().Query failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}
	schedulesMap := map[string]*entity.SchedulePlain{}
	for _, s := range schedules {
		schedulesMap[s.ID] = s
	}

	checked := true
	assessmentOutcomes, err := da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs, &checked)
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceReport: da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDs: get assessment outcomes failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
			log.Bool("checked", checked),
		)
		return nil, err
	}

	assessmentAttendances, err := rm.getCheckedAssessmentAttendance(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceReport: call getCheckedAssessmentAttendance failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	outcomeAttendances, err := da.GetOutcomeAttendanceDA().BatchGetByAssessmentIDsAndAttendanceID(ctx, tx, assessmentIDs, req.StudentID)
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceReport: call BatchGetByAssessmentIDsAndAttendanceID failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	outcomeIDs := rm.getOutcomeIDs(assessmentOutcomes)
	outcomeNamesMap, err := rm.getOutcomeNamesMap(ctx, tx, operator, outcomeIDs)
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceReport: call getOutcomeNamesMap failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("outcome_ids", outcomeIDs),
		)
		return nil, err
	}
	trOutcomeNames := func(ids []string) []string {
		var names []string
		for _, id := range ids {
			names = append(names, outcomeNamesMap[id])
		}
		return names
	}

	tr, err := rm.makeLatestOutcomeIDsTranslator(ctx, tx, operator, outcomeIDs)
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceReport: call makeLatestOutcomeIDsTranslator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("outcome_ids", outcomeIDs),
		)
		return nil, err
	}

	assessmentID2OutcomeIDsMap := rm.getAssessmentID2OutcomeIDsMap(assessmentOutcomes)
	attendanceIDExistsMap := rm.getAttendanceIDsExistMap(assessmentAttendances)
	achievedAssessmentID2OutcomeIDsMap := rm.getAchievedAssessmentID2OutcomeIDsMap(student.ID, assessmentOutcomes, outcomeAttendances)
	skipAssessmentID2OutcomeIDsMap := rm.getSkipAssessmentID2OutcomeIDsMap(student.ID, assessmentAttendances, assessmentOutcomes)
	notAchievedAssessmentID2OutcomeIDsMap := rm.getNotAchievedAssessmentID2OutcomeIDsMap(assessmentID2OutcomeIDsMap, achievedAssessmentID2OutcomeIDsMap, skipAssessmentID2OutcomeIDsMap)
	log.Debug(ctx, "GetStudentPerformanceReport: print all map",
		log.Any("assessmentID2OutcomeIDsMap", assessmentID2OutcomeIDsMap),
		log.Any("attendance_id_exists_map", attendanceIDExistsMap),
		log.Any("achievedAssessmentID2OutcomeIDsMap", achievedAssessmentID2OutcomeIDsMap),
		log.Any("skipAssessmentID2OutcomeIDsMap", skipAssessmentID2OutcomeIDsMap),
		log.Any("notAchievedAssessmentID2OutcomeIDsMap", notAchievedAssessmentID2OutcomeIDsMap),
	)

	result := entity.GetStudentPerformanceReportResponse{AssessmentIDs: assessmentIDs}
	for _, a := range assessments {
		newItem := entity.StudentPerformanceReportItem{
			StudentID:   student.ID,
			StudentName: student.Name,
			Attend:      attendanceIDExistsMap[req.StudentID],
			ScheduleID:  a.ScheduleID,
		}
		if s := schedulesMap[a.ScheduleID]; s != nil {
			newItem.ScheduleStartTime = s.StartAt
		}
		if newItem.Attend {
			newItem.AchievedNames = trOutcomeNames(tr(achievedAssessmentID2OutcomeIDsMap[a.ID]))
			newItem.NotAchievedNames = trOutcomeNames(tr(notAchievedAssessmentID2OutcomeIDsMap[a.ID]))
			newItem.NotAttemptedNames = trOutcomeNames(tr(skipAssessmentID2OutcomeIDsMap[a.ID]))
		}
		result.Items = append(result.Items, &newItem)
	}

	return &result, nil
}

func (rm *reportModel) getCompletedAssessmentIDs(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, classID string, lessonPlanID string) ([]string, error) {
	ids, err := rm.getAssessmentIDs(ctx, tx, operator, classID, lessonPlanID)
	if err != nil {
		log.Error(ctx, "get assessment ids failed",
			log.Err(err),
			log.String("class_id", classID),
			log.String("lesson_plan_id", lessonPlanID),
			log.Any("operator", operator),
		)
		return nil, err
	}
	result, err := da.GetAssessmentDA().FilterCompletedAssessmentIDs(ctx, tx, ids)
	if err != nil {
		log.Error(ctx, "filter completed assessment ids failed",
			log.Any("operator", operator),
			log.Strings("ids", ids),
		)
		return nil, err
	}
	return result, nil
}

func (rm *reportModel) getCompletedAssessments(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, classID string, lessonPlanID string) ([]*entity.Assessment, error) {
	ids, err := rm.getAssessmentIDs(ctx, tx, operator, classID, lessonPlanID)
	if err != nil {
		log.Error(ctx, "getCompletedAssessments: call getAssessmentIDs failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("class_id", classID),
			log.String("lesson_plan_id", lessonPlanID),
		)
		return nil, err
	}
	result, err := da.GetAssessmentDA().FilterCompletedAssessments(ctx, tx, ids)
	if err != nil {
		log.Error(ctx, "getCompletedAssessments: call FilterCompletedAssessments failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("class_id", classID),
			log.String("lesson_plan_id", lessonPlanID),
			log.Strings("ids", ids),
		)
		return nil, err
	}
	return result, nil
}

func (rm *reportModel) getAssessmentIDs(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, classID string, lessonPlanID string) ([]string, error) {
	scheduleIDs, err := rm.getScheduleIDs(ctx, tx, operator, classID, lessonPlanID)
	if err != nil {
		log.Error(ctx, "get assessment ids: get schedule ids failed",
			log.Err(err),
			log.String("class_id", classID),
			log.String("lesson_plan_id", lessonPlanID),
		)
		return nil, err
	}
	assessments, err := da.GetAssessmentDA().BatchGetAssessmentsByScheduleIDs(ctx, tx, scheduleIDs)
	if err != nil {
		log.Error(ctx, "get assessment ids: batch get assessment failed by schedule ids",
			log.Err(err),
			log.Any("cmd", "cmd"),
		)
		return nil, err
	}
	var result []string
	for _, assessment := range assessments {
		result = append(result, assessment.ID)
	}
	return result, nil
}

func (rm *reportModel) getScheduleIDs(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, classID string, lessonPlanID string) ([]string, error) {
	log.Debug(ctx, "get schedule ids: before call GetScheduleModel().Query()")
	result, err := GetScheduleModel().GetScheduleIDsByCondition(ctx, tx, operator, &entity.ScheduleIDsCondition{
		ClassID:      classID,
		LessonPlanID: lessonPlanID,
		StartAt:      time.Now().Add(constant.ScheduleAllowGoLiveTime).Unix(),
	})
	log.Debug(ctx, "get schedule ids: after call GetScheduleModel().Query()")
	if err != nil {
		log.Error(ctx, "get schedule ids: query failed",
			log.Err(err),
			log.String("class_id", classID),
			log.String("lesson_plan_id", lessonPlanID),
		)
		return nil, err
	}
	log.Debug(ctx, "get schedule ids success", log.Any("result", result))
	return result, nil
}

func (rm *reportModel) getCheckedAssessmentAttendance(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string) ([]*entity.AssessmentAttendance, error) {
	var (
		result  []*entity.AssessmentAttendance
		checked = true
	)
	if err := da.GetAssessmentAttendanceDA().QueryTx(ctx, tx, &da.AssessmentAttendanceCondition{
		AssessmentIDs: assessmentIDs,
		Checked:       &checked,
	}, &result); err != nil {
		log.Error(ctx, "getCheckedAssessmentAttendance: query assessment attendances failed",
			log.Err(err),
			log.Any("assessment_ids", assessmentIDs),
		)
		return nil, err
	}
	return result, nil
}

//func (rm *reportModel) getAssessmentOutcomes(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string) ([]*entity.AssessmentOutcome, error) {
//	result, err := da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs)
//	if err != nil {
//		log.Error(ctx, "getAssessmentOutcomes: batch get assessment outcomes failed",
//			log.Err(err),
//			log.Any("assessment_ids", assessmentIDs),
//		)
//		return nil, err
//	}
//	return result, nil
//}

func (rm *reportModel) getOutcomeAttendances(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string) ([]*entity.OutcomeAttendance, error) {
	result, err := da.GetOutcomeAttendanceDA().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "getOutcomeAttendances: call BatchGetByAssessmentIDs failed",
			log.Err(err),
			log.Any("assessment_ids", assessmentIDs),
		)
		return nil, err
	}
	return result, nil
}

func (rm *reportModel) getOutcomeIDs(assessmentOutcomes []*entity.AssessmentOutcome) []string {
	result := make([]string, 0, len(assessmentOutcomes))
	for _, v := range assessmentOutcomes {
		result = append(result, v.OutcomeID)
	}
	return utils.SliceDeduplication(result)
}

func (rm *reportModel) getOutcomesMap(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, outcomeIDs []string) (map[string]*entity.Outcome, error) {
	outcomes, err := GetOutcomeModel().GetLearningOutcomesByIDs(ctx, tx, outcomeIDs, operator)
	if err != nil {
		log.Error(ctx, "get student detail report: get learning outcome failed by ids",
			log.Err(err),
			log.Any("operator", operator),
			log.Strings("outcome_ids", outcomeIDs),
		)
		return nil, err
	}
	m := map[string]*entity.Outcome{}
	for _, outcome := range outcomes {
		m[outcome.ID] = outcome
	}
	return m, nil
}

func (rm *reportModel) getOutcomeNamesMap(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, outcomeIDs []string) (map[string]string, error) {
	outcomes, err := GetOutcomeModel().GetLearningOutcomesByIDs(ctx, tx, outcomeIDs, operator)
	if err != nil {
		log.Error(ctx, "get student detail report: get learning outcome failed by ids",
			log.Err(err),
			log.Any("operator", operator),
			log.Strings("outcome_ids", outcomeIDs),
		)
		return nil, err
	}
	m := map[string]string{}
	for _, outcome := range outcomes {
		m[outcome.ID] = outcome.Name
	}
	return m, nil
}

func (rm *reportModel) getAttendanceIDsExistMap(assessmentAttendances []*entity.AssessmentAttendance) map[string]bool {
	result := make(map[string]bool, len(assessmentAttendances))
	for _, assessmentAttendance := range assessmentAttendances {
		result[assessmentAttendance.AttendanceID] = true
	}
	return result
}

func (rm *reportModel) getAttendanceID2AssessmentOutcomesMap(assessmentAttendances []*entity.AssessmentAttendance, assessmentOutcomes []*entity.AssessmentOutcome) map[string][]*entity.AssessmentOutcome {
	result := map[string][]*entity.AssessmentOutcome{}

	var attendanceID2AssessmentIDsMap = map[string][]string{}
	for _, assessmentAttendance := range assessmentAttendances {
		attendanceID2AssessmentIDsMap[assessmentAttendance.AttendanceID] = append(attendanceID2AssessmentIDsMap[assessmentAttendance.AttendanceID], assessmentAttendance.AssessmentID)
	}

	var assessmentID2AssessmentOutcomesMap = map[string][]*entity.AssessmentOutcome{}
	for _, assessmentOutcome := range assessmentOutcomes {
		assessmentID2AssessmentOutcomesMap[assessmentOutcome.AssessmentID] = append(assessmentID2AssessmentOutcomesMap[assessmentOutcome.AssessmentID], assessmentOutcome)
	}

	for attendanceID, assessmentIDs := range attendanceID2AssessmentIDsMap {
		for _, assessmentID := range assessmentIDs {
			result[attendanceID] = append(result[attendanceID], assessmentID2AssessmentOutcomesMap[assessmentID]...)
		}
	}

	return result
}

func (rm *reportModel) getAttendanceID2OutcomeIDsMap(assessmentAttendances []*entity.AssessmentAttendance, assessmentOutcomes []*entity.AssessmentOutcome) map[string][]string {
	attendanceID2AssessmentOutcomesMap := rm.getAttendanceID2AssessmentOutcomesMap(assessmentAttendances, assessmentOutcomes)
	result := map[string][]string{}
	for attendanceID, assessmentOutcomes := range attendanceID2AssessmentOutcomesMap {
		for _, assessmentOutcome := range assessmentOutcomes {
			result[attendanceID] = append(result[attendanceID], assessmentOutcome.OutcomeID)
		}
	}
	for attendanceID, outcomeIDs := range result {
		result[attendanceID] = utils.SliceDeduplication(outcomeIDs)
	}
	return result
}

func (rm *reportModel) getAchievedAttendanceID2OutcomeIDsMap(outcomeAttendances []*entity.OutcomeAttendance) map[string][]string {
	result := map[string][]string{}
	for _, outcomeAttendance := range outcomeAttendances {
		result[outcomeAttendance.AttendanceID] = append(result[outcomeAttendance.AttendanceID], outcomeAttendance.OutcomeID)
	}
	for k, v := range result {
		result[k] = utils.SliceDeduplication(v)
	}
	return result
}

func (rm *reportModel) getSkipAttendanceID2OutcomeIDsMap(assessmentAttendances []*entity.AssessmentAttendance, assessmentOutcomes []*entity.AssessmentOutcome) map[string][]string {
	attendanceID2AssessmentOutcomesMap := rm.getAttendanceID2AssessmentOutcomesMap(assessmentAttendances, assessmentOutcomes)
	result := map[string][]string{}
	for attendanceID, assessmentOutcomes := range attendanceID2AssessmentOutcomesMap {
		skipOutcomeIDsMap := map[string]bool{}
		for _, assessmentOutcome := range assessmentOutcomes {
			_, exists := skipOutcomeIDsMap[assessmentOutcome.OutcomeID]
			if !exists && assessmentOutcome.Skip {
				skipOutcomeIDsMap[assessmentOutcome.OutcomeID] = true
			}
			if !assessmentOutcome.Skip {
				skipOutcomeIDsMap[assessmentOutcome.OutcomeID] = false
			}
		}
		for skipOutcomeID, ok := range skipOutcomeIDsMap {
			if ok {
				result[attendanceID] = append(result[attendanceID], skipOutcomeID)
			}
		}
	}
	return result
}

func (rm *reportModel) getNotAchievedAttendanceID2OutcomeIDsMap(attendanceID2OutcomeIDsMap, achievedAttendanceID2OutcomeIDsMap, skipAttendanceID2OutcomeIDsMap map[string][]string) map[string][]string {
	result := map[string][]string{}
	for attendanceID, outcomeIDs := range attendanceID2OutcomeIDsMap {
		var excludeOutcomeIDs []string
		excludeOutcomeIDs = append(excludeOutcomeIDs, achievedAttendanceID2OutcomeIDsMap[attendanceID]...)
		excludeOutcomeIDs = append(excludeOutcomeIDs, skipAttendanceID2OutcomeIDsMap[attendanceID]...)
		result[attendanceID] = utils.ExcludeStrings(outcomeIDs, excludeOutcomeIDs)
	}
	return result
}

func (rm *reportModel) getAssessmentID2OutcomeIDsMap(assessmentOutcomes []*entity.AssessmentOutcome) map[string][]string {
	result := map[string][]string{}
	for _, ao := range assessmentOutcomes {
		result[ao.AssessmentID] = append(result[ao.AssessmentID], ao.OutcomeID)
	}
	return result
}

func (rm *reportModel) getAchievedAssessmentID2OutcomeIDsMap(attendanceID string, assessmentOutcomes []*entity.AssessmentOutcome, outcomeAttendances []*entity.OutcomeAttendance) map[string][]string {
	achievedOutcomeIDsMap := map[string]map[string]bool{}
	for _, oa := range outcomeAttendances {
		if oa.AttendanceID != attendanceID {
			continue
		}
		if achievedOutcomeIDsMap[oa.AssessmentID] == nil {
			achievedOutcomeIDsMap[oa.AssessmentID] = map[string]bool{}
		}
		achievedOutcomeIDsMap[oa.AssessmentID][oa.OutcomeID] = true
	}

	result := map[string][]string{}
	for _, ao := range assessmentOutcomes {
		if achievedOutcomeIDsMap[ao.AssessmentID] != nil && achievedOutcomeIDsMap[ao.AssessmentID][ao.OutcomeID] {
			result[ao.AssessmentID] = append(result[ao.AssessmentID], ao.OutcomeID)
		}
	}
	for k, v := range result {
		result[k] = utils.SliceDeduplication(v)
	}
	return result
}

func (rm *reportModel) getSkipAssessmentID2OutcomeIDsMap(attendanceID string, assessmentAttendances []*entity.AssessmentAttendance, assessmentOutcomes []*entity.AssessmentOutcome) map[string][]string {
	attendanceID2AssessmentOutcomesMap := rm.getAttendanceID2AssessmentOutcomesMap(assessmentAttendances, assessmentOutcomes)
	assessmentOutcomes = attendanceID2AssessmentOutcomesMap[attendanceID]
	result := map[string][]string{}
	for _, ao := range assessmentOutcomes {
		if ao.Skip {
			result[ao.AssessmentID] = append(result[ao.AssessmentID], ao.OutcomeID)
		}
	}
	return result
}

func (rm *reportModel) getNotAchievedAssessmentID2OutcomeIDsMap(assessmentID2OutcomeIDsMap, achievedAssessmentID2OutcomeIDsMap, skipAssessmentID2OutcomeIDsMap map[string][]string) map[string][]string {
	result := map[string][]string{}
	for assessmentID, outcomeIDs := range assessmentID2OutcomeIDsMap {
		var excludeOutcomeIDs []string
		excludeOutcomeIDs = append(excludeOutcomeIDs, achievedAssessmentID2OutcomeIDsMap[assessmentID]...)
		excludeOutcomeIDs = append(excludeOutcomeIDs, skipAssessmentID2OutcomeIDsMap[assessmentID]...)
		result[assessmentID] = utils.ExcludeStrings(outcomeIDs, excludeOutcomeIDs)
	}
	return result
}

func (rm *reportModel) makeLatestOutcomeIDsTranslator(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, outcomeIDs []string) (func([]string) []string, error) {
	m, err := GetOutcomeModel().GetLatestOutcomesByIDsMapResult(ctx, tx, outcomeIDs, operator)
	if err != nil {
		if err != constant.ErrRecordNotFound {
			log.Error(ctx, "make latest outcome id translator: call outcome model failed",
				log.Err(err),
				log.Any("outcome_ids", outcomeIDs),
				log.Any("operator", operator),
			)
			return nil, err
		} else {
			m = map[string]*entity.Outcome{}
		}
	}
	return func(ids []string) []string {
		if len(ids) == 0 {
			return nil
		}
		var result []string
		for _, id := range ids {
			if v, ok := m[id]; ok {
				result = append(result, v.ID)
			} else {
				result = append(result, id)
			}
		}
		return utils.SliceDeduplication(result)
	}, nil
}

func (rm *reportModel) hasReportPermission(ctx context.Context, operator *entity.Operator, teacherID string) (bool, error) {
	checkP603, err := rm.checkPermission603(ctx, operator, teacherID)
	if err != nil {
		return false, err
	}
	if !checkP603 {
		return false, nil
	}

	optionalCheckers := []func(context.Context, *entity.Operator, string) (bool, error){
		rm.checkPermission614,
		rm.checkPermission610,
		rm.checkPermission611,
		rm.checkPermission612,
	}
	for _, check := range optionalCheckers {
		ok, err := check(ctx, operator, teacherID)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	return false, nil
}

func (rm *reportModel) checkPermission603(ctx context.Context, operator *entity.Operator, teacherID string) (bool, error) {
	hasP603, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.ReportTeacherReports603)
	if err != nil {
		log.Error(ctx, "check permission 603 failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("teacher_id", teacherID),
		)
		return false, err
	}
	if hasP603 {
		return true, nil
	}
	return false, nil
}

func (rm *reportModel) checkPermission614(ctx context.Context, operator *entity.Operator, teacherID string) (bool, error) {
	hasP614, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.ReportViewMyReports614)
	if err != nil {
		log.Error(ctx, "check permission 614 failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("teacher_id", teacherID),
		)
		return false, err
	}
	if hasP614 && operator.UserID == teacherID {
		return true, nil
	}
	return false, nil
}

func (rm *reportModel) checkPermission610(ctx context.Context, operator *entity.Operator, teacherID string) (bool, error) {
	hasP610, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.ReportViewReports610)
	if err != nil {
		log.Error(ctx, "check permission 610 failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("teacher_id", teacherID),
		)
		return false, err
	}

	if hasP610 {
		var validTeacherIDs []string
		teachers, err := external.GetTeacherServiceProvider().GetByOrganization(ctx, operator, operator.OrgID)
		if err != nil {
			log.Error(ctx, "check permission 610: call external \"GetByOrganization()\" failed",
				log.Err(err),
				log.Any("operator", operator),
				log.Any("teacher_id", teacherID),
			)
			return false, err
		}
		for _, teacher := range teachers {
			validTeacherIDs = append(validTeacherIDs, teacher.ID)
		}
		var schoolIDs []string
		schools, err := external.GetSchoolServiceProvider().GetByOperator(ctx, operator)
		for _, school := range schools {
			schoolIDs = append(schoolIDs, school.ID)
		}
		schoolID2TeachersMap, err := external.GetTeacherServiceProvider().GetBySchools(ctx, operator, schoolIDs)
		if err != nil {
			log.Error(ctx, "check permission 610: call external \"GetBySchools()\" failed",
				log.Err(err),
				log.Any("operator", operator),
				log.Any("teacher_id", teacherID),
			)
			return false, err
		}
		for _, teachers := range schoolID2TeachersMap {
			for _, teacher := range teachers {
				validTeacherIDs = append(validTeacherIDs, teacher.ID)
			}
		}
		for _, item := range validTeacherIDs {
			if item == teacherID {
				return true, nil
			}
		}
	}

	return false, nil
}

func (rm *reportModel) checkPermission611(ctx context.Context, operator *entity.Operator, teacherID string) (bool, error) {
	hasP611, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.ReportViewMySchoolReports611)
	if err != nil {
		log.Error(ctx, "check permission 611 failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("teacher_id", teacherID),
		)
		return false, err
	}

	if hasP611 {
		var validTeacherIDs []string
		var schoolIDs []string
		schools, err := external.GetSchoolServiceProvider().GetByOperator(ctx, operator)
		for _, school := range schools {
			schoolIDs = append(schoolIDs, school.ID)
		}
		schoolID2TeachersMap, err := external.GetTeacherServiceProvider().GetBySchools(ctx, operator, schoolIDs)
		if err != nil {
			log.Error(ctx, "check permission 611: call external \"GetBySchools()\" failed",
				log.Err(err),
				log.Any("operator", operator),
				log.Any("teacher_id", teacherID),
			)
			return false, err
		}
		for _, teachers := range schoolID2TeachersMap {
			for _, teacher := range teachers {
				validTeacherIDs = append(validTeacherIDs, teacher.ID)
			}
		}
		for _, item := range validTeacherIDs {
			if item == teacherID {
				return true, nil
			}
		}
	}

	return false, nil
}

func (rm *reportModel) checkPermission612(ctx context.Context, operator *entity.Operator, teacherID string) (bool, error) {
	hasP612, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.ReportViewMyOrganizationsReports612)
	if err != nil {
		log.Error(ctx, "check permission 612 failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("teacher_id", teacherID),
		)
		return false, err
	}

	if hasP612 {
		var validTeacherIDs []string
		teachers, err := external.GetTeacherServiceProvider().GetByOrganization(ctx, operator, operator.OrgID)
		if err != nil {
			log.Error(ctx, "check permission 612: call external \"GetByOrganization()\" failed",
				log.Err(err),
				log.Any("operator", operator),
				log.Any("teacher_id", teacherID),
			)
			return false, err
		}
		for _, teacher := range teachers {
			validTeacherIDs = append(validTeacherIDs, teacher.ID)
		}
		for _, item := range validTeacherIDs {
			if item == teacherID {
				return true, nil
			}
		}
	}

	return false, nil
}

func (rm *reportModel) getStudentsInClass(ctx context.Context, operator *entity.Operator, classID string) ([]*external.Student, error) {
	result, err := external.GetStudentServiceProvider().GetByClassID(ctx, operator, classID)
	if err != nil {
		log.Error(ctx, "getStudentsInClass: call GetByClassID failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("class_id", classID),
		)
		return nil, err
	}
	return result, nil
}

func (rm *reportModel) getStudentInClass(ctx context.Context, operator *entity.Operator, classID string, studentID string) (*external.Student, error) {
	students, err := rm.getStudentsInClass(ctx, operator, classID)
	if err != nil {
		log.Error(ctx, "getStudentInClass: call getStudentsInClass failed",
			log.Err(err),
			log.String("class_id", classID),
			log.String("student_id", studentID),
		)
		return nil, err
	}
	for _, item := range students {
		if item.ID == studentID {
			return item, nil
		}
	}
	return nil, constant.ErrRecordNotFound
}

// endregion

// region h5p

func (rm *reportModel) ListStudentsPerformanceH5PReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.ListStudentsPerformanceH5PReportRequest) (*entity.ListStudentsPerformanceH5PReportResponse, error) {
	{
		if req.ClassID == "" {
			log.Error(ctx, "ListStudentsPerformanceH5PReport: require class id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		if req.TeacherID == "" {
			log.Error(ctx, "ListStudentsPerformanceH5PReport: require teacher id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		if req.LessonPlanID == "" {
			log.Error(ctx, "ListStudentsPerformanceH5PReport: require lesson plan id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		allowed, err := rm.hasReportPermission(ctx, operator, req.TeacherID)
		if err != nil {
			log.Error(ctx, "ListStudentsPerformanceH5PReport: check report report permission failed",
				log.Any("req", req),
				log.Any("operator", operator),
			)
			return nil, err
		}
		if !allowed {
			log.Error(ctx, "ListStudentsPerformanceH5PReport: no permission",
				log.Any("req", req),
				log.Any("operator", operator),
			)
			return nil, constant.ErrForbidden
		}
	}

	pastLessonPlanIDs, err := GetContentModel().GetPastContentIDByID(ctx, tx, req.LessonPlanID)
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceH5PReport: call getLessonPlanH5PMaterials failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.String("lesson_plan_id", req.LessonPlanID),
		)
		return nil, err
	}
	finalLessonPlainIDs := make([]string, 0, len(pastLessonPlanIDs)+1)
	finalLessonPlainIDs = append(finalLessonPlainIDs, req.LessonPlanID)
	finalLessonPlainIDs = append(finalLessonPlainIDs, pastLessonPlanIDs...)

	materialIDs, err := rm.getLessonPlanH5PMaterialIDs(ctx, tx, operator, finalLessonPlainIDs)
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceH5PReport: call getLessonPlanH5PMaterialIDs failed",
			log.Err(err),
			log.Any("req", req),
			log.Strings("lesson_plan_ids", finalLessonPlainIDs),
		)
		return nil, err
	}

	students, err := rm.getStudentsInClass(ctx, operator, req.ClassID)
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceH5PReport: call getStudentsInClass failed",
			log.Err(err),
			log.Any("req", req),
			log.String("class_id", req.ClassID),
		)
		return nil, err
	}
	studentIDs := make([]string, 0, len(students))
	studentNamesMap := map[string]string{}
	for _, s := range students {
		studentIDs = append(studentIDs, s.ID)
		studentNamesMap[s.ID] = s.Name
	}

	attendanceIDsExistMap, err := rm.getAttendanceIDsExistMapByClassIDAndLessonPlanID(ctx, tx, operator, req.ClassID, req.LessonPlanID)
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceH5PReport: call getAttendanceIDsExistMapByClassIDAndLessonPlanID failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.String("class_id", req.ClassID),
			log.String("lesson_plan_id", req.LessonPlanID),
		)
		return nil, err
	}

	h5pEventCond := da.H5PEventCondition{
		LessonPlanIDs: finalLessonPlainIDs,
		MaterialIDs:   materialIDs,
		UserIDs:       studentIDs,
		VerbIDs:       []string{constant.ActivityEventVerbIDInitGame, constant.ActivityEventVerbIDAnswered, constant.ActivityEventVerbIDCompleted},
	}
	events, err := GetH5PEventModel().Query(ctx, tx, operator, h5pEventCond)
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceH5PReport: call GetH5PEventModel().Query failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Any("h5p_cond", h5pEventCond),
		)
		return nil, err
	}
	usersSpentTimeMap := rm.calculateUsersSpentTimeMap(events)

	r := entity.ListStudentsPerformanceH5PReportResponse{}
	for uid, spentTime := range usersSpentTimeMap {
		v := entity.StudentsPerformanceH5PReportItem{
			StudentID:   uid,
			StudentName: studentNamesMap[uid],
			Attend:      attendanceIDsExistMap[uid],
			SpentTime:   spentTime,
		}
		r.Items = append(r.Items, &v)
	}

	return &r, nil
}

func (rm *reportModel) GetStudentPerformanceH5PReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.GetStudentPerformanceH5PReportRequest) (*entity.GetStudentPerformanceH5PReportResponse, error) {
	{
		if req.ClassID == "" {
			log.Error(ctx, "GetStudentPerformanceReport: require class id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		if req.TeacherID == "" {
			log.Error(ctx, "GetStudentPerformanceReport: require teacher id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		if req.LessonPlanID == "" {
			log.Error(ctx, "GetStudentPerformanceReport: require lesson plan id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		if req.StudentID == "" {
			log.Error(ctx, "GetStudentPerformanceReport: require student id", log.Any("req", req))
			return nil, constant.ErrInvalidArgs
		}
		allowed, err := rm.hasReportPermission(ctx, operator, req.TeacherID)
		if err != nil {
			log.Error(ctx, "GetStudentPerformanceReport: check report report permission failed",
				log.Any("operator", operator),
				log.Any("req", req),
			)
			return nil, err
		}
		if !allowed {
			log.Error(ctx, "GetStudentPerformanceReport: no permission",
				log.Any("operator", operator),
				log.Any("req", req),
			)
			return nil, constant.ErrUnAuthorized
		}
	}

	pastLessonPlanIDs, err := GetContentModel().GetPastContentIDByID(ctx, tx, req.LessonPlanID)
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceH5PReport: call getLessonPlanH5PMaterials failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.String("lesson_plan_id", req.LessonPlanID),
		)
		return nil, err
	}
	finalLessonPlainIDs := make([]string, 0, len(pastLessonPlanIDs)+1)
	finalLessonPlainIDs = append(finalLessonPlainIDs, req.LessonPlanID)
	finalLessonPlainIDs = append(finalLessonPlainIDs, pastLessonPlanIDs...)

	materials, err := rm.getLessonPlanH5PMaterials(ctx, tx, operator, finalLessonPlainIDs)
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceH5PReport: call getLessonPlanH5PMaterials failed",
			log.Err(err),
			log.Any("req", req),
			log.Strings("lesson_plan_ids", finalLessonPlainIDs),
		)
		return nil, err
	}
	var materialIDs []string
	for _, m := range materials {
		materialIDs = append(materialIDs, m.ID)
	}

	h5pEventCond := da.H5PEventCondition{
		LessonPlanIDs: finalLessonPlainIDs,
		MaterialIDs:   materialIDs,
		UserIDs:       []string{req.StudentID},
		VerbIDs:       []string{constant.ActivityEventVerbIDInitGame, constant.ActivityEventVerbIDAnswered, constant.ActivityEventVerbIDCompleted},
	}
	events, err := GetH5PEventModel().Query(ctx, tx, operator, h5pEventCond)
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceH5PReport: call GetH5PEventModel().Query failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Any("h5p_cond", h5pEventCond),
		)
		return nil, err
	}
	h5pEventCond2 := da.H5PEventCondition{
		LessonPlanIDs:     h5pEventCond.LessonPlanIDs,
		MaterialIDs:       h5pEventCond.MaterialIDs,
		UserIDs:           h5pEventCond.UserIDs,
		VerbIDs:           []string{constant.ActivityEventVerbIDInteracted},
		LocalLibraryNames: []string{string(entity.ActivityTypeMemoryGame)},
	}
	events2, err := GetH5PEventModel().Query(ctx, tx, operator, h5pEventCond2)
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceH5PReport: call GetH5PEventModel().Query failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Any("h5p_cond_2", h5pEventCond2),
		)
		return nil, err
	}
	events = append(events, events2...)
	log.Debug(ctx, "GetStudentPerformanceH5PReport: print events data", log.Any("events", events))

	r := entity.GetStudentPerformanceH5PReportResponse{}
	for _, m := range materials {
		v := entity.StudentPerformanceH5PReportItem{
			MaterialID:   m.ID,
			MaterialName: m.Name,
		}
		meta := (m.Data.(*MaterialData)).Content
		v.ActivityType = entity.ParseActivityType(gjson.Get(meta, constant.H5PGJSONPathLibrary).String())
		log.Debug(ctx, "GetStudentPerformanceH5PReport: loop print activity type and meta",
			log.String("activity_type", string(v.ActivityType)),
			log.String("meta", meta),
		)
		switch v.ActivityType {
		case entity.ActivityTypeImageSequencing:
			data, err := rm.getActivityImageSequencing(m.ID, meta, events)
			if err != nil {
				log.Error(ctx, "GetStudentPerformanceH5PReport: call getActivityImageSequencing failed",
					log.Err(err),
					log.Any("operator", operator),
					log.Any("req", req),
					log.String("meta", meta),
					log.Any("events", events),
				)
				return nil, err
			}
			v.ActivityImageSequencing = data
			for _, r := range data.PlayRecords {
				v.TotalSpentTime += r.Duration
			}
			if len(data.PlayRecords) != 0 {
				v.AvgSpentTime = v.TotalSpentTime / int64(len(data.PlayRecords))
			}
		case entity.ActivityTypeMemoryGame:
			data, err := rm.getActivityMemoryGame(m.ID, meta, events)
			if err != nil {
				log.Error(ctx, "GetStudentPerformanceH5PReport: call getActivityMemoryGame failed",
					log.Err(err),
					log.Any("operator", operator),
					log.Any("req", req),
					log.String("meta", meta),
					log.Any("events", events),
				)
				return nil, err
			}
			v.ActivityMemoryGame = data
			for _, r := range data.PlayRecords {
				v.TotalSpentTime += r.Duration
			}
			if len(data.PlayRecords) != 0 {
				v.AvgSpentTime = v.TotalSpentTime / int64(len(data.PlayRecords))
			}
		case entity.ActivityTypeImagePair:
			data, err := rm.getActivityImagePair(m.ID, meta, events)
			if err != nil {
				log.Error(ctx, "GetStudentPerformanceH5PReport: call getActivityImagePair failed",
					log.Err(err),
					log.Any("operator", operator),
					log.Any("req", req),
					log.String("meta", meta),
					log.Any("events", events),
				)
				return nil, err
			}
			v.ActivityImagePair = data
			for _, r := range data.PlayRecords {
				v.TotalSpentTime += r.Duration
			}
			if len(data.PlayRecords) != 0 {
				v.AvgSpentTime = v.TotalSpentTime / int64(len(data.PlayRecords))
			}
		case entity.ActivityTypeFlashCards:
			data, err := rm.getActivityFlashCards(m.ID, meta, events)
			if err != nil {
				log.Error(ctx, "GetStudentPerformanceH5PReport: call ActivityTypeFlashCards failed",
					log.Err(err),
					log.Any("operator", operator),
					log.Any("req", req),
					log.String("meta", meta),
					log.Any("events", events),
				)
				return nil, err
			}
			v.ActivityFlashCards = data
			for _, r := range data.PlayRecords {
				v.TotalSpentTime += r.Duration
			}
			if len(data.PlayRecords) != 0 {
				v.AvgSpentTime = v.TotalSpentTime / int64(len(data.PlayRecords))
			}
		}

		r.Items = append(r.Items, &v)
	}

	return &r, nil
}

func (rm *reportModel) calculateUsersSpentTimeMap(events []*entity.H5PEvent) map[string]int64 {
	eventsMap := map[string]map[string][]*entity.H5PEvent{}
	for _, e := range events {
		if eventsMap[e.UserID] == nil {
			eventsMap[e.UserID] = map[string][]*entity.H5PEvent{}
		}
		eventsMap[e.UserID][e.PlayID] = append(eventsMap[e.UserID][e.PlayID], e)
	}
	r := map[string]int64{}
	for uid, playID2EventsMap := range eventsMap {
		var totalSpent int64 = 0
		for _, events := range playID2EventsMap {
			sort.Sort(entity.H5PEventsSortByTime(events))
			var (
				lastEndTime int64 = -1
				startTime   int64 = -1
				endTime     int64 = -1
			)
			for _, e := range events {
				switch e.VerbID {
				case constant.ActivityEventVerbIDInitGame:
					if lastEndTime == -1 {
						lastEndTime = e.EventTime
					}
				case constant.ActivityEventVerbIDAnswered, constant.ActivityEventVerbIDCompleted:
					if lastEndTime == -1 {
						lastEndTime = e.EventTime
						continue
					}
					startTime = lastEndTime
					endTime = e.EventTime
					lastEndTime = endTime
				}
				if startTime != -1 && endTime != -1 {
					totalSpent += endTime - startTime
				}
			}
		}
		r[uid] = totalSpent
	}
	return r
}

func (rm *reportModel) getAttendanceIDsExistMapByClassIDAndLessonPlanID(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, classID string, lessonPlanID string) (map[string]bool, error) {
	aids, err := rm.getCompletedAssessmentIDs(ctx, tx, operator, classID, lessonPlanID)
	if err != nil {
		log.Error(ctx, "getAttendanceIDsExistMapByClassIDAndLessonPlanID: call getCompletedAssessmentIDs failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("class_id", classID),
			log.String("lesson_plan_id", lessonPlanID),
		)
		return nil, err
	}
	assessmentAttendances, err := rm.getCheckedAssessmentAttendance(ctx, tx, aids)
	if err != nil {
		log.Error(ctx, "getAttendanceIDsExistMapByClassIDAndLessonPlanID: call getCheckedAssessmentAttendance failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("class_id", classID),
			log.String("lesson_plan_id", lessonPlanID),
		)
		return nil, err
	}
	return rm.getAttendanceIDsExistMap(assessmentAttendances), nil
}

func (rm *reportModel) getLessonPlanH5PMaterials(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, lessonPlanIDs []string) ([]*entity.SubContentsWithName, error) {
	materialsMap, err := GetContentModel().GetContentsSubContentsMapByIDList(ctx, tx, lessonPlanIDs, operator)
	switch {
	case err == dbo.ErrRecordNotFound:
		log.Error(ctx, "getLessonPlanH5PMaterialIDs: call GetContentsSubContentsMapByIDList no record",
			log.Err(err),
			log.Any("operator", operator),
			log.Strings("lesson_plan_ids", lessonPlanIDs),
		)
		return nil, constant.ErrRecordNotFound
	case err != nil:
		log.Error(ctx, "getLessonPlanH5PMaterials: call GetContentsSubContentsMapByIDList failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Strings("lesson_plan_ids", lessonPlanIDs),
		)
		return nil, err
	}
	var (
		materials           []*entity.SubContentsWithName
		materialIDsExistMap = map[string]bool{}
	)
	for _, items := range materialsMap {
		for _, item := range items {
			if materialIDsExistMap[item.ID] {
				continue
			} else {
				materialIDsExistMap[item.ID] = true
			}
			materials = append(materials, item)
		}
	}
	log.Debug(ctx, "getLessonPlanH5PMaterials: print materials", log.Any("materials", materials))
	var result []*entity.SubContentsWithName
	for _, m := range materials {
		if m == nil {
			continue
		}
		if v, ok := m.Data.(*MaterialData); ok && v.FileType == entity.FileTypeH5pExtend {
			result = append(result, m)
		}
	}
	log.Debug(ctx, "getLessonPlanH5PMaterials: filtered materials result", log.Any("result", result))
	return result, nil
}

func (rm *reportModel) getLessonPlanH5PMaterialIDs(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, lessonPlanIDs []string) ([]string, error) {
	materials, err := rm.getLessonPlanH5PMaterials(ctx, tx, operator, lessonPlanIDs)
	if err != nil {
		log.Error(ctx, "getLessonPlanH5PMaterialIDs: call GetContentSubContentsByID failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("lesson_plan_id", lessonPlanIDs),
		)
		return nil, err
	}
	var result []string
	for _, m := range materials {
		result = append(result, m.ID)
	}
	return result, nil
}

func (rm *reportModel) getActivityImageSequencing(materialID string, meta string, events []*entity.H5PEvent) (*entity.ActivityImageSequencing, error) {
	r := entity.ActivityImageSequencing{CardsNumber: int(gjson.Get(meta, constant.H5PGJSONPathSequenceImagesCardsNumber).Int())}

	playID2EventsMap := map[string][]*entity.H5PEvent{}
	for _, e := range events {
		if entity.ParseActivityType(e.LocalLibraryName) != entity.ActivityTypeImageSequencing {
			continue
		}
		if e.MaterialID != materialID {
			continue
		}
		if e.VerbID != constant.ActivityEventVerbIDInitGame && e.VerbID != constant.ActivityEventVerbIDAnswered {
			continue
		}
		playID2EventsMap[e.PlayID] = append(playID2EventsMap[e.PlayID], e)
	}

	for _, events := range playID2EventsMap {
		sort.Sort(entity.H5PEventsSortByTime(events))
		var (
			lastEndTime int64 = -1
			startTime   int64 = -1
			endTime     int64 = -1
		)
		for _, e := range events {
			switch e.VerbID {
			case constant.ActivityEventVerbIDInitGame:
				lastEndTime = e.EventTime
				continue
			case constant.ActivityEventVerbIDAnswered:
				if lastEndTime == -1 {
					lastEndTime = e.EventTime
					continue
				}
				startTime = lastEndTime
				endTime = e.EventTime
				lastEndTime = endTime
			}
			if e.VerbID == constant.ActivityEventVerbIDAnswered && startTime != -1 && endTime != -1 {
				record := entity.ActivityImageSequencingPlayRecord{
					StartTime:         startTime,
					EndTime:           endTime,
					Duration:          endTime - startTime,
					CorrectCardsCount: int(gjson.Get(e.Extends, constant.H5PGJSONPathSequenceImagesCorrectCardsCount).Int()),
				}
				r.PlayRecords = append(r.PlayRecords, &record)
			}
		}
	}

	r.PlayTimes = len(r.PlayRecords)

	return &r, nil
}

func (rm *reportModel) getActivityMemoryGame(materialID string, meta string, events []*entity.H5PEvent) (*entity.ActivityMemoryGame, error) {
	r := entity.ActivityMemoryGame{PairsNumber: int(gjson.Get(meta, constant.H5PGJSONPathMemoryGamePairsNumber).Int())}

	playID2EventsMap := map[string][]*entity.H5PEvent{}
	for _, e := range events {
		if entity.ParseActivityType(e.LocalLibraryName) != entity.ActivityTypeMemoryGame {
			continue
		}
		if e.MaterialID != materialID {
			continue
		}
		if e.VerbID != constant.ActivityEventVerbIDInitGame &&
			e.VerbID != constant.ActivityEventVerbIDCompleted &&
			e.VerbID != constant.ActivityEventVerbIDInteracted {
			continue
		}
		playID2EventsMap[e.PlayID] = append(playID2EventsMap[e.PlayID], e)
	}

	for _, ee := range playID2EventsMap {
		sort.Sort(entity.H5PEventsSortByTime(ee))
		var (
			lastEndTime int64 = -1
			startTime   int64 = -1
			endTime     int64 = -1
			clicksCount       = 0
		)
		for _, e := range ee {
			switch e.VerbID {
			case constant.ActivityEventVerbIDInitGame:
				lastEndTime = e.EventTime
				continue
			case constant.ActivityEventVerbIDCompleted:
				if lastEndTime == -1 {
					lastEndTime = e.EventTime
					continue
				}
				startTime = lastEndTime
				endTime = e.EventTime
				lastEndTime = endTime
			case constant.ActivityEventVerbIDInteracted:
				clicksCount++
			}
			if e.VerbID == constant.ActivityEventVerbIDCompleted && startTime != -1 && endTime != -1 {
				record := entity.ActivityMemoryGamePlayRecord{
					StartTime:   startTime,
					EndTime:     endTime,
					Duration:    endTime - startTime,
					ClicksCount: clicksCount,
				}
				r.PlayRecords = append(r.PlayRecords, &record)
				clicksCount = 0
			}
		}
	}

	r.PlayTimes = len(r.PlayRecords)

	return &r, nil
}

func (rm *reportModel) getActivityImagePair(materialID string, meta string, events []*entity.H5PEvent) (*entity.ActivityImagePair, error) {
	r := entity.ActivityImagePair{ParisNumber: int(gjson.Get(meta, constant.H5PGJSONPathImagePairPairsNumber).Int())}

	playID2EventsMap := map[string][]*entity.H5PEvent{}
	for _, e := range events {
		if entity.ParseActivityType(e.LocalLibraryName) != entity.ActivityTypeImagePair {
			continue
		}
		if e.MaterialID != materialID {
			continue
		}
		if e.VerbID != constant.ActivityEventVerbIDInitGame && e.VerbID != constant.ActivityEventVerbIDCompleted {
			continue
		}
		playID2EventsMap[e.PlayID] = append(playID2EventsMap[e.PlayID], e)
	}

	for _, ee := range playID2EventsMap {
		sort.Sort(entity.H5PEventsSortByTime(ee))
		var (
			lastEndTime int64 = -1
			startTime   int64 = -1
			endTime     int64 = -1
		)
		for _, e := range ee {
			switch e.VerbID {
			case constant.ActivityEventVerbIDInitGame:
				lastEndTime = e.EventTime
				continue
			case constant.ActivityEventVerbIDCompleted:
				if lastEndTime == -1 {
					lastEndTime = e.EventTime
					continue
				}
				startTime = lastEndTime
				endTime = e.EventTime
				lastEndTime = endTime
			}
			if e.VerbID == constant.ActivityEventVerbIDCompleted && startTime != -1 && endTime != -1 {
				record := entity.ActivityImagePairPlayRecord{
					StartTime:         startTime,
					EndTime:           endTime,
					Duration:          endTime - startTime,
					CorrectPairsCount: int(gjson.Get(e.Extends, constant.H5PGJSONPathImagePairCorrectPairsCount).Int()),
				}
				r.PlayRecords = append(r.PlayRecords, &record)
			}
		}
	}

	r.PlayTimes = len(r.PlayRecords)

	return &r, nil
}

func (rm *reportModel) getActivityFlashCards(materialID string, meta string, events []*entity.H5PEvent) (*entity.ActivityFlashCards, error) {
	r := entity.ActivityFlashCards{CardsNumber: int(gjson.Get(meta, constant.H5PGJSONPathFlashCardsCardsNumber).Int())}

	playID2EventsMap := map[string][]*entity.H5PEvent{}
	for _, e := range events {
		if entity.ParseActivityType(e.LocalLibraryName) != entity.ActivityTypeFlashCards {
			continue
		}
		if e.MaterialID != materialID {
			continue
		}
		if e.VerbID != constant.ActivityEventVerbIDInitGame && e.VerbID != constant.ActivityEventVerbIDCompleted {
			continue
		}
		playID2EventsMap[e.PlayID] = append(playID2EventsMap[e.PlayID], e)
	}

	for _, ee := range playID2EventsMap {
		sort.Sort(entity.H5PEventsSortByTime(ee))
		var (
			lastEndTime int64 = -1
			startTime   int64 = -1
			endTime     int64 = -1
		)
		for _, e := range ee {
			switch e.VerbID {
			case constant.ActivityEventVerbIDInitGame:
				lastEndTime = e.EventTime
				continue
			case constant.ActivityEventVerbIDCompleted:
				if lastEndTime == -1 {
					lastEndTime = e.EventTime
					continue
				}
				startTime = lastEndTime
				endTime = e.EventTime
				lastEndTime = endTime
			}
			if e.VerbID == constant.ActivityEventVerbIDCompleted && startTime != -1 && endTime != -1 {
				record := entity.ActivityFlashCardsPlayRecord{
					StartTime:         startTime,
					EndTime:           endTime,
					Duration:          endTime - startTime,
					CorrectCardsCount: int(gjson.Get(e.Extends, constant.H5PGJSONPathFlashCardsCorrectCardsCount).Int()),
				}
				r.PlayRecords = append(r.PlayRecords, &record)
			}
		}
	}

	return &r, nil
}

// endregion
