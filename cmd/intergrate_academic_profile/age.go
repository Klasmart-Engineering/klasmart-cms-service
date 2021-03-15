package intergrate_academic_profile

import (
	"context"
	"log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

var AgeNameMapper = map[string]string{
	"3-4": "3 - 4 year(s)",
	"4-5": "4 - 5 year(s)",
	"5-6": "5 - 6 year(s)",
	"6-7": "6 - 7 year(s)",
	"7-8": "7 - 8 year(s)",
}

func (s *MapperImpl) initAgeMapper(ctx context.Context) error {
	log.Println("start init age mapper")
	s.MapperAge.ageMapping = make(map[string]string)

	err := s.loadAmsAges(ctx)
	if err != nil {
		return err
	}

	err = s.loadOurAges(ctx)
	if err != nil {
		log.Println("init age mapper error")
		return err
	}
	log.Println("completed init age mapper")
	return err
}

func (s *MapperImpl) loadAmsAges(ctx context.Context) error {
	s.MapperAge.amsAges = make(map[string]map[string]*external.Age, len(s.amsPrograms))
	for _, amsProgram := range s.amsPrograms {
		amsAges, err := external.GetAgeServiceProvider().GetByProgram(ctx, s.operator, amsProgram.ID)
		if err != nil {
			return err
		}
		s.MapperAge.amsAges[amsProgram.ID] = make(map[string]*external.Age, len(amsAges))
		for _, amsAge := range amsAges {
			s.MapperAge.amsAges[amsProgram.ID][amsAge.Name] = amsAge
		}
	}
	return nil
}

func (s *MapperImpl) loadOurAges(ctx context.Context) error {
	var ourAges []*entity.Age
	err := da.GetAgeDA().Query(ctx, &da.AgeCondition{}, &ourAges)
	if err != nil {
		return err
	}

	s.MapperAge.ourAges = make(map[string]*entity.Age, len(ourAges))
	for _, age := range ourAges {
		s.MapperAge.ourAges[age.ID] = age
	}
	return nil
}

func (s *MapperImpl) Age(ctx context.Context, organizationID, programID, ageID string) (string, error) {
	s.MapperAge.amsAgeMutex.Lock()
	defer s.MapperAge.amsAgeMutex.Unlock()

	cacheID, found := s.MapperAge.ageMapping[ageID]
	if found {
		return cacheID, nil
	}

	// our
	ourAges, found := s.MapperAge.ourAges[ageID]
	if !found {
		return s.defaultAgeID()
	}

	// ams
	amsProgramID, err := s.Program(ctx, organizationID, programID)
	if err != nil {
		return "", err
	}
	amsAges, found := s.MapperAge.amsAges[amsProgramID]
	if !found {
		return s.defaultAgeID()
	}
	amsAge, found := amsAges[AgeNameMapper[ourAges.Name]]
	if !found {
		for _, item := range amsAges {
			return item.ID, nil
		}
		return s.defaultAgeID()
	}

	// mapping
	s.MapperAge.ageMapping[ageID] = amsAge.ID

	return amsAge.ID, nil
}

func (s *MapperImpl) defaultAgeID() (string, error) {
	noneName := "None Specified"
	program, found := s.amsPrograms[noneName]
	if !found {
		return "", constant.ErrRecordNotFound
	}
	age, found := s.MapperAge.amsAges[program.ID][noneName]
	if !found {
		return "", constant.ErrRecordNotFound
	}
	return age.ID, nil
}
