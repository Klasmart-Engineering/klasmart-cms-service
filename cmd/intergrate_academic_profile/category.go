package intergrate_academic_profile

import (
	"context"
)

func (s *MapperImpl) initCategoryMapper(ctx context.Context) error {
	err := s.loadAmsCategorys(ctx)
	if err != nil {
		return err
	}

	return s.loadOurCategorys(ctx)
}

func (s *MapperImpl) loadAmsCategorys(ctx context.Context) error {
	return nil
}

func (s *MapperImpl) loadOurCategorys(ctx context.Context) error {
	return nil
}

func (s *MapperImpl) Category(ctx context.Context, organizationID, programID, categoryID string) (string, error) {
	return "", nil
}
