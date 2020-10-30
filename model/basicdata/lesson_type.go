package basicdata

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type ILessonTypeModel interface {
	Query(ctx context.Context, condition *da.LessonTypeCondition) ([]*entity.LessonType, error)
	GetByID(ctx context.Context, id string) (*entity.LessonType, error)
}

type lessonTypeModel struct {
}

func (m *lessonTypeModel) Query(ctx context.Context, condition *da.LessonTypeCondition) ([]*entity.LessonType, error) {
	var result []*entity.LessonType
	err := da.GetLessonTypeDA().Query(ctx, condition, &result)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	return result, nil
}

func (m *lessonTypeModel) GetByID(ctx context.Context, id string) (*entity.LessonType, error) {
	var result = new(entity.LessonType)
	err := da.GetLessonTypeDA().Get(ctx, id, result)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "GetByID:not found", log.Err(err), log.String("id", id))
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "GetByID error", log.Err(err), log.String("id", id))
		return nil, err
	}
	return result, nil
}

var (
	_lessonTypeOnce  sync.Once
	_lessonTypeModel ILessonTypeModel
)

func GetLessonTypeModel() ILessonTypeModel {
	_lessonTypeOnce.Do(func() {
		_lessonTypeModel = &lessonTypeModel{}
	})
	return _lessonTypeModel
}
