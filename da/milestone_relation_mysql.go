package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type MilestoneRelationSQLDA struct {
	dbo.BaseDA
}

func (MilestoneRelationSQLDA) TableName() string {
	return entity.MilestoneRelationTable
}
func (MilestoneRelationSQLDA) MasterType() entity.RelationType {
	return entity.MilestoneType
}

func (mas MilestoneRelationSQLDA) DeleteTx(ctx context.Context, tx *dbo.DBContext, masterIDs []string) error {
	return GetRelationDA().DeleteTx(ctx, tx, mas.TableName(), mas.MasterType(), masterIDs)
}

func (mas MilestoneRelationSQLDA) InsertTx(ctx context.Context, tx *dbo.DBContext, relations []*entity.Relation) error {
	return GetRelationDA().InsertTx(ctx, tx, mas.TableName(), relations)
}

func (mas MilestoneRelationSQLDA) SearchTx(ctx context.Context, tx *dbo.DBContext, condition *RelationCondition) ([]*entity.Relation, error) {
	var relation []*entity.MilestoneRelation
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

var milestoneRelationDA *MilestoneRelationSQLDA
var _milestoneRelationDAOnce sync.Once

func GetMilestoneRelationDA() *MilestoneRelationSQLDA {
	_milestoneRelationDAOnce.Do(func() {
		milestoneRelationDA = new(MilestoneRelationSQLDA)
	})
	return milestoneRelationDA
}
