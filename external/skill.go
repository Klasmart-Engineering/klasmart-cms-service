package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/basicdata"
)

type SkillServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*entity.Skill, error)
}

func GetSkillServiceProvider() SkillServiceProvider {
	return &skillService{}
}

type skillService struct{}

func (s skillService) BatchGet(ctx context.Context, ids []string) ([]*entity.Skill, error) {
	result, err := basicdata.GetSkillModel().Query(ctx, &da.SkillCondition{
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

//type Skill struct {
//	ID   string `json:"id"`
//	Name string `json:"name"`
//}
//
//type mockSkillService struct{}
//
//func (s mockSkillService) BatchGet(ctx context.Context, ids []string) ([]*Skill, error) {
//	var skills []*Skill
//	for _, option := range GetMockData().Options {
//		for _, development := range option.Developmental {
//			skills = append(skills, development.Skill...)
//		}
//	}
//	return skills, nil
//}
