package intergrate_academic_profile

import (
	"context"

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
	s.MapperAge.ageMapping = make(map[string]string)

	err := s.loadAmsAges(ctx)
	if err != nil {
		return err
	}

	return s.loadOurAges(ctx)
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
	ageID, found := s.MapperAge.ageMapping[ageID]
	if found {
		return ageID, nil
	}

	s.MapperAge.amsAgeMutex.Lock()
	defer s.MapperAge.amsAgeMutex.Unlock()

	// our
	ourAges, found := s.MapperAge.ourAges[ageID]
	if !found {
		return "", constant.ErrRecordNotFound
	}

	// ams
	amsProgramID, err := s.Program(ctx, organizationID, programID)
	if err != nil {
		return "", err
	}
	amsAges, found := s.MapperAge.amsAges[amsProgramID]
	if !found {
		return "", constant.ErrRecordNotFound
	}
	amsAge, found := amsAges[AgeNameMapper[ourAges.Name]]
	if !found {
		return "", constant.ErrRecordNotFound
	}

	// mapping
	s.MapperAge.ageMapping[ageID] = amsAge.ID

	return amsAge.ID, nil
}
