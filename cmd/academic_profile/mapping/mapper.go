package mapping

import "context"

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
