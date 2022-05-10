package da

import (
	"context"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
)

type ISkillCoverage interface {
	GetTeacherReportItems(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, teacherIDs ...string) (items []*entity.TeacherReportItem, err error)
}

func (r *ReportDA) GetTeacherReportItems(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, teacherIDs ...string) (items []*entity.TeacherReportItem, err error) {
	if !config.Get().RedisConfig.OpenCache {
		return r.getTeacherReportItemsByMySQL(ctx, tx, op, teacherIDs...)
	}

	request := &getSkillCoverageQueryCondition{
		Operator:   op,
		TeacherIDs: teacherIDs,
	}

	items = []*entity.TeacherReportItem{}
	err = r.skillCoverageCache.Get(ctx, request, &items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

type getSkillCoverageQueryCondition struct {
	Operator   *entity.Operator `json:"operator"`
	TeacherIDs []string         `json:"teacher_ids"`
}

func (r *ReportDA) getSkillCoverage(ctx context.Context, condition interface{}) (interface{}, error) {
	request, ok := condition.(*getSkillCoverageQueryCondition)
	if !ok {
		log.Error(ctx, "invalid request", log.Any("condition", condition))
		return nil, constant.ErrInvalidArgs
	}

	return r.getTeacherReportItemsByMySQL(ctx, dbo.MustGetDB(ctx), request.Operator, request.TeacherIDs...)
}

func (r *ReportDA) getTeacherReportItemsByMySQL(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, teacherIDs ...string) (items []*entity.TeacherReportItem, err error) {
	items = []*entity.TeacherReportItem{}
	sql := `
select 	 
	or2.relation_id as category_id,
	lo.id as outcome_id,
	lo.name as outcome_name
from assessments_users_outcomes_v2 auov 
inner join assessments_users_v2 auv1 on auov.assessment_user_id =auv1.id 
inner join learning_outcomes lo on lo.id = auov.outcome_id 
inner join outcomes_relations or2 on or2.master_id = auov.outcome_id and or2.relation_type = ?
where auv1.assessment_id in (
	select 
		av.id 
	from assessments_users_v2 auv
	inner join assessments_v2 av on auv.assessment_id =av.id 
	where auv.user_id in (?)
	and auv.user_type = ?
	and (
	    (auv.status_by_user  = ? and av.assessment_type  in (?) ) 
	    or (auv.status_by_system != ? and av.assessment_type  in (?))
	)
	and av.org_id = ?
	and av.delete_at =0 
) 
`
	args := []interface{}{
		entity.CategoryType,
		teacherIDs,
		v2.AssessmentUserTypeTeacher,
		v2.AssessmentUserStatusParticipate,
		v2.AssessmentTypeOnlineStudy,
		v2.AssessmentUserSystemStatusNotStarted,
		[]interface{}{
			v2.AssessmentTypeOnlineClass,
			v2.AssessmentTypeOfflineClass,
			v2.AssessmentTypeOfflineStudy,
		},
		op.OrgID,
	}
	err = r.QueryRawSQL(ctx, &items, sql, args...)
	if err != nil {
		return
	}
	return
}
