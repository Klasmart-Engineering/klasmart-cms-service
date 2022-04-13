package assessmentV2

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
)

type IAssessmentUserResultDA interface {
	dbo.DataAccesser
	PageByCondition(ctx context.Context, condition *UserResultPageCondition) (int64, []*v2.AssessmentUserResultDBView, error)
	GetAssessmentUserResultDBView(ctx context.Context, condition *AssessmentUserResultDBViewCondition) (int64, []*v2.AssessmentUserResultDBView, error)
}

type AssessmentUserResultDA struct {
	dbo.BaseDA
}

type AssessmentUserResultDBViewCondition struct {
	OrgID        sql.NullString
	ScheduleIDs  entity.NullStrings
	UserIDs      entity.NullStrings
	CompleteAtGe sql.NullInt64
	CompleteAtLe sql.NullInt64

	OrderBy AssessmentUserResultOrderBy
	Pager   dbo.Pager
}

func (a *AssessmentUserResultDA) GetAssessmentUserResultDBView(ctx context.Context, condition *AssessmentUserResultDBViewCondition) (int64, []*v2.AssessmentUserResultDBView, error) {
	tx := dbo.MustGetDB(ctx)
	tx.ResetCondition()

	commonSql := fmt.Sprintf(` 
from %s t1 
inner join %s t2 
on t1.assessment_user_id = t2.id 
inner join %s t3 on t2.assessment_id = t3.id
where 
`,
		constant.TableNameAssessmentReviewerFeedbackV2,
		constant.TableNameAssessmentsUsersV2,
		constant.TableNameAssessmentV2,
	)

	var wheres []string
	var params []interface{}
	wheres = append(wheres, "t3.assessment_type = ?")
	params = append(params, v2.AssessmentTypeOfflineStudy.String())

	if condition.OrgID.Valid {
		wheres = append(wheres, "t3.org_id = ?")
		params = append(params, condition.OrgID.String)
	}

	if condition.UserIDs.Valid {
		wheres = append(wheres, "t2.user_id in (?)")
		params = append(params, condition.UserIDs.Strings)
	}
	if condition.ScheduleIDs.Valid {
		wheres = append(wheres, "t3.schedule_id in (?)")
		params = append(params, condition.ScheduleIDs.Strings)
	}

	if condition.CompleteAtGe.Valid{
		wheres = append(wheres, "t1.complete_at >= ?")
		params = append(params, condition.CompleteAtGe.Int64)
	}
	if condition.CompleteAtLe.Valid{
		wheres = append(wheres, "t1.complete_at <= ?")
		params = append(params, condition.CompleteAtLe.Int64)
	}


	commonSql += strings.Join(wheres, " and ")

	countSql := fmt.Sprintf("%s %s", "select count(*)", commonSql)
	dataSql := fmt.Sprintf("%s %s", "select t1.*,t2.user_id,t3.id assessment_id,t3.schedule_id,t3.title", commonSql)

	var result []*v2.AssessmentUserResultDBView
	var err error
	var total int64
	if condition.Pager.Enable() {
		err := tx.Raw(countSql, params...).Scan(&total).Error
		if err != nil {
			return 0, nil, err
		}
		if total <= 0 {
			return 0, nil, nil
		}

		offset, limit := condition.Pager.Offset()
		dataSql += fmt.Sprintf(" order by %s LIMIT %d OFFSET %d ", condition.OrderBy.ToSQL(), limit, offset)

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

func (a *AssessmentUserResultDA) getPageByConditionSql(selectSql string, condition *UserResultPageCondition) string {
	sql := fmt.Sprintf(`
%s from %s as t4 inner join
(SELECT t1.id assessment_id,t1.schedule_id schedule_id, t2.user_id as user_id,t2.id as assessment_user_id
FROM %s t1 inner join %s t2 
on t1.id = t2.assessment_id 
where
`,
		selectSql,
		constant.TableNameAssessmentReviewerFeedbackV2,
		constant.TableNameAssessmentV2,
		constant.TableNameAssessmentsUsersV2,
	)
	return sql
}

func (a *AssessmentUserResultDA) PageByCondition(ctx context.Context, condition *UserResultPageCondition) (int64, []*v2.AssessmentUserResultDBView, error) {
	tx := dbo.MustGetDB(ctx)
	tx.ResetCondition()

	countSelect := "select count(*) "
	dataSelect := "select t4.*,t5.assessment_id,t5.schedule_id,t5.user_id "

	var wheres []string
	var params []interface{}
	wheres = append(wheres, "t1.org_id = ?")
	params = append(params, condition.OrgID)
	wheres = append(wheres, "t1.assessment_type = ?")
	params = append(params, v2.AssessmentTypeOfflineStudy)

	if condition.TeacherIDs.Valid {
		wheres = append(wheres, " t2.assessment_id in (select assessment_id from assessments_users_v2 t3 where t3.user_id in (?) and t3.user_type = ?)")
		params = append(params, condition.TeacherIDs.Strings, v2.AssessmentUserTypeTeacher)
	}

	joinSql := " ) as t5 on t5.assessment_user_id = t4.assessment_user_id"

	countSql := fmt.Sprintf("%s %s %s", a.getPageByConditionSql(countSelect, condition), strings.Join(wheres, " and "), joinSql)
	dataSql := fmt.Sprintf("%s %s %s", a.getPageByConditionSql(dataSelect, condition), strings.Join(wheres, " and "), joinSql)

	if condition.Status.Valid {
		countSql += " where t4.status in (?)"
		dataSql += " where t4.status in (?)"
		params = append(params, condition.Status.Strings)
	}

	var total int64
	err := tx.Raw(countSql, params...).Count(&total).Error
	if err != nil {
		log.Error(ctx, "count error", log.Err(err), log.String("countSql", countSql), log.Any("params", params), log.Any("condition", condition))
		return 0, nil, err
	}

	if total <= 0 {
		return 0, nil, nil
	}

	var result []*v2.AssessmentUserResultDBView
	if condition.Pager.Enable() {
		offset, limit := condition.Pager.Offset()
		dataSql += fmt.Sprintf(" order by %s LIMIT %d OFFSET %d ", condition.OrderBy.ToSQL(), limit, offset)

		err := tx.ResetCondition().Raw(dataSql, params...).Scan(&result).Error
		if err != nil {
			log.Error(ctx, "query from db error", log.Err(err), log.String("countSql", countSql), log.Any("params", params), log.Any("condition", condition))
			return 0, nil, err
		}
	}

	return total, result, nil
}

var (
	_assessmentUserResultOnce sync.Once
	_assessmentUserResultDA   IAssessmentUserResultDA
)

func GetAssessmentUserResultDA() IAssessmentUserResultDA {
	_assessmentUserResultOnce.Do(func() {
		_assessmentUserResultDA = &AssessmentUserResultDA{}
	})
	return _assessmentUserResultDA
}

type UserResultPageCondition struct {
	OrgID      string
	TeacherIDs entity.NullStrings
	Status     entity.NullStrings
	OrderBy    AssessmentUserResultOrderBy
	Pager      dbo.Pager
}

type AssessmentUserResultCondition struct {
	AssessmentUserID  sql.NullString
	AssessmentUserIDs entity.NullStrings
	Status            sql.NullString

	OrderBy  AssessmentUserResultOrderBy
	Pager    dbo.Pager
	DeleteAt sql.NullInt64
}

func (c AssessmentUserResultCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.AssessmentUserID.Valid {
		wheres = append(wheres, "assessment_user_id = (?)")
		params = append(params, c.AssessmentUserID.String)
	}

	if c.AssessmentUserIDs.Valid {
		wheres = append(wheres, "assessment_user_id in (?)")
		params = append(params, c.AssessmentUserIDs.Strings)
	}

	if c.Status.Valid {
		wheres = append(wheres, "status = (?)")
		params = append(params, c.Status.String)
	}

	if c.DeleteAt.Valid {
		wheres = append(wheres, "delete_at>0")
	} else {
		wheres = append(wheres, "(delete_at=0)")
	}

	return wheres, params
}

func (c AssessmentUserResultCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c AssessmentUserResultCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type AssessmentUserResultOrderBy int

const (
	AssessmentUserResultOrderByCreateAtAsc AssessmentUserResultOrderBy = iota + 1
	AssessmentUserResultOrderByCreateAtDesc
	AssessmentUserResultOrderByCompleteAtAsc
	AssessmentUserResultOrderByCompleteAtDesc
)

func NewAssessmentUserResultOrderBy(orderBy string) AssessmentUserResultOrderBy {
	switch orderBy {
	case "complete_at":
		return AssessmentUserResultOrderByCompleteAtAsc
	case "-complete_at":
		return AssessmentUserResultOrderByCompleteAtDesc
	case "submit_at":
		return AssessmentUserResultOrderByCreateAtAsc
	case "-submit_at":
		return AssessmentUserResultOrderByCreateAtDesc
	}
	return AssessmentUserResultOrderByCreateAtDesc
}

func (c AssessmentUserResultOrderBy) ToSQL() string {
	switch c {
	case AssessmentUserResultOrderByCompleteAtAsc:
		return "complete_at"
	case AssessmentUserResultOrderByCompleteAtDesc:
		return "complete_at desc"
	case AssessmentUserResultOrderByCreateAtAsc:
		return "create_at"
	case AssessmentUserResultOrderByCreateAtDesc:
		return "create_at desc"
	default:
		return "create_at desc"
	}
}
