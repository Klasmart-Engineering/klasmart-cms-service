package model

import (
	"context"
	"sort"
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

type IReportModel interface {
	ListStudentsReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.ListStudentsAchievementReportRequest) (*entity.StudentsAchievementReportResponse, error)
	GetStudentReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.GetStudentAchievementReportRequest) (*entity.StudentAchievementReportResponse, error)
	GetTeacherReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, teacherID string) (*entity.TeacherReport, error)

	ListStudentsPerformanceReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.ListStudentsPerformanceReportRequest) (*entity.ListStudentsPerformanceReportResponse, error)
	GetStudentPerformanceReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.GetStudentPerformanceReportRequest) (*entity.GetStudentPerformanceReportResponse, error)
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

func (m *reportModel) ListStudentsReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.ListStudentsAchievementReportRequest) (*entity.StudentsAchievementReportResponse, error) {
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
		allowed, err := m.hasReportPermission(ctx, operator, req.TeacherID)
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

	students, err := m.getStudentsInClass(ctx, operator, req.ClassID)
	if err != nil {
		log.Error(ctx, "list students report: get students",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}

	assessmentIDs, err := m.getCompletedAssessmentIDs(ctx, tx, operator, req.ClassID, req.LessonPlanID)
	if err != nil {
		log.Error(ctx, "list student report: get assessment ids failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}

	assessmentAttendances, err := m.getCheckedAssessmentAttendance(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "list student report: get checked assessment attendance failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	var assessmentOutcomes []*entity.AssessmentOutcome
	if err := da.GetAssessmentOutcomeDA().QueryTx(ctx, tx, &da.QueryAssessmentOutcomeConditions{
		AssessmentIDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   true,
		},
		Checked: entity.NullBool{
			Bool:  true,
			Valid: true,
		},
	}, &assessmentOutcomes); err != nil {
		log.Error(ctx, "ListStudentsReport: da.GetAssessmentOutcomeDA().QueryTx: get assessment outcomes failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	outcomeAttendances, err := m.getOutcomeAttendances(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "list student report: call getOutcomeAttendances failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	tr, err := m.makeLatestOutcomeIDsTranslator(ctx, tx, operator, m.getOutcomeIDs(assessmentOutcomes))
	if err != nil {
		log.Error(ctx, "list student report: make latest outcome ids translator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Any("assessment_outcomes", assessmentOutcomes),
		)
		return nil, err
	}

	attendanceIDExistsMap := m.getAttendanceIDsExistMap(assessmentAttendances)
	attendanceID2OutcomeIDsMap := m.getAttendanceID2OutcomeIDsMap(assessmentAttendances, assessmentOutcomes)
	achievedAttendanceID2OutcomeIDsMap := m.getAchievedAttendanceID2OutcomeIDsMap(outcomeAttendances)
	skipAttendanceID2OutcomeIDsMap := m.getSkipAttendanceID2OutcomeIDsMap(assessmentAttendances, assessmentOutcomes)
	notAchievedAttendanceID2OutcomeIDsMap := m.getNotAchievedAttendanceID2OutcomeIDsMap(attendanceID2OutcomeIDsMap, achievedAttendanceID2OutcomeIDsMap, skipAttendanceID2OutcomeIDsMap)
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

func (m *reportModel) GetStudentReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.GetStudentAchievementReportRequest) (*entity.StudentAchievementReportResponse, error) {
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
		allowed, err := m.hasReportPermission(ctx, operator, req.TeacherID)
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

	student, err := m.getStudentInClass(ctx, operator, req.ClassID, req.StudentID)
	if err != nil {
		log.Error(ctx, "list students report: call getStudentInClass failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}

	assessmentIDs, err := m.getCompletedAssessmentIDs(ctx, tx, operator, req.ClassID, req.LessonPlanID)
	if err != nil {
		log.Error(ctx, "get student detail report: get assessment ids failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}

	var assessmentOutcomes []*entity.AssessmentOutcome
	if err := da.GetAssessmentOutcomeDA().QueryTx(ctx, tx, &da.QueryAssessmentOutcomeConditions{
		AssessmentIDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   true,
		},
		Checked: entity.NullBool{
			Bool:  true,
			Valid: true,
		},
	}, &assessmentOutcomes); err != nil {
		log.Error(ctx, "GetStudentDetailReport: da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDs: get assessment outcomes failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	assessmentAttendances, err := m.getCheckedAssessmentAttendance(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "get student detail report: get checked assessment attendances failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	outcomeAttendances, err := m.getOutcomeAttendances(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "get student detail report: call getOutcomeAttendances failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	outcomeIDs := m.getOutcomeIDs(assessmentOutcomes)
	tr, err := m.makeLatestOutcomeIDsTranslator(ctx, tx, operator, outcomeIDs)
	if err != nil {
		log.Error(ctx, "get student detail report: make latest outcome ids translator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("outcome_ids", outcomeIDs),
		)
		return nil, err
	}

	latestOutcomeIDs := tr(outcomeIDs)
	outcomesMap, err := m.getOutcomesMap(ctx, tx, operator, latestOutcomeIDs)
	if err != nil {
		log.Error(ctx, "get student detail report: get outcomes map failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("outcome_ids", outcomeIDs),
			log.Strings("latest_outcome_ids", latestOutcomeIDs),
		)
		return nil, err
	}

	categories, err := external.GetCategoryServiceProvider().GetByOrganization(ctx, operator)
	if err != nil {
		log.Error(ctx, "get student detail report: query all category failed",
			log.Err(err),
			log.Any("req", req),
			log.Any("operator", operator),
		)
		return nil, err
	}

	attendanceIDExistsMap := m.getAttendanceIDsExistMap(assessmentAttendances)
	attendanceID2OutcomeIDsMap := m.getAttendanceID2OutcomeIDsMap(assessmentAttendances, assessmentOutcomes)
	achievedAttendanceID2OutcomeIDsMap := m.getAchievedAttendanceID2OutcomeIDsMap(outcomeAttendances)
	skipAttendanceID2OutcomeIDsMap := m.getSkipAttendanceID2OutcomeIDsMap(assessmentAttendances, assessmentOutcomes)
	notAchievedAttendanceID2OutcomeIDsMap := m.getNotAchievedAttendanceID2OutcomeIDsMap(attendanceID2OutcomeIDsMap, achievedAttendanceID2OutcomeIDsMap, skipAttendanceID2OutcomeIDsMap)
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
	for _, category := range categories {
		c := entity.StudentAchievementReportCategoryItem{Name: category.Name}
		achievedOIDs := tr(achievedAttendanceID2OutcomeIDsMap[req.StudentID])
		for _, oid := range achievedOIDs {
			o := outcomesMap[oid]
			if o == nil {
				log.Debug(ctx, "get student report: not found achieved outcome",
					log.String("outcome_id", oid),
				)
				continue
			}
			if utils.ContainsStr(o.Categories, category.ID) {
				c.AchievedItems = append(c.AchievedItems, o.Name)
			}
		}
		notAchievedOIDs := tr(notAchievedAttendanceID2OutcomeIDsMap[req.StudentID])
		for _, oid := range notAchievedOIDs {
			o := outcomesMap[oid]
			if o == nil {
				log.Debug(ctx, "get student report: not found not achieved outcome",
					log.String("outcome_id", oid),
				)
				continue
			}
			if utils.ContainsStr(o.Categories, category.ID) {
				c.NotAchievedItems = append(c.NotAchievedItems, o.Name)
			}
		}
		skipOIDs := tr(skipAttendanceID2OutcomeIDsMap[req.StudentID])
		for _, oid := range skipOIDs {
			o := outcomesMap[oid]
			if o == nil {
				log.Debug(ctx, "get student report: not found skip outcome",
					log.String("outcome_id", oid),
				)
				continue
			}
			if utils.ContainsStr(o.Categories, category.ID) {
				c.NotAttemptedItems = append(c.NotAttemptedItems, o.Name)
			}
		}
		result.Categories = append(result.Categories, &c)
	}

	return &result, nil
}

func (m *reportModel) GetTeacherReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, teacherID string) (*entity.TeacherReport, error) {
	var assessmentIDs []string
	{
		var assessments []*entity.Assessment
		if err := da.GetAssessmentDA().Query(ctx, &da.QueryAssessmentConditions{
			OrgID: entity.NullString{
				String: operator.OrgID,
				Valid:  true,
			},
			TeacherIDs: entity.NullStrings{
				Strings: []string{teacherID},
				Valid:   true,
			},
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
		var assessmentOutcomes []*entity.AssessmentOutcome
		if err := da.GetAssessmentOutcomeDA().QueryTx(ctx, tx, &da.QueryAssessmentOutcomeConditions{
			AssessmentIDs: entity.NullStrings{
				Strings: assessmentIDs,
				Valid:   true,
			},
			Checked: entity.NullBool{
				Bool:  true,
				Valid: true,
			},
		}, &assessmentOutcomes); err != nil {
			log.Error(ctx, "GetTeacherReport: da.GetAssessmentOutcomeDA().QueryTx: get assessment outcomes failed",
				log.Err(err),
				log.Any("operator", operator),
				log.Strings("assessment_ids", assessmentIDs),
			)
			return nil, err
		}

		outcomeIDs := m.getOutcomeIDs(assessmentOutcomes)
		oidTr, err := m.makeLatestOutcomeIDsTranslator(ctx, tx, operator, outcomeIDs)
		if err != nil {
			log.Error(ctx, "GetTeacherReport: make latest outcome ids translator failed",
				log.Err(err),
				log.Any("outcome_ids", outcomeIDs),
				log.Any("operator", operator),
			)
		}
		outcomeIDs = oidTr(outcomeIDs)
		outcomes, err = GetOutcomeModel().GetByIDs(ctx, operator, tx, outcomeIDs)
		if err != nil {
			log.Error(ctx, "get teacher report: get learning outcome failed by ids",
				log.Err(err),
				log.Any("operator", operator),
				log.String("teacher_id", teacherID),
			)
			return nil, err
		}
	}
	categoryIDToNameMap := map[string]string{}
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
			categoryIDToNameMap[item.ID] = item.Name
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
				Name: categoryIDToNameMap[developmentalID],
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

func (m *reportModel) ListStudentsPerformanceReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.ListStudentsPerformanceReportRequest) (*entity.ListStudentsPerformanceReportResponse, error) {
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
		allowed, err := m.hasReportPermission(ctx, operator, req.TeacherID)
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

	students, err := m.getStudentsInClass(ctx, operator, req.ClassID)
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceReport: call getStudentsInClass failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}

	assessmentIDs, err := m.getCompletedAssessmentIDs(ctx, tx, operator, req.ClassID, req.LessonPlanID)
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceReport: call getCompletedAssessmentIDs failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}

	assessmentAttendances, err := m.getCheckedAssessmentAttendance(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceReport: call getCheckedAssessmentAttendance failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	var assessmentOutcomes []*entity.AssessmentOutcome
	if err := da.GetAssessmentOutcomeDA().QueryTx(ctx, tx, &da.QueryAssessmentOutcomeConditions{
		AssessmentIDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   true,
		},
		Checked: entity.NullBool{
			Bool:  true,
			Valid: true,
		},
	}, &assessmentOutcomes); err != nil {
		log.Error(ctx, "ListStudentsPerformanceReport: da.GetAssessmentOutcomeDA().QueryTx: get assessment outcomes failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	outcomeAttendances, err := m.getOutcomeAttendances(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceReport: call getOutcomeAttendances failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	outcomeNamesMap, err := m.getOutcomeNamesMap(ctx, tx, operator, m.getOutcomeIDs(assessmentOutcomes))
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

	trLatestOutcomeIDs, err := m.makeLatestOutcomeIDsTranslator(ctx, tx, operator, m.getOutcomeIDs(assessmentOutcomes))
	if err != nil {
		log.Error(ctx, "ListStudentsPerformanceReport: call makeLatestOutcomeIDsTranslator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Any("assessment_outcomes", assessmentOutcomes),
		)
		return nil, err
	}

	attendanceIDExistsMap := m.getAttendanceIDsExistMap(assessmentAttendances)
	attendanceID2OutcomeIDsMap := m.getAttendanceID2OutcomeIDsMap(assessmentAttendances, assessmentOutcomes)
	achievedAttendanceID2OutcomeIDsMap := m.getAchievedAttendanceID2OutcomeIDsMap(outcomeAttendances)
	skipAttendanceID2OutcomeIDsMap := m.getSkipAttendanceID2OutcomeIDsMap(assessmentAttendances, assessmentOutcomes)
	notAchievedAttendanceID2OutcomeIDsMap := m.getNotAchievedAttendanceID2OutcomeIDsMap(attendanceID2OutcomeIDsMap, achievedAttendanceID2OutcomeIDsMap, skipAttendanceID2OutcomeIDsMap)
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

func (m *reportModel) GetStudentPerformanceReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, req entity.GetStudentPerformanceReportRequest) (*entity.GetStudentPerformanceReportResponse, error) {
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
		allowed, err := m.hasReportPermission(ctx, operator, req.TeacherID)
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

	student, err := m.getStudentInClass(ctx, operator, req.ClassID, req.StudentID)
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceReport: call getStudentInClass failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}

	assessments, err := m.getCompletedAssessments(ctx, tx, operator, req.ClassID, req.LessonPlanID)
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

	schedules, err := GetScheduleModel().GetVariableDataByIDs(ctx, operator, scheduleIDs, &entity.ScheduleInclude{Subject: true})
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceReport: call GetScheduleDA().Query failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
		)
		return nil, err
	}
	schedulesMap := map[string]*entity.ScheduleVariable{}
	for _, s := range schedules {
		schedulesMap[s.ID] = s
	}

	var (
		assessmentOutcomes []*entity.AssessmentOutcome
		checked            = true
	)
	if err := da.GetAssessmentOutcomeDA().QueryTx(ctx, tx, &da.QueryAssessmentOutcomeConditions{
		AssessmentIDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   true,
		},
		Checked: entity.NullBool{
			Bool:  true,
			Valid: true,
		},
	}, &assessmentOutcomes); err != nil {
		log.Error(ctx, "GetStudentPerformanceReport: da.GetAssessmentOutcomeDA().QueryTx: get assessment outcomes failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("assessment_ids", assessmentIDs),
			log.Bool("checked", checked),
		)
		return nil, err
	}

	assessmentAttendances, err := m.getCheckedAssessmentAttendance(ctx, tx, assessmentIDs)
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

	outcomeIDs := m.getOutcomeIDs(assessmentOutcomes)
	outcomeNamesMap, err := m.getOutcomeNamesMap(ctx, tx, operator, outcomeIDs)
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

	tr, err := m.makeLatestOutcomeIDsTranslator(ctx, tx, operator, outcomeIDs)
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceReport: call makeLatestOutcomeIDsTranslator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Strings("outcome_ids", outcomeIDs),
		)
		return nil, err
	}

	assessmentID2OutcomeIDsMap := m.getAssessmentID2OutcomeIDsMap(assessmentOutcomes)
	attendanceIDExistsMap := m.getAttendanceIDsExistMap(assessmentAttendances)
	achievedAssessmentID2OutcomeIDsMap := m.getAchievedAssessmentID2OutcomeIDsMap(student.ID, assessmentOutcomes, outcomeAttendances)
	skipAssessmentID2OutcomeIDsMap := m.getSkipAssessmentID2OutcomeIDsMap(student.ID, assessmentAttendances, assessmentOutcomes)
	notAchievedAssessmentID2OutcomeIDsMap := m.getNotAchievedAssessmentID2OutcomeIDsMap(assessmentID2OutcomeIDsMap, achievedAssessmentID2OutcomeIDsMap, skipAssessmentID2OutcomeIDsMap)
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

func (m *reportModel) getCompletedAssessmentIDs(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, classID string, lessonPlanID string) ([]string, error) {
	assessments, err := m.getCompletedAssessments(ctx, tx, operator, classID, lessonPlanID)
	if err != nil {
		log.Error(ctx, "GetStudentPerformanceReport: call getCompletedAssessments failed",
			log.Err(err),
			log.String("class_id", classID),
			log.String("lesson_plan_id", lessonPlanID),
			log.Any("operator", operator),
		)
		return nil, err
	}
	assessmentIDs := make([]string, 0, len(assessments))
	for _, a := range assessments {
		assessmentIDs = append(assessmentIDs, a.ID)
	}
	return assessmentIDs, nil
}

func (m *reportModel) getCompletedAssessments(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, classID string, lessonPlanID string) ([]*entity.Assessment, error) {
	ids, err := m.getAssessmentIDs(ctx, tx, operator, classID, lessonPlanID)
	if err != nil {
		log.Error(ctx, "getCompletedAssessments: call getAssessmentIDs failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("class_id", classID),
			log.String("lesson_plan_id", lessonPlanID),
		)
		return nil, err
	}
	result := make([]*entity.Assessment, 0, len(ids))
	if err := da.GetAssessmentDA().Query(ctx, &da.QueryAssessmentConditions{
		IDs: entity.NullStrings{
			Strings: ids,
			Valid:   true,
		},
		ClassTypes: entity.NullScheduleClassTypes{
			Value: []entity.ScheduleClassType{entity.ScheduleClassTypeOnlineClass, entity.ScheduleClassTypeOfflineClass},
			Valid: true,
		},
		Status: entity.NullAssessmentStatus{
			Value: entity.AssessmentStatusComplete,
			Valid: true,
		},
	}, &result); err != nil {
		log.Error(ctx, "da.GetAssessmentDA().Query: call FilterCompletedAssessments failed",
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

func (m *reportModel) getAssessmentIDs(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, classID string, lessonPlanID string) ([]string, error) {
	scheduleIDs, err := m.getScheduleIDs(ctx, tx, operator, classID, lessonPlanID)
	if err != nil {
		log.Error(ctx, "get assessment ids: get schedule ids failed",
			log.Err(err),
			log.String("class_id", classID),
			log.String("lesson_plan_id", lessonPlanID),
		)
		return nil, err
	}
	assessments := make([]*entity.Assessment, 0, len(scheduleIDs))
	if err := da.GetAssessmentDA().Query(ctx, &da.QueryAssessmentConditions{
		ClassTypes: entity.NullScheduleClassTypes{
			Value: []entity.ScheduleClassType{entity.ScheduleClassTypeOnlineClass, entity.ScheduleClassTypeOfflineClass},
			Valid: true,
		},
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	}, &assessments); err != nil {
		log.Error(ctx, "da.GetAssessmentDA().Query: batch get assessment failed by schedule ids",
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

func (m *reportModel) getScheduleIDs(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, classID string, lessonPlanID string) ([]string, error) {
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

func (m *reportModel) getCheckedAssessmentAttendance(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string) ([]*entity.AssessmentAttendance, error) {
	var result []*entity.AssessmentAttendance
	if err := da.GetAssessmentAttendanceDA().QueryTx(ctx, tx, &da.QueryAssessmentAttendanceConditions{
		AssessmentIDs: entity.NullStrings{Strings: assessmentIDs, Valid: true},
		Checked:       entity.NullBool{Bool: true, Valid: true},
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

func (m *reportModel) getOutcomeAttendances(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string) ([]*entity.OutcomeAttendance, error) {
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

func (m *reportModel) getOutcomeIDs(assessmentOutcomes []*entity.AssessmentOutcome) []string {
	result := make([]string, 0, len(assessmentOutcomes))
	for _, v := range assessmentOutcomes {
		result = append(result, v.OutcomeID)
	}
	return utils.SliceDeduplication(result)
}

func (rm *reportModel) getOutcomesMap(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, outcomeIDs []string) (map[string]*entity.Outcome, error) {
	outcomes, err := GetOutcomeModel().GetByIDs(ctx, operator, tx, outcomeIDs)
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
	outcomes, err := GetOutcomeModel().GetByIDs(ctx, operator, tx, outcomeIDs)
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

func (m *reportModel) getAttendanceIDsExistMap(assessmentAttendances []*entity.AssessmentAttendance) map[string]bool {
	result := make(map[string]bool, len(assessmentAttendances))
	for _, assessmentAttendance := range assessmentAttendances {
		result[assessmentAttendance.AttendanceID] = true
	}
	return result
}

func (m *reportModel) getAttendanceID2AssessmentOutcomesMap(assessmentAttendances []*entity.AssessmentAttendance, assessmentOutcomes []*entity.AssessmentOutcome) map[string][]*entity.AssessmentOutcome {
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

func (m *reportModel) getAttendanceID2OutcomeIDsMap(assessmentAttendances []*entity.AssessmentAttendance, assessmentOutcomes []*entity.AssessmentOutcome) map[string][]string {
	attendanceID2AssessmentOutcomesMap := m.getAttendanceID2AssessmentOutcomesMap(assessmentAttendances, assessmentOutcomes)
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

func (m *reportModel) getAchievedAttendanceID2OutcomeIDsMap(outcomeAttendances []*entity.OutcomeAttendance) map[string][]string {
	result := map[string][]string{}
	for _, outcomeAttendance := range outcomeAttendances {
		result[outcomeAttendance.AttendanceID] = append(result[outcomeAttendance.AttendanceID], outcomeAttendance.OutcomeID)
	}
	for k, v := range result {
		result[k] = utils.SliceDeduplication(v)
	}
	return result
}

func (m *reportModel) getSkipAttendanceID2OutcomeIDsMap(assessmentAttendances []*entity.AssessmentAttendance, assessmentOutcomes []*entity.AssessmentOutcome) map[string][]string {
	attendanceID2AssessmentOutcomesMap := m.getAttendanceID2AssessmentOutcomesMap(assessmentAttendances, assessmentOutcomes)
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

func (m *reportModel) getNotAchievedAttendanceID2OutcomeIDsMap(attendanceID2OutcomeIDsMap, achievedAttendanceID2OutcomeIDsMap, skipAttendanceID2OutcomeIDsMap map[string][]string) map[string][]string {
	result := map[string][]string{}
	for attendanceID, outcomeIDs := range attendanceID2OutcomeIDsMap {
		var excludeOutcomeIDs []string
		excludeOutcomeIDs = append(excludeOutcomeIDs, achievedAttendanceID2OutcomeIDsMap[attendanceID]...)
		excludeOutcomeIDs = append(excludeOutcomeIDs, skipAttendanceID2OutcomeIDsMap[attendanceID]...)
		result[attendanceID] = utils.ExcludeStrings(outcomeIDs, excludeOutcomeIDs)
	}
	return result
}

func (m *reportModel) getAssessmentID2OutcomeIDsMap(assessmentOutcomes []*entity.AssessmentOutcome) map[string][]string {
	result := map[string][]string{}
	for _, ao := range assessmentOutcomes {
		result[ao.AssessmentID] = append(result[ao.AssessmentID], ao.OutcomeID)
	}
	return result
}

func (m *reportModel) getAchievedAssessmentID2OutcomeIDsMap(attendanceID string, assessmentOutcomes []*entity.AssessmentOutcome, outcomeAttendances []*entity.OutcomeAttendance) map[string][]string {
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

func (m *reportModel) getSkipAssessmentID2OutcomeIDsMap(attendanceID string, assessmentAttendances []*entity.AssessmentAttendance, assessmentOutcomes []*entity.AssessmentOutcome) map[string][]string {
	attendanceID2AssessmentOutcomesMap := m.getAttendanceID2AssessmentOutcomesMap(assessmentAttendances, assessmentOutcomes)
	assessmentOutcomes = attendanceID2AssessmentOutcomesMap[attendanceID]
	result := map[string][]string{}
	for _, ao := range assessmentOutcomes {
		if ao.Skip {
			result[ao.AssessmentID] = append(result[ao.AssessmentID], ao.OutcomeID)
		}
	}
	return result
}

func (m *reportModel) getNotAchievedAssessmentID2OutcomeIDsMap(assessmentID2OutcomeIDsMap, achievedAssessmentID2OutcomeIDsMap, skipAssessmentID2OutcomeIDsMap map[string][]string) map[string][]string {
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
	m, err := GetOutcomeModel().GetLatestByIDsMapResult(ctx, operator, tx, outcomeIDs)
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
	log.Debug(ctx, "get latest outcome ids",
		log.Strings("outcome_ids", outcomeIDs),
		log.Any("latest_outcome_map", m),
	)
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

func (m *reportModel) hasReportPermission(ctx context.Context, operator *entity.Operator, teacherID string) (bool, error) {
	checkP603, err := m.checkPermission603(ctx, operator, teacherID)
	if err != nil {
		return false, err
	}
	if !checkP603 {
		return false, nil
	}

	optionalCheckers := []func(context.Context, *entity.Operator, string) (bool, error){
		m.checkPermission614,
		m.checkPermission610,
		m.checkPermission611,
		m.checkPermission612,
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

func (m *reportModel) checkPermission603(ctx context.Context, operator *entity.Operator, teacherID string) (bool, error) {
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

func (m *reportModel) checkPermission614(ctx context.Context, operator *entity.Operator, teacherID string) (bool, error) {
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

func (m *reportModel) checkPermission610(ctx context.Context, operator *entity.Operator, teacherID string) (bool, error) {
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

func (m *reportModel) checkPermission611(ctx context.Context, operator *entity.Operator, teacherID string) (bool, error) {
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

func (m *reportModel) checkPermission612(ctx context.Context, operator *entity.Operator, teacherID string) (bool, error) {
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

func (m *reportModel) getStudentsInClass(ctx context.Context, operator *entity.Operator, classID string) ([]*external.Student, error) {
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

func (m *reportModel) getStudentInClass(ctx context.Context, operator *entity.Operator, classID string, studentID string) (*external.Student, error) {
	students, err := m.getStudentsInClass(ctx, operator, classID)
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
