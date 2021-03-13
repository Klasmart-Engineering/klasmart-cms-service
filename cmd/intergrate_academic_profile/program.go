package intergrate_academic_profile

import (
	"context"

	"github.com/nzai/qr/constants"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

func (s *MapperImpl) Program(ctx context.Context, organizationID, programID string) (string, error) {
	programID, found := s.programMapping[programID]
	if found {
		return programID, nil
	}

	s.amsProgramMutex.Lock()
	defer s.amsProgramMutex.Unlock()

	ourProgram, found := s.ourPrograms[programID]
	if !found {
		log.Error(ctx, "unknown program id", log.String("programID", programID), log.String("organizationID", organizationID))
		return "", constants.ErrRecordNotFound
	}

	amsProgram, found := s.amsPrograms[ourProgram.Name]
	if !found {
		log.Error(ctx, "unknown program id", log.Any("program", ourProgram), log.String("organizationID", organizationID))
		return "", constants.ErrRecordNotFound
	}

	s.programMapping[ourProgram.ID] = amsProgram.ID

	return amsProgram.ID, nil
}

func (s *MapperImpl) initProgramMapper(ctx context.Context) error {
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

	s.amsPrograms = make(map[string]*external.Program, len(programs))
	for _, program := range programs {
		s.amsPrograms[program.Name] = program
	}

	return nil
}

func (s *MapperImpl) loadOurPrograms(ctx context.Context) error {
	var programs []*entity.Program
	err := da.GetProgramDA().Query(ctx, &da.ProgramCondition{}, &programs)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err))
		return err
	}

	// key: program name
	s.ourPrograms = make(map[string]*entity.Program, len(programs))
	for _, program := range programs {
		s.ourPrograms[program.Name] = program
	}

	return nil
}
