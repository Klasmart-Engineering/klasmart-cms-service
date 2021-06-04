package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
