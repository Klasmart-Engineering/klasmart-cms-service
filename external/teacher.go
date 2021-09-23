package external

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"
	"strings"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type TeacherServiceProvider interface {
	cache.IDataSource
	Get(ctx context.Context, operator *entity.Operator, id string) (*Teacher, error)
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableTeacher, error)
	BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*NullableTeacher, error)
	BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error)
	GetByOrganization(ctx context.Context, operator *entity.Operator, organizationID string) ([]*Teacher, error)
	GetByOrganizations(ctx context.Context, operator *entity.Operator, organizationIDs []string) (map[string][]*Teacher, error)
	GetBySchool(ctx context.Context, operator *entity.Operator, schoolID string) ([]*Teacher, error)
	GetBySchools(ctx context.Context, operator *entity.Operator, schoolIDs []string) (map[string][]*Teacher, error)
	GetByClasses(ctx context.Context, operator *entity.Operator, classIDs []string) (map[string][]*Teacher, error)
	Query(ctx context.Context, operator *entity.Operator, organizationID, keyword string) ([]*Teacher, error)
	FilterByPermission(ctx context.Context, operator *entity.Operator, userIDs []string, permissionName PermissionName) ([]string, error)
}

type Teacher struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type NullableTeacher struct {
	Valid bool `json:"-"`
	*Teacher
}

func (n *NullableTeacher) StringID() string {
	return n.Teacher.ID
}
func (n *NullableTeacher) RelatedIDs() []*cache.RelatedEntity {
	return nil
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

func (s AmsTeacherService) Get(ctx context.Context, operator *entity.Operator, id string) (*Teacher, error) {
	teachers, err := s.BatchGet(ctx, operator, []string{id})
	if err != nil {
		return nil, err
	}

	if teachers[0].Teacher == nil || !teachers[0].Valid {
		return nil, constant.ErrRecordNotFound
	}

	return teachers[0].Teacher, nil
}

func (s AmsTeacherService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableTeacher, error) {
	if len(ids) == 0 {
		return []*NullableTeacher{}, nil
	}
	res := make([]*NullableTeacher, 0, len(ids))
	err := cache.GetPassiveCacheRefresher().BatchGet(ctx, s.Name(), ids, &res, operator)
	if err != nil {
		return nil, err
	}

	return res, nil
}
func (s AmsTeacherService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	operator, err := optionsWithOperator(ctx, options...)
	if err != nil {
		return nil, err
	}

	users, err := GetUserServiceProvider().BatchGet(ctx, operator, ids)
	if err != nil {
		return nil, err
	}

	teachers := make([]cache.Object, len(users))
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

func (s AmsTeacherService) BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*NullableTeacher, error) {
	teachers, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]*NullableTeacher{}, err
	}

	dict := make(map[string]*NullableTeacher, len(teachers))
	for _, teacher := range teachers {
		if teacher.Teacher == nil || !teacher.Valid {
			continue
		}
		dict[teacher.ID] = teacher
	}

	return dict, nil
}

func (s AmsTeacherService) BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error) {
	return GetUserServiceProvider().BatchGetNameMap(ctx, operator, ids)
}

func (s AmsTeacherService) GetByOrganization(ctx context.Context, operator *entity.Operator, organizationID string) ([]*Teacher, error) {
	request := chlorine.NewRequest(`
	query ($organization_id: ID!) {
		organization(organization_id: $organization_id) {
			classes{
				teachers{
					id: user_id
					name: user_name
				}
			}    
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("organization_id", organizationID)

	data := &struct {
		Organization struct {
			Classes []struct {
				Teachers []*Teacher `json:"teachers"`
			} `json:"classes"`
		} `json:"organization"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get teachers by org failed",
			log.Err(err),
			log.String("organizationID", organizationID))
		return nil, err
	}

	teachers := make([]*Teacher, 0, len(data.Organization.Classes))
	for _, class := range data.Organization.Classes {
		teachers = append(teachers, class.Teachers...)
	}

	log.Info(ctx, "get teachers by org success",
		log.String("organizationID", organizationID),
		log.Any("teachers", teachers))

	return teachers, nil
}

//TODO:No Test Program
func (s AmsTeacherService) GetByOrganizations(ctx context.Context, operator *entity.Operator, organizationIDs []string) (map[string][]*Teacher, error) {
	if len(organizationIDs) == 0 {
		return map[string][]*Teacher{}, nil
	}

	sb := new(strings.Builder)
	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$organization_id_", ": ID!", len(organizationIDs)))
	for index := range organizationIDs {
		fmt.Fprintf(sb, "q%d: organization(organization_id: $organization_id_%d) {classes{teachers{id:user_id name:user_name}}}\n", index, index)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	for index, id := range organizationIDs {
		request.Var(fmt.Sprintf("organization_id_%d", index), id)
	}

	data := map[string]*struct {
		Classes []struct {
			Teachers []*Teacher `json:"teachers"`
		} `json:"classes"`
	}{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get users by organization ids failed", log.Err(err), log.Strings("ids", organizationIDs))
		return nil, err
	}

	teachers := make(map[string][]*Teacher, len(organizationIDs))
	for index, organizationID := range organizationIDs {
		classes := data[fmt.Sprintf("q%d", index)]
		if classes == nil {
			continue
		}

		for _, class := range classes.Classes {
			teachers[organizationID] = append(teachers[organizationID], class.Teachers...)
		}
	}

	log.Info(ctx, "get teachers by orgs success",
		log.Strings("organizationIDs", organizationIDs),
		log.Any("teachers", teachers))

	return teachers, nil
}

func (s AmsTeacherService) GetBySchool(ctx context.Context, operator *entity.Operator, schoolID string) ([]*Teacher, error) {
	request := chlorine.NewRequest(`
	query ($school_id: ID!) {
		school(school_id: $school_id) {
			classes{
				teachers{
					id: user_id
					name: user_name
				}
			}    
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("school_id", schoolID)

	data := &struct {
		School struct {
			Classes []struct {
				Teachers []*Teacher `json:"teachers"`
			} `json:"classes"`
		} `json:"school"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get teachers by school failed",
			log.Err(err),
			log.String("schoolID", schoolID))
		return nil, err
	}

	teachers := make([]*Teacher, 0, len(data.School.Classes))
	for _, class := range data.School.Classes {
		teachers = append(teachers, class.Teachers...)
	}

	log.Info(ctx, "get teachers by school success",
		log.String("schoolID", schoolID),
		log.Any("teachers", teachers))

	return teachers, nil
}

//TODO:No Test Program
func (s AmsTeacherService) GetBySchools(ctx context.Context, operator *entity.Operator, schoolIDs []string) (map[string][]*Teacher, error) {
	if len(schoolIDs) == 0 {
		return map[string][]*Teacher{}, nil
	}

	sb := new(strings.Builder)
	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$school_id_", ": ID!", len(schoolIDs)))
	for index := range schoolIDs {
		fmt.Fprintf(sb, "q%d: school(school_id: $school_id_%d) {classes{teachers{id:user_id name:user_name}}}\n", index, index)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	for index, id := range schoolIDs {
		request.Var(fmt.Sprintf("school_id_%d", index), id)
	}

	data := map[string]*struct {
		Classes []struct {
			Teachers []*Teacher `json:"teachers"`
		} `json:"classes"`
	}{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get users by school ids failed", log.Err(err), log.Strings("ids", schoolIDs))
		return nil, err
	}

	teachers := make(map[string][]*Teacher, len(schoolIDs))
	for index, schoolID := range schoolIDs {
		classes := data[fmt.Sprintf("q%d", index)]
		if classes == nil {
			continue
		}

		for _, class := range classes.Classes {
			teachers[schoolID] = append(teachers[schoolID], class.Teachers...)
		}
	}

	log.Info(ctx, "get teachers by schools success",
		log.Strings("schoolIDs", schoolIDs),
		log.Any("teachers", teachers))

	return teachers, nil
}

//TODO:No Test Program
func (s AmsTeacherService) GetByClasses(ctx context.Context, operator *entity.Operator, classIDs []string) (map[string][]*Teacher, error) {
	if len(classIDs) == 0 {
		return map[string][]*Teacher{}, nil
	}

	sb := new(strings.Builder)
	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$class_id_", ": ID!", len(classIDs)))
	for index := range classIDs {
		fmt.Fprintf(sb, "q%d: class(class_id: $class_id_%d) {teachers{id:user_id name:user_name}}\n", index, index)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	for index, id := range classIDs {
		request.Var(fmt.Sprintf("class_id_%d", index), id)
	}

	data := map[string]*struct {
		Teachers []*Teacher `json:"teachers"`
	}{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get users by class ids failed", log.Err(err), log.Strings("ids", classIDs))
		return nil, err
	}

	teachers := make(map[string][]*Teacher, len(classIDs))
	for index, classID := range classIDs {
		query, found := data[fmt.Sprintf("q%d", index)]
		if !found || query == nil {
			log.Warn(ctx, "classes not found", log.Strings("classIDs", classIDs), log.String("id", classIDs[index]))
			continue
		}

		teachers[classID] = append(teachers[classID], query.Teachers...)
	}

	log.Info(ctx, "get teachers by classes success",
		log.Strings("classIDs", classIDs),
		log.Any("teachers", teachers))

	return teachers, nil
}

func (s AmsTeacherService) Query(ctx context.Context, operator *entity.Operator, organizationID, keyword string) ([]*Teacher, error) {
	users, err := GetUserServiceProvider().Query(ctx, operator, organizationID, keyword)
	if err != nil {
		return nil, err
	}

	teachers := make([]*Teacher, len(users))
	for index, user := range users {
		teachers[index] = &Teacher{
			ID:   user.ID,
			Name: user.Name,
		}
	}

	return teachers, nil
}

func (s AmsTeacherService) FilterByPermission(ctx context.Context, operator *entity.Operator, userIDs []string, permissionName PermissionName) ([]string, error) {
	return GetUserServiceProvider().FilterByPermission(ctx, operator, userIDs, permissionName)
}

func (s AmsTeacherService) Name() string {
	return "ams_teacher_service"
}
