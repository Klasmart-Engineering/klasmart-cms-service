package external

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type StudentServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Student, error)
	GetByClassID(ctx context.Context, classID string) ([]*Student, error)
}

type Student struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var (
	_amsStudentService *AmsStudentService
	_amsStudentOnce    sync.Once
)

func GetStudentServiceProvider() StudentServiceProvider {
	_amsStudentOnce.Do(func() {
		_amsStudentService = &AmsStudentService{}
	})

	return _amsStudentService
}

type AmsStudentService struct{}

func (s AmsStudentService) BatchGet(ctx context.Context, ids []string) ([]*Student, error) {
	if len(ids) == 0 {
		return []*Student{}, nil
	}

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range ids {
		fmt.Fprintf(sb, "u%d: user(user_id: \"%s\") {user_id user_name}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String())

	data := map[string]struct {
		UserID   string `json:"user_id"`
		UserName string `json:"user_name"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get students by ids failed", log.Err(err), log.Strings("ids", ids))
		return nil, err
	}

	var queryAlias string
	students := make([]*Student, 0, len(data))
	for index := range ids {
		queryAlias = fmt.Sprintf("u%d", index)
		user, found := data[queryAlias]
		if !found {
			log.Error(ctx, "students not found", log.Strings("ids", ids), log.String("id", ids[index]))
			return nil, constant.ErrRecordNotFound
		}

		students = append(students, &Student{
			ID:   user.UserID,
			Name: user.UserName,
		})
	}

	return students, nil
}

func (s AmsStudentService) GetByClassID(ctx context.Context, classID string) ([]*Student, error) {
	q := `query ($classID: ID!){
	class(class_id: $classID){
		students{
			id: user_id
			name: user_name
		}
  	}
}`
	req := chlorine.NewRequest(q)
	req.Var("classID", classID)
	var payload []*Student
	res := chlorine.Response{
		Data: &struct {
			Class struct {
				Students *[]*Student `json:"students"`
			} `json:"class"`
		}{Class: struct {
			Students *[]*Student `json:"students"`
		}{Students: &payload}},
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
