package model

import (
	"context"
	"database/sql"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
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
			result.Items = append(result.Items, newItem)
			continue
		}

		newItem.AllAchievedCount = len(data.AchievedAttendanceID2OutcomeIDsMap[student.ID])
		newItem.NotAchievedCount = len(data.SkipAttendanceID2OutcomeIDsMap[student.ID])
		newItem.NotAchievedCount = len(data.NotAchievedAttendanceID2OutcomeIDsMap[student.ID])

		result.Items = append(result.Items, newItem)
	}

	switch cmd.SortBy {
	case entity.ReportSortByDescending:
		sort.Sort(sort.Reverse(entity.StudentReportItemSortByName(result.Items)))
	case entity.ReportSortByAscending:
		fallthrough
	default:
		sort.Sort(entity.StudentReportItemSortByName(result.Items))
	}

	return &result, nil
}

func (r *reportModel) GetStudentDetailReport(ctx context.Context, tx *dbo.DBContext, cmd entity.GetStudentDetailReportCommand) (*entity.StudentDetailReport, error) {
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
			return nil, errors.New("get student detail report: not found student in class")
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
			return &result, nil
		}

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

		for _, category := range categories {
			newItem := entity.StudentReportCategory{Name: category}
			{
				outcomeIDs := data.AchievedAttendanceID2OutcomeIDsMap[cmd.StudentID]
				for _, outcomeID := range outcomeIDs {
					outcome := outcomeID2OutcomeMap[outcomeID]
					if outcome == nil {
						continue
					}
					if outcome.Developmental == string(category) {
						newItem.AllAchievedItems = append(newItem.AllAchievedItems, outcome.Name)
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
					if outcome.Developmental == string(category) {
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
					if outcome.Developmental == string(category) {
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
	var (
		allOutcomeIDs                         []string
		allAttendanceID2AssessmentOutcomesMap = map[string][]*entity.AssessmentOutcome{}
		attendanceIDExistsMap                 = map[string]bool{}
	)
	{
		assessmentAttendances, err := da.GetAssessmentAttendanceDA().BatchGetByAssessmentIDs(ctx, tx, assessmentIDs)
		if err != nil {
			log.Error(ctx, "list students report: batch get assessment attendances failed",
				log.Err(err),
				log.Any("assessment_ids", assessmentIDs),
			)
			return nil, err
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
			return nil, err
		}
		var assessmentID2AssessmentOutcomesMap = map[string][]*entity.AssessmentOutcome{}
		for _, assessmentOutcome := range assessmentOutcomes {
			allOutcomeIDs = append(allOutcomeIDs, assessmentOutcome.OutcomeID)
			assessmentID2AssessmentOutcomesMap[assessmentOutcome.AssessmentID] = append(assessmentID2AssessmentOutcomesMap[assessmentOutcome.AssessmentID], assessmentOutcome)
		}
		allOutcomeIDs = r.uniqueStrings(allOutcomeIDs)
		for attendanceID, assessmentIDs := range attendanceID2AssessmentIDsMap {
			for _, assessmentID := range assessmentIDs {
				allAttendanceID2AssessmentOutcomesMap[attendanceID] = append(allAttendanceID2AssessmentOutcomesMap[attendanceID], assessmentID2AssessmentOutcomesMap[assessmentID]...)
			}
		}
	}

	allAttendanceID2OutcomeIDsMap := map[string][]string{}
	{
		for attendanceID, assessmentOutcomes := range allAttendanceID2AssessmentOutcomesMap {
			for _, assessmentOutcome := range assessmentOutcomes {
				allAttendanceID2OutcomeIDsMap[attendanceID] = append(allAttendanceID2OutcomeIDsMap[attendanceID], assessmentOutcome.OutcomeID)
			}
		}
		for attendanceID, outcomeIDs := range allAttendanceID2OutcomeIDsMap {
			allAttendanceID2OutcomeIDsMap[attendanceID] = r.uniqueStrings(outcomeIDs)
		}
	}

	skipAttendanceID2OutcomeIDsMap := map[string][]string{}
	{
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
			for skipOutcomeID := range skipOutcomeIDMap {
				skipAttendanceID2OutcomeIDsMap[attendanceID] = append(skipAttendanceID2OutcomeIDsMap[attendanceID], skipOutcomeID)
			}
		}
	}

	achievedAttendanceID2OutcomeIDsMap := map[string][]string{}
	{
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
			achievedAttendanceID2OutcomeIDsMap[k] = r.uniqueStrings(v)
		}
	}

	notAchievedAttendanceID2OutcomeIDsMap := map[string][]string{}
	for attendanceID, outcomeIDs := range allAttendanceID2OutcomeIDsMap {
		var excludeOutcomeIDs []string
		excludeOutcomeIDs = append(excludeOutcomeIDs, achievedAttendanceID2OutcomeIDsMap[attendanceID]...)
		excludeOutcomeIDs = append(excludeOutcomeIDs, skipAttendanceID2OutcomeIDsMap[attendanceID]...)
		notAchievedAttendanceID2OutcomeIDsMap[attendanceID] = r.excludeStrings(outcomeIDs, excludeOutcomeIDs)
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

func (r *reportModel) uniqueStrings(input []string) []string {
	m := map[string]bool{}
	for _, item := range input {
		m[item] = true
	}
	var result []string
	for k := range m {
		result = append(result, k)
	}
	return result
}

func (r *reportModel) excludeStrings(source []string, targets []string) []string {
	var result []string
	for _, item := range source {
		find := false
		for _, target := range targets {
			if item == target {
				find = true
			}
		}
		if !find {
			result = append(result, item)
		}
	}
	return result
}
