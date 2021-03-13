package intergrate_academic_profile

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

func (s *MapperImpl) initSubCategoryMapper(ctx context.Context) error {
	s.MapperSubCategory.subCategoryMapping = make(map[string]string)

	err := s.loadAmsSubCategorys(ctx)
	if err != nil {
		return err
	}

	return s.loadOurSubCategorys(ctx)
}

func (s *MapperImpl) loadAmsSubCategorys(ctx context.Context) error {
	s.MapperSubCategory.amsSubCategorys = make(map[string]map[string]*external.SubCategory, len(s.amsPrograms))
	for _, amsProgram := range s.amsPrograms {
		categoryMap, found := s.MapperCategory.amsCategorys[amsProgram.ID]
		if !found {
			return constant.ErrRecordNotFound
		}
		s.MapperSubCategory.amsSubCategorys[amsProgram.ID] = make(map[string]*external.SubCategory)
		for _, amsCategory := range categoryMap {
			subCatetorys, err := external.GetSubCategoryServiceProvider().GetByCategory(ctx, s.operator, amsCategory.ID)
			if err != nil {
				return err
			}

			for _, sub := range subCatetorys {
				s.MapperSubCategory.amsSubCategorys[amsProgram.ID][amsCategory.ID+":"+sub.Name] = sub
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
	subID, found := s.MapperSubCategory.subCategoryMapping[subCategoryID]
	if found {
		return subID, nil
	}

	s.MapperSubCategory.amsSubCategoryMutex.Lock()
	defer s.MapperSubCategory.amsSubCategoryMutex.Unlock()

	// our
	ourSub, found := s.MapperSubCategory.ourSubCategorys[subCategoryID]
	if !found {
		return s.defaultSubCategoryID()
	}

	// ams
	amsProgramID, err := s.Program(ctx, organizationID, programID)
	if err != nil {
		return "", err
	}
	amsCategory, err := s.Category(ctx, organizationID, programID, categoryID)
	if err != nil {
		return "", err
	}
	amsSubCategorys, found := s.MapperSubCategory.amsSubCategorys[amsProgramID]
	if !found {
		return s.defaultSubCategoryID()
	}

	amsSubCategory, found := amsSubCategorys[amsCategory+":"+ourSub.Name]
	if !found {
		for _, item := range amsSubCategorys {
			return item.ID, nil
		}
		return s.defaultSubCategoryID()
	}

	// mapping
	s.MapperCategory.categoryMapping[subCategoryID] = amsSubCategory.ID

	return amsSubCategory.ID, nil
}

func (s *MapperImpl) defaultSubCategoryID() (string, error) {
	noneName := "None Specified"
	program, found := s.amsPrograms[noneName]
	if !found {
		return "", constant.ErrRecordNotFound
	}
	categoryID, err := s.defaultCategoryID()
	if err != nil {
		return "", constant.ErrRecordNotFound
	}
	sub, found := s.MapperSubCategory.amsSubCategorys[program.ID][categoryID+":"+noneName]
	if !found {
		return "", constant.ErrRecordNotFound
	}
	return sub.ID, nil
}
