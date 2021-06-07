package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"sync"
)

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

type IAssessmentModel interface {
	Query(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, conditions *da.QueryAssessmentConditions) ([]*entity.Assessment, error)
	Summary(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.QueryAssessmentsSummaryArgs) (*entity.AssessmentsSummary, error)
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
		teachers    []*external.Teacher
		scheduleIDs []string
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
	if scheduleIDs, err = GetScheduleModel().GetScheduleIDsByOrgID(ctx, tx, operator, operator.OrgID); err != nil {
		log.Error(ctx, "summary: get schedule ids failed",
			log.Err(err),
			log.String("org_id", operator.OrgID),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}
	cond.ScheduleIDs = entity.NullStrings{
		Strings: scheduleIDs,
		Valid:   true,
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
