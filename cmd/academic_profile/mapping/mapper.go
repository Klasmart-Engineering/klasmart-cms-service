package mapping

import (
	"context"
)

type Mapper interface {
	Program(ctx context.Context, organizationID, programID string) string
	Subject(ctx context.Context, organizationID, programID, subjectID string) string
	// developmental
	Category(ctx context.Context, organizationID, programID, categoryID string) string
	// skills
	SubCategory(ctx context.Context, organizationID, programID, categoryID, subCategoryID string) string
	Age(ctx context.Context, organizationID, programID, AgeID string) string
	Grade(ctx context.Context, organizationID, programID, gradeID string) string
}

func GetMapper() Mapper {
	return nil
}

type mapperImpl struct {
}

func NewMapperImpl() *mapperImpl {
	return &mapperImpl{}
}
