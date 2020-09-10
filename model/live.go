package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type ILiveModel interface {
	MakeLiveToken(ctx context.Context, op *entity.Operator, data *entity.LiveToken) (string, error)
}

func (s *liveModel) MakeLiveToken(ctx context.Context, op *entity.Operator, data *entity.LiveToken) (string, error) {
	//payload := op.ToLivePayload()
	//
	//liveContent := entity.LiveContentInfo{
	//	UserID: op.UserID,
	//}
	//schedule, err := GetScheduleModel().GetPlainByID(ctx, data.ScheduleID)
	//if err != nil {
	//	return "", err
	//}
	//liveContent.RoomID = schedule.ID
	//var h5pList []*entity.H5pFileInfo

	return "", nil
}

func (s *liveModel) getLiveMaterials(ctx context.Context, data *entity.LiveToken) ([]*entity.LiveMaterial, error) {
	return nil, nil
}

type liveModel struct{}

var (
	_liveOnce  sync.Once
	_liveModel ILiveModel
)

func GetLiveModel() ILiveModel {
	_liveOnce.Do(func() {
		_liveModel = &liveModel{}
	})
	return _liveModel
}
