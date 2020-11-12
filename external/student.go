package external

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type StudentServiceProvider interface {
	Get(ctx context.Context, id string) (*Student, error)
	BatchGet(ctx context.Context, ids []string) ([]*NullableStudent, error)
	GetByClassID(ctx context.Context, classID string) ([]*Student, error)
}

type Student struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type NullableStudent struct {
	Valid bool `json:"-"`
	*Student
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

func (s AmsStudentService) Get(ctx context.Context, id string) (*Student, error) {
	students, err := s.BatchGet(ctx, []string{id})
	if err != nil {
		return nil, err
	}

	if students[0].Valid {
		return nil, constant.ErrRecordNotFound
	}

	return students[0].Student, nil
}

func (s AmsStudentService) BatchGet(ctx context.Context, ids []string) ([]*NullableStudent, error) {
	if len(ids) == 0 {
		return []*NullableStudent{}, nil
	}

	users, err := GetUserServiceProvider().BatchGet(ctx, ids)
	if err != nil {
		return nil, err
	}

	students := make([]*NullableStudent, len(users))
	for index, user := range users {
		students[index] = &NullableStudent{
			Valid: user.Valid,
			Student: &Student{
				ID:   user.User.ID,
				Name: user.User.Name,
			},
		}
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
