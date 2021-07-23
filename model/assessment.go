package model

import (
	"context"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

var (
	assessmentModelInstance     IAssessmentModel
	assessmentModelInstanceOnce = sync.Once{}
)

var (
	ErrInvalidOrderByValue = errors.New("invalid order by value")
)

func GetAssessmentModel() IAssessmentModel {
	assessmentModelInstanceOnce.Do(func() {
		assessmentModelInstance = &assessmentModel{}
	})
	return assessmentModelInstance
}

type IAssessmentModel interface {
	Query(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, conditions *da.QueryAssessmentConditions) ([]*entity.Assessment, error)
	Summary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.QueryAssessmentsSummaryArgs) (*entity.AssessmentsSummary, error)

	StudentQuery(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, conditions *da.StudentQueryAssessmentConditions) (int, []*entity.StudentAssessment, error)
}

type assessmentModel struct {
	assessmentBase
}

func (m *assessmentModel) Query(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, conditions *da.QueryAssessmentConditions) ([]*entity.Assessment, error) {
	var r []*entity.Assessment
	if err := da.GetAssessmentDA().QueryTx(ctx, tx, conditions, &r); err != nil {
		log.Error(ctx, "query assessments failed",
			log.Err(err),
			log.Any("conditions", conditions),
			log.Any("operator", operator),
		)
		return nil, err
	}
	return r, nil
}

func (m *assessmentModel) Summary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.QueryAssessmentsSummaryArgs) (*entity.AssessmentsSummary, error) {
	// check permission
	var (
		checker = NewAssessmentPermissionChecker(operator)
		err     error
	)
	if err = checker.SearchAllPermissions(ctx); err != nil {
		return nil, err
	}
	if args.Status.Valid && !checker.CheckStatus(args.Status.Value) {
		log.Error(ctx, "summary: check status failed",
			log.Any("args", args),
			log.Any("checker", checker),
		)
		return nil, constant.ErrForbidden
	}

	// get assessment list
	var (
		assessments []*entity.Assessment
		cond        = da.QueryAssessmentConditions{
			OrgID: entity.NullString{
				String: operator.OrgID,
				Valid:  true,
			},
			Status: args.Status,
			AllowTeacherIDs: entity.NullStrings{
				Strings: checker.AllowTeacherIDs(),
				Valid:   true,
			},
			AllowTeacherIDAndStatusPairs: entity.NullAssessmentAllowTeacherIDAndStatusPairs{
				Values: checker.allowPairs,
				Valid:  len(checker.allowPairs) > 0,
			},
		}
		teachers []*external.Teacher
	)
	if args.TeacherName.Valid {
		if teachers, err = external.GetTeacherServiceProvider().Query(ctx, operator, operator.OrgID, args.TeacherName.String); err != nil {
			log.Error(ctx, "List: external.GetTeacherServiceProvider().Query: query failed",
				log.Err(err),
				log.String("org_id", operator.OrgID),
				log.String("teacher_name", args.TeacherName.String),
				log.Any("args", args),
				log.Any("operator", operator),
			)
			return nil, err
		}
		log.Debug(ctx, "summary: query teachers success",
			log.String("org_id", operator.OrgID),
			log.String("teacher_name", args.TeacherName.String),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		if len(teachers) > 0 {
			cond.TeacherIDs.Valid = true
			for _, item := range teachers {
				cond.TeacherIDs.Strings = append(cond.TeacherIDs.Strings, item.ID)
			}
		} else {
			cond.TeacherIDs.Valid = false
		}
	}

	if err := da.GetAssessmentDA().QueryTx(ctx, tx, &cond, &assessments); err != nil {
		log.Error(ctx, "summary: query assessments failed",
			log.Err(err),
			log.Any("cond", cond),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}
	if len(assessments) == 0 {
		return nil, nil
	}

	r := entity.AssessmentsSummary{}
	for _, a := range assessments {
		switch a.Status {
		case entity.AssessmentStatusComplete:
			r.Complete++
		case entity.AssessmentStatusInProgress:
			r.InProgress++
		}
	}

	// merge home fun study summary
	r2, err := GetHomeFunStudyModel().Summary(ctx, tx, operator, args)
	if err != nil {
		log.Error(ctx, "summary: get home fun study summary",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return nil, err
	}
	r.InProgress += r2.InProgress
	r.Complete += r2.Complete

	return &r, nil
}

func (m *assessmentModel) StudentQuery(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, condition *da.StudentQueryAssessmentConditions) (int, []*entity.StudentAssessment, error) {
	scheduleClassType := condition.ClassType.ToScheduleClassType()
	if scheduleClassType.IsHomeFun {
		//Query assessments
		return m.studentsAssessmentQuery(ctx, operator, tx, scheduleClassType, condition)
	}
	//Query home fun study
	return m.studentsHomeFunStudyQuery(ctx, operator, tx, scheduleClassType, condition)
}

func (m *assessmentModel) studentsAssessmentQuery(ctx context.Context,
	operator *entity.Operator, tx *dbo.DBContext, scheduleClassType entity.ScheduleClassTypeDesc,
	condition *da.StudentQueryAssessmentConditions) (int, []*entity.StudentAssessment, error) {
	studentIDs := entity.NullStrings{
		Strings: []string{condition.StudentID},
		Valid:   true,
	}
	classType := entity.NullString{
		String: scheduleClassType.ClassType.String(),
		Valid:  scheduleClassType.ClassType != "",
	}
	orderBy := entity.NullAssessmentsOrderBy{
		Value: entity.AssessmentOrderBy(condition.OrderBy),
		Valid: condition.OrderBy != "",
	}
	if orderBy.Valid && !orderBy.Value.Valid() {
		log.Error(ctx, "orderBy invalid",
			log.Any("orderBy", orderBy),
			log.Any("condition", condition),
		)
		return 0, nil, ErrInvalidOrderByValue
	}

	var r []*entity.Assessment
	conditions := &da.QueryAssessmentConditions{
		IDs:                          condition.IDs,
		Status:                       condition.Status,
		StudentIDs:                   studentIDs,
		TeacherIDs:                   condition.TeacherIDs,
		AllowTeacherIDAndStatusPairs: entity.NullAssessmentAllowTeacherIDAndStatusPairs{},
		CreatedBetween:               condition.CreatedBetween,
		UpdateBetween:                condition.UpdateBetween,
		CompleteBetween:              condition.CompleteBetween,
		ClassType:                    classType,
		OrderBy:                      orderBy,
		Pager:                        condition.Pager,
	}
	total, err := da.GetAssessmentDA().PageTx(ctx, tx, conditions, &r)
	if err != nil {
		log.Error(ctx, "StudentQuery:GetAssessmentDA.QueryTx failed",
			log.Err(err),
			log.Any("condition", condition),
		)
		return 0, nil, err
	}
	res, err := m.assessmentsToStudentAssessments(ctx, operator, tx, r)
	if err != nil {
		log.Error(ctx, "StudentQuery:assessmentsToStudentAssessments failed",
			log.Err(err),
			log.Any("assessment", r),
		)
		return 0, nil, err
	}
	return total, res, nil
}

func (m *assessmentModel) assessmentsToStudentAssessments(ctx context.Context,
	operator *entity.Operator, tx *dbo.DBContext, r []*entity.Assessment) ([]*entity.StudentAssessment, error) {
	res := make([]*entity.StudentAssessment, len(r))
	ids := make([]string, len(r))
	scheduleIDs := make([]string, len(r))
	for i := range r {
		res[i] = &entity.StudentAssessment{
			ID:         r[i].ID,
			Title:      r[i].Title,
			Status:     string(r[i].Status),
			CreateAt:   r[i].CreateAt,
			CompleteAt: r[i].CompleteTime,
			UpdateAt:   r[i].UpdateAt,
			ScheduleID: r[i].ScheduleID,
			//Comment:             r[i].CompleteTime,
			//Score:               r[i].Sco,
			//Teacher:             nil,
			//Schedule:            nil,
			//FeedbackAttachments: nil,
		}
		ids[i] = r[i].ID
		scheduleIDs[i] = r[i].ScheduleID
	}
	err := m.fillStudentAssessments(ctx, operator, tx, res)
	if err != nil {
		log.Error(ctx, "fillStudentAssessments failed",
			log.Err(err),
			log.Any("res", res),
		)
		return nil, err
	}

	return res, nil
}

func (m *assessmentModel) studentsHomeFunStudyQuery(ctx context.Context,
	operator *entity.Operator, tx *dbo.DBContext, scheduleClassType entity.ScheduleClassTypeDesc,
	condition *da.StudentQueryAssessmentConditions) (int, []*entity.StudentAssessment, error) {
	studentIDs := entity.NullStrings{
		Strings: []string{condition.StudentID},
		Valid:   true,
	}
	classType := entity.NullString{
		String: scheduleClassType.ClassType.String(),
		Valid:  scheduleClassType.ClassType != "",
	}
	teacherIDs := utils.NullSQLJSONStringArray{
		Values: utils.SQLJSONStringArray(condition.TeacherIDs.Strings),
		Valid:  condition.TeacherIDs.Valid,
	}
	orderBy := entity.NullListHomeFunStudiesOrderBy{
		Value: entity.ListHomeFunStudiesOrderBy(condition.OrderBy),
		Valid: condition.OrderBy != "",
	}
	if orderBy.Valid && !orderBy.Value.Valid() {
		log.Error(ctx, "orderBy invalid",
			log.Any("orderBy", orderBy),
			log.Any("condition", condition),
		)
		return 0, nil, ErrInvalidOrderByValue
	}
	//Query home fun study
	var r []*entity.HomeFunStudy

	conditions := &da.QueryHomeFunStudyCondition{
		IDs:             condition.IDs,
		Status:          condition.Status,
		StudentIDs:      studentIDs,
		TeacherIDs:      teacherIDs,
		CreatedBetween:  condition.CreatedBetween,
		UpdateBetween:   condition.UpdateBetween,
		CompleteBetween: condition.CompleteBetween,
		ClassType:       classType,
		OrderBy:         orderBy,
		Pager:           condition.Pager,
	}
	total, err := da.GetHomeFunStudyDA().PageTx(ctx, tx, conditions, &r)
	if err != nil {
		log.Error(ctx, "StudentQuery:GetHomeFunStudyDA.QueryTx failed",
			log.Err(err),
			log.Any("condition", condition),
		)
		return 0, nil, err
	}
	res, err := m.homeFunStudyToStudentAssessments(ctx, operator, tx, r)
	if err != nil {
		log.Error(ctx, "StudentQuery:assessmentsToStudentAssessments failed",
			log.Err(err),
			log.Any("assessment", r),
		)
		return 0, nil, err
	}
	return total, res, nil
}

func (m *assessmentModel) homeFunStudyToStudentAssessments(ctx context.Context,
	operator *entity.Operator, tx *dbo.DBContext, r []*entity.HomeFunStudy) ([]*entity.StudentAssessment, error) {
	res := make([]*entity.StudentAssessment, len(r))
	for i := range r {
		res[i] = &entity.StudentAssessment{
			ID:         r[i].ID,
			Title:      r[i].Title,
			Status:     string(r[i].Status),
			CreateAt:   r[i].CreateAt,
			UpdateAt:   r[i].UpdateAt,
			CompleteAt: r[i].CompleteAt,
			ScheduleID: r[i].ScheduleID,
			Comment:    r[i].AssessComment,
			Score:      int(r[i].AssessScore),
			//Teacher:             nil,
			//Schedule:            nil,
			//FeedbackAttachments: nil,
		}
	}
	err := m.fillStudentAssessments(ctx, operator, tx, res)
	if err != nil {
		log.Error(ctx, "fillStudentAssessments failed",
			log.Err(err),
			log.Any("res", res),
		)
		return nil, err
	}
	return res, nil
}

func (m *assessmentModel) fillStudentAssessments(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, assesments []*entity.StudentAssessment) error {
	//GetScheduleModel().QueryByCondition(ctx, operator, )
	//external.GetTeacherServiceProvider().BatchGetMap(ctx, )
	return nil
}
