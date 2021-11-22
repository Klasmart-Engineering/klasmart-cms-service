package external

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"sync"
	"text/template"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type TeacherLoadServiceProvider interface {
	BatchGetClassWithStudent(ctx context.Context, operator *entity.Operator, teacherIDs []string) (map[string]*TeacherClassWithStudent, error)
	BatchGetActiveClassWithStudent(ctx context.Context, operator *entity.Operator, teacherIDs []string) (map[string]*TeacherClassWithStudent, error)
}

type AmsTeacherLoadService struct{}

var (
	_amsTeacherLoadService     *AmsTeacherLoadService
	_amsTeacherLoadServiceOnce sync.Once
)

func GetTeacherLoadServiceProvider() TeacherLoadServiceProvider {
	_amsTeacherLoadServiceOnce.Do(func() {
		_amsTeacherLoadService = &AmsTeacherLoadService{}
	})

	return _amsTeacherLoadService
}

func (t *AmsTeacherLoadService) BatchGetClassWithStudent(ctx context.Context, operator *entity.Operator, teacherIDs []string) (map[string]*TeacherClassWithStudent, error) {
	ids := utils.SliceDeduplicationExcludeEmpty(teacherIDs)
	if len(ids) == 0 {
		return map[string]*TeacherClassWithStudent{}, nil
	}

	query := `
query {
	{{range $i, $e := .}}
	q{{$i}}: user(user_id: "{{$e}}") {
		user_id
		classesTeaching {
			class_id
			students{
				user_id
			}
		}
	}
	{{end}}
}`

	temp, err := template.New("").Parse(query)
	if err != nil {
		log.Error(ctx, "BatchGetClassWithStudent: init template failed", log.Err(err))
		return nil, err
	}
	buffer := new(bytes.Buffer)
	err = temp.Execute(buffer, ids)
	if err != nil {
		log.Error(ctx, "BatchGetClassWithStudent: execute template failed", log.Err(err), log.Strings("teacher_ids", ids))
		return nil, err
	}

	request := chlorine.NewRequest(buffer.String(), chlorine.ReqToken(operator.Token))
	data := map[string]*TeacherClassWithStudent{}
	response := &chlorine.Response{
		Data: &data,
	}

	statusCode, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "BatchGetClassWithStudent: run failed", log.Err(err), log.Strings("teacher_ids", ids))
		return nil, err
	}
	if statusCode != http.StatusOK {
		err = &entity.ExternalError{
			Err:  errors.New("response data contains err"),
			Type: constant.InternalErrorTypeAms,
		}
		log.Error(ctx, "BatchGetClassWithStudent: run failed", log.Int("status_code", statusCode), log.Err(err), log.Strings("teacher_ids", ids))
		return nil, err
	}
	result := make(map[string]*TeacherClassWithStudent, len(data))
	for i, v := range data {
		log.Debug(ctx, "BatchGetClassWithStudent: extract result",
			log.String("index", i),
			log.Any("data", v))
		if v != nil {
			result[v.UserID] = v
		}
	}
	return result, nil
}

type UserFilter struct {
	UserID struct {
		Operator OperatorType `json:"operator"`
		Value    string       `json:"value"`
	} `json:"userId"`
}

type UsersConnection struct {
	Edges []struct {
		Node struct {
			ID                        string `json:"id"`
			ClassesTeachingConnection struct {
				Edges []struct {
					Node struct {
						ID                 string   `json:"id"`
						Name               string   `json:"name"`
						Status             APStatus `json:"status"`
						StudentsConnection struct {
							Edges []struct {
								Node struct {
									ID         string   `json:"id"`
									GivenName  string   `json:"givenName"`
									FamilyName string   `json:"familyName"`
									Status     APStatus `json:"status"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"studentsConnection"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"classesTeachingConnection"`
		} `json:"node"`
	} `json:"edges"`
}

func (t *AmsTeacherLoadService) BatchGetActiveClassWithStudent(ctx context.Context, operator *entity.Operator, teacherIDs []string) (map[string]*TeacherClassWithStudent, error) {

	ids := utils.SliceDeduplicationExcludeEmpty(teacherIDs)
	if len(ids) == 0 {
		return map[string]*TeacherClassWithStudent{}, nil
	}
	orFilters := make([]UserFilter, len(ids))
	for i, v := range ids {
		uf := UserFilter{}
		uf.UserID.Operator = OperatorTypeEq
		uf.UserID.Value = v
		orFilters[i] = uf
	}

	query := `
query ($or:[UserFilter!]){
	usersConnection(direction:BACKWARD, filter:{OR:$or}) {
		edges{
		  node{
			id
			classesTeachingConnection(filter:{status: {operator: eq, value: "active"}}){
			  edges{
				node{
				  id
				  name
				  status
				  studentsConnection(filter:{userStatus:{operator: eq, value: "active"}, organizationUserStatus:{operator:eq, value: "active"}}){
					edges{
					  node{
						id
						givenName
						familyName
						status
					  }
					}
				  }
				}
			  }
			}
		  }
		}
	} 
}`

	request := chlorine.NewRequest(query, chlorine.ReqToken(operator.Token))
	request.Var("or", orFilters)

	var data UsersConnection
	response := &chlorine.Response{
		Data: &struct {
			*UsersConnection `json:"usersConnection"`
		}{
			&data,
		},
	}

	statusCode, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "BatchGetClassWithStudent: run failed", log.Err(err), log.Strings("teacher_ids", ids))
		return nil, err
	}
	if statusCode != http.StatusOK {
		err = &entity.ExternalError{
			Err:  constant.ErrAmsHttpFailed,
			Type: constant.InternalErrorTypeAms,
		}
		log.Warn(ctx, "BatchGetClassWithStudent: run failed", log.Int("status_code", statusCode), log.Strings("teacher_ids", ids))
		return nil, err
	}
	result := make(map[string]*TeacherClassWithStudent, len(data.Edges))
	for _, teacher := range data.Edges {
		tcs := TeacherClassWithStudent{}
		tcs.UserID = teacher.Node.ID
		tcs.ClassesTeaching = make([]*ClassStudents, len(teacher.Node.ClassesTeachingConnection.Edges))
		for i, class := range teacher.Node.ClassesTeachingConnection.Edges {
			classStudents := ClassStudents{ClassID: class.Node.ID}
			classStudents.Students = make([]*StudentInClass, len(class.Node.StudentsConnection.Edges))
			for j, student := range class.Node.StudentsConnection.Edges {
				classStudents.Students[j] = &StudentInClass{student.Node.ID}
			}
			tcs.ClassesTeaching[i] = &classStudents
		}
		result[tcs.UserID] = &tcs
	}
	return result, err
}

type StudentInClass struct {
	UserID string `json:"user_id"`
}
type ClassStudents struct {
	ClassID  string            `json:"class_id"`
	Students []*StudentInClass `json:"students"`
}

type TeacherClassWithStudent struct {
	UserID          string           `json:"user_id"`
	ClassesTeaching []*ClassStudents `json:"classesTeaching"`
}

type TeacherClassWithStudentCounter struct {
	Class   int
	Student int
}

func (tcs TeacherClassWithStudent) CountClassAndStudent(ctx context.Context, classIDList []string) TeacherClassWithStudentCounter {
	classIDs := make([]string, 0, len(classIDList))
	studentIDs := make([]string, 0)
	for _, class := range tcs.ClassesTeaching {
		if !utils.ContainsString(classIDList, class.ClassID) {
			continue
		}
		classIDs = append(classIDs, class.ClassID)
		for _, student := range class.Students {
			studentIDs = append(studentIDs, student.UserID)
		}

		log.Debug(ctx, "CountClassAndStudent",
			log.String("teacher_id", tcs.UserID),
			log.String("class_id", class.ClassID),
			log.Any("students", class.Students))
	}

	return TeacherClassWithStudentCounter{
		Class:   len(utils.SliceDeduplication(classIDs)),
		Student: len(utils.SliceDeduplication(studentIDs)),
	}
}
