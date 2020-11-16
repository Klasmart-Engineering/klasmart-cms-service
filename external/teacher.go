package external

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type TeacherServiceProvider interface {
	Get(ctx context.Context, id string) (*Teacher, error)
	BatchGet(ctx context.Context, ids []string) ([]*NullableTeacher, error)
	GetByOrganization(ctx context.Context, organizationID string) ([]*Teacher, error)
	GetByOrganizations(ctx context.Context, organizationIDs []string) (map[string][]*Teacher, error)
	GetBySchool(ctx context.Context, schoolID string) ([]*Teacher, error)
	GetBySchools(ctx context.Context, schoolIDs []string) (map[string][]*Teacher, error)
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

func (s AmsTeacherService) GetByOrganizations(ctx context.Context, organizationIDs []string) (map[string][]*Teacher, error) {
	if len(organizationIDs) == 0 {
		return map[string][]*Teacher{}, nil
	}

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range organizationIDs {
		fmt.Fprintf(sb, "q%d: organization(organization_id: \"%s\") {classes{teachers{id:user_id name:user_name}}}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String())

	data := map[string]*struct {
		Classes []struct {
			Teachers []*Teacher `json:"teachers"`
		} `json:"classes"`
	}{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get users by organization ids failed", log.Err(err), log.Strings("ids", organizationIDs))
		return nil, err
	}

	teachers := make(map[string][]*Teacher, len(organizationIDs))
	for index, organizationID := range organizationIDs {
		classes := data[fmt.Sprintf("q%d", index)]
		for _, class := range classes.Classes {
			teachers[organizationID] = append(teachers[organizationID], class.Teachers...)
		}
	}

	return teachers, nil
}

func (s AmsTeacherService) GetBySchool(ctx context.Context, schoolID string) ([]*Teacher, error) {
	request := chlorine.NewRequest(`
	query ($school_id: ID!) {
		school(school_id: $school_id) {
			classes{
				teachers{
					id: user_id
					name: user_name
				}
			}    
		}
	}`)
	request.Var("school_id", schoolID)

	data := &struct {
		School struct {
			Classes []struct {
				Teachers []*Teacher `json:"teachers"`
			} `json:"classes"`
		} `json:"school"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get teachers by school failed",
			log.Err(err),
			log.String("schoolID", schoolID))
		return nil, err
	}

	teachers := make([]*Teacher, 0, len(data.School.Classes))
	for _, class := range data.School.Classes {
		teachers = append(teachers, class.Teachers...)
	}

	return teachers, nil
}

func (s AmsTeacherService) GetBySchools(ctx context.Context, schoolIDs []string) (map[string][]*Teacher, error) {
	return nil, nil
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
