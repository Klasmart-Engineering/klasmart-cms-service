package intergrate_academic_profile

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

func (s *MapperImpl) initCategoryMapper(ctx context.Context) error {
	s.MapperCategory.categoryMapping = make(map[string]string)

	err := s.loadAmsCategorys(ctx)
	if err != nil {
		return err
	}

	return s.loadOurCategorys(ctx)
}

func (s *MapperImpl) loadAmsCategorys(ctx context.Context) error {
	s.MapperCategory.amsCategoryIDMap = make(map[string]*external.Category)
	s.MapperCategory.amsCategorys = make(map[string]map[string]*external.Category, len(s.amsPrograms))
	for _, amsProgram := range s.amsPrograms {
		amsCategorys, err := external.GetCategoryServiceProvider().GetByProgram(ctx, s.operator, amsProgram.ID)
		if err != nil {
			return err
		}
		s.MapperCategory.amsCategorys[amsProgram.ID] = make(map[string]*external.Category, len(amsCategorys))
		for _, amsCategory := range amsCategorys {
			s.MapperCategory.amsCategoryIDMap[amsCategory.ID] = amsCategory
			s.MapperCategory.amsCategorys[amsProgram.ID][amsCategory.Name] = amsCategory
		}
	}

	return nil
}

func (s *MapperImpl) loadOurCategorys(ctx context.Context) error {
	ourCategorys, err := model.GetDevelopmentalModel().Query(ctx, &da.DevelopmentalCondition{})
	if err != nil {
		return err
	}

	s.MapperCategory.ourCategorys = make(map[string]*entity.Developmental, len(ourCategorys))
	for _, category := range ourCategorys {
		s.MapperCategory.ourCategorys[category.ID] = category
	}
	return nil
}

func (s *MapperImpl) Category(ctx context.Context, organizationID, programID, categoryID string) (string, error) {
	id, found := s.MapperCategory.categoryMapping[categoryID]
	if found {
		return id, nil
	}

	s.MapperCategory.amsCategoryMutex.Lock()
	defer s.MapperCategory.amsCategoryMutex.Unlock()

	// our
	ourCategorys, found := s.MapperCategory.ourCategorys[categoryID]
	if !found {
		return "", constant.ErrRecordNotFound
	}

	// ams
	amsProgramID, err := s.Program(ctx, organizationID, programID)
	if err != nil {
		return "", err
	}
	amsCategorys, found := s.MapperCategory.amsCategorys[amsProgramID]
	if !found {
		return "", constant.ErrRecordNotFound
	}
	amsCategory, found := amsCategorys[ourCategorys.Name]
	if !found {
		return "", constant.ErrRecordNotFound
	}

	// mapping
	s.MapperCategory.categoryMapping[categoryID] = amsCategory.ID

	return amsCategory.ID, nil
}
