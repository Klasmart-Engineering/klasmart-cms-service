package external

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type TeacherServiceProvider interface {
	Get(ctx context.Context, id string) (*Teacher, error)
	BatchGet(ctx context.Context, ids []string) ([]*NullableTeacher, error)
	GetByOrganization(ctx context.Context, organizationID string) ([]*Teacher, error)
	Query(ctx context.Context, organizationID, keyword string) ([]*Teacher, error)
}

type Teacher struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type NullableTeacher struct {
	Valid bool `json:"-"`
	*Teacher
}

var (
	_amsTeacherService *AmsTeacherService
	_amsTeacherOnce    sync.Once
)

func GetTeacherServiceProvider() TeacherServiceProvider {
	_amsTeacherOnce.Do(func() {
		_amsTeacherService = &AmsTeacherService{}
	})

	return _amsTeacherService
}

type AmsTeacherService struct{}

func (s AmsTeacherService) Get(ctx context.Context, id string) (*Teacher, error) {
	teachers, err := s.BatchGet(ctx, []string{id})
	if err != nil {
		return nil, err
	}

	if !teachers[0].Valid {
		return nil, constant.ErrRecordNotFound
	}

	return teachers[0].Teacher, nil
}

func (s AmsTeacherService) BatchGet(ctx context.Context, ids []string) ([]*NullableTeacher, error) {
	if len(ids) == 0 {
		return []*NullableTeacher{}, nil
	}

	users, err := GetUserServiceProvider().BatchGet(ctx, ids)
	if err != nil {
		return nil, err
	}

	teachers := make([]*NullableTeacher, len(users))
	for index, user := range users {
		teachers[index] = &NullableTeacher{
			Valid: user.Valid,
			Teacher: &Teacher{
				ID:   user.User.ID,
				Name: user.User.Name,
			},
		}
	}

	return teachers, nil
}

func (s AmsTeacherService) GetByOrganization(ctx context.Context, organizationID string) ([]*Teacher, error) {
	request := chlorine.NewRequest(`
	query ($organization_id: ID!) {
		organization(organization_id: $organization_id) {
			classes{
				teachers{
					id: user_id
					name: user_name
				}
			}    
		}
	}`)
	request.Var("organization_id", organizationID)

	data := &struct {
		Organization struct {
			Classes []struct {
				Teachers []*Teacher `json:"teachers"`
			} `json:"classes"`
		} `json:"organization"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get teachers by org failed",
			log.Err(err),
			log.String("organizationID", organizationID))
		return nil, err
	}

	teachers := make([]*Teacher, 0, len(data.Organization.Classes))
	for _, class := range data.Organization.Classes {
		teachers = append(teachers, class.Teachers...)
	}

	return teachers, nil
}

func (s AmsTeacherService) Query(ctx context.Context, organizationID, keyword string) ([]*Teacher, error) {
	users, err := GetUserServiceProvider().Query(ctx, organizationID, keyword)
	if err != nil {
		return nil, err
	}

	teachers := make([]*Teacher, len(users))
	for index, user := range users {
		teachers[index] = &Teacher{
			ID:   user.ID,
			Name: user.Name,
		}
	}

	return teachers, nil
}
