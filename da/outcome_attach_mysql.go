package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type OutcomeAttachSqlDA struct {
	dbo.BaseDA
}

func (OutcomeAttachSqlDA) TableName() string {
	return entity.AttachOutcomeTable
}
func (OutcomeAttachSqlDA) MasterType() string {
	return entity.OutcomeType
}

func (mas OutcomeAttachSqlDA) DeleteTx(ctx context.Context, tx *dbo.DBContext, masterIDs []string) error {
	return GetAttachDA().DeleteTx(ctx, tx, mas.TableName(), mas.MasterType(), masterIDs)
}

func (mas OutcomeAttachSqlDA) InsertTx(ctx context.Context, tx *dbo.DBContext, attaches []*entity.Attach) error {
	return GetAttachDA().InsertTx(ctx, tx, mas.TableName(), attaches)
}

func (mas OutcomeAttachSqlDA) SearchTx(ctx context.Context, tx *dbo.DBContext, condition *AttachCondition) ([]*entity.Attach, error) {
	var attach []*entity.OutcomeAttach
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

var outcomeAttachDA *OutcomeAttachSqlDA
var _outcomeAttachDAOnce sync.Once

func GetOutcomeAttachDA() *OutcomeAttachSqlDA {
	_outcomeAttachDAOnce.Do(func() {
		outcomeAttachDA = new(OutcomeAttachSqlDA)
	})
	return outcomeAttachDA
}
