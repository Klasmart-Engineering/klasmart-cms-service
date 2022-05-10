package external

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/KL-Engineering/kidsloop-cache/cache"

	"github.com/KL-Engineering/chlorine"
	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

type StudentServiceProvider interface {
	cache.IDataSource
	Get(ctx context.Context, operator *entity.Operator, id string) (*Student, error)
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableStudent, error)
	BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*NullableStudent, error)
	BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error)
	GetByClassID(ctx context.Context, operator *entity.Operator, classID string) ([]*Student, error)
	GetByClassIDs(ctx context.Context, operator *entity.Operator, classIDs []string) (map[string][]*Student, error)
	Query(ctx context.Context, operator *entity.Operator, organizationID, keyword string) ([]*Student, error)
	FilterByPermission(ctx context.Context, operator *entity.Operator, userIDs []string, permissionName PermissionName) ([]string, error)
}

type Student struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type NullableStudent struct {
	Valid bool   `json:"valid"`
	StrID string `json:"str_id"`
	*Student
}

func (n *NullableStudent) StringID() string {
	return n.StrID
}
func (n *NullableStudent) RelatedIDs() []*cache.RelatedEntity {
	return nil
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

func (s AmsStudentService) Get(ctx context.Context, operator *entity.Operator, id string) (*Student, error) {
	students, err := s.BatchGet(ctx, operator, []string{id})
	if err != nil {
		return nil, err
	}

	if students[0].Student == nil || !students[0].Valid {
		return nil, constant.ErrRecordNotFound
	}

	return students[0].Student, nil
}

func (s AmsStudentService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableStudent, error) {
	log.Info(ctx, "Doing BatchGet student",
		log.Strings("ids", ids))
	if len(ids) == 0 {
		return []*NullableStudent{}, nil
	}

	students, err := s.QueryByIDs(ctx, ids, operator)
	if err != nil {
		return nil, err
	}
	res := make([]*NullableStudent, 0, len(ids))
	for i := range students {
		res = append(res, students[i].(*NullableStudent))
	}
	return res, nil
}

func (s AmsStudentService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
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

	students := make([]cache.Object, len(users))
	for index, user := range users {
		student := &NullableStudent{
			Valid: user.Valid,
			StrID: user.ID,
		}
		if user.Valid {
			student.Student = &Student{
				ID:   user.User.ID,
				Name: user.User.Name,
			}
		}
		students[index] = student
	}

	return students, nil
}

func (s AmsStudentService) BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*NullableStudent, error) {
	students, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]*NullableStudent{}, err
	}
	log.Info(ctx, "BatchGetMap.BatchGet students success",
		log.Strings("ids", ids),
		log.Any("students", students))
	dict := make(map[string]*NullableStudent, len(students))
	for _, student := range students {
		if students[0].Student == nil || !student.Valid {
			continue
		}
		dict[student.ID] = student
	}
	log.Info(ctx, "BatchGetMap.BatchGet students map success",
		log.Strings("ids", ids),
		log.Any("dict", dict))

	return dict, nil
}

func (s AmsStudentService) BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error) {
	return GetUserServiceProvider().BatchGetNameMap(ctx, operator, ids)
}

func (s AmsStudentService) GetByClassID(ctx context.Context, operator *entity.Operator, classID string) ([]*Student, error) {
	q := `query ($classID: ID!){
	class(class_id: $classID){
		students{
			id: user_id
			name: user_name
		}
  	}
}`
	req := chlorine.NewRequest(q, chlorine.ReqToken(operator.Token))
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
	_, err := GetAmsClient().Run(ctx, req, &res)
	if err != nil {
		log.Error(ctx, "Run error", log.String("q", q), log.Any("res", res), log.Err(err))
		return nil, err
	}
	if len(res.Errors) > 0 {
		log.Error(ctx, "Res error", log.String("q", q), log.Any("res", res), log.Err(res.Errors))
		return nil, res.Errors
	}

	log.Info(ctx, "get students by class success",
		log.String("classID", classID),
		log.Any("students", payload))

	return payload, nil
}

//TODO:No Test Program
func (s AmsStudentService) GetByClassIDs(ctx context.Context, operator *entity.Operator, classIDs []string) (map[string][]*Student, error) {
	if len(classIDs) == 0 {
		return map[string][]*Student{}, nil
	}

	sb := new(strings.Builder)

	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$class_id_", ": ID!", len(classIDs)))
	for index := range classIDs {
		fmt.Fprintf(sb, "q%d: class(class_id: $class_id_%d) {students{id:user_id name:user_name}}\n", index, index)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	for index, id := range classIDs {
		request.Var(fmt.Sprintf("class_id_%d", index), id)
	}

	data := map[string]*struct {
		Students []*Student `json:"students"`
	}{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get users by class ids failed", log.Err(err), log.Strings("ids", classIDs))
		return nil, err
	}

	students := make(map[string][]*Student, len(classIDs))
	for index, classID := range classIDs {
		query, found := data[fmt.Sprintf("q%d", index)]
		if !found || query == nil {
			log.Warn(ctx, "classes not found", log.Strings("classIDs", classIDs), log.String("id", classIDs[index]))
			continue
		}

		students[classID] = append(students[classID], query.Students...)
	}

	log.Info(ctx, "get students by classes success",
		log.Strings("classIDs", classIDs),
		log.Any("students", students))

	return students, nil
}

func (s AmsStudentService) Query(ctx context.Context, operator *entity.Operator, organizationID, keyword string) ([]*Student, error) {
	users, err := GetUserServiceProvider().Query(ctx, operator, organizationID, keyword)
	if err != nil {
		return nil, err
	}

	students := make([]*Student, len(users))
	for index, user := range users {
		students[index] = &Student{
			ID:   user.ID,
			Name: user.Name,
		}
	}

	return students, nil
}

func (s AmsStudentService) FilterByPermission(ctx context.Context, operator *entity.Operator, userIDs []string, permissionName PermissionName) ([]string, error) {
	return GetUserServiceProvider().FilterByPermission(ctx, operator, userIDs, permissionName)
}
func (s AmsStudentService) Name() string {
	return "ams_student_service"
}
