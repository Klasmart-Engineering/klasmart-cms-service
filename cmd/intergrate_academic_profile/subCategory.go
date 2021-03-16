package intergrate_academic_profile

import (
	"context"
	"log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

func (s *MapperImpl) initSubCategoryMapper(ctx context.Context) error {
	log.Println("start init subCategory mapper")
	err := s.loadAmsSubCategorys(ctx)
	if err != nil {
		return err
	}

	err = s.loadOurSubCategorys(ctx)
	if err != nil {
		log.Println("init subCategory mapper error")
		return err
	}
	log.Println("completed init subCategory mapper")
	return nil
}

func (s *MapperImpl) loadAmsSubCategorys(ctx context.Context) error {
	s.MapperSubCategory.amsSubCategorys = make(map[string]map[string]*external.SubCategory, len(s.amsPrograms))
	for _, amsProgram := range s.amsPrograms {
		categoryMap, found := s.MapperCategory.amsCategorys[amsProgram.ID]
		if !found {
			return constant.ErrRecordNotFound
		}
		for _, amsCategory := range categoryMap {
			subCatetorys, err := external.GetSubCategoryServiceProvider().GetByCategory(ctx, s.operator, amsCategory.ID)
			if err != nil {
				return err
			}
			s.MapperSubCategory.amsSubCategorys[amsProgram.ID+amsCategory.ID] = make(map[string]*external.SubCategory)
			for _, sub := range subCatetorys {
				s.MapperSubCategory.amsSubCategorys[amsProgram.ID+amsCategory.ID][sub.Name] = sub
			}
		}
	}
	return nil
}

func (s *MapperImpl) loadOurSubCategorys(ctx context.Context) error {
	var ourSubs []*entity.Skill
	err := da.GetSkillDA().Query(ctx, &da.SkillCondition{}, &ourSubs)
	if err != nil {
		return err
	}

	s.MapperSubCategory.ourSubCategorys = make(map[string]*entity.Skill, len(ourSubs))
	for _, sub := range ourSubs {
		s.MapperSubCategory.ourSubCategorys[sub.ID] = sub
	}
	return nil
}

func (s *MapperImpl) SubCategory(ctx context.Context, organizationID, programID, categoryID, subCategoryID string) (string, error) {
	s.MapperSubCategory.amsSubCategoryMutex.Lock()
	defer s.MapperSubCategory.amsSubCategoryMutex.Unlock()

	// ams
	amsProgramID, err := s.Program(ctx, organizationID, programID)
	if err != nil {
		return "", err
	}
	amsCategoryID, err := s.Category(ctx, organizationID, programID, categoryID)
	if err != nil {
		return "", err
	}

	// our
	ourSub, found := s.MapperSubCategory.ourSubCategorys[subCategoryID]
	if !found {
		return s.defaultSubCategoryID(amsProgramID, amsCategoryID)
	}

	amsSubCategorys, found := s.MapperSubCategory.amsSubCategorys[amsProgramID+amsCategoryID]
	if !found {
		return s.defaultSubCategoryID(amsProgramID, amsCategoryID)
	}

	amsSubCategory, found := amsSubCategorys[ourSub.Name]
	if !found {
		return s.defaultSubCategoryID(amsProgramID, amsCategoryID)
	}

	return amsSubCategory.ID, nil
}

func (s *MapperImpl) defaultSubCategoryID(amsProgramID string, amsCategoryID string) (string, error) {
	amsSubCategorys := s.MapperSubCategory.amsSubCategorys[amsProgramID+amsCategoryID]
	noneName := "None Specified"
	subCategory, found := amsSubCategorys[noneName]
	if !found {
		for _, item := range amsSubCategorys {
			return item.ID, nil
		}
		return "", nil
	}
	return subCategory.ID, nil
}
