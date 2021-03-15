package intergrate_academic_profile

import (
	"context"
	"log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

func (s *MapperImpl) initGradeMapper(ctx context.Context) error {
	log.Println("start init grade mapper")
	s.MapperGrade.gradeMapping = make(map[string]string)
	err := s.loadAmsGrades(ctx)
	if err != nil {
		return err
	}

	err = s.loadOurGrades(ctx)
	if err != nil {
		log.Println("init grade mapper error")
		return err
	}
	log.Println("completed init grade mapper")
	return nil
}

func (s *MapperImpl) loadAmsGrades(ctx context.Context) error {
	s.MapperGrade.amsGrades = make(map[string]map[string]*external.Grade, len(s.amsPrograms))
	for _, amsProgram := range s.amsPrograms {
		amsGrades, err := external.GetGradeServiceProvider().GetByProgram(ctx, s.operator, amsProgram.ID)
		if err != nil {
			return err
		}
		s.MapperGrade.amsGrades[amsProgram.ID] = make(map[string]*external.Grade, len(amsGrades))
		for _, amsGrade := range amsGrades {
			s.MapperGrade.amsGrades[amsProgram.ID][amsGrade.Name] = amsGrade
		}
	}
	return nil
}

func (s *MapperImpl) loadOurGrades(ctx context.Context) error {
	var ourGrades []*entity.Grade
	err := da.GetGradeDA().Query(ctx, &da.GradeCondition{}, &ourGrades)
	if err != nil {
		return err
	}

	s.MapperGrade.ourGrades = make(map[string]*entity.Grade, len(ourGrades))
	for _, grade := range ourGrades {
		s.MapperGrade.ourGrades[grade.ID] = grade
	}
	return nil
}

func (s *MapperImpl) Grade(ctx context.Context, organizationID, programID, gradeID string) (string, error) {
	cacheID, found := s.MapperGrade.gradeMapping[gradeID]
	if found {
		return cacheID, nil
	}

	s.MapperGrade.amsGradeMutex.Lock()
	defer s.MapperGrade.amsGradeMutex.Unlock()

	// our
	ourGrades, found := s.MapperGrade.ourGrades[gradeID]
	if !found {
		return s.defaultGradeID()
	}

	// ams
	amsProgramID, err := s.Program(ctx, organizationID, programID)
	if err != nil {
		return "", err
	}
	amsGrades, found := s.MapperGrade.amsGrades[amsProgramID]
	if !found {
		return s.defaultGradeID()
	}
	amsGrade, found := amsGrades[ourGrades.Name]
	if !found {
		for _, item := range amsGrades {
			return item.ID, nil
		}
		return s.defaultGradeID()
	}

	// mapping
	s.MapperGrade.gradeMapping[gradeID] = amsGrade.ID

	return amsGrade.ID, nil
}

func (s *MapperImpl) defaultGradeID() (string, error) {
	noneName := "None Specified"
	program, found := s.amsPrograms[noneName]
	if !found {
		return "", constant.ErrRecordNotFound
	}
	grade, found := s.MapperGrade.amsGrades[program.ID][noneName]
	if !found {
		return "", constant.ErrRecordNotFound
	}
	return grade.ID, nil
}
