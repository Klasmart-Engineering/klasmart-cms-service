package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type OutcomeRelationSqlDA struct {
	dbo.BaseDA
}

func (OutcomeRelationSqlDA) TableName() string {
	return entity.OutcomeRelationTable
}
func (OutcomeRelationSqlDA) MasterType() entity.RelationType {
	return entity.OutcomeType
}

func (mas OutcomeRelationSqlDA) DeleteTx(ctx context.Context, tx *dbo.DBContext, masterIDs []string) error {
	return GetRelationDA().DeleteTx(ctx, tx, mas.TableName(), mas.MasterType(), masterIDs)
}

func (mas OutcomeRelationSqlDA) InsertTx(ctx context.Context, tx *dbo.DBContext, relations []*entity.Relation) error {
	return GetRelationDA().InsertTx(ctx, tx, mas.TableName(), relations)
}

func (mas OutcomeRelationSqlDA) SearchTx(ctx context.Context, tx *dbo.DBContext, condition *RelationCondition) ([]*entity.Relation, error) {
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

var outcomeRelationDA *OutcomeRelationSqlDA
var _outcomeRelationDAOnce sync.Once

func GetOutcomeRelationDA() *OutcomeRelationSqlDA {
	_outcomeRelationDAOnce.Do(func() {
		outcomeRelationDA = new(OutcomeRelationSqlDA)
	})
	return outcomeRelationDA
}
