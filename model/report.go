package model

import (
	"context"
	"sort"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type IReportModel interface {
	ListStudentsReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, cmd entity.ListStudentsReportCommand) (*entity.StudentsReport, error)
	GetStudentReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, cmd entity.GetStudentReportCommand) (*entity.StudentReport, error)
	GetTeacherReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, teacherID string) (*entity.TeacherReport, error)
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

func (r *reportModel) ListStudentsReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, cmd entity.ListStudentsReportCommand) (*entity.StudentsReport, error) {
	{
		if !cmd.Status.Valid() {
			log.Error(ctx, "list students report: invalid status", log.Any("cmd", cmd))
			return nil, constant.ErrInvalidArgs
		}
		if !cmd.SortBy.Valid() {
			log.Error(ctx, "list students report: invalid sort by", log.Any("cmd", cmd))
			return nil, constant.ErrInvalidArgs
		}
		if cmd.ClassID == "" {
			log.Error(ctx, "list students report: require class id", log.Any("cmd", cmd))
			return nil, constant.ErrInvalidArgs
		}
		if cmd.TeacherID == "" {
			log.Error(ctx, "list students report: require teacher id", log.Any("cmd", cmd))
			return nil, constant.ErrInvalidArgs
		}
		if cmd.LessonPlanID == "" {
			log.Error(ctx, "list students report: require lesson plan id", log.Any("cmd", cmd))
			return nil, constant.ErrInvalidArgs
		}
		allowed, err := r.hasReportPermission(ctx, operator, cmd.TeacherID)
		if err != nil {
			log.Error(ctx, "list students report: check report report permission failed",
				log.Any("cmd", cmd),
				log.Any("operator", operator),
			)
			return nil, err
		}
		if !allowed {
			log.Error(ctx, "list students report: no permission",
				log.Any("cmd", cmd),
				log.Any("operator", operator),
			)
			return nil, constant.ErrForbidden
		}
	}

	var students []*external.Student
	{
		var err error
		log.Debug(ctx, "list students report: before call GetClassServiceProvider().getStudents()")
		students, err = external.GetStudentServiceProvider().GetByClassID(ctx, operator, cmd.ClassID)
		log.Debug(ctx, "list students report: after call GetClassServiceProvider().getStudents()")
		if err != nil {
			log.Error(ctx, "list students report: get students",
				log.Err(err),
				log.Any("cmd", cmd),
			)
			return nil, err
		}
	}

	log.Debug(ctx, "list students report: before call getAssessmentIDs()")
	assessmentIDs, err := r.getAssessmentIDs(ctx, tx, operator, cmd.ClassID, cmd.LessonPlanID)
	log.Debug(ctx, "list students report: after call getAssessmentIDs()")
	if err != nil {
		log.Error(ctx, "list student report: get assessment ids failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
		return nil, err
	}

	data, err := r.getAttendanceOutcomeData(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "list student report: get assessment outcome data failed",
			log.Err(err),
			log.Any("assessment_ids", assessmentIDs),
			log.Any("cmd", cmd),
		)
		return nil, err
	}
	outcomeIDsTranslatorFunc, err := makeLatestOutcomeIDsTranslator(ctx, tx, data.AllOutcomeIDs, operator)
	if err != nil {
		log.Error(ctx, "list student report: make latest outcome ids translator failed",
			log.Err(err),
			log.Any("data", data),
			log.Any("cmd", cmd),
			log.Any("operator", operator),
		)
	}
	var result = entity.StudentsReport{AssessmentIDs: assessmentIDs}
	for _, student := range students {
		newItem := entity.StudentReportItem{StudentID: student.ID, StudentName: student.Name}

		if !data.AllAttendanceIDExistsMap[student.ID] {
			newItem.Attend = false
			result.Items = append(result.Items, &newItem)
			continue
		}

		newItem.Attend = true
		newItem.AchievedCount = len(outcomeIDsTranslatorFunc(data.AchievedAttendanceID2OutcomeIDsMap[student.ID]))
		newItem.NotAttemptedCount = len(outcomeIDsTranslatorFunc(data.SkipAttendanceID2OutcomeIDsMap[student.ID]))
		newItem.NotAchievedCount = len(outcomeIDsTranslatorFunc(data.NotAchievedAttendanceID2OutcomeIDsMap[student.ID]))

		result.Items = append(result.Items, &newItem)
	}

	sortInterface := entity.NewSortingStudentReportItems(result.Items, cmd.Status)
	switch cmd.SortBy {
	case entity.ReportSortByDesc:
		sort.Sort(sort.Reverse(sortInterface))
	case entity.ReportSortByAsc:
		fallthrough
	default:
		sort.Sort(sortInterface)
	}

	return &result, nil
}

func (r *reportModel) GetStudentReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, cmd entity.GetStudentReportCommand) (*entity.StudentReport, error) {
	{
		if cmd.ClassID == "" {
			log.Error(ctx, "get student detail report: require class id", log.Any("cmd", cmd))
			return nil, constant.ErrInvalidArgs
		}
		if cmd.TeacherID == "" {
			log.Error(ctx, "get student detail report: require teacher id", log.Any("cmd", cmd))
			return nil, constant.ErrInvalidArgs
		}
		if cmd.LessonPlanID == "" {
			log.Error(ctx, "get student detail report: require lesson plan id", log.Any("cmd", cmd))
			return nil, constant.ErrInvalidArgs
		}
		if cmd.StudentID == "" {
			log.Error(ctx, "get student detail report: require student id", log.Any("cmd", cmd))
			return nil, constant.ErrInvalidArgs
		}
		allowed, err := r.hasReportPermission(ctx, operator, cmd.TeacherID)
		if err != nil {
			log.Error(ctx, "get student detail report: check report report permission failed",
				log.Any("cmd", cmd),
				log.Any("operator", operator),
			)
			return nil, err
		}
		if !allowed {
			log.Error(ctx, "get student detail report: no permission",
				log.Any("cmd", cmd),
				log.Any("operator", operator),
			)
			return nil, constant.ErrUnAuthorized
		}
	}

	var student *external.Student
	{
		log.Debug(ctx, "get student detail report: before call GetClassServiceProvider().getStudents()")
		students, err := external.GetStudentServiceProvider().GetByClassID(ctx, operator, cmd.ClassID)
		log.Debug(ctx, "get student detail report: after call GetClassServiceProvider().getStudents()")
		if err != nil {
			log.Error(ctx, "list students report: get students",
				log.Err(err),
				log.Any("cmd", cmd),
			)
			return nil, err
		}
		for _, item := range students {
			if item.ID == cmd.StudentID {
				student = item
				break
			}
		}
		if student == nil {
			log.Error(ctx, "get student detail report: not found student in class", log.Any("cmd", cmd))
			return nil, constant.ErrRecordNotFound
		}
	}

	log.Debug(ctx, "get student detail report: before call getAssessmentIDs()")
	assessmentIDs, err := r.getAssessmentIDs(ctx, tx, operator, cmd.ClassID, cmd.LessonPlanID)
	log.Debug(ctx, "get student detail report: after call getAssessmentIDs()")
	if err != nil {
		log.Error(ctx, "list student report: get assessment ids failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
		return nil, err
	}

	data, err := r.getAttendanceOutcomeData(ctx, tx, assessmentIDs)
	if err != nil {
		return nil, err
	}

	outcomeID2OutcomeMap := map[string]*entity.Outcome{}
	{
		outcomes, err := GetOutcomeModel().GetLearningOutcomesByIDs(ctx, tx, data.AllOutcomeIDs, cmd.Operator)
		if err != nil {
			log.Error(ctx, "get student detail report: get learning outcome failed by ids",
				log.Err(err),
				log.Strings("outcome_ids", data.AllOutcomeIDs),
				log.Any("cmd", cmd),
			)
			return nil, err
		}
		for _, outcome := range outcomes {
			outcomeID2OutcomeMap[outcome.ID] = outcome
		}
	}

	var result = entity.StudentReport{StudentName: student.Name, AssessmentIDs: assessmentIDs}
	{
		if !data.AllAttendanceIDExistsMap[cmd.StudentID] {
			result.Attend = false
			return &result, nil
		}

		result.Attend = true

		developmentalList, err := GetDevelopmentalModel().Query(ctx, &da.DevelopmentalCondition{})
		if err != nil {
			log.Error(ctx, "get student detail report: query all developmental failed",
				log.Err(err),
				log.Any("cmd", cmd),
				log.Any("operator", operator),
			)
			return nil, err
		}
		outcomeIDsTranslatorFunc, err := makeLatestOutcomeIDsTranslator(ctx, tx, data.AllOutcomeIDs, operator)
		if err != nil {
			log.Error(ctx, "get student detail report: make latest outcome ids translator failed",
				log.Err(err),
				log.Any("cmd", cmd),
				log.Any("data", data),
				log.Any("operator", operator),
			)
		}
		for _, developmental := range developmentalList {
			newItem := entity.StudentReportCategory{Name: developmental.Name}
			{
				outcomeIDs := outcomeIDsTranslatorFunc(data.AchievedAttendanceID2OutcomeIDsMap[cmd.StudentID])
				for _, outcomeID := range outcomeIDs {
					outcome := outcomeID2OutcomeMap[outcomeID]
					if outcome == nil {
						continue
					}
					if outcome.Developmental == developmental.ID {
						newItem.AchievedItems = append(newItem.AchievedItems, outcome.Name)
					}
				}
			}
			{
				outcomeIDs := outcomeIDsTranslatorFunc(data.NotAchievedAttendanceID2OutcomeIDsMap[cmd.StudentID])
				for _, outcomeID := range outcomeIDs {
					outcome := outcomeID2OutcomeMap[outcomeID]
					if outcome == nil {
						continue
					}
					if outcome.Developmental == developmental.ID {
						newItem.NotAchievedItems = append(newItem.NotAchievedItems, outcome.Name)
					}
				}
			}
			{
				outcomeIDs := outcomeIDsTranslatorFunc(data.SkipAttendanceID2OutcomeIDsMap[cmd.StudentID])
				for _, outcomeID := range outcomeIDs {
					outcome := outcomeID2OutcomeMap[outcomeID]
					if outcome == nil {
						continue
					}
					if outcome.Developmental == developmental.ID {
						newItem.NotAttemptedItems = append(newItem.NotAttemptedItems, outcome.Name)
					}
				}
			}
			result.Categories = append(result.Categories, &newItem)
		}
	}

	return &result, nil
}

func (r *reportModel) getAssessmentIDs(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, classID string, lessonPlanID string) ([]string, error) {
	scheduleIDs, err := r.getScheduleIDs(ctx, tx, operator, classID, lessonPlanID)
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

func (r *reportModel) getScheduleIDs(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, classID string, lessonPlanID string) ([]string, error) {
	log.Debug(ctx, "get schedule ids: before call GetScheduleModel().Query()")
	result, err := GetScheduleModel().GetScheduleIDsByCondition(ctx, tx, operator, &entity.ScheduleIDsCondition{
		ClassID:      classID,
		LessonPlanID: lessonPlanID,
		Status:       entity.ScheduleStatusClosed,
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

func (r *reportModel) getAttendanceOutcomeData(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string) (*entity.ReportAttendanceOutcomeData, error) {
	attendanceIDExistsMap, allOutcomeIDs, allAttendanceID2AssessmentOutcomesMap, err := r.getOutcomeRelatedData(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "get attendance outcome data: get outcome related data failed",
			log.Err(err),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	allAttendanceID2OutcomeIDsMap := r.getAllAttendanceID2OutcomeIDsMap(allAttendanceID2AssessmentOutcomesMap)

	skipAttendanceID2OutcomeIDsMap := r.getSkipAttendanceID2OutcomeIDsMap(allAttendanceID2AssessmentOutcomesMap)

	achievedAttendanceID2OutcomeIDsMap, err := r.getAchievedAttendanceID2OutcomeIDsMap(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "get attendance outcome data: get achieved attendance id to outcome ids map failed",
			log.Err(err),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}

	notAchievedAttendanceID2OutcomeIDsMap := map[string][]string{}
	for attendanceID, outcomeIDs := range allAttendanceID2OutcomeIDsMap {
		var excludeOutcomeIDs []string
		excludeOutcomeIDs = append(excludeOutcomeIDs, achievedAttendanceID2OutcomeIDsMap[attendanceID]...)
		excludeOutcomeIDs = append(excludeOutcomeIDs, skipAttendanceID2OutcomeIDsMap[attendanceID]...)
		notAchievedAttendanceID2OutcomeIDsMap[attendanceID] = utils.ExcludeStrings(outcomeIDs, excludeOutcomeIDs)
	}

	return &entity.ReportAttendanceOutcomeData{
		AllOutcomeIDs:                         allOutcomeIDs,
		AllAttendanceIDExistsMap:              attendanceIDExistsMap,
		AllAttendanceID2OutcomeIDsMap:         allAttendanceID2OutcomeIDsMap,
		SkipAttendanceID2OutcomeIDsMap:        skipAttendanceID2OutcomeIDsMap,
		AchievedAttendanceID2OutcomeIDsMap:    achievedAttendanceID2OutcomeIDsMap,
		NotAchievedAttendanceID2OutcomeIDsMap: notAchievedAttendanceID2OutcomeIDsMap,
	}, nil
}

func (r *reportModel) getOutcomeRelatedData(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string) (map[string]bool, []string, map[string][]*entity.AssessmentOutcome, error) {
	var (
		attendanceIDExistsMap                 = map[string]bool{}
		allOutcomeIDs                         []string
		allAttendanceID2AssessmentOutcomesMap = map[string][]*entity.AssessmentOutcome{}
	)

	checked := true
	var assessmentAttendances []*entity.AssessmentAttendance
	if err := da.GetAssessmentAttendanceDA().QueryTx(ctx, tx, &da.AssessmentAttendanceCondition{
		AssessmentIDs: assessmentIDs,
		Checked:       &checked,
	}, &assessmentAttendances); err != nil {
		log.Error(ctx, "list students report: query assessment attendances failed",
			log.Err(err),
			log.Any("assessment_ids", assessmentIDs),
		)
		return nil, nil, nil, err
	}
	var attendanceID2AssessmentIDsMap = map[string][]string{}
	for _, assessmentAttendance := range assessmentAttendances {
		attendanceID2AssessmentIDsMap[assessmentAttendance.AttendanceID] = append(attendanceID2AssessmentIDsMap[assessmentAttendance.AttendanceID], assessmentAttendance.AssessmentID)
		attendanceIDExistsMap[assessmentAttendance.AttendanceID] = true
	}
	assessmentOutcomes, err := da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "list students report: batch get assessment outcomes failed",
			log.Err(err),
			log.Any("assessment_ids", assessmentIDs),
		)
		return nil, nil, nil, err
	}
	var assessmentID2AssessmentOutcomesMap = map[string][]*entity.AssessmentOutcome{}
	for _, assessmentOutcome := range assessmentOutcomes {
		allOutcomeIDs = append(allOutcomeIDs, assessmentOutcome.OutcomeID)
		assessmentID2AssessmentOutcomesMap[assessmentOutcome.AssessmentID] = append(assessmentID2AssessmentOutcomesMap[assessmentOutcome.AssessmentID], assessmentOutcome)
	}
	allOutcomeIDs = utils.SliceDeduplication(allOutcomeIDs)
	for attendanceID, assessmentIDs := range attendanceID2AssessmentIDsMap {
		for _, assessmentID := range assessmentIDs {
			allAttendanceID2AssessmentOutcomesMap[attendanceID] = append(allAttendanceID2AssessmentOutcomesMap[attendanceID], assessmentID2AssessmentOutcomesMap[assessmentID]...)
		}
	}

	return attendanceIDExistsMap, allOutcomeIDs, allAttendanceID2AssessmentOutcomesMap, nil
}

func (r *reportModel) getAllAttendanceID2OutcomeIDsMap(allAttendanceID2AssessmentOutcomesMap map[string][]*entity.AssessmentOutcome) map[string][]string {
	allAttendanceID2OutcomeIDsMap := map[string][]string{}
	for attendanceID, assessmentOutcomes := range allAttendanceID2AssessmentOutcomesMap {
		for _, assessmentOutcome := range assessmentOutcomes {
			allAttendanceID2OutcomeIDsMap[attendanceID] = append(allAttendanceID2OutcomeIDsMap[attendanceID], assessmentOutcome.OutcomeID)
		}
	}
	for attendanceID, outcomeIDs := range allAttendanceID2OutcomeIDsMap {
		allAttendanceID2OutcomeIDsMap[attendanceID] = utils.SliceDeduplication(outcomeIDs)
	}
	return allAttendanceID2OutcomeIDsMap
}

func (r *reportModel) getSkipAttendanceID2OutcomeIDsMap(allAttendanceID2AssessmentOutcomesMap map[string][]*entity.AssessmentOutcome) map[string][]string {
	skipAttendanceID2OutcomeIDsMap := map[string][]string{}
	for attendanceID, assessmentOutcomes := range allAttendanceID2AssessmentOutcomesMap {
		skipOutcomeIDMap := map[string]bool{}
		for _, assessmentOutcome := range assessmentOutcomes {
			_, exists := skipOutcomeIDMap[assessmentOutcome.OutcomeID]
			if !exists && assessmentOutcome.Skip {
				skipOutcomeIDMap[assessmentOutcome.OutcomeID] = true
			}
			if !assessmentOutcome.Skip {
				skipOutcomeIDMap[assessmentOutcome.OutcomeID] = false
			}
		}
		for skipOutcomeID, ok := range skipOutcomeIDMap {
			if ok {
				skipAttendanceID2OutcomeIDsMap[attendanceID] = append(skipAttendanceID2OutcomeIDsMap[attendanceID], skipOutcomeID)
			}
		}
	}
	return skipAttendanceID2OutcomeIDsMap
}

func (r *reportModel) getAchievedAttendanceID2OutcomeIDsMap(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string) (map[string][]string, error) {
	achievedAttendanceID2OutcomeIDsMap := map[string][]string{}
	outcomeAttendances, err := da.GetOutcomeAttendanceDA().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "list students report: batch get outcome attendance failed",
			log.Err(err),
			log.Any("assessment_ids", assessmentIDs),
		)
		return nil, err
	}
	for _, outcomeAttendance := range outcomeAttendances {
		achievedAttendanceID2OutcomeIDsMap[outcomeAttendance.AttendanceID] = append(achievedAttendanceID2OutcomeIDsMap[outcomeAttendance.AttendanceID], outcomeAttendance.OutcomeID)
	}
	for k, v := range achievedAttendanceID2OutcomeIDsMap {
		achievedAttendanceID2OutcomeIDsMap[k] = utils.SliceDeduplication(v)
	}
	return achievedAttendanceID2OutcomeIDsMap, err
}

func (r *reportModel) hasReportPermission(ctx context.Context, operator *entity.Operator, teacherID string) (bool, error) {
	hasP603, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.ReportTeacherReports603)
	if err != nil {
		log.Error(ctx, "has report permission: check permission 603 failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("teacher_id", teacherID),
		)
		return false, err
	}
	if !hasP603 {
		return false, nil
	}

	hasP614, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.ReportViewMyReports614)
	if err != nil {
		log.Error(ctx, "has report permission: check permission 614 failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("teacher_id", teacherID),
		)
		return false, err
	}
	if hasP614 && operator.UserID == teacherID {
		return true, nil
	}

	hasP610, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.ReportViewReports610)
	if err != nil {
		log.Error(ctx, "has report permission: check permission 610 failed",
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
			log.Error(ctx, "has report permission: call external \"GetByOrganization()\" failed",
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
		schools, err := external.GetSchoolServiceProvider().GetSchoolsAssociatedWithUserID(ctx, operator, operator.UserID)
		for _, school := range schools {
			schoolIDs = append(schoolIDs, school.ID)
		}
		schoolID2TeachersMap, err := external.GetTeacherServiceProvider().GetBySchools(ctx, operator, schoolIDs)
		if err != nil {
			log.Error(ctx, "has report permission: call external \"GetBySchools()\" failed",
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

func (r *reportModel) GetTeacherReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, teacherID string) (*entity.TeacherReport, error) {
	var assessmentIDs []string
	{
		var assessments []*entity.Assessment
		if err := da.GetAssessmentDA().Query(ctx, &da.QueryAssessmentsCondition{
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
		assessmentOutcomes, err := da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs)
		if err != nil {
			log.Error(ctx, "get teacher report: batch get failed by assessment ids",
				log.Err(err),
				log.Any("operator", operator),
				log.String("teacher_id", teacherID),
			)
			return nil, err
		}
		var outcomeIDs []string
		for _, item := range assessmentOutcomes {
			outcomeIDs = append(outcomeIDs, item.OutcomeID)
		}
		utils.SliceDeduplication(outcomeIDs)
		outcomeIDsTranslatorFunc, err := makeLatestOutcomeIDsTranslator(ctx, tx, outcomeIDs, operator)
		if err != nil {
			log.Error(ctx, "get student detail report: make latest outcome ids translator failed",
				log.Err(err),
				log.Any("outcome_ids", outcomeIDs),
				log.Any("operator", operator),
			)
		}
		outcomeIDs = outcomeIDsTranslatorFunc(outcomeIDs)
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
		developmentalList, err := GetDevelopmentalModel().Query(ctx, &da.DevelopmentalCondition{})
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

func makeLatestOutcomeIDsTranslator(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string, operator *entity.Operator) (func([]string) []string, error) {
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
			}
		}
		return utils.SliceDeduplication(result)
	}, nil
}
