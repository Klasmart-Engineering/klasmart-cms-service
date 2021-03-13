package intergrate_academic_profile

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

func (s *MapperImpl) Program(ctx context.Context, organizationID, programID string) (string, error) {
	amsProgramID, found := s.programMapping[programID]
	if found {
		return amsProgramID, nil
	}

	s.programMutex.Lock()
	defer s.programMutex.Unlock()

	ourProgram, found := s.ourPrograms[programID]
	if !found {
		log.Error(ctx, "unknown program id", log.String("programID", programID), log.String("organizationID", organizationID))
		return "", constant.ErrRecordNotFound
	}

	amsProgram, found := s.amsPrograms[ourProgram.Name]
	if !found {
		log.Error(ctx, "unknown program id", log.Any("program", ourProgram), log.String("organizationID", organizationID))
		return "", constant.ErrRecordNotFound
	}

	s.programMapping[ourProgram.ID] = amsProgram.ID

	return amsProgram.ID, nil
}

func (s *MapperImpl) initProgramMapper(ctx context.Context) error {
	s.programMapping = make(map[string]string)

	err := s.loadAmsPrograms(ctx)
	if err != nil {
		return err
	}

	return s.loadOurPrograms(ctx)
}

func (s *MapperImpl) loadAmsPrograms(ctx context.Context) error {
	programs, err := external.GetProgramServiceProvider().GetByOrganization(ctx, s.operator)
	if err != nil {
		return err
	}

	if len(programs) == 0 {
		log.Panic(ctx, "get program by organization failed", log.Any("operator", s.operator))
	}

	s.amsPrograms = make(map[string]*external.Program, len(programs))
	for _, program := range programs {
		s.amsPrograms[program.Name] = program
	}

	return nil
}

func (s *MapperImpl) loadOurPrograms(ctx context.Context) error {
	s.ourPrograms = map[string]*entity.Program{
		"5fd9ddface9660cbc5f667d8": {
			ID:   "5fd9ddface9660cbc5f667d8",
			Name: "None Specified",
			// Name: "Non specified",
		},
		"5fdac06ea878718a554ff00d": {
			ID:   "5fdac06ea878718a554ff00d",
			Name: "ESL",
		},
		"5fdac0f61f066722a1351adb": {
			ID:   "5fdac0f61f066722a1351adb",
			Name: "Math",
		},
		"5fdac0fe1f066722a1351ade": {
			ID:   "5fdac0fe1f066722a1351ade",
			Name: "Science",
		},
		"program0": {
			ID:   "program0",
			Name: "None Specified",
			// Name: "Non specified",
		},
		"program1": {
			ID:   "program1",
			Name: "Bada Talk",
		},
		"program2": {
			ID:   "program2",
			Name: "Bada Math",
		},
		"program3": {
			ID:   "program3",
			Name: "Bada STEM",
		},
		"program4": {
			ID:   "program4",
			Name: "Bada Genius",
		},
		"program5": {
			ID:   "program5",
			Name: "Bada Read",
		},
		"program6": {
			ID:   "program6",
			Name: "Bada Sound",
		},
		"program7": {
			ID:   "program7",
			Name: "Bada Rhyme",
		},
	}
	// var programs []*entity.Program
	// err := da.GetProgramDA().Query(ctx, &da.ProgramCondition{}, &programs)
	// if err != nil {
	// 	log.Error(ctx, "query error", log.Err(err))
	// 	return err
	// }

	// // key: program name
	// s.ourPrograms = make(map[string]*entity.Program, len(programs))
	// for _, program := range programs {
	// 	s.ourPrograms[program.Name] = program
	// }

	return nil
}
