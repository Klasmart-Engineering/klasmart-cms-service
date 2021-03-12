package intergrate_academic_profile

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
	// TODO
	return &MapperImpl{}
}

type MapperImpl struct{}

func (s MapperImpl) Program(ctx context.Context, organizationID, programID string) (string, error) {
	// TODO
	return "", nil
}

func (s MapperImpl) Subject(ctx context.Context, organizationID, programID, subjectID string) (string, error) {
	// TODO
	return "", nil
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
