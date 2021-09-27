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
	for _, v := range data {
		result[v.UserID] = v
	}
	return result, nil
}

type TeacherClassWithStudent struct {
	UserID          string `json:"user_id"`
	ClassesTeaching []struct {
		ClassID  string `json:"class_id"`
		Students []struct {
			UserID string `json:"user_id"`
		} `json:"students"`
	} `json:"classesTeaching"`
}

type TeacherClassWithStudentCounter struct {
	Class   int64
	Student int64
}

func (tcs TeacherClassWithStudent) CountClassAndStudent(ctx context.Context) TeacherClassWithStudentCounter {
	classIDs := make([]string, 0, len(tcs.ClassesTeaching))
	studentIDs := make([]string, 0)
	for _, class := range tcs.ClassesTeaching {
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
		Class:   int64(len(utils.SliceDeduplication(classIDs))),
		Student: int64(len(utils.SliceDeduplication(studentIDs))),
	}
}
