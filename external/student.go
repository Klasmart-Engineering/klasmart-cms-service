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
		log.Error(ctx, "get students by ids failed", log.Strings("ids", ids))
		return nil, err
	}

	var queryAlias string
	students := make([]*Student, 0, len(data))
	for index := range ids {
		queryAlias = fmt.Sprintf("u%d", index)
		user, found := data[queryAlias]
		if !found {
			return nil, constant.ErrRecordNotFound
		}

		students = append(students, &Student{
			ID:   user.UserID,
			Name: user.UserName,
		})
	}

	return students, nil
}
