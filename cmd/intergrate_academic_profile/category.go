package intergrate_academic_profile

import (
	"context"
	"log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

func (s MapperImpl) Category(ctx context.Context, organizationID, programID, categoryID string) (string, error) {
	oldCategory, err := model.GetDevelopmentalModel().GetByID(ctx, categoryID)
	if err != nil {
		return "", err
	}

	newProgramID, err := s.Program(ctx, organizationID, programID)
	newCategorys, err := external.GetCategoryServiceProvider().GetByProgram(ctx, &entity.Operator{
		OrgID: organizationID,
	}, newProgramID)
	if err != nil {
		return "", err
	}

	var result string
	var count int
	for _, item := range newCategorys {
		if oldCategory.Name == item.Name {
			count++
			if count > 1 {
				break
			}
			result = item.ID
		}
	}
	if count <= 0 {
		log.Printf("category not found;orgID:%s,newProgramID:%s,categoryName:%s", organizationID, newProgramID, oldCategory.Name)
		return "", constant.ErrRecordNotFound
	}
	if count > 1 {
		log.Printf("has many categorys;orgID:%s,newProgramID:%s,categoryName:%s", organizationID, newProgramID, oldCategory.Name)
		return "", constant.ErrConflict
	}

	return result, nil
}
