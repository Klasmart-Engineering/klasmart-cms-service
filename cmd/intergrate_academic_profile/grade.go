package intergrate_academic_profile

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

func (s *MapperImpl) Grade(ctx context.Context, organizationID, programID, gradeID string) (string, error) {
	s.gradeMutex.Lock()
	defer s.gradeMutex.Unlock()

	amsProgramID, err := s.Program(ctx, organizationID, programID)
	if err != nil {
		return "", err
	}

	ourGrade, found := s.ourGrades[gradeID]
	if !found {
		log.Error(ctx, "unknown grade id", log.String("gradeID", gradeID), log.String("organizationID", organizationID))
		return s.defaultGrade(amsProgramID)
	}

	amsGrade, found := s.amsGrades[amsProgramID+":"+ourGrade.Name]
	if found {
		return amsGrade.ID, nil
	}

	// new grade id should be mapping
	grades, err := external.GetGradeServiceProvider().GetByProgram(ctx, s.operator, amsProgramID)
	if err != nil {
		return "", err
	}

	// s.amsGrades = make(map[string]*external.Grade, len(grades))
	for _, grade := range grades {
		s.amsGrades[amsProgramID+":"+grade.Name] = grade
	}

	amsGrade, found = s.amsGrades[amsProgramID+":"+ourGrade.Name]
	if !found {
		log.Error(ctx, "unknown grade id",
			log.String("organizationID", organizationID),
			log.String("programID", programID),
			log.Any("grade", ourGrade))

		return s.defaultGrade(amsProgramID)
	}

	return amsGrade.ID, nil
}

func (s *MapperImpl) initGradeMapper(ctx context.Context) error {
	log.Info(ctx, "init prgram cache start")
	defer log.Info(ctx, "init program cache end")

	s.amsGrades = make(map[string]*external.Grade)

	return s.loadOurGrades(ctx)
}

func (s *MapperImpl) loadOurGrades(ctx context.Context) error {
	s.ourGrades = map[string]*entity.Grade{
		"grade0": {
			ID:   "grade0",
			Name: "None Specified",
			// Name: "Non specified",
		},
		"grade1": {
			ID:   "grade1",
			Name: "Not Specific",
		},
		"grade2": {
			ID:   "grade2",
			Name: "PreK-1",
		},
		"grade3": {
			ID:   "grade3",
			Name: "PreK-2",
		},
		"grade4": {
			ID:   "grade4",
			Name: "K",
		},
		"grade5": {
			ID:   "grade5",
			Name: "Grade 1",
		},
		"grade6": {
			ID:   "grade6",
			Name: "Grade 2",
		},
		"grade7": {
			ID:   "grade7",
			Name: "PreK-3",
		},
		"grade8": {
			ID:   "grade8",
			Name: "PreK-4",
		},
		"grade9": {
			ID:   "grade9",
			Name: "PreK-5",
		},
		"grade10": {
			ID:   "grade10",
			Name: "PreK-6",
		},
		"grade11": {
			ID:   "grade11",
			Name: "PreK-7",
		},
		"grade12": {
			ID:   "grade12",
			Name: "Kindergarten",
		},
	}

	return nil
}

func (s MapperImpl) defaultGrade(programID string) (string, error) {
	switch programID {
	case "7565ae11-8130-4b7d-ac24-1d9dd6f792f2":
		// None Specified
		return "98461ca1-06a1-432a-97d0-4e1dff33e1a5", nil
	case "75004121-0c0d-486c-ba65-4c57deacb44b":
		// ESL
		return "0ecb8fa9-d77e-4dd3-b220-7e79704f1b03", nil
	case "14d350f1-a7ba-4f46-bef9-dc847f0cbac5":
		// Math
		return "abc900b9-5b8c-4e54-a4a8-54f102b2c1c6", nil
	case "04c630cc-fabe-4176-80f2-30a029907a33":
		// Science
		return "abc900b9-5b8c-4e54-a4a8-54f102b2c1c6", nil
	case "f6617737-5022-478d-9672-0354667e0338":
		// Bada Talk
		return "98461ca1-06a1-432a-97d0-4e1dff33e1a5", nil
	case "4591423a-2619-4ef8-a900-f5d924939d02":
		// Bada Math
		return "d7e2e258-d4b3-4e95-b929-49ae702de4be", nil
	case "d1bbdcc5-0d80-46b0-b98e-162e7439058f":
		// Bada STEM
		return "d7e2e258-d4b3-4e95-b929-49ae702de4be", nil
	case "b39edb9a-ab91-4245-94a4-eb2b5007c033":
		// Bada Genius
		return "98461ca1-06a1-432a-97d0-4e1dff33e1a5", nil
	case "7a8c5021-142b-44b1-b60b-275c29d132fe":
		// Bada Read
		return "98461ca1-06a1-432a-97d0-4e1dff33e1a5", nil
	case "56e24fa0-e139-4c80-b365-61c9bc42cd3f":
		// Bada Sound
		return "98461ca1-06a1-432a-97d0-4e1dff33e1a5", nil
	case "93f293e8-2c6a-47ad-bc46-1554caac99e4":
		// Bada Rhyme
		return "98461ca1-06a1-432a-97d0-4e1dff33e1a5", nil
	default:
		// None Specified
		return "98461ca1-06a1-432a-97d0-4e1dff33e1a5", nil
	}
}

// func (s *MapperImpl) initGradeMapper(ctx context.Context) error {
// 	log.Println("start init grade mapper")
// 	s.MapperGrade.gradeMapping = make(map[string]string)
// 	err := s.loadAmsGrades(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	err = s.loadOurGrades(ctx)
// 	if err != nil {
// 		log.Println("init grade mapper error")
// 		return err
// 	}
// 	log.Println("completed init grade mapper")
// 	return nil
// }

// func (s *MapperImpl) loadAmsGrades(ctx context.Context) error {
// 	s.MapperGrade.amsGrades = make(map[string]map[string]*external.Grade, len(s.amsPrograms))
// 	for _, amsProgram := range s.amsPrograms {
// 		amsGrades, err := external.GetGradeServiceProvider().GetByProgram(ctx, s.operator, amsProgram.ID)
// 		if err != nil {
// 			return err
// 		}
// 		s.MapperGrade.amsGrades[amsProgram.ID] = make(map[string]*external.Grade, len(amsGrades))
// 		for _, amsGrade := range amsGrades {
// 			s.MapperGrade.amsGrades[amsProgram.ID][amsGrade.Name] = amsGrade
// 		}
// 	}
// 	return nil
// }

// func (s *MapperImpl) loadOurGrades(ctx context.Context) error {
// 	var ourGrades []*entity.Grade
// 	err := da.GetGradeDA().Query(ctx, &da.GradeCondition{}, &ourGrades)
// 	if err != nil {
// 		return err
// 	}

// 	s.MapperGrade.ourGrades = make(map[string]*entity.Grade, len(ourGrades))
// 	for _, grade := range ourGrades {
// 		s.MapperGrade.ourGrades[grade.ID] = grade
// 	}
// 	return nil
// }

// func (s *MapperImpl) Grade(ctx context.Context, organizationID, programID, gradeID string) (string, error) {
// 	s.MapperGrade.amsGradeMutex.Lock()
// 	defer s.MapperGrade.amsGradeMutex.Unlock()

// 	cacheID, found := s.MapperGrade.gradeMapping[gradeID]
// 	if found {
// 		return cacheID, nil
// 	}

// 	// our
// 	ourGrades, found := s.MapperGrade.ourGrades[gradeID]
// 	if !found {
// 		return s.defaultGradeID()
// 	}

// 	// ams
// 	amsProgramID, err := s.Program(ctx, organizationID, programID)
// 	if err != nil {
// 		return "", err
// 	}
// 	amsGrades, found := s.MapperGrade.amsGrades[amsProgramID]
// 	if !found {
// 		return s.defaultGradeID()
// 	}
// 	amsGrade, found := amsGrades[ourGrades.Name]
// 	if !found {
// 		for _, item := range amsGrades {
// 			return item.ID, nil
// 		}
// 		return s.defaultGradeID()
// 	}

// 	// mapping
// 	s.MapperGrade.gradeMapping[gradeID] = amsGrade.ID

// 	return amsGrade.ID, nil
// }

// func (s *MapperImpl) defaultGradeID() (string, error) {
// 	noneName := "None Specified"
// 	program, found := s.amsPrograms[noneName]
// 	if !found {
// 		return "", constant.ErrRecordNotFound
// 	}
// 	grade, found := s.MapperGrade.amsGrades[program.ID][noneName]
// 	if !found {
// 		return "", constant.ErrRecordNotFound
// 	}
// 	return grade.ID, nil
// }
