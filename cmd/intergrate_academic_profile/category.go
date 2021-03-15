package intergrate_academic_profile

import (
	"context"
	"log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

func (s *MapperImpl) initCategoryMapper(ctx context.Context) error {
	log.Println("start init category mapper")
	s.MapperCategory.categoryMapping = make(map[string]string)

	err := s.loadAmsCategorys(ctx)
	if err != nil {
		return err
	}

	err = s.loadOurCategorys(ctx)
	if err != nil {
		log.Println("init category mapper error")
		return err
	}
	log.Println("completed init category mapper")
	return nil
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
	var ourCategorys []*entity.Developmental
	err := da.GetDevelopmentalDA().Query(ctx, &da.DevelopmentalCondition{}, &ourCategorys)
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
	s.MapperCategory.amsCategoryMutex.Lock()
	defer s.MapperCategory.amsCategoryMutex.Unlock()

	id, found := s.MapperCategory.categoryMapping[categoryID]
	if found {
		return id, nil
	}

	// our
	ourCategorys, found := s.MapperCategory.ourCategorys[categoryID]
	if !found {
		return s.defaultCategoryID()
	}

	// ams
	amsProgramID, err := s.Program(ctx, organizationID, programID)
	if err != nil {
		return s.defaultCategoryID()
	}
	amsCategorys, found := s.MapperCategory.amsCategorys[amsProgramID]
	if !found {
		return s.defaultCategoryID()
	}
	amsCategory, found := amsCategorys[ourCategorys.Name]
	if !found {
		for _, item := range amsCategorys {
			return item.ID, nil
		}
		return s.defaultCategoryID()
	}

	// mapping
	s.MapperCategory.categoryMapping[categoryID] = amsCategory.ID

	return amsCategory.ID, nil
}

func (s *MapperImpl) defaultCategoryID() (string, error) {
	noneName := "None Specified"
	program, found := s.amsPrograms[noneName]
	if !found {
		return "", constant.ErrRecordNotFound
	}
	category, found := s.MapperCategory.amsCategorys[program.ID][noneName]
	if !found {
		return "", constant.ErrRecordNotFound
	}
	return category.ID, nil
}
