package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
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
	Add(ctx context.Context, tx *dbo.DBContext, cmd entity.AddAssessmentCommand) (string, error)
	Update(ctx context.Context, tx *dbo.DBContext, cmd entity.UpdateAssessmentCommand) error
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
		subjectService, err := external.GetSubjectServiceProvider()
		if err != nil {
			log.Error(ctx, "get assessment detail: get subject service failed",
				log.Err(err),
				log.String("id", id),
			)
			return nil, err
		}
		subjects, err := subjectService.BatchGet(ctx, []string{item.SubjectID})
		if err != nil {
			log.Error(ctx, "get assessment detail: batch get subject failed",
				log.Err(err),
				log.String("id", id),
				log.String("subject_id", item.SubjectID),
			)
			return nil, err
		}
		subject := subjects[0]
		result.Subject = entity.AssessmentSubject{
			ID:   subject.ID,
			Name: subject.Name,
		}
	}

	// fill teacher
	{
		teacherService, err := external.GetTeacherServiceProvider()
		if err != nil {
			log.Error(ctx, "get assessment detail: get student service failed",
				log.Err(err),
				log.String("id", id),
			)
			return nil, err
		}
		teachers, err := teacherService.BatchGet(ctx, []string{item.SubjectID})
		if err != nil {
			log.Error(ctx, "get assessment detail: batch get teacher failed",
				log.Err(err),
				log.String("id", id),
				log.String("teacher_id", item.TeacherID),
			)
			return nil, err
		}
		subject := teachers[0]
		result.Teacher = entity.AssessmentTeacher{
			ID:   subject.ID,
			Name: subject.Name,
		}
	}

	// TODO: fill number of activities (white)
	// TODO: fill number of outcomes (white)
	// TODO: fill outcome attendance maps (brilliant)

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
			for _, item := range items {
				*cond.TeacherIDs = append(*cond.TeacherIDs, item.ID)
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

	subjectNameMap := map[string]string{}
	{
		subjectService, err := external.GetSubjectServiceProvider()
		if err != nil {
			log.Error(ctx, "list assessments: get subject service failed",
				log.Err(err),
				log.Any("cmd", cmd),
			)
			return nil, err
		}
		items, err := subjectService.BatchGet(ctx, subjectIDs)
		if err != nil {
			log.Error(ctx, "list assessments: batch get subject failed",
				log.Err(err),
				log.Any("cmd", cmd),
			)
			return nil, err
		}
		for _, item := range items {
			subjectNameMap[item.ID] = item.Name
		}
	}

	programNameMap := map[string]string{}
	{
		programService, err := external.GetProgramServiceProvider()
		if err != nil {
			log.Error(ctx, "list assessments: get program service failed",
				log.Err(err),
				log.Any("cmd", cmd),
			)
			return nil, err
		}
		items, err := programService.BatchGet(ctx, subjectIDs)
		if err != nil {
			log.Error(ctx, "list assessments: batch get program failed",
				log.Err(err),
				log.Any("cmd", cmd),
			)
			return nil, err
		}
		for _, item := range items {
			programNameMap[item.ID] = item.Name
		}
	}

	teacherNameMap := map[string]string{}
	{
		teacherNameService, err := external.GetTeacherServiceProvider()
		if err != nil {
			log.Error(ctx, "list assessments: get teacher service failed",
				log.Err(err),
				log.Any("cmd", cmd),
			)
			return nil, err
		}
		items, err := teacherNameService.BatchGet(ctx, subjectIDs)
		if err != nil {
			log.Error(ctx, "list assessments: batch get teacher failed",
				log.Err(err),
				log.Any("cmd", cmd),
			)
			return nil, err
		}
		for _, item := range items {
			teacherNameMap[item.ID] = item.Name
		}
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

func (a *assessmentModel) Add(ctx context.Context, tx *dbo.DBContext, cmd entity.AddAssessmentCommand) (string, error) {
	var (
		nowUnix = time.Now().Unix()
		newID   = utils.NewID()
	)
	newItem := entity.Assessment{
		ID:           newID,
		ScheduleID:   cmd.ScheduleID,
		Title:        cmd.Title(),
		ProgramID:    cmd.ProgramID,
		SubjectID:    cmd.SubjectID,
		TeacherID:    cmd.TeacherID,
		ClassLength:  cmd.ClassLength,
		ClassEndTime: cmd.ClassEndTime,
		CompleteTime: cmd.CompleteTime,
		Status:       cmd.Status,
		CreateTime:   nowUnix,
		UpdateTime:   nowUnix,
	}
	if _, err := da.GetAssessmentDA().InsertTx(ctx, tx, &newItem); err != nil {
		log.Error(ctx, "add assessment: add failed",
			log.Err(err),
			log.Any("cmd", cmd),
			log.Any("new_item", newItem),
		)
		return "", err
	}
	return newID, nil
}

func (a *assessmentModel) Update(ctx context.Context, tx *dbo.DBContext, cmd entity.UpdateAssessmentCommand) error {
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
			if err := da.GetOutcomeAttendanceDA().BatchDeleteByOutcomeIDs(ctx, tx, outcomeIDs); err != nil {
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
						OutcomeID:    item.OutcomeID,
						AttendanceID: attendanceID,
					})
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
