package external

import (
	"context"
	"strconv"
	"text/template"

	"go.uber.org/zap/buffer"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	cl "gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type OrganizationServiceProvider interface {
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableOrganization, error)
	GetMine(ctx context.Context, operator *entity.Operator, userID string) ([]*Organization, error)
	GetParents(ctx context.Context, operator *entity.Operator, orgID string) ([]*Organization, error)
	GetChildren(ctx context.Context, operator *entity.Operator, orgID string) ([]*Organization, error)
	GetOrganizationOrSchoolName(ctx context.Context, operator *entity.Operator, id []string) ([]string, error)
	GetByPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*Organization, error)
	GetOrganizationsAssociatedWithUserID(ctx context.Context, operator *entity.Operator, id string) ([]*Organization, error)
}

type Organization struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parent_id"`
}

type NullableOrganization struct {
	Organization
	Valid bool `json:"-"`
}

func GetOrganizationServiceProvider() OrganizationServiceProvider {
	return &AmsOrganizationService{}
}

type AmsOrganizationService struct{}

func (s AmsOrganizationService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableOrganization, error) {
	q := `query orgs($orgIDs: [ID!]){
	organizations(organization_ids: $orgIDs){
    	id: organization_id
    	name: organization_name
  	}
}`
	req := cl.NewRequest(q, chlorine.ReqToken(operator.Token))
	req.Var("orgIDs", ids)
	payload := make([]*Organization, len(ids))
	res := cl.Response{
		Data: &struct {
			Organizations []*Organization `json:"organizations"`
		}{Organizations: payload},
	}
	_, err := GetAmsClient().Run(ctx, req, &res)
	if err != nil {
		log.Error(ctx, "Run error", log.String("q", q), log.Any("res", res), log.Err(err))
		return nil, err
	}
	if len(res.Errors) > 0 {
		log.Error(ctx, "Res error", log.String("q", q), log.Any("res", res), log.Err(res.Errors))
		return nil, res.Errors
	}
	nullableOrganizations := make([]*NullableOrganization, len(payload))
	for i := range payload {
		if payload[i] == nil {
			nullableOrganizations[i] = &NullableOrganization{Valid: false}
		} else {
			nullableOrganizations[i] = &NullableOrganization{*payload[i], true}
		}
	}

	log.Info(ctx, "get orgs by ids success",
		log.Strings("ids", ids),
		log.Any("orgs", nullableOrganizations))

	return nullableOrganizations, nil
}

func (s AmsOrganizationService) GetMine(ctx context.Context, operator *entity.Operator, userID string) ([]*Organization, error) {
	// TODO: Maybe don't need
	return []*Organization{}, nil
}

func (s AmsOrganizationService) GetParents(ctx context.Context, operator *entity.Operator, orgID string) ([]*Organization, error) {
	return []*Organization{}, nil
}

func (s AmsOrganizationService) GetChildren(ctx context.Context, operator *entity.Operator, orgID string) ([]*Organization, error) {
	return []*Organization{}, nil
}

func (s AmsOrganizationService) GetOrganizationOrSchoolName(ctx context.Context, operator *entity.Operator, ids []string) ([]string, error) {
	raw := `query{
	{{range $i, $e := .}}
	org_{{$i}}: organization(organization_id: "{{$e}}"){
		id: organization_id
    	name: organization_name
  	}
	sch_{{$i}}: school(school_id: "{{$e}}"){
		id: school_id
    	name: school_name
  	}
	{{end}}
}`
	temp, err := template.New("OrgSch").Parse(raw)
	if err != nil {
		log.Error(ctx, "temp error", log.String("raw", raw), log.Err(err))
		return nil, err
	}
	buf := buffer.Buffer{}
	err = temp.Execute(&buf, ids)
	if err != nil {
		log.Error(ctx, "temp execute failed", log.String("raw", raw), log.Err(err))
		return nil, err
	}
	req := chlorine.NewRequest(buf.String(), chlorine.ReqToken(operator.Token))
	type Payload struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	payload := make(map[string]*Payload, len(ids))
	res := chlorine.Response{
		Data: &payload,
	}

	_, err = GetAmsClient().Run(ctx, req, &res)
	if err != nil {
		log.Error(ctx, "Run error", log.String("q", buf.String()), log.Any("res", res), log.Err(err))
		return nil, err
	}
	if len(res.Errors) > 0 {
		log.Error(ctx, "Res error", log.String("q", buf.String()), log.Any("res", res), log.Err(res.Errors))
		return nil, res.Errors
	}
	nameList := make([]string, len(ids))
	for k, v := range payload {
		index, err := strconv.Atoi(k[len("org_"):])
		if err != nil {
			log.Error(ctx, "Res error", log.String("q", buf.String()), log.Any("res", res), log.Err(res.Errors))
			return nil, err
		}
		if v != nil && nameList[index] == "" {
			nameList[index] = v.Name
		}

	}

	log.Info(ctx, "get names by org or school ids success",
		log.Strings("ids", ids),
		log.Strings("names", nameList))

	return nameList, nil
}

func (s AmsOrganizationService) GetByPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*Organization, error) {
	request := chlorine.NewRequest(`
	query(
		$user_id: ID!
		$permission_name: String!
	) {
		user(user_id: $user_id) {
			organizationsWithPermission(permission_name: $permission_name) {
				organization {
					organization_id
					organization_name
				}
			}
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("user_id", operator.UserID)
	request.Var("permission_name", permissionName.String())

	data := &struct {
		User struct {
			OrganizationsWithPermission []struct {
				Organization struct {
					OrganizationID   string `json:"organization_id"`
					OrganizationName string `json:"organization_name"`
				} `json:"organization"`
			} `json:"organizationsWithPermission"`
		} `json:"user"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get has permission organizations failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("permissionName", permissionName.String()))
		return nil, err
	}

	orgs := make([]*Organization, 0, len(data.User.OrganizationsWithPermission))
	for _, membership := range data.User.OrganizationsWithPermission {
		orgs = append(orgs, &Organization{
			ID:   membership.Organization.OrganizationID,
			Name: membership.Organization.OrganizationName,
		})
	}

	log.Info(ctx, "get orgs by permission success",
		log.Any("operator", operator),
		log.String("permissionName", permissionName.String()),
		log.Any("orgs", orgs))

	return orgs, nil
}

func (s AmsOrganizationService) GetOrganizationsAssociatedWithUserID(ctx context.Context, operator *entity.Operator, id string) ([]*Organization, error) {
	request := chlorine.NewRequest(`
	query($user_id: ID!) {
		user(user_id: $user_id) {
			memberships{
				organization{
					id:organization_id
					name:organization_name
				}
			}
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("user_id", id)

	data := &struct {
		User struct {
			Memberships []struct {
				Organization Organization `json:"organization"`
			} `json:"memberships"`
		} `json:"user"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get orgs by user failed",
			log.Err(err),
			log.String("userID", id))
		return nil, err
	}

	orgs := make([]*Organization, 0)
	for _, membership := range data.User.Memberships {
		orgs = append(orgs, &Organization{
			ID:   membership.Organization.ID,
			Name: membership.Organization.Name,
		})
	}

	log.Info(ctx, "get orgs by user success",
		log.String("userID", id),
		log.Any("orgs", orgs))

	return orgs, nil
}
