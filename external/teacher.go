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

type TeacherServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*Teacher, error)
	Query(ctx context.Context, organizationID, keyword string) ([]*Teacher, error)
}

type Teacher struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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

func (s AmsTeacherService) BatchGet(ctx context.Context, ids []string) ([]*Teacher, error) {
	if len(ids) == 0 {
		return []*Teacher{}, nil
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
		log.Error(ctx, "get teachers by ids failed", log.Err(err), log.Strings("ids", ids))
		return nil, err
	}

	var queryAlias string
	teachers := make([]*Teacher, 0, len(data))
	for index := range ids {
		queryAlias = fmt.Sprintf("u%d", index)
		user, found := data[queryAlias]
		if !found {
			log.Error(ctx, "teachers not found", log.Strings("ids", ids), log.String("id", ids[index]))
			return nil, constant.ErrRecordNotFound
		}

		teachers = append(teachers, &Teacher{
			ID:   user.UserID,
			Name: user.UserName,
		})
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
