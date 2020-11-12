package external

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type TeacherServiceProvider interface {
	Get(ctx context.Context, id string) (*Teacher, error)
	BatchGet(ctx context.Context, ids []string) ([]*NullableTeacher, error)
	Query(ctx context.Context, organizationID, keyword string) ([]*Teacher, error)
}

type Teacher struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type NullableTeacher struct {
	Valid bool `json:"-"`
	*Teacher
}

var (
	_amsTeacherService *AmsTeacherService
	_amsTeacherOnce    sync.Once
)

func GetTeacherServiceProvider() TeacherServiceProvider {
	_amsTeacherOnce.Do(func() {
		_amsTeacherService = &AmsTeacherService{}
	})

	return _amsTeacherService
}

type AmsTeacherService struct{}

func (s AmsTeacherService) Get(ctx context.Context, id string) (*Teacher, error) {
	teachers, err := s.BatchGet(ctx, []string{id})
	if err != nil {
		return nil, err
	}

	if teachers[0].Valid {
		return nil, constant.ErrRecordNotFound
	}

	return teachers[0].Teacher, nil
}

func (s AmsTeacherService) BatchGet(ctx context.Context, ids []string) ([]*NullableTeacher, error) {
	if len(ids) == 0 {
		return []*NullableTeacher{}, nil
	}

	users, err := GetUserServiceProvider().BatchGet(ctx, ids)
	if err != nil {
		return nil, err
	}

	teachers := make([]*NullableTeacher, len(users))
	for index, user := range users {
		teachers[index] = &NullableTeacher{
			Valid: user.Valid,
			Teacher: &Teacher{
				ID:   user.User.ID,
				Name: user.User.Name,
			},
		}
	}

	return teachers, nil
}

func (s AmsTeacherService) Query(ctx context.Context, organizationID, keyword string) ([]*Teacher, error) {
	request := chlorine.NewRequest(`
	query(
		$organization_id: ID!
		$keyword: String!
	) {
		organization(organization_id: $organization_id) {
			findMembers(search_query: $keyword) {
				user{
					user_id
					user_name
				}
			}
		}
	}`)
	request.Var("organization_id", organizationID)
	request.Var("keyword", keyword)

	data := &struct {
		Organization struct {
			FindMembers []struct {
				User struct {
					UserID   string `json:"user_id"`
					UserName string `json:"user_name"`
				} `json:"user"`
			} `json:"findMembers"`
		} `json:"organization"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetChlorine().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get teachers by permission failed",
			log.Err(err),
			log.String("organizationID", organizationID),
			log.String("keyword", keyword))
		return nil, err
	}

	teachers := make([]*Teacher, 0, len(data.Organization.FindMembers))
	for _, member := range data.Organization.FindMembers {
		teachers = append(teachers, &Teacher{
			ID:   member.User.UserID,
			Name: member.User.UserName,
		})
	}

	return teachers, nil
}
