package external

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"go.uber.org/zap/buffer"
)

type ClassServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*NullableClass, error)
	GetByUserID(ctx context.Context, userID string) ([]*Class, error)
	GetByOrganizationIDs(ctx context.Context, orgIDs []string) (map[string][]*Class, error)
	GetBySchoolIDs(ctx context.Context, schoolIDs []string) (map[string][]*Class, error)
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

func (s AmsClassService) BatchGet(ctx context.Context, ids []string) ([]*NullableClass, error) {
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
	buf := buffer.Buffer{}
	err = temp.Execute(&buf, ids)
	if err != nil {
		log.Error(ctx, "temp execute failed", log.String("raw", raw), log.Err(err))
		return nil, err
	}
	req := chlorine.NewRequest(buf.String())
	payload := make(map[string]*Class, len(ids))
	res := chlorine.Response{
		Data: &payload,
	}

	_, err = GetChlorine().Run(ctx, req, &res)
	if err != nil {
		log.Error(ctx, "Run error", log.String("q", buf.String()), log.Any("res", res), log.Err(err))
		return nil, err
	}
	if len(res.Errors) > 0 {
		log.Error(ctx, "Res error", log.String("q", buf.String()), log.Any("res", res), log.Err(res.Errors))
		return nil, res.Errors
	}
	var classes []*NullableClass
	for _, v := range payload {
		if v == nil {
			classes = append(classes, &NullableClass{Valid: false})
		} else {
			classes = append(classes, &NullableClass{*v, true})
		}
	}

	log.Info(ctx, "get classes by ids success",
		log.Strings("ids", ids),
		log.Any("classes", classes))

	return classes, nil
}

func (s AmsClassService) GetByUserID(ctx context.Context, userID string) ([]*Class, error) {
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
	}`)
	request.Var("user_id", userID)

	data := &struct {
		User struct {
			ClassesTeaching []*Class `json:"classesTeaching"`
			ClassesStudying []*Class `json:"classesStudying"`
		} `json:"user"`
	}{}

	_, err := GetChlorine().Run(ctx, request, &chlorine.Response{Data: data})
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

func (s AmsClassService) GetByOrganizationIDs(ctx context.Context, organizationIDs []string) (map[string][]*Class, error) {
	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range organizationIDs {
		fmt.Fprintf(sb, "q%d: organization(organization_id: \"%s\") {classes{id: class_id name: class_name }}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String())

	data := map[string]*struct {
		Classes []*Class `json:"classes"`
	}{}

	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
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

func (s AmsClassService) GetBySchoolIDs(ctx context.Context, schoolIDs []string) (map[string][]*Class, error) {
	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range schoolIDs {
		fmt.Fprintf(sb, "q%d: school(school_id: \"%s\") {classes{id: class_id name: class_name }}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String())

	data := map[string]*struct {
		Classes []*Class `json:"classes"`
	}{}

	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get classes by org ids failed", log.Err(err), log.Strings("ids", schoolIDs))
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
