package external

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type SchoolServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*NullableSchool, error)
	GetByOrganizationID(ctx context.Context, organizationID string) ([]*School, error)
	GetByPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*School, error)
}

type School struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type NullableSchool struct {
	School
	Valid bool `json:"-"`
}

var (
	_amsSchoolService *AmsSchoolService
	_amsSchoolOnce    sync.Once
)

func GetSchoolServiceProvider() SchoolServiceProvider {
	_amsSchoolOnce.Do(func() {
		_amsSchoolService = &AmsSchoolService{}
	})

	return _amsSchoolService
}

type AmsSchoolService struct{}

func (s AmsSchoolService) BatchGet(ctx context.Context, ids []string) ([]*NullableSchool, error) {
	if len(ids) == 0 {
		return []*NullableSchool{}, nil
	}

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range ids {
		fmt.Fprintf(sb, "u%d: user(user_id: \"%s\") {user_id user_name}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String())

	data := map[string]*struct {
		SchoolID   string `json:"school_id"`
		SchoolName string `json:"school_name"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get schools by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	var queryAlias string
	schools := make([]*NullableSchool, 0, len(data))
	for index := range ids {
		queryAlias = fmt.Sprintf("u%d", index)
		school, found := data[queryAlias]

		if !found || school == nil {
			schools = append(schools, &NullableSchool{Valid: false})
			continue
			// log.Error(ctx, "schools not found", log.Strings("ids", ids), log.String("id", ids[index]))
			// return nil, constant.ErrRecordNotFound
		}

		schools = append(schools, &NullableSchool{
			Valid: true,
			School: School{
				ID:   school.SchoolID,
				Name: school.SchoolName,
			},
		})
	}

	return schools, nil
}

func (s AmsSchoolService) GetByOrganizationID(ctx context.Context, organizationID string) ([]*School, error) {
	request := chlorine.NewRequest(`
	query($organization_id: ID!) {
		organization(organization_id: $organization_id) {
			schools{
				school_id
				school_name
			}
		}
	}`)
	request.Var("organization_id", organizationID)

	data := &struct {
		Organization struct {
			Schools []struct {
				SchoolID   string `json:"school_id"`
				SchoolName string `json:"school_name"`
			} `json:"schools"`
		} `json:"organization"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query schools by organization failed",
			log.Err(err),
			log.String("organizationID", organizationID))
		return nil, err
	}

	schools := make([]*School, 0, len(data.Organization.Schools))
	for _, school := range data.Organization.Schools {
		schools = append(schools, &School{
			ID:   school.SchoolID,
			Name: school.SchoolName,
		})
	}

	return schools, nil
}

func (s AmsSchoolService) GetByPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*School, error) {
	request := chlorine.NewRequest(`
	query(
		$user_id: ID!
		$permission_name: String!
	) {
		user(user_id: $user_id) {
			schoolsWithPermission(permission_name: $permission_name) {
				school {
					school_id
					school_name
				}
			}
		}
	}`)
	request.Var("user_id", operator.UserID)
	request.Var("permission_name", permissionName.String())

	data := &struct {
		User struct {
			SchoolsWithPermission []struct {
				School struct {
					SchoolID   string `json:"school_id"`
					SchoolName string `json:"school_name"`
				} `json:"school"`
			} `json:"schoolsWithPermission"`
		} `json:"user"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get schools by permission failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("permissionName", permissionName.String()))
		return nil, err
	}

	schools := make([]*School, 0, len(data.User.SchoolsWithPermission))
	for _, membership := range data.User.SchoolsWithPermission {
		schools = append(schools, &School{
			ID:   membership.School.SchoolID,
			Name: membership.School.SchoolName,
		})
	}

	return schools, nil
}
