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
		log.Panic(ctx, "init program mapping failed", log.Err(err))
	}

	err = impl.initSubjectMapper(ctx)
	if err != nil {
		log.Panic(ctx, "init subject mapping failed", log.Err(err))
	}

	return impl
}

type MapperImpl struct {
	operator *entity.Operator

	programMutex sync.Mutex
	// key: ams program name
	amsPrograms map[string]*external.Program
	// key: our program id
	ourPrograms map[string]*entity.Program
	// key: our program id
	// value: ams program id
	programMapping map[string]string

	subjectMutex sync.Mutex
	// key: {ams program id}:{ams subject name}
	amsSubjects map[string]*external.Subject
	// key: our subject id
	ourSubjects map[string]*entity.Subject
	// key:  {our program id}:{our subject id}
	// value: ams subject id
	subjectMapping map[string]string
}

func (s MapperImpl) SubCategory(ctx context.Context, organizationID, programID, categoryID, subCategoryID string) (string, error) {
	// TODO
	return "", nil
}

func (s MapperImpl) Age(ctx context.Context, organizationID, programID, AgeID string) (string, error) {
	// TODO
	return "", nil
}

func (s MapperImpl) Grade(ctx context.Context, organizationID, programID, gradeID string) (string, error) {
	// TODO
	return "", nil
}
