package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/basicdata"
)

type LessonTypeProvider interface {
	BatchGet(ctx context.Context, ids []int) ([]*LessonType, error)
}
type LessonType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type mockLessonTypeService struct{}

func (m *mockLessonTypeService) BatchGet(ctx context.Context, ids []int) ([]*LessonType, error) {
	//TODO: to add publish scope
	return []*LessonType{
		{
			ID:   0,
			Name: "No",
		},
		{
			ID:   1,
			Name: "Test",
		},
		{
			ID:   2,
			Name: "NoTest",
		},
	}, nil
}

func GetLessonTypeProvider() LessonTypeProvider {
	return &mockLessonTypeService{}
}

type lessonTypeService struct{}

func (s lessonTypeService) BatchGet(ctx context.Context, ids []string) ([]*entity.LessonType, error) {
	result, err := basicdata.GetLessonTypeModel().Query(ctx, &da.LessonTypeCondition{
		IDs: entity.NullStrings{
			Strings: ids,
			Valid:   len(ids) != 0,
		},
	})
	if err != nil {
		log.Error(ctx, "BatchGet:error", log.Err(err), log.Strings("ids", ids))
		return nil, err
	}
	return result, nil
}
