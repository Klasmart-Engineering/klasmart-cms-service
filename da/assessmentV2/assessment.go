package assessmentV2

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
)

type IAssessmentDA interface {
	dbo.DataAccesser

	DeleteByScheduleIDsTx(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) error
}

var (
	assessmentDAInstance     IAssessmentDA
	assessmentDAInstanceOnce = sync.Once{}
)

func GetAssessmentDA() IAssessmentDA {
	assessmentDAInstanceOnce.Do(func() {
		assessmentDAInstance = &assessmentDA{}
	})
	return assessmentDAInstance
}

type assessmentDA struct {
	dbo.BaseDA
}

func (a *assessmentDA) DeleteByScheduleIDsTx(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) error {
	tx.ResetCondition()

	if err := tx.Unscoped().
		Where("schedule_id in (?)", scheduleIDs).
		Delete(&v2.Assessment{}).Error; err != nil {
		log.Error(ctx, "delete assessment failed",
			log.Strings("scheduleIDs", scheduleIDs),
		)
		return err
	}

	return nil
}

func (a *assessmentDA) First(ctx context.Context, condition *AssessmentCondition) (*v2.Assessment, error) {
	panic("implement me")
}

type AssessmentCondition struct {
	OrgID          sql.NullString
	ScheduleID     sql.NullString
	ScheduleIDs    entity.NullStrings
	Status         entity.NullStrings
	AssessmentType sql.NullString
	TeacherIDs     entity.NullStrings

	OrderBy AssessmentOrderBy
	Pager   dbo.Pager

	DeleteAt sql.NullInt64
}

func (c AssessmentCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.OrgID.Valid {
		wheres = append(wheres, "org_id = ?")
		params = append(params, c.OrgID.String)
	}

	if c.ScheduleID.Valid {
		wheres = append(wheres, "schedule_id = ?")
		params = append(params, c.ScheduleID.String)
	}
	if c.ScheduleIDs.Valid {
		wheres = append(wheres, "schedule_id in (?)")
		params = append(params, c.ScheduleIDs.Strings)
	}

	if c.Status.Valid {
		wheres = append(wheres, "status in (?)")
		params = append(params, c.Status.Strings)
	}

	if c.AssessmentType.Valid {
		wheres = append(wheres, "assessment_type = ?")
		params = append(params, c.AssessmentType.String)
	}

	if c.TeacherIDs.Valid {
		sql := fmt.Sprintf(`
exists(select 1 from %s where 
user_id in (?) and 
%s.assessment_id = %s.id and 
%s.user_type = ? and
%s.status_by_user = ?)`,
			constant.TableNameAssessmentsUsersV2,
			constant.TableNameAssessmentsUsersV2,
			constant.TableNameAssessmentV2,
			constant.TableNameAssessmentsUsersV2,
			constant.TableNameAssessmentsUsersV2,
		)
		wheres = append(wheres, sql)
		params = append(params, c.TeacherIDs.Strings, v2.AssessmentUserTypeTeacher, v2.AssessmentUserStatusParticipate)
	}

	if c.DeleteAt.Valid {
		wheres = append(wheres, "delete_at>0")
	} else {
		wheres = append(wheres, "(delete_at=0)")
	}

	return wheres, params
}

func (c AssessmentCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c AssessmentCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type AssessmentOrderBy int

const (
	AssessmentOrderByNameAsc AssessmentOrderBy = iota + 1
	AssessmentOrderByClassEndAtAsc
	AssessmentOrderByClassEndAtDesc
	AssessmentOrderByCompleteAtAsc
	AssessmentOrderByCompleteAtDesc
	AssessmentOrderByCreateAtAsc
	AssessmentOrderByCreateAtDesc
)

func NewAssessmentOrderBy(orderBy string) AssessmentOrderBy {
	switch orderBy {
	case "class_end_at":
		return AssessmentOrderByClassEndAtAsc
	case "-class_end_at":
		return AssessmentOrderByClassEndAtDesc
	case "complete_at":
		return AssessmentOrderByCompleteAtAsc
	case "-complete_at":
		return AssessmentOrderByCompleteAtDesc
	case "create_at":
		return AssessmentOrderByCreateAtAsc
	case "-create_at":
		return AssessmentOrderByCreateAtDesc
	}
	return AssessmentOrderByCreateAtDesc
}

func (c AssessmentOrderBy) ToSQL() string {
	switch c {
	case AssessmentOrderByClassEndAtAsc:
		return "class_end_at"
	case AssessmentOrderByClassEndAtDesc:
		return "class_end_at desc"
	case AssessmentOrderByCompleteAtAsc:
		return "complete_at"
	case AssessmentOrderByCompleteAtDesc:
		return "complete_at desc"
	case AssessmentOrderByCreateAtAsc:
		return "create_at"
	case AssessmentOrderByCreateAtDesc:
		return "create_at desc"
	default:
		return "create_at desc"
	}
}
