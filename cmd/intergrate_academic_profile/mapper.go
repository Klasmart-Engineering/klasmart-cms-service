package intergrate_academic_profile

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

type Mapper interface {
	Program(ctx context.Context, organizationID, programID string) (string, error)
	Subject(ctx context.Context, organizationID, programID, subjectID string) (string, error)
	// developmental
	Category(ctx context.Context, organizationID, programID, categoryID string) (string, error)
	// skills
	SubCategory(ctx context.Context, organizationID, programID, categoryID, subCategoryID string) (string, error)
	Age(ctx context.Context, organizationID, programID, AgeID string) (string, error)
	Grade(ctx context.Context, organizationID, programID, gradeID string) (string, error)
}

func NewMapper(operator *entity.Operator) Mapper {
	ctx := context.TODO()

	impl := &MapperImpl{operator: operator}

	err := impl.initProgramMapper(ctx)
	if err != nil {
		log.Fatal(ctx, "init program mapping failed", log.Err(err))
	}

	return impl
}

type MapperImpl struct {
	operator *entity.Operator

	amsProgramMutex sync.Mutex
	// key: ams program name
	amsPrograms map[string]*external.Program
	// key: our program id
	ourPrograms map[string]*entity.Program
	// key: our program id
	// value: ams program id
	programMapping map[string]string

	MapperAge   MapperAge
	MapperGrade MapperGrade
}
type MapperAge struct {
	amsAgeMutex sync.Mutex

	// key:program id
	// val:age map(name:Age)
	amsAges map[string]map[string]*external.Age
	// key: our age id
	ourAges map[string]*entity.Age

	// key: our age id
	// value: ams age id
	ageMapping map[string]string
}

type MapperGrade struct {
	amsGradeMutex sync.Mutex

	// key:program id
	// val:age map(name:Age)
	amsGrades map[string]map[string]*external.Grade
	// key: our age id
	ourGrades map[string]*entity.Grade

	// key: our age id
	// value: ams age id
	gradeMapping map[string]string
}

func (s MapperImpl) Subject(ctx context.Context, organizationID, programID, subjectID string) (string, error) {
	// TODO
	return "", nil
}
