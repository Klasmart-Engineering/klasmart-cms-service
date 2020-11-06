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
	BatchGet(ctx context.Context, ids []string) ([]*Class, error)
	GetByUserID(ctx context.Context, userID string) ([]*Class, error)
	GetByOrganizationIDs(ctx context.Context, orgIDs []string) (map[string][]*Class, error)
	GetBySchoolIDs(ctx context.Context, schoolIDs []string) (map[string][]*Class, error)
}

type Class struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetClassServiceProvider() ClassServiceProvider {
	return &AmsClassService{}
}

type AmsClassService struct{}

func (s AmsClassService) BatchGet(ctx context.Context, ids []string) ([]*Class, error) {
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
	var classes []*Class
	for _, v := range payload {
		classes = append(classes, v)
	}
	return classes, nil
}

func (s AmsClassService) GetByUserID(ctx context.Context, userID string) ([]*Class, error) {
	request := chlorine.NewRequest(`
	query($user_id: ID!){
		user(user_id: $user_id) {
			classesTeaching{
				class_id
				class_name
			}
			classesStudying{
				class_id
				class_name
			}
		}
	}`)
	request.Var("user_id", userID)

	data := &struct {
		User struct {
			ClassesTeaching []struct {
				ClassID   string `json:"class_id"`
				ClassName string `json:"class_name"`
			} `json:"classesTeaching"`
			ClassesStudying []struct {
				ClassID   string `json:"class_id"`
				ClassName string `json:"class_name"`
			} `json:"classesStudying"`
		} `json:"user"`
	}{}

	_, err := GetChlorine().Run(ctx, request, &chlorine.Response{Data: data})
	if err != nil {
		log.Error(ctx, "query classes by user id failed", log.String("userID", userID))
		return nil, err
	}

	classes := make([]*Class, 0, len(data.User.ClassesTeaching)+len(data.User.ClassesStudying))
	for _, class := range data.User.ClassesTeaching {
		classes = append(classes, &Class{
			ID:   class.ClassID,
			Name: class.ClassName,
		})
	}

	for _, class := range data.User.ClassesStudying {
		classes = append(classes, &Class{
			ID:   class.ClassID,
			Name: class.ClassName,
		})
	}

	return classes, nil
}

func (s AmsClassService) GetByOrganizationIDs(ctx context.Context, organizationIDs []string) (map[string][]*Class, error) {
	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range organizationIDs {
		fmt.Fprintf(sb, "q%d: organization(organization_id: \"%s\") {classes{class_id class_name }}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String())

	data := map[string]*struct {
		Classes []struct {
			ClassID   string `json:"class_id"`
			ClassName string `json:"class_name"`
		} `json:"classes"`
	}{}

	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get users by ids failed", log.Strings("ids", organizationIDs))
		return nil, err
	}

	orgs := make(map[string][]*Class, len(organizationIDs))
	var queryAlias string
	for index := range organizationIDs {
		queryAlias = fmt.Sprintf("q%d", index)
		org, found := data[queryAlias]
		if !found || org == nil {
			log.Error(ctx, "user not found", log.String("id", organizationIDs[index]))
			return nil, constant.ErrRecordNotFound
		}

		orgs[organizationIDs[index]] = make([]*Class, 0, len(org.Classes))
		for _, class := range org.Classes {
			orgs[organizationIDs[index]] = append(orgs[organizationIDs[index]], &Class{
				ID:   class.ClassID,
				Name: class.ClassName,
			})
		}
	}

	return orgs, nil
}

func (s AmsClassService) GetBySchoolIDs(ctx context.Context, schoolIDs []string) (map[string][]*Class, error) {
	// TODO
	return nil, nil
}
