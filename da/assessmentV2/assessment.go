package assessmentV2

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"

	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
)

type IAssessmentDA interface {
	dbo.DataAccesser

	DeleteByScheduleIDsTx(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) error
	PageStudentAssessment(ctx context.Context, condition *StudentAssessmentCondition) (int, []*v2.StudentAssessmentDBView, error)
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

func (a *assessmentDA) PageStudentAssessment(ctx context.Context, condition *StudentAssessmentCondition) (int, []*v2.StudentAssessmentDBView, error) {
	tx := dbo.MustGetDB(ctx)
	tx.ResetCondition()

	commonSql := fmt.Sprintf(` 
from %s t1
inner join %s t2 on t1.assessment_id = t2.id
where 
`,
		constant.TableNameAssessmentsUsersV2,
		constant.TableNameAssessmentV2,
	)

	var wheres []string
	var params []interface{}
	wheres = append(wheres, "t2.assessment_type = ?")
	params = append(params, condition.AssessmentType.String)

	wheres = append(wheres, "t2.org_id = ?")
	params = append(params, condition.OrgID.String)

	wheres = append(wheres, "t1.user_id = ?")
	params = append(params, condition.StudentID.String)

	if condition.ScheduleIDs.Valid {
		wheres = append(wheres, "t2.schedule_id in (?)")
		params = append(params, condition.ScheduleIDs.Strings)
	}

	if condition.Status.Valid {
		wheres = append(wheres, "t2.status_by_system in (?)")
		params = append(params, condition.Status.Strings)
	}

	if condition.TeacherIDs.Valid {
		sql := fmt.Sprintf(`
exists(select 1 from %s where 
user_id in (?)
and 
%s.assessment_id = t1.assessment_id
%s.user_type = ?)`,
			constant.TableNameAssessmentsUsersV2,
			constant.TableNameAssessmentsUsersV2,
			constant.TableNameAssessmentsUsersV2,
		)
		wheres = append(wheres, sql)
		params = append(params, condition.TeacherIDs.Strings, v2.AssessmentUserTypeTeacher)
	}

	if condition.CreatedAtGe.Valid {
		wheres = append(wheres, "t1.create_at >= ?")
		params = append(params, condition.CompleteAtGe.Int64)
	}
	if condition.CreatedAtLe.Valid {
		wheres = append(wheres, "t1.create_at <= ?")
		params = append(params, condition.CompleteAtLe.Int64)
	}

	if condition.InProgressAtGe.Valid {
		wheres = append(wheres, "t1.in_progress_at >= ?")
		params = append(params, condition.InProgressAtGe.Int64)
	}
	if condition.InProgressAtLe.Valid {
		wheres = append(wheres, "t1.in_progress_at <= ?")
		params = append(params, condition.InProgressAtGe.Int64)
	}

	if condition.DoneAtGe.Valid {
		wheres = append(wheres, "t1.done_at >= ?")
		params = append(params, condition.DoneAtGe.Int64)
	}
	if condition.DoneAtLe.Valid {
		wheres = append(wheres, "t1.done_at <= ?")
		params = append(params, condition.DoneAtGe.Int64)
	}

	if condition.ResubmittedAtGe.Valid {
		wheres = append(wheres, "t1.resubmitted_at >= ?")
		params = append(params, condition.ResubmittedAtGe.Int64)
	}
	if condition.ResubmittedAtLe.Valid {
		wheres = append(wheres, "t1.resubmitted_at <= ?")
		params = append(params, condition.ResubmittedAtGe.Int64)
	}

	if condition.CompleteAtGe.Valid {
		wheres = append(wheres, "t1.completed_at >= ?")
		params = append(params, condition.CompleteAtGe.Int64)
	}
	if condition.CompleteAtLe.Valid {
		wheres = append(wheres, "t1.completed_at <= ?")
		params = append(params, condition.CompleteAtLe.Int64)
	}

	if condition.DeleteAt.Valid {
		wheres = append(wheres, "t1.delete_at>0")
	} else {
		wheres = append(wheres, "t1.delete_at=0")
	}

	commonSql += strings.Join(wheres, " and ")

	countSql := fmt.Sprintf("%s %s", "select count(*)", commonSql)
	dataSql := fmt.Sprintf("%s %s", "select t1.*,t2.id assessment_id,t2.title,t2.assessment_type,t2.schedule_id", commonSql)

	var result []*v2.StudentAssessmentDBView
	var err error
	var total int
	if condition.Pager.Enable() {
		err := tx.Raw(countSql, params...).Scan(&total).Error
		if err != nil {
			return 0, nil, err
		}
		if total <= 0 {
			return 0, nil, nil
		}

		offset, limit := condition.Pager.Offset()
		dataSql += fmt.Sprintf(" order by t1.%s LIMIT %d OFFSET %d ", condition.OrderBy.ToSQL(), limit, offset)

		err = tx.Raw(dataSql, params...).Scan(&result).Error
		if err != nil {
			return 0, nil, err
		}
	} else {
		err = tx.Raw(dataSql, params...).Scan(&result).Error
	}

	if err != nil {
		log.Error(ctx, "GetAssessmentUserResultDBView error", log.Err(err), log.String("sql", dataSql), log.Any("params", params))
		return 0, nil, err
	}

	return total, result, nil
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

type StudentAssessmentCondition struct {
	OrgID           sql.NullString
	ScheduleIDs     entity.NullStrings
	StudentID       sql.NullString
	Status          entity.NullStrings
	AssessmentType  sql.NullString
	TeacherIDs      entity.NullStrings
	CreatedAtGe     sql.NullInt64
	CreatedAtLe     sql.NullInt64
	InProgressAtGe  sql.NullInt64
	InProgressAtLe  sql.NullInt64
	DoneAtGe        sql.NullInt64
	DoneAtLe        sql.NullInt64
	ResubmittedAtGe sql.NullInt64
	ResubmittedAtLe sql.NullInt64
	CompleteAtGe    sql.NullInt64
	CompleteAtLe    sql.NullInt64

	OrderBy StudentAssessmentOrderBy
	Pager   dbo.Pager

	DeleteAt sql.NullInt64
}

type StudentAssessmentOrderBy string

func (o StudentAssessmentOrderBy) ToSQL() string {
	switch o {
	case "create_at":
		return "create_at"
	case "-create_at":
		return "create_at desc"
	case "in_progress_at":
		return "in_progress_at"
	case "-in_progress_at":
		return "in_progress_at desc"
	case "done_at":
		return "done_at"
	case "-done_at":
		return "done_at desc"
	case "resubmitted_at":
		return "resubmitted_at"
	case "-resubmitted_at":
		return "resubmitted_at desc"
	case "completed_at":
		return "completed_at"
	case "-completed_at":
		return "completed_at desc"
	default:
		return "create_at desc"
	}
}

type StudentIDsAndStatus struct {
	StudentID sql.NullString
	Status    sql.NullString
}
type AssessmentCondition struct {
	OrgID           sql.NullString
	ScheduleID      sql.NullString
	ScheduleIDs     entity.NullStrings
	Status          entity.NullStrings
	AssessmentType  sql.NullString
	AssessmentTypes entity.NullStrings
	TeacherIDs      entity.NullStrings

	UpdateAtGe   sql.NullInt64
	UpdateAtLe   sql.NullInt64
	CreatedAtGe  sql.NullInt64
	CreatedAtLe  sql.NullInt64
	CompleteAtGe sql.NullInt64
	CompleteAtLe sql.NullInt64

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

	if c.AssessmentTypes.Valid {
		wheres = append(wheres, "assessment_type in (?)")
		params = append(params, c.AssessmentTypes.Strings)
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

	if c.CreatedAtGe.Valid {
		wheres = append(wheres, "create_at >= ?")
		params = append(params, c.CreatedAtGe.Int64)
	}
	if c.CreatedAtLe.Valid {
		wheres = append(wheres, "create_at <= ?")
		params = append(params, c.CreatedAtLe.Int64)
	}

	if c.CompleteAtGe.Valid {
		wheres = append(wheres, "complete_at >= ?")
		params = append(params, c.CompleteAtGe.Int64)
	}
	if c.CompleteAtLe.Valid {
		wheres = append(wheres, "complete_at <= ?")
		params = append(params, c.CompleteAtLe.Int64)
	}

	if c.UpdateAtGe.Valid {
		wheres = append(wheres, "update_at >= ?")
		params = append(params, c.UpdateAtGe.Int64)
	}
	if c.UpdateAtLe.Valid {
		wheres = append(wheres, "update_at <= ?")
		params = append(params, c.UpdateAtLe.Int64)
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
