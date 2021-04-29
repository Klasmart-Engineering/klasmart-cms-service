package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type OutcomeRelationSQLDA struct {
	dbo.BaseDA
}

func (OutcomeRelationSQLDA) TableName() string {
	return entity.OutcomeRelationTable
}
func (OutcomeRelationSQLDA) MasterType() entity.RelationType {
	return entity.OutcomeType
}

func (mas OutcomeRelationSQLDA) DeleteTx(ctx context.Context, tx *dbo.DBContext, masterIDs []string) error {
	return GetRelationDA().DeleteTx(ctx, tx, mas.TableName(), mas.MasterType(), masterIDs)
}

func (mas OutcomeRelationSQLDA) InsertTx(ctx context.Context, tx *dbo.DBContext, relations []*entity.Relation) error {
	return GetRelationDA().InsertTx(ctx, tx, mas.TableName(), relations)
}

func (mas OutcomeRelationSQLDA) SearchTx(ctx context.Context, tx *dbo.DBContext, condition *RelationCondition) ([]*entity.Relation, error) {
	var relation []*entity.OutcomeRelation
	_, err := GetRelationDA().BaseDA.PageTx(ctx, tx, condition, &relation)
	if err != nil {
		return nil, err
	}
	relations := make([]*entity.Relation, len(relation))
	for i := range relation {
		relations[i] = &relation[i].Relation
	}
	return relations, nil
}

var outcomeRelationDA *OutcomeRelationSQLDA
var _outcomeRelationDAOnce sync.Once

func GetOutcomeRelationDA() *OutcomeRelationSQLDA {
	_outcomeRelationDAOnce.Do(func() {
		outcomeRelationDA = new(OutcomeRelationSQLDA)
	})
	return outcomeRelationDA
}
