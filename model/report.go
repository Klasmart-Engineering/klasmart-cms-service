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
)

type IReportModel interface {
	ListStudentsReport(ctx context.Context, tx *dbo.DBContext, cmd entity.ListStudentsReportCommand) (*entity.StudentsReport, error)
	GetStudentDetailReport(ctx context.Context, tx *dbo.DBContext, cmd entity.GetStudentDetailReportCommand) (*entity.StudentDetailReport, error)
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

func (r *reportModel) ListStudentsReport(ctx context.Context, tx *dbo.DBContext, cmd entity.ListStudentsReportCommand) (*entity.StudentsReport, error) {
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
	}

	var students []*external.Student
	{
		var err error
		students, err = external.GetClassServiceProvider().GetStudents(ctx, cmd.ClassID)
		if err != nil {
			log.Error(ctx, "list students report: get students",
				log.Err(err),
				log.Any("cmd", cmd),
			)
			return nil, err
		}
	}

	assessmentIDs, err := r.getAssessmentIDs(ctx, tx, cmd.TeacherID, cmd.ClassID, cmd.LessonPlanID)
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
	var result entity.StudentsReport
	for _, student := range students {
		newItem := entity.StudentReportItem{StudentID: student.ID, StudentName: student.Name}

		if !data.AllAttendanceIDExistsMap[student.ID] {
			newItem.Attend = false
			result.Items = append(result.Items, newItem)
			continue
		}

		newItem.Attend = true
		newItem.AchievedCount = len(data.AchievedAttendanceID2OutcomeIDsMap[student.ID])
		newItem.NotAttemptedCount = len(data.SkipAttendanceID2OutcomeIDsMap[student.ID])
		newItem.NotAchievedCount = len(data.NotAchievedAttendanceID2OutcomeIDsMap[student.ID])

		result.Items = append(result.Items, newItem)
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

func (r *reportModel) GetStudentDetailReport(ctx context.Context, tx *dbo.DBContext, cmd entity.GetStudentDetailReportCommand) (*entity.StudentDetailReport, error) {
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
	}

	var student *external.Student
	{
		students, err := external.GetClassServiceProvider().GetStudents(ctx, cmd.ClassID)
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

	assessmentIDs, err := r.getAssessmentIDs(ctx, tx, cmd.TeacherID, cmd.ClassID, cmd.LessonPlanID)
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

	var result = entity.StudentDetailReport{StudentName: student.Name}
	{
		if !data.AllAttendanceIDExistsMap[cmd.StudentID] {
			result.Attend = false
			return &result, nil
		}

		result.Attend = true

		categories := []entity.ReportCategory{
			entity.ReportCategorySpeechLanguagesSkills,
			entity.ReportCategoryFineMotorSkills,
			entity.ReportCategoryGrossMotorSkills,
			entity.ReportCategoryCognitiveSkills,
			entity.ReportCategoryPersonalDevelopment,
			entity.ReportCategoryLanguageAndNumeracySkills,
			entity.ReportCategorySocialAndEmotional,
			entity.ReportCategoryOral,
			entity.ReportCategoryLiteracy,
			entity.ReportCategoryWholeChild,
			entity.ReportCategoryKnowledge,
		}

		categoryMap := map[string]string{}
		{
			var categoryIDs []string
			for _, outcomeID := range data.AllOutcomeIDs {
				outcome := outcomeID2OutcomeMap[outcomeID]
				if outcome == nil {
					continue
				}
				categoryIDs = append(categoryIDs, outcome.Developmental)
			}
			categories, err := GetDevelopmentalModel().Query(ctx, &da.DevelopmentalCondition{
				IDs: entity.NullStrings{
					Strings: categoryIDs,
					Valid:   len(categoryIDs) != 0,
				},
			})
			if err != nil {
				log.Error(ctx, "get student detail report: batch get developmental failed",
					log.Err(err),
					log.Strings("category_ids", categoryIDs),
					log.Any("cmd", cmd),
				)
				return nil, err
			}
			for _, item := range categories {
				categoryMap[item.ID] = item.Name
			}
		}

		for _, category := range categories {
			newItem := entity.StudentReportCategory{Name: category}
			{
				outcomeIDs := data.AchievedAttendanceID2OutcomeIDsMap[cmd.StudentID]
				for _, outcomeID := range outcomeIDs {
					outcome := outcomeID2OutcomeMap[outcomeID]
					if outcome == nil {
						continue
					}
					if categoryMap[outcome.Developmental] == string(category) {
						newItem.AchievedItems = append(newItem.AchievedItems, outcome.Name)
					}
				}
			}
			{
				outcomeIDs := data.NotAchievedAttendanceID2OutcomeIDsMap[cmd.StudentID]
				for _, outcomeID := range outcomeIDs {
					outcome := outcomeID2OutcomeMap[outcomeID]
					if outcome == nil {
						continue
					}
					if categoryMap[outcome.Developmental] == string(category) {
						newItem.NotAchievedItems = append(newItem.NotAchievedItems, outcome.Name)
					}
				}
			}
			{
				outcomeIDs := data.SkipAttendanceID2OutcomeIDsMap[cmd.StudentID]
				for _, outcomeID := range outcomeIDs {
					outcome := outcomeID2OutcomeMap[outcomeID]
					if outcome == nil {
						continue
					}
					if categoryMap[outcome.Developmental] == string(category) {
						newItem.NotAttemptedItems = append(newItem.NotAttemptedItems, outcome.Name)
					}
				}
			}
			result.Categories = append(result.Categories, newItem)
		}
	}

	return &result, nil
}

func (r *reportModel) getAssessmentIDs(ctx context.Context, tx *dbo.DBContext, classID string, teacherID string, lessonPlanID string) ([]string, error) {
	scheduleIDs, err := r.getScheduleIDs(ctx, tx, teacherID, classID, lessonPlanID)
	if err != nil {
		log.Error(ctx, "get assessment ids: get schedule ids failed",
			log.Err(err),
			log.String("class_id", classID),
			log.String("teacher_id", teacherID),
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

func (r *reportModel) getScheduleIDs(ctx context.Context, tx *dbo.DBContext, classID string, teacherID string, lessonPlanID string) ([]string, error) {
	items, err := GetScheduleModel().Query(ctx, &da.ScheduleCondition{
		TeacherID: sql.NullString{
			String: teacherID,
			Valid:  true,
		},
		LessonPlanID: sql.NullString{
			String: lessonPlanID,
			Valid:  true,
		},
		ClassID: sql.NullString{
			String: classID,
			Valid:  true,
		},
		Status: sql.NullString{
			String: string(entity.ScheduleStatusClosed),
			Valid:  true,
		},
	})
	if err != nil {
		return nil, err
	}
	var result []string
	for _, item := range items {
		result = append(result, item.ID)
	}
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

	assessmentAttendances, err := da.GetAssessmentAttendanceDA().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "list students report: batch get assessment attendances failed",
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
