package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/basicdata"
)

type SubjectServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*entity.Subject, error)
}

func GetSubjectServiceProvider() SubjectServiceProvider {
	return &subjectService{}
}

type mockSubjectService struct{}

func (s mockSubjectService) BatchGet(ctx context.Context, ids []string) ([]*entity.Subject, error) {
	var subjects []*entity.Subject
	for _, option := range GetMockData().Options {
		subjects = append(subjects, option.Subject...)
	}
	return subjects, nil
}

type subjectService struct{}

func (s subjectService) BatchGet(ctx context.Context, ids []string) ([]*entity.Subject, error) {
	result, err := basicdata.GetSubjectModel().Query(ctx, &da.SubjectCondition{
		IDs: entity.NullStrings{
			Strings: ids,
			Valid:   len(ids) != 0,
		},
	})
	if err != nil {
		log.Error(ctx, "BatchGet:error", log.Err(err), log.Strings("ids", ids))
		return nil, err
	}
	return result, nil
}
