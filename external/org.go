package external

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	cl "gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type OrganizationServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Organization, error)
	GetMine(ctx context.Context, userID string) ([]*Organization, error)
	GetParents(ctx context.Context, orgID string) ([]*Organization, error)
	GetChildren(ctx context.Context, orgID string) ([]*Organization, error)
	GetOrganizationOrSchoolName(ctx context.Context, id []string) (map[string]string, error)
	GetByPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*Organization, error)
}

type Organization struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parent_id"`
}

func GetOrganizationServiceProvider() OrganizationServiceProvider {
	return &AmsOrganizationService{}
}

type AmsOrganizationService struct{}

func (s AmsOrganizationService) BatchGet(ctx context.Context, ids []string) ([]*Organization, error) {
	q := `query orgs($orgIDs: [ID!]){
	organizations(organization_ids: $orgIDs){
    	id: organization_id
    	name: organization_name
  	}
}`
	req := cl.NewRequest(q)
	req.Var("orgIDs", ids)
	payload := make([]*Organization, len(ids))
	res := cl.Response{
		Data: &struct {
			Organizations []*Organization `json:"organizations"`
		}{Organizations: payload},
	}
	_, err := GetChlorine().Run(ctx, req, &res)
	if err != nil {
		log.Error(ctx, "Run error", log.String("q", q), log.Any("res", res), log.Err(err))
		return nil, err
	}
	if len(res.Errors) > 0 {
		log.Error(ctx, "Res error", log.String("q", q), log.Any("res", res), log.Err(res.Errors))
		return nil, res.Errors
	}
	return payload, nil
}

func (s AmsOrganizationService) GetMine(ctx context.Context, userID string) ([]*Organization, error) {
	// TODO: Maybe don't need
	return []*Organization{}, nil
}

func (s AmsOrganizationService) GetParents(ctx context.Context, orgID string) ([]*Organization, error) {
	return []*Organization{}, nil
}

func (s AmsOrganizationService) GetChildren(ctx context.Context, orgID string) ([]*Organization, error) {
	return []*Organization{}, nil
}

func (s AmsOrganizationService) GetOrganizationOrSchoolName(ctx context.Context, id []string) (map[string]string, error){
	return "", nil
}

func (s AmsOrganizationService) GetByPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*Organization, error) {
	request := chlorine.NewRequest(`
	query(
		$user_id: ID!
		$permission_name: ID!
	){
		user(user_id: $user_id) {
			memberships {
				organization{
					organization_id
					organization_name        
				}
				checkAllowed(permission_name: $permission_name)
			}
		}
	}`)
	request.Var("user_id", operator.UserID)
	request.Var("permission_name", permissionName.String())

	data := &struct {
		User struct {
			Memberships []struct {
				Organization struct {
					OrganizationID   string `json:"organization_id"`
					OrganizationName string `json:"organization_name"`
				} `json:"organization"`
				CheckAllowed bool `json:"checkAllowed"`
			} `json:"memberships"`
		} `json:"user"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get has permission organizations failed",
			log.Any("operator", operator),
			log.String("permissionName", permissionName.String()))
		return nil, err
	}

	orgs := make([]*Organization, 0, len(data.User.Memberships))
	for _, membership := range data.User.Memberships {
		if !membership.CheckAllowed {
			continue
		}

		orgs = append(orgs, &Organization{
			ID:   membership.Organization.OrganizationID,
			Name: membership.Organization.OrganizationName,
		})
	}

	return orgs, nil
}
