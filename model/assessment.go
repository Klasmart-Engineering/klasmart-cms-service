package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
	"time"
)

type IAssessmentModel interface {
	Detail(ctx context.Context, tx *dbo.DBContext, id string) (*entity.AssessmentDetailView, error)
	List(ctx context.Context, tx *dbo.DBContext, cmd entity.ListAssessmentsCommand) (*entity.ListAssessmentsResult, error)
	Add(ctx context.Context, cmd entity.AddAssessmentCommand) (string, error)
	Update(ctx context.Context, cmd entity.UpdateAssessmentCommand) error
}

var (
	assessmentModelInstance     IAssessmentModel
	assessmentModelInstanceOnce = sync.Once{}
)

func GetAssessmentModel() IAssessmentModel {
	assessmentModelInstanceOnce.Do(func() {
		assessmentModelInstance = &assessmentModel{}
	})
	return assessmentModelInstance
}

type assessmentModel struct{}

func (a *assessmentModel) Detail(ctx context.Context, tx *dbo.DBContext, id string) (*entity.AssessmentDetailView, error) {
	var result entity.AssessmentDetailView

	item, err := da.GetAssessmentDA().GetExcludeSoftDeleted(ctx, tx, id)
	if err != nil {
		log.Error(ctx, "get assessment detail: get from da failed",
			log.Err(err),
			log.String("id", "id"),
		)
		return nil, err
	}
	result = entity.AssessmentDetailView{
		ID:           item.ID,
		Title:        item.Title,
		ClassEndTime: item.ClassEndTime,
		ClassLength:  item.ClassLength,
		CompleteTime: item.CompleteTime,
	}

	// fill attendances
	{
		attendanceIDs, err := da.GetAssessmentAttendanceDA().GetAttendanceIDsByAssessmentID(ctx, tx, id)
		if err != nil {
			log.Error(ctx, "get assessment detail: get assessment attendances failed",
				log.Err(err),
				log.String("id", "id"),
			)
			return nil, err
		}
		studentService, err := external.GetStudentServiceProvider()
		if err != nil {
			log.Error(ctx, "get assessment detail: get student service failed",
				log.Err(err),
				log.String("id", id),
			)
			return nil, err
		}
		students, err := studentService.BatchGet(ctx, attendanceIDs)
		if err != nil {
			log.Error(ctx, "get assessment detail: batch get student failed",
				log.Err(err),
				log.String("id", id),
				log.Strings("attendance_ids", attendanceIDs),
			)
			return nil, err
		}
		for _, student := range students {
			result.Attendances = append(result.Attendances, entity.AssessmentAttendanceView{
				ID:   student.ID,
				Name: student.Name,
			})
		}
	}

	// fill subject
	{
		nameMap, err := a.getSubjectNameMap(ctx, []string{item.SubjectID})
		if err != nil {
			log.Error(ctx, "get assessment detail: get subject name map failed",
				log.Err(err),
				log.String("id", id),
				log.String("subject_id", item.SubjectID),
			)
			return nil, err
		}
		result.Subject = entity.AssessmentSubject{
			ID:   item.SubjectID,
			Name: nameMap[item.SubjectID],
		}
	}

	// fill teacher
	{
		nameMap, err := a.getTeacherNameMap(ctx, []string{item.TeacherID})
		if err != nil {
			log.Error(ctx, "get assessment detail: get teacher name map failed",
				log.Err(err),
				log.String("id", id),
				log.String("teacher_id", item.TeacherID),
			)
			return nil, err
		}
		result.Teacher = entity.AssessmentTeacher{
			ID:   item.TeacherID,
			Name: nameMap[item.TeacherID],
		}
	}

	// fill number of outcomes and activities
	{
		schedule, err := GetScheduleModel().GetPlainByID(ctx, item.ScheduleID)
		if err != nil {
			log.Error(ctx, "add assessment: get schedule failed by id",
				log.Err(err),
				log.Any("id", id),
				log.Any("schedule_id", item.ScheduleID),
			)
			return nil, err
		}
		counts, err := GetContentModel().ContentDataCount(ctx, tx, schedule.LessonPlanID)
		if err != nil {
			log.Error(ctx, "add assessment: get number of activities and outcomes failed",
				log.Err(err),
				log.Any("id", id),
				log.Any("lesson_plan_id", schedule.LessonPlanID),
			)
		}
		result.NumberOfActivities = counts.SubContentCount
		result.NumberOfOutcomes = counts.OutcomesCount
	}

	// fill outcome attendance maps
	{
		outcomeMap := map[string]*entity.Outcome{}
		outcomeIDs, err := da.GetAssessmentOutcomeDA().GetOutcomeIDsByAssessmentID(ctx, tx, id)
		if err != nil {
			log.Error(ctx, "get assessment detail: get outcome ids failed by assessment id",
				log.Err(err),
				log.String("id", id),
			)
			return nil, err
		}
		if len(outcomeIDs) > 0 {
			outcomes, err := GetOutcomeModel().GetLearningOutcomesByIDs(ctx, tx, outcomeIDs, &entity.Operator{})
			if err != nil {
				log.Error(ctx, "get assessment detail: batch get outcomes failed by outcome ids",
					log.Err(err),
					log.Strings("outcome_ids", outcomeIDs),
				)
				return nil, err
			}
			for _, outcome := range outcomes {
				outcomeMap[outcome.ID] = outcome
			}
			outcomeAttendances, err := da.GetOutcomeAttendanceDA().BatchGetByAssessmentIDAndOutcomeIDs(ctx, tx, id, outcomeIDs)
			outcomeAttendanceMap := map[string]*struct {
				AttendanceIDs []string
				Skip          bool
			}{}
			for _, item := range outcomeAttendances {
				v, ok := outcomeAttendanceMap[item.OutcomeID]
				if !ok {
					v = &struct {
						AttendanceIDs []string
						Skip          bool
					}{}
				}
				v.AttendanceIDs = append(outcomeAttendanceMap[item.OutcomeID].AttendanceIDs, item.AttendanceID)
			}
			if err != nil {
				log.Error(ctx, "get assessment detail: batch get outcomes failed by outcome ids",
					log.Err(err),
					log.Strings("outcome_ids", outcomeIDs),
				)
				return nil, err
			}
			for _, outcome := range outcomes {
				newItem := entity.OutcomeAttendanceMapView{
					OutcomeID:   outcome.ID,
					OutcomeName: outcome.Name,
					Assumed:     outcome.Assumed,
				}
				if outcomeAttendanceMap[outcome.ID] != nil {
					newItem.Skip = outcomeAttendanceMap[outcome.ID].Skip
					newItem.AttendanceIDs = outcomeAttendanceMap[outcome.ID].AttendanceIDs
				}
				result.OutcomeAttendanceMaps = append(result.OutcomeAttendanceMaps, newItem)
			}
		}
	}

	return &result, nil
}

func (a *assessmentModel) List(ctx context.Context, tx *dbo.DBContext, cmd entity.ListAssessmentsCommand) (*entity.ListAssessmentsResult, error) {
	cond := da.QueryAssessmentsCondition{
		Status:   cmd.Status,
		OrderBy:  cmd.OrderBy,
		Page:     cmd.Page,
		PageSize: cmd.PageSize,
	}
	{
		if cmd.TeacherName != nil {
			teacherService, err := external.GetTeacherServiceProvider()
			if err != nil {
				log.Error(ctx, "list assessments: get teacher service failed",
					log.Err(err),
					log.Any("cmd", cmd),
				)
				return nil, err
			}
			items, err := teacherService.Query(ctx, *cmd.TeacherName)
			if err != nil {
				log.Error(ctx, "list assessments: query teacher service failed",
					log.Err(err),
					log.Any("cmd", cmd),
				)
				return nil, err
			}
			if len(items) > 0 {
				if cond.TeacherIDs == nil {
					cond.TeacherIDs = &[]string{}
				}
				for _, item := range items {
					*cond.TeacherIDs = append(*cond.TeacherIDs, item.ID)
				}
			}
		}
	}

	var items []*entity.Assessment
	if err := da.GetAssessmentDA().QueryTx(ctx, tx, &cond, &items); err != nil {
		log.Error(ctx, "list assessments: query tx failed",
			log.Err(err),
			log.Any("cmd", cmd),
			log.Any("cond", cond),
		)
		return nil, err
	}

	total, err := da.GetAssessmentDA().CountTx(ctx, tx, &cond, &entity.Assessment{})
	if err != nil {
		log.Error(ctx, "list assessments: count tx failed",
			log.Err(err),
			log.Any("cmd", cmd),
			log.Any("cond", cond),
		)
		return nil, err
	}

	var subjectIDs, programIDs, teacherIDs []string
	for _, item := range items {
		subjectIDs = append(subjectIDs, item.SubjectID)
		programIDs = append(programIDs, item.ProgramID)
		teacherIDs = append(teacherIDs, item.TeacherID)
	}

	subjectNameMap, err := a.getSubjectNameMap(ctx, subjectIDs)
	if err != nil {
		log.Error(ctx, "detail: get subject name map failed",
			log.Err(err),
			log.Strings("subject_ids", subjectIDs),
		)
		return nil, err
	}

	programNameMap, err := a.getProgramNameMap(ctx, programIDs)
	if err != nil {
		log.Error(ctx, "detail: get program name map failed",
			log.Err(err),
			log.Strings("program_ids", programIDs),
		)
		return nil, err
	}

	teacherNameMap, err := a.getTeacherNameMap(ctx, teacherIDs)
	if err != nil {
		log.Error(ctx, "detail: get teacher name map failed",
			log.Err(err),
			log.Strings("teacher_ids", teacherIDs),
		)
		return nil, err
	}

	result := entity.ListAssessmentsResult{Total: total}
	for _, item := range items {
		result.Items = append(result.Items, &entity.AssessmentListView{
			ID:    item.ID,
			Title: item.Title,
			Subject: entity.AssessmentSubject{
				ID:   item.SubjectID,
				Name: subjectNameMap[item.SubjectID],
			},
			Program: entity.AssessmentProgram{
				ID:   item.ProgramID,
				Name: programNameMap[item.ProgramID],
			},
			Teacher: entity.AssessmentTeacher{
				ID:   item.TeacherID,
				Name: teacherNameMap[item.TeacherID],
			},
			ClassEndTime: item.ClassEndTime,
			CompleteTime: item.CompleteTime,
			Status:       item.Status,
		})
	}

	return &result, err
}

func (a *assessmentModel) getProgramNameMap(ctx context.Context, programIDs []string) (map[string]string, error) {
	programNameMap := map[string]string{}
	programService, err := external.GetProgramServiceProvider()
	if err != nil {
		log.Error(ctx, "list assessments: get program service failed",
			log.Err(err),
			log.Strings("program_ids", programIDs),
		)
		return nil, err
	}
	items, err := programService.BatchGet(ctx, programIDs)
	if err != nil {
		log.Error(ctx, "list assessments: batch get program failed",
			log.Err(err),
			log.Strings("program_ids", programIDs),
		)
		return nil, err
	}
	for _, item := range items {
		programNameMap[item.ID] = item.Name
	}
	return programNameMap, nil
}

func (a *assessmentModel) getSubjectNameMap(ctx context.Context, subjectIDs []string) (map[string]string, error) {
	subjectNameMap := map[string]string{}
	subjectService, err := external.GetSubjectServiceProvider()
	if err != nil {
		log.Error(ctx, "list assessments: get subject service failed",
			log.Err(err),
			log.Strings("subject_ids", subjectIDs),
		)
		return nil, err
	}
	items, err := subjectService.BatchGet(ctx, subjectIDs)
	if err != nil {
		log.Error(ctx, "list assessments: batch get subject failed",
			log.Err(err),
			log.Strings("subject_ids", subjectIDs),
		)
		return nil, err
	}
	for _, item := range items {
		subjectNameMap[item.ID] = item.Name
	}
	return subjectNameMap, nil
}

func (a *assessmentModel) getTeacherNameMap(ctx context.Context, teacherIDs []string) (map[string]string, error) {
	teacherNameMap := map[string]string{}
	teacherNameService, err := external.GetTeacherServiceProvider()
	if err != nil {
		log.Error(ctx, "list assessments: get teacher service failed",
			log.Err(err),
			log.Strings("teacher_ids", teacherIDs),
		)
		return nil, err
	}
	items, err := teacherNameService.BatchGet(ctx, teacherIDs)
	if err != nil {
		log.Error(ctx, "list assessments: batch get teacher failed",
			log.Err(err),
			log.Strings("teacher_ids", teacherIDs),
		)
		return nil, err
	}
	for _, item := range items {
		teacherNameMap[item.ID] = item.Name
	}
	return teacherNameMap, nil
}

func (a *assessmentModel) Add(ctx context.Context, cmd entity.AddAssessmentCommand) (string, error) {
	// validate teacher id
	{
		ok, err := a.existsTeacherByID(ctx, cmd.TeacherID)
		if err != nil {
			log.Error(ctx, "add assessment: check teacher exists failed",
				log.Err(err),
				log.Any("cmd", cmd),
				log.String("teacher_id", cmd.TeacherID),
			)
			return "", err
		}
		if !ok {
			log.Info(ctx, "add assessment: teacher not found by id",
				log.Err(err),
				log.Any("cmd", cmd),
				log.String("teacher_id", cmd.TeacherID),
			)
			return "", constant.ErrInvalidArgs
		}
	}

	// validate subject id
	{
		ok, err := a.existsSubjectByID(ctx, cmd.SubjectID)
		if err != nil {
			log.Error(ctx, "add assessment: check subject exists failed",
				log.Err(err),
				log.Any("cmd", cmd),
				log.String("subject_id", cmd.SubjectID),
			)
			return "", err
		}
		if !ok {
			log.Info(ctx, "add assessment: subject not found by id",
				log.Err(err),
				log.Any("cmd", cmd),
				log.String("subject_id", cmd.SubjectID),
			)
			return "", constant.ErrInvalidArgs
		}
	}

	// validate program id
	{
		ok, err := a.existsProgramByID(ctx, cmd.ProgramID)
		if err != nil {
			log.Error(ctx, "add assessment: check program exists failed",
				log.Err(err),
				log.Any("cmd", cmd),
				log.String("program_id", cmd.ProgramID),
			)
			return "", err
		}
		if !ok {
			log.Info(ctx, "add assessment: program not found by id",
				log.Err(err),
				log.Any("cmd", cmd),
				log.String("program_id", cmd.ProgramID),
			)
			return "", constant.ErrInvalidArgs
		}
	}

	var outcomeIDs []string
	{
		schedule, err := GetScheduleModel().GetPlainByID(ctx, cmd.ScheduleID)
		if err != nil {
			log.Error(ctx, "add assessment: get schedule failed by id",
				log.Err(err),
				log.Any("cmd", cmd),
				log.Any("schedule_id", cmd.ScheduleID),
			)
			switch err {
			case constant.ErrRecordNotFound, dbo.ErrRecordNotFound:
				return "", constant.ErrInvalidArgs
			default:
				return "", err
			}
		}
		outcomeIDs, err = GetContentModel().GetContentOutcomeByID(ctx, dbo.MustGetDB(ctx), schedule.LessonPlanID)
		if err != nil {
			log.Error(ctx, "add assessment: get outcome failed by id",
				log.Err(err),
				log.Any("cmd", cmd),
				log.Any("lesson_plan_id", schedule.LessonPlanID),
			)
			return "", err
		}
	}

	var newID = utils.NewID()
	if err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		{
			nowUnix := time.Now().Unix()
			newItem := entity.Assessment{
				ID:           newID,
				ScheduleID:   cmd.ScheduleID,
				Title:        cmd.Title(),
				ProgramID:    cmd.ProgramID,
				SubjectID:    cmd.SubjectID,
				TeacherID:    cmd.TeacherID,
				ClassLength:  cmd.ClassLength,
				ClassEndTime: cmd.ClassEndTime,
				CreateTime:   nowUnix,
				UpdateTime:   nowUnix,
			}
			if len(outcomeIDs) == 0 {
				newItem.Status = entity.AssessmentStatusComplete
				newItem.CompleteTime = time.Now().Unix()
			} else {
				newItem.Status = entity.AssessmentStatusInProgress
			}
			if _, err := da.GetAssessmentDA().InsertTx(ctx, tx, &newItem); err != nil {
				log.Error(ctx, "add assessment: add failed",
					log.Err(err),
					log.Any("cmd", cmd),
					log.Any("new_item", newItem),
				)
				return err
			}
		}
		{
			var items []*entity.AssessmentOutcome
			for _, outcomeID := range outcomeIDs {
				items = append(items, &entity.AssessmentOutcome{
					ID:           utils.NewID(),
					AssessmentID: newID,
					OutcomeID:    outcomeID,
					Skip:         false,
				})
			}
			if len(items) > 0 {
				if err := da.GetAssessmentOutcomeDA().BatchInsert(ctx, tx, items); err != nil {
					log.Error(ctx, "add assessment: batch insert assessment outcome map failed",
						log.Err(err),
						log.Any("cmd", cmd),
						log.Any("items", items),
					)
					return err
				}
			}
		}
		{
			outcomeMap := map[string]*entity.Outcome{}
			outcomes, err := GetOutcomeModel().GetLearningOutcomesByIDs(ctx, tx, outcomeIDs, &entity.Operator{})
			if err != nil {
				log.Error(ctx, "get assessment detail: batch get outcomes failed by outcome ids",
					log.Err(err),
					log.Strings("outcome_ids", outcomeIDs),
				)
				return err
			}
			for _, outcome := range outcomes {
				outcomeMap[outcome.ID] = outcome
			}

			var items []*entity.OutcomeAttendance
			for _, outcomeID := range outcomeIDs {
				if outcomeMap[outcomeID] == nil {
					continue
				}
				if !outcomeMap[outcomeID].Assumed {
					continue
				}
				for _, attendanceID := range cmd.AttendanceIDs {
					items = append(items, &entity.OutcomeAttendance{
						ID:           utils.NewID(),
						AssessmentID: newID,
						OutcomeID:    outcomeID,
						AttendanceID: attendanceID,
					})
				}
			}
			if len(items) > 0 {
				if err := da.GetOutcomeAttendanceDA().BatchInsert(ctx, tx, items); err != nil {
					log.Error(ctx, "add assessment: batch insert outcome attendance map failed",
						log.Err(err),
						log.Any("cmd", cmd),
						log.Any("items", items),
					)
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return "", err
	}

	return newID, nil
}

func (a *assessmentModel) Update(ctx context.Context, cmd entity.UpdateAssessmentCommand) error {
	if err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		if cmd.Action == entity.UpdateAssessmentActionComplete {
			if err := da.GetAssessmentDA().UpdateStatus(ctx, tx, cmd.ID, entity.AssessmentStatusComplete); err != nil {
				log.Error(ctx, "update assessment: update status to complete failed",
					log.Err(err),
					log.Any("cmd", cmd),
				)
				return err
			}
		}
		if cmd.AttendanceIDs != nil {
			if err := da.GetAssessmentAttendanceDA().DeleteByAssessmentID(ctx, tx, cmd.ID); err != nil {
				log.Error(ctx, "update assessment: delete assessment attendance failed by assessment id",
					log.Err(err),
					log.Any("cmd", cmd),
				)
				return err
			}
			var items []*entity.AssessmentAttendance
			for _, attendanceID := range *cmd.AttendanceIDs {
				items = append(items, &entity.AssessmentAttendance{
					ID:           utils.NewID(),
					AssessmentID: cmd.ID,
					AttendanceID: attendanceID,
				})
			}
			if err := da.GetAssessmentAttendanceDA().BatchInsert(ctx, tx, items); err != nil {
				log.Error(ctx, "update assessment: batch insert assessment attendance map failed",
					log.Err(err),
					log.Any("items", items),
					log.Any("cmd", cmd),
				)
				return err
			}
		}
		if cmd.OutcomeAttendanceMaps != nil {
			var outcomeIDs []string
			for _, item := range *cmd.OutcomeAttendanceMaps {
				outcomeIDs = append(outcomeIDs, item.OutcomeID)
			}
			if err := da.GetOutcomeAttendanceDA().BatchDeleteByAssessmentIDAndOutcomeIDs(ctx, tx, cmd.ID, outcomeIDs); err != nil {
				log.Error(ctx, "update assessment: batch delete outcome attendance map failed by outcome ids",
					log.Err(err),
					log.Strings("outcome_ids", outcomeIDs),
					log.Any("cmd", cmd),
				)
				return err
			}
			var items []*entity.OutcomeAttendance
			for _, item := range *cmd.OutcomeAttendanceMaps {
				for _, attendanceID := range item.AttendanceIDs {
					items = append(items, &entity.OutcomeAttendance{
						ID:           utils.NewID(),
						AssessmentID: cmd.ID,
						OutcomeID:    item.OutcomeID,
						AttendanceID: attendanceID,
					})
				}
				if err := da.GetAssessmentOutcomeDA().UpdateSkipField(ctx, tx, cmd.ID, item.OutcomeID, item.Skip); err != nil {
					log.Error(ctx, "update assessment: batch insert outcome attendance map failed",
						log.Err(err),
						log.Any("cmd", cmd),
						log.String("id", cmd.ID),
						log.String("outcome_id", item.OutcomeID),
						log.Bool("skip", item.Skip),
					)
				}
			}
			if err := da.GetOutcomeAttendanceDA().BatchInsert(ctx, tx, items); err != nil {
				log.Error(ctx, "update assessment: batch insert outcome attendance map failed",
					log.Err(err),
					log.Any("items", items),
					log.Any("cmd", cmd),
				)
				return err
			}
		}
		return nil
	}); err != nil {
		log.Error(ctx, "update assessment: tx failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
	}
	return nil
}

func (a *assessmentModel) existsTeacherByID(ctx context.Context, id string) (bool, error) {
	teacherService, err := external.GetTeacherServiceProvider()
	if err != nil {
		log.Error(ctx, "check teacher exists: get teacher service failed",
			log.Err(err),
			log.String("id", id),
		)
	}
	if _, err := teacherService.BatchGet(ctx, []string{id}); err != nil {
		switch err {
		case dbo.ErrRecordNotFound, constant.ErrRecordNotFound:
			log.Info(ctx, "check teacher exists: not found teachers",
				log.Err(err),
				log.String("id", id),
			)
			return false, nil
		default:
			log.Error(ctx, "check teacher exists: batch get teachers failed",
				log.Err(err),
				log.String("id", id),
			)
			return false, err
		}
	}
	return true, nil
}

func (a *assessmentModel) existsSubjectByID(ctx context.Context, id string) (bool, error) {
	subjectService, err := external.GetSubjectServiceProvider()
	if err != nil {
		log.Error(ctx, "check subject exists: get subject service failed",
			log.Err(err),
			log.String("id", id),
		)
	}
	if _, err := subjectService.BatchGet(ctx, []string{id}); err != nil {
		switch err {
		case dbo.ErrRecordNotFound, constant.ErrRecordNotFound:
			log.Info(ctx, "check subject exists: not found subjects",
				log.Err(err),
				log.String("id", id),
			)
			return false, nil
		default:
			log.Error(ctx, "check subject exists: batch get subjects failed",
				log.Err(err),
				log.String("id", id),
			)
			return false, err
		}
	}
	return true, nil
}

func (a *assessmentModel) existsProgramByID(ctx context.Context, id string) (bool, error) {
	programService, err := external.GetProgramServiceProvider()
	if err != nil {
		log.Error(ctx, "check program exists: get program service failed",
			log.Err(err),
			log.String("id", id),
		)
	}
	if _, err := programService.BatchGet(ctx, []string{id}); err != nil {
		switch err {
		case dbo.ErrRecordNotFound, constant.ErrRecordNotFound:
			log.Info(ctx, "check program exists: not found programs",
				log.Err(err),
				log.String("id", id),
			)
			return false, nil
		default:
			log.Error(ctx, "check program exists: batch get programs failed",
				log.Err(err),
				log.String("id", id),
			)
			return false, err
		}
	}
	return true, nil
}
