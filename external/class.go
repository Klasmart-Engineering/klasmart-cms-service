package external

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

	"go.uber.org/zap/buffer"
)

type ClassServiceProvider interface {
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableClass, error)
	BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*NullableClass, error)
	BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error)
	GetByUserID(ctx context.Context, operator *entity.Operator, userID string, options ...APOption) ([]*Class, error)
	GetByUserIDs(ctx context.Context, operator *entity.Operator, userIDs []string, options ...APOption) (map[string][]*Class, error)
	GetByOrganizationIDs(ctx context.Context, operator *entity.Operator, orgIDs []string, options ...APOption) (map[string][]*Class, error)
	GetBySchoolIDs(ctx context.Context, operator *entity.Operator, schoolIDs []string, options ...APOption) (map[string][]*Class, error)
	GetOnlyUnderOrgClasses(ctx context.Context, operator *entity.Operator, orgID string) ([]*NullableClass, error)
}

type Class struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Status APStatus `json:"status"`
}

type NullableClass struct {
	Class
	Valid bool `json:"-"`
}

func GetClassServiceProvider() ClassServiceProvider {
	return &AmsClassService{}
}

type AmsClassService struct{}

func (s AmsClassService) GetOnlyUnderOrgClasses(ctx context.Context, operator *entity.Operator, orgID string) ([]*NullableClass, error) {
	orgClassMap, err := s.GetByOrganizationIDs(ctx, operator, []string{orgID})
	if err != nil {
		log.Error(ctx, "GetClassServiceProvider.GetByOrganizationIDs error", log.Any("op", operator))
		return nil, err
	}
	orgClassList, ok := orgClassMap[orgID]
	if !ok || len(orgClassList) <= 0 {
		log.Info(ctx, "no classes under the organization", log.Any("op", operator))
		return nil, constant.ErrRecordNotFound
	}
	orgClassIDs := make([]string, len(orgClassList))
	for i, item := range orgClassList {
		orgClassIDs[i] = item.ID
	}
	classSchoolMap, err := GetSchoolServiceProvider().GetByClasses(ctx, operator, orgClassIDs)
	if err != nil {
		log.Error(ctx, "GetSchoolServiceProvider.GetByClasses error", log.Any("op", operator), log.Strings("orgClassIDs", orgClassIDs))
		return nil, err
	}

	underOrgClassIDs := make([]string, 0)
	for key, schools := range classSchoolMap {
		if len(schools) == 0 {
			underOrgClassIDs = append(underOrgClassIDs, key)
		}
	}
	classInfos, err := s.BatchGet(ctx, operator, underOrgClassIDs)
	if err != nil {
		log.Error(ctx, "GetClassServiceProvider.BatchGet error", log.Any("op", operator), log.Strings("underOrgClassIDs", underOrgClassIDs))
		return nil, err
	}

	return classInfos, nil
}

func (s AmsClassService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableClass, error) {
	if len(ids) == 0 {
		return []*NullableClass{}, nil
	}

	raw := `query{
	{{range $i, $e := .}}
	index_{{$i}}: class(class_id: "{{$e}}"){
		id: class_id
    	name: class_name
		status
		students{
			id: user_id
			name: user_name			
		}
  	}
	{{end}}
}`
	temp, err := template.New("Classes").Parse(raw)
	if err != nil {
		log.Error(ctx, "temp error", log.String("raw", raw), log.Err(err))
		return nil, err
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	buf := buffer.Buffer{}
	err = temp.Execute(&buf, _ids)
	if err != nil {
		log.Error(ctx, "temp execute failed", log.String("raw", raw), log.Err(err))
		return nil, err
	}
	req := chlorine.NewRequest(buf.String(), chlorine.ReqToken(operator.Token))
	payload := make(map[string]*Class, len(ids))
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
	var classes []*NullableClass
	for index := range ids {
		class := payload[fmt.Sprintf("index_%d", indexMapping[index])]
		if class == nil {
			classes = append(classes, &NullableClass{Valid: false})
		} else {
			classes = append(classes, &NullableClass{*class, true})
		}
	}

	log.Info(ctx, "get classes by ids success",
		log.Strings("ids", ids),
		log.Any("classes", classes))

	return classes, nil
}

func (s AmsClassService) BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*NullableClass, error) {
	classes, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]*NullableClass{}, err
	}

	dict := make(map[string]*NullableClass, len(classes))
	for _, class := range classes {
		dict[class.ID] = class
	}

	return dict, nil
}

func (s AmsClassService) BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error) {
	classes, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]string{}, err
	}

	dict := make(map[string]string, len(classes))
	for _, class := range classes {
		dict[class.ID] = class.Name
	}

	return dict, nil
}

func (s AmsClassService) GetByUserID(ctx context.Context, operator *entity.Operator, userID string, options ...APOption) ([]*Class, error) {
	classes, err := s.GetByUserIDs(ctx, operator, []string{userID}, options...)
	if err != nil {
		return nil, err
	}

	return classes[userID], nil
}

func (s AmsClassService) GetByUserIDs(ctx context.Context, operator *entity.Operator, userIDs []string, options ...APOption) (map[string][]*Class, error) {
	_userIDs := utils.SliceDeduplicationExcludeEmpty(userIDs)

	if len(_userIDs) == 0 {
		return map[string][]*Class{}, nil
	}

	classes := make(map[string][]*Class, len(_userIDs))

	total := len(_userIDs)
	pageSize := constant.AMSRequestUserClassPageSize
	pageCount := (total + pageSize - 1) / pageSize

	condition := NewCondition(options...)
	data := map[string]*struct {
		ClassesTeaching []*Class `json:"classesTeaching"`
		ClassesStudying []*Class `json:"classesStudying"`
	}{}

	cerr := make(chan error, pageCount)
	for i := 0; i < pageCount; i++ {
		go func(j int) {
			start := j * pageSize
			end := (j + 1) * pageSize
			if end >= total {
				end = total
			}
			pageUserIDs := _userIDs[start:end]

			sb := new(strings.Builder)
			sb.WriteString("query {")
			for index, id := range pageUserIDs {
				fmt.Fprintf(sb, "q%d: user(user_id: \"%s\") {\n", index, id)
				fmt.Fprintln(sb, "classesTeaching {id:class_id name:class_name status}")
				fmt.Fprintln(sb, "classesStudying {id:class_id name:class_name status}}")
			}
			sb.WriteString("}")

			request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))

			response := &chlorine.Response{
				Data: &data,
			}

			_, err := GetAmsClient().Run(ctx, request, response)
			if err != nil {
				log.Error(ctx, "get classes by users failed", log.Err(err), log.Strings("pageUserIDs", pageUserIDs))
				cerr <- err
				return
			}

			var queryAlias string
			for index,userID := range pageUserIDs {
				queryAlias = fmt.Sprintf("q%d", index)
				query, found := data[queryAlias]
				if !found || query == nil {
					log.Error(ctx, "classes not found", log.Strings("pageUserIDs", pageUserIDs), log.String("id", userID))
					cerr <- constant.ErrRecordNotFound
					return
				}

				allClasses := append(query.ClassesTeaching, query.ClassesStudying...)
				classes[userID] = make([]*Class, 0, len(allClasses))
				for _, class := range allClasses {
					if condition.Status.Valid {
						if condition.Status.Status != class.Status {
							continue
						}
					} else {
						// only status = "Active" data is returned by default
						if class.Status != Active {
							continue
						}
					}

					classes[userID] = append(classes[userID], class)
				}
			}
			cerr <- nil
		}(i)
	}

	for i := 0; i < pageCount; i++ {
		if err := <-cerr; err != nil {
			return nil, err
		}
	}

	log.Info(ctx, "get classes by users success",
		log.Strings("userIDs", userIDs),
		log.Any("classes", classes))

	return classes, nil
}

func (s AmsClassService) GetByOrganizationIDs(ctx context.Context, operator *entity.Operator, organizationIDs []string, options ...APOption) (map[string][]*Class, error) {
	if len(organizationIDs) == 0 {
		return map[string][]*Class{}, nil
	}

	condition := NewCondition(options...)

	_organizationIDs, indexMapping := utils.SliceDeduplicationMap(organizationIDs)

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range _organizationIDs {
		fmt.Fprintf(sb, "q%d: organization(organization_id: \"%s\") {classes{id: class_id name: class_name status}}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))

	data := map[string]*struct {
		Classes []*Class `json:"classes"`
	}{}

	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get classes by org ids failed", log.Err(err), log.Strings("ids", organizationIDs))
		return nil, err
	}
	log.Info(ctx, "GetByOrganizationIDs", log.Any("data", data))
	classes := make(map[string][]*Class, len(organizationIDs))
	var queryAlias string
	for index := range organizationIDs {
		queryAlias = fmt.Sprintf("q%d", indexMapping[index])
		org, found := data[queryAlias]
		if !found || org == nil {
			log.Error(ctx, "classes not found", log.String("id", organizationIDs[index]))
			return nil, constant.ErrRecordNotFound
		}

		classes[organizationIDs[index]] = make([]*Class, 0, len(org.Classes))
		for _, class := range org.Classes {
			if condition.Status.Valid {
				if condition.Status.Status != class.Status {
					continue
				}
			} else {
				// only status = "Active" data is returned by default
				if class.Status != Active {
					continue
				}
			}

			classes[organizationIDs[index]] = append(classes[organizationIDs[index]], class)
		}
	}

	log.Info(ctx, "get classes by org ids success",
		log.Strings("organizationIDs", organizationIDs),
		log.Any("classes", classes))

	return classes, nil
}

func (s AmsClassService) GetBySchoolIDs(ctx context.Context, operator *entity.Operator, schoolIDs []string, options ...APOption) (map[string][]*Class, error) {
	if len(schoolIDs) == 0 {
		return map[string][]*Class{}, nil
	}

	condition := NewCondition(options...)

	_schoolIDs, indexMapping := utils.SliceDeduplicationMap(schoolIDs)

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range _schoolIDs {
		fmt.Fprintf(sb, "q%d: school(school_id: \"%s\") {classes{id: class_id name: class_name status}}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))

	data := map[string]*struct {
		Classes []*Class `json:"classes"`
	}{}

	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get classes by schools failed", log.Err(err), log.Strings("ids", schoolIDs))
		return nil, err
	}

	classes := make(map[string][]*Class, len(schoolIDs))
	var queryAlias string
	for index := range schoolIDs {
		queryAlias = fmt.Sprintf("q%d", indexMapping[index])
		org, found := data[queryAlias]
		if !found || org == nil {
			log.Error(ctx, "classes not found", log.Strings("schoolIDs", schoolIDs), log.String("id", schoolIDs[index]))
			return nil, constant.ErrRecordNotFound
		}

		classes[schoolIDs[index]] = make([]*Class, 0, len(org.Classes))
		for _, class := range org.Classes {
			if condition.Status.Valid {
				if condition.Status.Status != class.Status {
					continue
				}
			} else {
				// only status = "Active" data is returned by default
				if class.Status != Active {
					continue
				}
			}

			classes[schoolIDs[index]] = append(classes[schoolIDs[index]], class)
		}
	}

	log.Info(ctx, "get classes by schools success",
		log.Strings("schoolIDs", schoolIDs),
		log.Any("classes", classes))

	return classes, nil
}
