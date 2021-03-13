package intergrate_academic_profile

import (
	"context"
	"log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

func (s MapperImpl) SubCategory(ctx context.Context, organizationID, programID, categoryID, subCategoryID string) (string, error) {
	oldSub, err := model.GetSkillModel().GetByID(ctx, subCategoryID)
	if err != nil {
		return "", err
	}

	newCategoryID, err := s.Category(ctx, organizationID, programID, categoryID)
	subs, err := external.GetSubCategoryServiceProvider().GetByCategory(ctx, &entity.Operator{
		OrgID: organizationID,
	}, newCategoryID)
	if err != nil {
		return "", err
	}

	var subID string
	var count int
	for _, item := range subs {
		if oldSub.Name == item.Name {
			count++
			if count > 1 {
				break
			}
			subID = item.ID
		}
	}
	if count <= 0 {
		log.Printf("age not found;orgID:%s,newCategoryID:%s,gradeName:%s", organizationID, newCategoryID, oldSub.Name)
		return "", constant.ErrRecordNotFound
	}
	if count > 1 {
		log.Printf("age has many;orgID:%s,newCategoryID:%s,gradeName:%s", organizationID, newCategoryID, oldSub.Name)
		return "", constant.ErrConflict
	}
	return subID, nil
}
