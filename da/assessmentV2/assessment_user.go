package assessmentV2

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
)

type IAssessmentUserDA interface {
	dbo.DataAccesser
	//GetUserIDsByCondition(ctx context.Context, condition *AssessmentUserCondition) ([]string, error)
	UpdateStatusTx(ctx context.Context, tx *dbo.DBContext, condition *AssessmentUserCondition, status v2.AssessmentUserStatus) error
	UpdateSystemStatusTx(ctx context.Context, tx *dbo.DBContext, condition *AssessmentUserCondition, status v2.AssessmentUserStatus) error
	DeleteByAssessmentIDsTx(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string) error
}

type assessmentUserDA struct {
	dbo.BaseDA
}

func (a *assessmentUserDA) DeleteByAssessmentIDsTx(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string) error {
	tx.ResetCondition()

	if err := tx.Unscoped().
		Where("assessment_id in (?)", assessmentIDs).
		Delete(&v2.AssessmentUser{}).Error; err != nil {
		log.Error(ctx, "delete assessment user failed",
			log.Strings("assessmentIDs", assessmentIDs),
		)
		return err
	}

	return nil
}

//func (a *assessmentUserDA) GetUserIDsByCondition(ctx context.Context, condition *AssessmentUserCondition) ([]string, error) {
//	tx := dbo.MustGetDB(ctx)
//	tx.ResetCondition()
//
//	var userIDs []string
//	wheres, params := condition.GetConditions()
//
//	err := tx.Model(v2.AssessmentUser{}).Where(wheres, params...).Distinct().Pluck("user_id", userIDs).Error
//	if err != nil {
//		return nil, err
//	}
//
//	return userIDs, nil
//}

func (a *assessmentUserDA) UpdateStatusTx(ctx context.Context, tx *dbo.DBContext, condition *AssessmentUserCondition, status v2.AssessmentUserStatus) error {
	tx.ResetCondition()

	wheres, params := condition.GetConditions()
	if len(wheres) <= 0 {
		return errors.New("update required condition")
	}
	err := tx.Model(v2.AssessmentUser{}).
		Where(strings.Join(wheres, " and "), params...).
		Updates(map[string]interface{}{"status_by_system": status, "status_by_user": status, "update_at": time.Now().Unix()}).
		Error

	if err != nil {
		return err
	}

	return nil
}

func (a *assessmentUserDA) UpdateSystemStatusTx(ctx context.Context, tx *dbo.DBContext, condition *AssessmentUserCondition, status v2.AssessmentUserStatus) error {
	tx.ResetCondition()

	wheres, params := condition.GetConditions()
	if len(wheres) <= 0 {
		return errors.New("update required condition")
	}

	err := tx.Model(v2.AssessmentUser{}).
		Where(strings.Join(wheres, " and "), params...).
		UpdateColumn("status_by_system", status).
		UpdateColumn("update_at", time.Now().Unix()).
		Error

	if err != nil {
		log.Error(ctx, "update assessment user status error", log.Err(err))
		return err
	}

	return nil
}

var (
	_assessmentUserOnce sync.Once
	_assessmentUserDA   IAssessmentUserDA
)

func GetAssessmentUserDA() IAssessmentUserDA {
	_assessmentUserOnce.Do(func() {
		_assessmentUserDA = &assessmentUserDA{}
	})
	return _assessmentUserDA
}

type AssessmentUserCondition struct {
	AssessmentID  sql.NullString
	AssessmentIDs entity.NullStrings
	UserType      sql.NullString
	UserIDs       entity.NullStrings
	//StatusBySystem sql.NullString
	StatusByUser sql.NullString
}

func (c AssessmentUserCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.AssessmentID.Valid {
		wheres = append(wheres, "assessment_id = ?")
		params = append(params, c.AssessmentID.String)
	}

	if c.AssessmentIDs.Valid {
		wheres = append(wheres, "assessment_id in (?)")
		params = append(params, c.AssessmentIDs.Strings)
	}
	if c.UserType.Valid {
		wheres = append(wheres, "user_type = ?")
		params = append(params, c.UserType.String)
	}

	if c.UserIDs.Valid {
		wheres = append(wheres, "user_id in (?)")
		params = append(params, c.UserIDs.Strings)
	}
	//if c.StatusBySystem.Valid {
	//	wheres = append(wheres, "status_by_system = (?)")
	//	params = append(params, c.StatusBySystem.String)
	//}
	if c.StatusByUser.Valid {
		wheres = append(wheres, "status_by_user = (?)")
		params = append(params, c.StatusByUser.String)
	}

	return wheres, params
}

func (c AssessmentUserCondition) GetOrderBy() string {
	return ""
}

func (c AssessmentUserCondition) GetPager() *dbo.Pager {
	return nil
}

//type AssessmentUserOrderBy int
//
//const (
//	AssessmentUserOrderByNameAsc = iota + 1
//)
//
//func NewAssessmentUserOrderBy(orderBy string) AssessmentUserOrderBy {
//	return AssessmentUserOrderByNameAsc
//}
//
//func (c AssessmentUserOrderBy) ToSQL() string {
//	switch c {
//	default:
//		return "number desc, name asc"
//	}
//}
