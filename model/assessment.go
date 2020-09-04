package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"sync"
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

	item, err := da.GetAssessmentDA().Get(ctx, tx, id)
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
	panic("implement me")
}

func (a *assessmentModel) Add(ctx context.Context, tx *dbo.DBContext, cmd entity.AddAssessmentCommand) (string, error) {
	panic("implement me")
}

func (a *assessmentModel) Update(ctx context.Context, tx *dbo.DBContext, cmd entity.UpdateAssessmentCommand) error {
	panic("implement me")
}
