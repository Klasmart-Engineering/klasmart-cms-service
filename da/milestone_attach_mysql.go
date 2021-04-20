package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type MilestoneAttachSqlDA struct {
	dbo.BaseDA
}

func (MilestoneAttachSqlDA) TableName() string {
	return entity.AttachMilestoneTable
}
func (MilestoneAttachSqlDA) MasterType() string {
	return entity.MilestoneType
}

func (mas MilestoneAttachSqlDA) DeleteTx(ctx context.Context, tx *dbo.DBContext, masterIDs []string) error {
	return GetAttachDA().DeleteTx(ctx, tx, mas.TableName(), mas.MasterType(), masterIDs)
}

func (mas MilestoneAttachSqlDA) InsertTx(ctx context.Context, tx *dbo.DBContext, attaches []*entity.Attach) error {
	return GetAttachDA().InsertTx(ctx, tx, mas.TableName(), attaches)
}

func (mas MilestoneAttachSqlDA) SearchTx(ctx context.Context, tx *dbo.DBContext, condition *AttachCondition) ([]*entity.Attach, error) {
	var attach []*entity.MilestoneAttach
	_, err := GetAttachDA().BaseDA.PageTx(ctx, tx, condition, &attach)
	if err != nil {
		return nil, err
	}
	attaches := make([]*entity.Attach, len(attach))
	for i := range attach {
		attaches[i] = &attach[i].Attach
	}
	return attaches, nil
}

var milestoneAttachDA *MilestoneAttachSqlDA
var _milestoneAttachDAOnce sync.Once

func GetMilestoneAttachDA() *MilestoneAttachSqlDA {
	_milestoneAttachDAOnce.Do(func() {
		milestoneAttachDA = new(MilestoneAttachSqlDA)
	})
	return milestoneAttachDA
}
