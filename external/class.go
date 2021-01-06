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
	GetByUserID(ctx context.Context, operator *entity.Operator, userID string) ([]*Class, error)
	GetByUserIDs(ctx context.Context, operator *entity.Operator, userIDs []string) (map[string][]*Class, error)
	GetByOrganizationIDs(ctx context.Context, operator *entity.Operator, orgIDs []string) (map[string][]*Class, error)
	GetBySchoolIDs(ctx context.Context, operator *entity.Operator, schoolIDs []string) (map[string][]*Class, error)
}

type Class struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type NullableClass struct {
	Class
	Valid bool `json:"-"`
}

func GetClassServiceProvider() ClassServiceProvider {
	return &AmsClassService{}
}

type AmsClassService struct{}

func (s AmsClassService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableClass, error) {
	raw := `query{
	{{range $i, $e := .}}
	index_{{$i}}: class(class_id: "{{$e}}"){
		id: class_id
    	name: class_name
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

func (s AmsClassService) GetByUserID(ctx context.Context, operator *entity.Operator, userID string) ([]*Class, error) {
	request := chlorine.NewRequest(`
	query($user_id: ID!){
		user(user_id: $user_id) {
			classesTeaching{
				id: class_id
				name: class_name
			}
			classesStudying{
				id: class_id
				name: class_name
			}
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("user_id", userID)

	data := &struct {
		User struct {
			ClassesTeaching []*Class `json:"classesTeaching"`
			ClassesStudying []*Class `json:"classesStudying"`
		} `json:"user"`
	}{}

	_, err := GetAmsClient().Run(ctx, request, &chlorine.Response{Data: data})
	if err != nil {
		log.Error(ctx, "query classes by user id failed", log.Err(err), log.String("userID", userID))
		return nil, err
	}

	classes := make([]*Class, 0)
	classes = append(classes, data.User.ClassesTeaching...)
	classes = append(classes, data.User.ClassesStudying...)

	log.Info(ctx, "get classes by user success",
		log.String("userID", userID),
		log.Any("classes", classes))

	return classes, nil
}

func (s AmsClassService) GetByUserIDs(ctx context.Context, operator *entity.Operator, userIDs []string) (map[string][]*Class, error) {
	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range userIDs {
		fmt.Fprintf(sb, "q%d: user(user_id: \"%s\") {\n", index, id)
		fmt.Fprintln(sb, "classesTeaching {id:class_id name:class_name}")
		fmt.Fprintln(sb, "classesStudying {id:class_id name:class_name}}")
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))

	data := map[string]*struct {
		ClassesTeaching []*Class `json:"classesTeaching"`
		ClassesStudying []*Class `json:"classesStudying"`
	}{}

	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get classes by users failed", log.Err(err), log.Strings("ids", userIDs))
		return nil, err
	}

	classes := make(map[string][]*Class, len(userIDs))
	var queryAlias string
	for index := range userIDs {
		queryAlias = fmt.Sprintf("q%d", index)
		query, found := data[queryAlias]
		if !found || query == nil {
			log.Error(ctx, "classes not found", log.Strings("userIDs", userIDs), log.String("id", userIDs[index]))
			return nil, constant.ErrRecordNotFound
		}

		classes[userIDs[index]] = make([]*Class, 0, len(query.ClassesTeaching)+len(query.ClassesStudying))
		if query.ClassesTeaching != nil {
			classes[userIDs[index]] = append(classes[userIDs[index]], query.ClassesTeaching...)
		}

		if query.ClassesStudying != nil {
			classes[userIDs[index]] = append(classes[userIDs[index]], query.ClassesStudying...)
		}
	}

	log.Info(ctx, "get classes by users success",
		log.Strings("userIDs", userIDs),
		log.Any("classes", classes))

	return classes, nil
}

func (s AmsClassService) GetByOrganizationIDs(ctx context.Context, operator *entity.Operator, organizationIDs []string) (map[string][]*Class, error) {
	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range organizationIDs {
		fmt.Fprintf(sb, "q%d: organization(organization_id: \"%s\") {classes{id: class_id name: class_name }}\n", index, id)
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
		queryAlias = fmt.Sprintf("q%d", index)
		org, found := data[queryAlias]
		if !found || org == nil {
			log.Error(ctx, "classes not found", log.String("id", organizationIDs[index]))
			return nil, constant.ErrRecordNotFound
		}

		if org.Classes != nil {
			classes[organizationIDs[index]] = org.Classes
		} else {
			classes[organizationIDs[index]] = []*Class{}
		}
	}

	log.Info(ctx, "get classes by org success",
		log.Strings("organizationIDs", organizationIDs),
		log.Any("classes", classes))

	return classes, nil
}

func (s AmsClassService) GetBySchoolIDs(ctx context.Context, operator *entity.Operator, schoolIDs []string) (map[string][]*Class, error) {
	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range schoolIDs {
		fmt.Fprintf(sb, "q%d: school(school_id: \"%s\") {classes{id: class_id name: class_name }}\n", index, id)
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
		queryAlias = fmt.Sprintf("q%d", index)
		org, found := data[queryAlias]
		if !found || org == nil {
			log.Error(ctx, "classes not found", log.Strings("schoolIDs", schoolIDs), log.String("id", schoolIDs[index]))
			return nil, constant.ErrRecordNotFound
		}

		if org.Classes != nil {
			classes[schoolIDs[index]] = org.Classes
		} else {
			classes[schoolIDs[index]] = []*Class{}
		}
	}

	log.Info(ctx, "get classes by schools success",
		log.Strings("schoolIDs", schoolIDs),
		log.Any("classes", classes))

	return classes, nil
}
