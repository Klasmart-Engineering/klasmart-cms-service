package intergrate_academic_profile

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

// var AgeNameMapper = map[string]string{
// 	"3-4": "3 - 4 year(s)",
// 	"4-5": "4 - 5 year(s)",
// 	"5-6": "5 - 6 year(s)",
// 	"6-7": "6 - 7 year(s)",
// 	"7-8": "7 - 8 year(s)",
// }

func (s *MapperImpl) Age(ctx context.Context, organizationID, programID, ageID string) (string, error) {
	s.ageMutex.Lock()
	defer s.ageMutex.Unlock()

	amsProgramID, err := s.Program(ctx, organizationID, programID)
	if err != nil {
		return "", err
	}

	ourAge, found := s.ourAges[ageID]
	if !found {
		log.Error(ctx, "unknown age id", log.String("ageID", ageID), log.String("organizationID", organizationID))
		return s.defaultAge(amsProgramID)
	}

	amsAge, found := s.amsAges[amsProgramID+":"+ourAge.Name]
	if found {
		return amsAge.ID, nil
	}

	// new age id should be mapping
	ages, err := external.GetAgeServiceProvider().GetByProgram(ctx, s.operator, amsProgramID)
	if err != nil {
		return "", err
	}

	// s.amsAges = make(map[string]*external.Age, len(ages))
	for _, age := range ages {
		s.amsAges[amsProgramID+":"+age.Name] = age
	}

	amsAge, found = s.amsAges[amsProgramID+":"+ourAge.Name]
	if !found {
		log.Error(ctx, "unknown age id",
			log.String("organizationID", organizationID),
			log.String("programID", programID),
			log.Any("age", ourAge))

		return s.defaultAge(amsProgramID)
	}

	return amsAge.ID, nil
}

func (s *MapperImpl) initAgeMapper(ctx context.Context) error {
	log.Info(ctx, "init prgram cache start")
	defer log.Info(ctx, "init program cache end")

	s.amsAges = make(map[string]*external.Age)

	return s.loadOurAges(ctx)
}

// "3-4": "3 - 4 year(s)",
// "4-5": "4 - 5 year(s)",
// "5-6": "5 - 6 year(s)",
// "6-7": "6 - 7 year(s)",
// "7-8": "7 - 8 year(s)",
func (s *MapperImpl) loadOurAges(ctx context.Context) error {
	s.ourAges = map[string]*entity.Age{
		"age0": {
			ID:   "age0",
			Name: "None Specified",
			// Name: "Non specified",
		},
		"age1": {
			ID:   "age1",
			Name: "3 - 4 year(s)",
		},
		"age2": {
			ID:   "age2",
			Name: "4 - 5 year(s)",
		},
		"age3": {
			ID:   "age3",
			Name: "5 - 6 year(s)",
		},
		"age4": {
			ID:   "age4",
			Name: "6 - 7 year(s)",
		},
		"age5": {
			ID:   "age5",
			Name: "7 - 8 year(s)",
		},
	}
	// var ages []*entity.Age
	// err := da.GetAgeDA().Query(ctx, &da.AgeCondition{}, &ages)
	// if err != nil {
	// 	log.Error(ctx, "query error", log.Err(err))
	// 	return err
	// }

	// // key: age name
	// s.ourAges = make(map[string]*entity.Age, len(ages))
	// for _, age := range ages {
	// 	s.ourAges[age.Name] = age
	// }

	return nil
}

func (s *MapperImpl) defaultAge(programID string) (string, error) {
	switch programID {
	case "7565ae11-8130-4b7d-ac24-1d9dd6f792f2":
		// None Specified
		return "023eeeb1-5f72-4fa3-a2a7-63603607ac2b", nil
	case "75004121-0c0d-486c-ba65-4c57deacb44b":
		// ESL
		return "bb7982cd-020f-4e1a-93fc-4a6874917f07", nil
	case "14d350f1-a7ba-4f46-bef9-dc847f0cbac5":
		// Math
		return "bb7982cd-020f-4e1a-93fc-4a6874917f07", nil
	case "04c630cc-fabe-4176-80f2-30a029907a33":
		// Science
		return "bb7982cd-020f-4e1a-93fc-4a6874917f07", nil
	case "f6617737-5022-478d-9672-0354667e0338":
		// Bada Talk
		return "bb7982cd-020f-4e1a-93fc-4a6874917f07", nil
	case "4591423a-2619-4ef8-a900-f5d924939d02":
		// Bada Math
		return "bb7982cd-020f-4e1a-93fc-4a6874917f07", nil
	case "d1bbdcc5-0d80-46b0-b98e-162e7439058f":
		// Bada STEM
		return "bb7982cd-020f-4e1a-93fc-4a6874917f07", nil
	case "b39edb9a-ab91-4245-94a4-eb2b5007c033":
		// Bada Genius
		return "bb7982cd-020f-4e1a-93fc-4a6874917f07", nil
	case "7a8c5021-142b-44b1-b60b-275c29d132fe":
		// Bada Read
		return "bb7982cd-020f-4e1a-93fc-4a6874917f07", nil
	case "56e24fa0-e139-4c80-b365-61c9bc42cd3f":
		// Bada Sound
		return "bb7982cd-020f-4e1a-93fc-4a6874917f07", nil
	case "93f293e8-2c6a-47ad-bc46-1554caac99e4":
		// Bada Rhyme
		return "bb7982cd-020f-4e1a-93fc-4a6874917f07", nil
	default:
		// None Specified
		return "023eeeb1-5f72-4fa3-a2a7-63603607ac2b", nil
	}
}

// func (s *MapperImpl) initAgeMapper(ctx context.Context) error {
// 	log.Println("start init age mapper")
// 	s.MapperAge.ageMapping = make(map[string]string)

// 	err := s.loadAmsAges(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	err = s.loadOurAges(ctx)
// 	if err != nil {
// 		log.Println("init age mapper error")
// 		return err
// 	}
// 	log.Println("completed init age mapper")
// 	return err
// }

// func (s *MapperImpl) loadAmsAges(ctx context.Context) error {
// 	s.MapperAge.amsAges = make(map[string]map[string]*external.Age, len(s.amsPrograms))
// 	for _, amsProgram := range s.amsPrograms {
// 		amsAges, err := external.GetAgeServiceProvider().GetByProgram(ctx, s.operator, amsProgram.ID)
// 		if err != nil {
// 			return err
// 		}
// 		s.MapperAge.amsAges[amsProgram.ID] = make(map[string]*external.Age, len(amsAges))
// 		for _, amsAge := range amsAges {
// 			s.MapperAge.amsAges[amsProgram.ID][amsAge.Name] = amsAge
// 		}
// 	}
// 	return nil
// }

// func (s *MapperImpl) loadOurAges(ctx context.Context) error {
// 	var ourAges []*entity.Age
// 	err := da.GetAgeDA().Query(ctx, &da.AgeCondition{}, &ourAges)
// 	if err != nil {
// 		return err
// 	}

// 	s.MapperAge.ourAges = make(map[string]*entity.Age, len(ourAges))
// 	for _, age := range ourAges {
// 		s.MapperAge.ourAges[age.ID] = age
// 	}
// 	return nil
// }

// func (s *MapperImpl) Age(ctx context.Context, organizationID, programID, ageID string) (string, error) {
// 	s.MapperAge.amsAgeMutex.Lock()
// 	defer s.MapperAge.amsAgeMutex.Unlock()

// 	cacheID, found := s.MapperAge.ageMapping[ageID]
// 	if found {
// 		return cacheID, nil
// 	}

// 	// our
// 	ourAges, found := s.MapperAge.ourAges[ageID]
// 	if !found {
// 		return s.defaultAgeID()
// 	}

// 	// ams
// 	amsProgramID, err := s.Program(ctx, organizationID, programID)
// 	if err != nil {
// 		return "", err
// 	}
// 	amsAges, found := s.MapperAge.amsAges[amsProgramID]
// 	if !found {
// 		return s.defaultAgeID()
// 	}
// 	amsAge, found := amsAges[AgeNameMapper[ourAges.Name]]
// 	if !found {
// 		for _, item := range amsAges {
// 			return item.ID, nil
// 		}
// 		return s.defaultAgeID()
// 	}

// 	// mapping
// 	s.MapperAge.ageMapping[ageID] = amsAge.ID

// 	return amsAge.ID, nil
// }

// func (s *MapperImpl) defaultAgeID() (string, error) {
// 	noneName := "None Specified"
// 	program, found := s.amsPrograms[noneName]
// 	if !found {
// 		return "", constant.ErrRecordNotFound
// 	}
// 	age, found := s.MapperAge.amsAges[program.ID][noneName]
// 	if !found {
// 		return "", constant.ErrRecordNotFound
// 	}
// 	return age.ID, nil
// }
