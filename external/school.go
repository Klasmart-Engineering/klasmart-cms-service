package external

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type SchoolServiceProvider interface {
	Get(ctx context.Context, operator *entity.Operator, id string) (*School, error)
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableSchool, error)
	GetByOrganizationID(ctx context.Context, operator *entity.Operator, organizationID string) ([]*School, error)
	GetByPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*School, error)
	GetByOperator(ctx context.Context, operator *entity.Operator) ([]*School, error)
}

type School struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type NullableSchool struct {
	Valid bool `json:"-"`
	*School
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

func (s AmsSchoolService) Get(ctx context.Context, operator *entity.Operator, id string) (*School, error) {
	schools, err := s.BatchGet(ctx, operator, []string{id})
	if err != nil {
		return nil, err
	}

	if !schools[0].Valid {
		return nil, constant.ErrRecordNotFound
	}

	return schools[0].School, nil
}

func (s AmsSchoolService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableSchool, error) {
	if len(ids) == 0 {
		return []*NullableSchool{}, nil
	}

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range ids {
		fmt.Fprintf(sb, "q%d: school(school_id: \"%s\") {id:school_id name:school_name}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))

	data := map[string]*School{}

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

	schools := make([]*NullableSchool, 0, len(data))
	for index := range ids {
		school := data[fmt.Sprintf("q%d", index)]
		schools = append(schools, &NullableSchool{
			Valid:  school != nil,
			School: school,
		})
	}

	log.Info(ctx, "get schools by ids success",
		log.Strings("ids", ids),
		log.Any("schools", schools))

	return schools, nil
}

func (s AmsSchoolService) GetByOrganizationID(ctx context.Context, operator *entity.Operator, organizationID string) ([]*School, error) {
	request := chlorine.NewRequest(`
	query($organization_id: ID!) {
		organization(organization_id: $organization_id) {
			schools{
				school_id
				school_name
			}
		}
	}`, chlorine.ReqToken(operator.Token))
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

	log.Info(ctx, "query schools by organization success",
		log.String("organizationID", organizationID),
		log.Any("schools", schools))

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
					organization {
						organization_id
					}
				}
			}
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("user_id", operator.UserID)
	request.Var("permission_name", permissionName.String())

	data := &struct {
		User struct {
			SchoolsWithPermission []struct {
				School struct {
					SchoolID     string `json:"school_id"`
					SchoolName   string `json:"school_name"`
					Organization struct {
						OrganizationID string `json:"organization_id"`
					} `json:"organization"`
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
		// filtering by operator's org id
		if membership.School.Organization.OrganizationID != operator.OrgID {
			continue
		}

		schools = append(schools, &School{
			ID:   membership.School.SchoolID,
			Name: membership.School.SchoolName,
		})
	}

	log.Info(ctx, "get schools by permission",
		log.Any("operator", operator),
		log.String("permissionName", permissionName.String()),
		log.Any("schools", schools))

	return schools, nil
}

func (s AmsSchoolService) GetByOperator(ctx context.Context, operator *entity.Operator) ([]*School, error) {
	request := chlorine.NewRequest(`
	query($user_id: ID!) {
		user(user_id: $user_id) {
			school_memberships{
				school {
					school_id
					school_name
					organization {
						organization_id
					}
				}
			}
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("user_id", operator.UserID)

	data := &struct {
		User struct {
			SchoolMemberships []struct {
				School struct {
					SchoolID     string `json:"school_id"`
					SchoolName   string `json:"school_name"`
					Organization struct {
						OrganizationID string `json:"organization_id"`
					} `json:"organization"`
				} `json:"school"`
			} `json:"school_memberships"`
		} `json:"user"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get schools by operator failed",
			log.Err(err),
			log.Any("operator", operator))
		return nil, err
	}

	schools := make([]*School, 0)
	for _, membership := range data.User.SchoolMemberships {
		// filtering by operator's org id
		if membership.School.Organization.OrganizationID != operator.OrgID {
			continue
		}

		schools = append(schools, &School{
			ID:   membership.School.SchoolID,
			Name: membership.School.SchoolName,
		})
	}

	log.Info(ctx, "get schools by operator success",
		log.Any("operator", operator))

	return schools, nil
}
