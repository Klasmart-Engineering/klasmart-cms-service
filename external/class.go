package external

import (
	"context"
	"text/template"

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
	return &mockClassService{}
}

type mockClassService struct{}

func (s mockClassService) BatchGet(ctx context.Context, ids []string) ([]*Class, error) {
	client := cl.NewClient(url)
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
		return nil, err
	}
	buf := buffer.Buffer{}
	temp.Execute(&buf, ids)
	//fmt.Println(buf.String())
	req := cl.NewRequest(buf.String())
	payload := make(map[string]*Class, len(ids))
	res := cl.Response{
		Data: &payload,
	}
	_, err = client.Run(ctx, req, &res)
	if err != nil {
		return nil, err
	}
	if len(res.Errors) > 0 {
		return nil, res.Errors
	}
	var classes []*Class
	for _, v := range payload {
		classes = append(classes, v)
	}
	return classes, nil
	//return GetMockData().Classes, nil
}

func (s mockClassService) GetStudents(ctx context.Context, classID string) ([]*Student, error) {
	client := cl.NewClient(url)
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
	_, err := client.Run(ctx, req, &res)
	if err != nil {
		return nil, err
	}
	if len(res.Errors) > 0 {
		return nil, res.Errors
	}
	return payload, nil
	//classes := GetMockData().Classes
	//for _, class := range classes {
	//	if class.ID == classID {
	//		return class.Students, nil
	//	}
	//}
	//return nil, nil
}
