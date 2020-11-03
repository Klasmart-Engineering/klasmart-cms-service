package external

import (
	"context"
	"text/template"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"go.uber.org/zap/buffer"

	cl "gitlab.badanamu.com.cn/calmisland/chlorine"
)

type ClassServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Class, error)
	GetStudents(ctx context.Context, classID string) ([]*Student, error)
}

type Class struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Students []*Student `json:"students"`
}

func GetClassServiceProvider() ClassServiceProvider {
	//return &mockClassService{}
	return &graphqlClassService{}
}

type graphqlClassService struct{}

func (s graphqlClassService) BatchGet(ctx context.Context, ids []string) ([]*Class, error) {
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
	req := cl.NewRequest(buf.String())
	payload := make(map[string]*Class, len(ids))
	res := cl.Response{
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

func (s graphqlClassService) GetStudents(ctx context.Context, classID string) ([]*Student, error) {
	q := `query ($classID: ID!){
	class(class_id: $classID){
		students{
			id: user_id
			name: user_name
		}
  	}
}`
	req := cl.NewRequest(q)
	req.Var("classID", classID)
	var payload []*Student
	res := cl.Response{
		Data: &struct {
			Class struct{ Students *[]*Student }
		}{Class: struct{ Students *[]*Student }{Students: &payload}},
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

type mockClassService struct{}

func (s mockClassService) BatchGet(ctx context.Context, ids []string) ([]*Class, error) {
	return GetMockData().Classes, nil
}

func (s mockClassService) GetStudents(ctx context.Context, classID string) ([]*Student, error) {
	classes := GetMockData().Classes
	for _, class := range classes {
		if class.ID == classID {
			return class.Students, nil
		}
	}
	return nil, nil
}
