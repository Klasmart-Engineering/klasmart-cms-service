package external

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"go.uber.org/zap/buffer"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	cl "gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type OrganizationServiceProvider interface {
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableOrganization, error)
	BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*NullableOrganization, error)
	BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error)
	GetByClasses(ctx context.Context, operator *entity.Operator, classIDs []string, options ...APOption) (map[string]*Organization, error)
	GetNameByOrganizationOrSchool(ctx context.Context, operator *entity.Operator, id []string) ([]string, error)
	GetNameMapByOrganizationOrSchool(ctx context.Context, operator *entity.Operator, id []string) (map[string]string, error)
	GetByPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName, options ...APOption) ([]*Organization, error)
	GetByUserID(ctx context.Context, operator *entity.Operator, id string, options ...APOption) ([]*Organization, error)
	OrgAthPrgSjtCtgSubCtgGrdAge(ctx context.Context, op *entity.Operator, orgIDs, athIDs, prgIDs, subIDs, ctgIDs, subCtgIDs, grdIDs, ageIDs []string) (
		organization, author, program, subject, category, subCategory, grade, age map[string]string, err error)
}

type Organization struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Status APStatus `json:"status"`
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
	if len(ids) == 0 {
		return []*NullableOrganization{}, nil
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
	_, err := GetAmsClient().Run(ctx, req, &res)
	if err != nil {
		log.Error(ctx, "Run error", log.String("q", q), log.Any("res", res), log.Err(err))
		return nil, err
	}
	if len(res.Errors) > 0 {
		log.Error(ctx, "Res error", log.String("q", q), log.Any("res", res), log.Err(res.Errors))
		return nil, res.Errors
	}
	nullableOrganizations := make([]*NullableOrganization, len(ids))
	for index := range ids {
		if payload[indexMapping[index]] == nil {
			nullableOrganizations[index] = &NullableOrganization{Valid: false}
		} else {
			nullableOrganizations[index] = &NullableOrganization{*payload[indexMapping[index]], true}
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
		dict[organization.ID] = organization.Name
	}

	return dict, nil
}

func (s AmsOrganizationService) GetByClasses(ctx context.Context, operator *entity.Operator, classIDs []string, options ...APOption) (map[string]*Organization, error) {
	if len(classIDs) == 0 {
		return map[string]*Organization{}, nil
	}

	condition := NewCondition(options...)

	_classIDs, indexMapping := utils.SliceDeduplicationMap(classIDs)

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range _classIDs {
		fmt.Fprintf(sb, `q%d: class(class_id: "%s") {organization{id:organization_id name:organization_name status}}`, index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))

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
	err = temp.Execute(&buf, ids)
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

func (s AmsOrganizationService) OrgAthPrgSjtCtgSubCtgGrdAge(ctx context.Context, op *entity.Operator, orgIDs, athIDs, prgIDs, subIDs, ctgIDs, subCtgIDs, grdIDs, ageIDs []string) (
	organization, author, program, subject, category, subCategory, grade, age map[string]string, err error) {

	_orgIDs := utils.SliceDeduplication(orgIDs)
	_athIDs := utils.SliceDeduplication(athIDs)
	_prgIDs := utils.SliceDeduplication(prgIDs)
	_subIDs := utils.SliceDeduplication(subIDs)
	_ctgIDs := utils.SliceDeduplication(ctgIDs)
	_subCtgIDs := utils.SliceDeduplication(subCtgIDs)
	_grdIDs := utils.SliceDeduplication(grdIDs)
	_ageIDs := utils.SliceDeduplication(ageIDs)

	organization = make(map[string]string, len(_orgIDs))
	author = make(map[string]string, len(_athIDs))
	program = make(map[string]string, len(_prgIDs))
	subject = make(map[string]string, len(_subIDs))
	category = make(map[string]string, len(_ctgIDs))
	subCategory = make(map[string]string, len(_subCtgIDs))
	grade = make(map[string]string, len(_grdIDs))
	age = make(map[string]string, len(_ageIDs))

	if len(orgIDs) == 0 && len(athIDs) == 0 && len(prgIDs) == 0 && len(subIDs) == 0 && len(ctgIDs) == 0 && len(subCtgIDs) == 0 && len(grdIDs) == 0 && len(ageIDs) == 0 {
		return
	}

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range _orgIDs {
		fmt.Fprintf(sb, "org%d: organization(organization_id: \"%s\") {id:organization_id name:organization_name}\n", index, id)
	}

	for index, id := range _athIDs {
		fmt.Fprintf(sb, "ath%d: user(user_id: \"%s\") {id:user_id name:full_name}\n", index, id)
	}

	for index, id := range _prgIDs {
		fmt.Fprintf(sb, "prg%d: program(id: \"%s\") {id name}\n", index, id)
	}

	for index, id := range _subIDs {
		fmt.Fprintf(sb, "sbj%d: subject(id: \"%s\") {id name}\n", index, id)
	}

	for index, id := range _ctgIDs {
		fmt.Fprintf(sb, "ctg%d: category(id: \"%s\") {id name}\n", index, id)
	}

	for index, id := range _subCtgIDs {
		fmt.Fprintf(sb, "sbc%d: subcategory(id: \"%s\") {id name}\n", index, id)
	}

	for index, id := range _grdIDs {
		fmt.Fprintf(sb, "grd%d: grade(id: \"%s\") {id name}\n", index, id)
	}

	for index, id := range _ageIDs {
		fmt.Fprintf(sb, "age%d: age_range(id: \"%s\") {id name}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(op.Token))

	type Value struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	data := map[string]*Value{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err = GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "OrgAthPrgSjtCtgSubCtgGrdAge: Run failed",
			log.Err(err),
			log.Any("operator", op),
			log.Strings("program", prgIDs))
		return
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "OrgAthPrgSjtCtgSubCtgGrdAge: response error",
			log.Err(response.Errors),
			log.Any("operator", op),
			log.Strings("program", prgIDs))
		return
	}

	for k, v := range data {
		switch k[:3] {
		case "org":
			if v != nil {
				organization[v.ID] = v.Name
			}
		case "ath":
			if v != nil {
				author[v.ID] = v.Name
			}
		case "prg":
			if v != nil {
				program[v.ID] = v.Name
			}
		case "sbj":
			if v != nil {
				subject[v.ID] = v.Name
			}
		case "ctg":
			if v != nil {
				category[v.ID] = v.Name
			}
		case "sbc":
			if v != nil {
				subCategory[v.ID] = v.Name
			}
		case "grd":
			if v != nil {
				grade[v.ID] = v.Name
			}
		case "age":
			if v != nil {
				age[v.ID] = v.Name
			}
		}
	}
	log.Info(ctx, "OrgAthPrgSjtCtgSubCtgGrdAge: success",
		log.Any("data", data))
	return
}
