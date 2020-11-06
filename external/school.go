package external

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type SchoolServiceProvider interface {
	BatchGet(ctx context.Context, ids []string) ([]*School, error)
	GetByOrganizationID(ctx context.Context, organizationID string) ([]*School, error)
	GetByPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*School, error)
}

type School struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var (
	_amsSchoolService *AmsSchoolService
	_amsSchoolOnce    sync.Once
)

func GetSchoolServiceProvider() SchoolServiceProvider {
	_amsSchoolOnce.Do(func() {
		_amsSchoolService = &AmsSchoolService{}
	})

	return _amsSchoolService
}

type AmsSchoolService struct{}

func (s AmsSchoolService) BatchGet(ctx context.Context, ids []string) ([]*School, error) {
	if len(ids) == 0 {
		return []*School{}, nil
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
		log.Error(ctx, "get schools by ids failed", log.Strings("ids", ids))
		return nil, err
	}

	var queryAlias string
	schools := make([]*School, 0, len(data))
	for index := range ids {
		queryAlias = fmt.Sprintf("u%d", index)
		user, found := data[queryAlias]
		if !found {
			return nil, constant.ErrRecordNotFound
		}

		schools = append(schools, &School{
			ID:   user.UserID,
			Name: user.UserName,
		})
	}

	return schools, nil
}

func (s AmsSchoolService) GetByOrganizationID(ctx context.Context, organizationID string) ([]*School, error) {
	// TODO: add impl
	return nil, nil
}

func (s AmsSchoolService) GetByPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName) ([]*School, error) {
	// TODO: add impl
	return nil, nil
}
