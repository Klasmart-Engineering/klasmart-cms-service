package external

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"

	"go.uber.org/zap/buffer"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	cl "gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type OrganizationServiceProvider interface {
	cache.IDataSource
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableOrganization, error)
	BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*NullableOrganization, error)
	BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error)
	GetByClasses(ctx context.Context, operator *entity.Operator, classIDs []string, options ...APOption) (map[string]*Organization, error)
	GetNameByOrganizationOrSchool(ctx context.Context, operator *entity.Operator, id []string) ([]string, error)
	GetNameMapByOrganizationOrSchool(ctx context.Context, operator *entity.Operator, id []string) (map[string]string, error)
	GetByPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName, options ...APOption) ([]*Organization, error)
	GetByUserID(ctx context.Context, operator *entity.Operator, id string, options ...APOption) ([]*Organization, error)
}

type Organization struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Status APStatus `json:"status"`
}

type NullableOrganization struct {
	Organization
	StrID string `json:"str_id"`
	Valid bool   `json:"valid"`
}

func (n *NullableOrganization) StringID() string {
	return n.StrID
}
func (n *NullableOrganization) RelatedIDs() []*cache.RelatedEntity {
	return nil
}

func GetOrganizationServiceProvider() OrganizationServiceProvider {
	return &AmsOrganizationService{}
}

type AmsOrganizationService struct{}

func (s AmsOrganizationService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableOrganization, error) {
	if len(ids) == 0 {
		return []*NullableOrganization{}, nil
	}

	res := make([]*NullableOrganization, 0, len(ids))
	err := cache.GetPassiveCacheRefresher().BatchGet(ctx, s.Name(), ids, &res, operator)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s AmsOrganizationService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	operator, err := optionsWithOperator(ctx, options...)
	if err != nil {
		return nil, err
	}

	q := `query orgs($orgIDs: [ID!]){
	organizations(organization_ids: $orgIDs){
    	id: organization_id
    	name: organization_name
		status
  	}
}`

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	req := cl.NewRequest(q, chlorine.ReqToken(operator.Token))
	req.Var("orgIDs", _ids)
	payload := make([]*Organization, len(ids))
	res := cl.Response{
		Data: &struct {
			Organizations []*Organization `json:"organizations"`
		}{Organizations: payload},
	}
	_, err = GetAmsClient().Run(ctx, req, &res)
	if err != nil {
		log.Error(ctx, "Run error", log.String("q", q), log.Any("res", res), log.Err(err))
		return nil, err
	}
	if len(res.Errors) > 0 {
		log.Error(ctx, "Res error", log.String("q", q), log.Any("res", res), log.Err(res.Errors))
		return nil, res.Errors
	}
	nullableOrganizations := make([]cache.Object, len(ids))
	for index := range ids {
		if payload[indexMapping[index]] == nil {
			nullableOrganizations[index] = &NullableOrganization{
				StrID: ids[index],
				Valid: false,
			}
		} else {
			nullableOrganizations[index] = &NullableOrganization{
				Organization: *payload[indexMapping[index]],
				StrID:        ids[index],
				Valid:        true,
			}
		}
	}

	log.Info(ctx, "get orgs by ids success",
		log.Strings("ids", ids),
		log.Any("orgs", nullableOrganizations))

	return nullableOrganizations, nil
}
func (s AmsOrganizationService) BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*NullableOrganization, error) {
	organizations, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]*NullableOrganization{}, err
	}

	dict := make(map[string]*NullableOrganization, len(organizations))
	for _, organization := range organizations {
		if !organization.Valid {
			continue
		}
		dict[organization.ID] = organization
	}

	return dict, nil
}

func (s AmsOrganizationService) BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error) {
	organizations, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]string{}, err
	}

	dict := make(map[string]string, len(organizations))
	for _, organization := range organizations {
		if !organization.Valid {
			continue
		}
		dict[organization.ID] = organization.Name
	}

	return dict, nil
}

//TODO: No Test Program
func (s AmsOrganizationService) GetByClasses(ctx context.Context, operator *entity.Operator, classIDs []string, options ...APOption) (map[string]*Organization, error) {
	if len(classIDs) == 0 {
		return map[string]*Organization{}, nil
	}

	condition := NewCondition(options...)

	_classIDs, indexMapping := utils.SliceDeduplicationMap(classIDs)

	sb := new(strings.Builder)

	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$class_id_", ": ID!", len(_classIDs)))
	for index := range _classIDs {
		fmt.Fprintf(sb, `q%d: class(class_id: $class_id_%d) {organization{id:organization_id name:organization_name status}}`, index, index)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	for index, id := range _classIDs {
		request.Var(fmt.Sprintf("class_id_%d", index), id)
	}

	data := map[string]*struct {
		Organization Organization `json:"organization"`
	}{}

	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get organizations by classes failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Strings("classIDs", classIDs))
		return nil, err
	}

	orgs := make(map[string]*Organization, len(classIDs))
	for index, classID := range classIDs {
		class := data[fmt.Sprintf("q%d", indexMapping[index])]
		if class == nil {
			continue
		}

		if condition.Status.Valid {
			if condition.Status.Status != class.Organization.Status {
				continue
			}
		} else {
			// only status = "Active" data is returned by default
			if class.Organization.Status != Active {
				continue
			}
		}

		orgs[classID] = &class.Organization
	}

	log.Info(ctx, "get organizations by classes success",
		log.Any("operator", operator),
		log.Strings("classIDs", classIDs),
		log.Any("orgs", orgs))

	return orgs, nil
}

func (s AmsOrganizationService) GetNameByOrganizationOrSchool(ctx context.Context, operator *entity.Operator, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	raw := `query{
	{{range $i, $e := .}}
	org_{{$i}}: organization(organization_id: "{{$e}}"){
		id: organization_id
    	name: organization_name
		status
  	}
	sch_{{$i}}: school(school_id: "{{$e}}"){
		id: school_id
    	name: school_name
		status
  	}
	{{end}}
}`
	temp, err := template.New("OrgSch").Parse(raw)
	if err != nil {
		log.Error(ctx, "temp error", log.String("raw", raw), log.Err(err))
		return nil, err
	}
	buf := buffer.Buffer{}
	err = temp.Execute(&buf, _ids)
	if err != nil {
		log.Error(ctx, "temp execute failed", log.String("raw", raw), log.Err(err))
		return nil, err
	}
	req := chlorine.NewRequest(buf.String(), chlorine.ReqToken(operator.Token))
	type Payload struct {
		ID     string   `json:"id"`
		Name   string   `json:"name"`
		Status APStatus `json:"status"`
	}
	payload := make(map[string]*Payload, len(_ids))
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
	for i := range ids {
		orgKey := fmt.Sprintf("org_%d", indexMapping[i])
		schKey := fmt.Sprintf("sch_%d", indexMapping[i])
		if payload[orgKey] != nil && payload[orgKey].Name != "" {
			nameList[i] = payload[orgKey].Name
		}
		if payload[schKey] != nil && payload[schKey].Name != "" {
			nameList[i] = payload[schKey].Name
		}
	}

	log.Info(ctx, "get names by org or school ids success",
		log.Strings("ids", ids),
		log.Strings("names", nameList))

	return nameList, nil
}

func (s AmsOrganizationService) GetNameMapByOrganizationOrSchool(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error) {
	names, err := s.GetNameByOrganizationOrSchool(ctx, operator, ids)
	if err != nil {
		return map[string]string{}, err
	}

	dict := make(map[string]string, len(names))
	for index, name := range names {
		dict[ids[index]] = name
	}

	return dict, nil
}

func (s AmsOrganizationService) GetByPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName, options ...APOption) ([]*Organization, error) {
	condition := NewCondition(options...)

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
					status
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
					OrganizationID   string   `json:"organization_id"`
					OrganizationName string   `json:"organization_name"`
					Status           APStatus `json:"status"`
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
		if condition.Status.Valid {
			if condition.Status.Status != membership.Organization.Status {
				continue
			}
		} else {
			// only status = "Active" data is returned by default
			if membership.Organization.Status != Active {
				continue
			}
		}

		orgs = append(orgs, &Organization{
			ID:     membership.Organization.OrganizationID,
			Name:   membership.Organization.OrganizationName,
			Status: membership.Organization.Status,
		})
	}

	log.Info(ctx, "get orgs by permission success",
		log.Any("operator", operator),
		log.String("permissionName", permissionName.String()),
		log.Any("orgs", orgs))

	return orgs, nil
}

func (s AmsOrganizationService) GetByUserID(ctx context.Context, operator *entity.Operator, id string, options ...APOption) ([]*Organization, error) {
	condition := NewCondition(options...)

	request := chlorine.NewRequest(`
	query($user_id: ID!) {
		user(user_id: $user_id) {
			memberships{
				organization{
					id:organization_id
					name:organization_name
					status
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
		if condition.Status.Valid {
			if condition.Status.Status != membership.Organization.Status {
				continue
			}
		} else {
			// only status = "Active" data is returned by default
			if membership.Organization.Status != Active {
				continue
			}
		}

		orgs = append(orgs, &Organization{
			ID:     membership.Organization.ID,
			Name:   membership.Organization.Name,
			Status: membership.Organization.Status,
		})
	}

	log.Info(ctx, "get orgs by user success",
		log.String("userID", id),
		log.Any("orgs", orgs))

	return orgs, nil
}

func (s AmsOrganizationService) Name() string {
	return "ams_org_service"
}
