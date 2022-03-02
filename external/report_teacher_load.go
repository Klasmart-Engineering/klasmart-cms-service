package external

import (
	"context"
	"net/http"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type TeacherLoadServiceProvider interface {
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

func (t *AmsTeacherLoadService) BatchGetActiveClassWithStudent(ctx context.Context, operator *entity.Operator, teacherIDs []string) (map[string]*TeacherClassWithStudent, error) {

	ids := utils.SliceDeduplicationExcludeEmpty(teacherIDs)
	if len(ids) == 0 {
		return map[string]*TeacherClassWithStudent{}, nil
	}
	classOr := make([]ClassFilter, 0, len(ids))
	userOr := make([]UserFilter, 0, len(ids))
	for _, v := range ids {
		cf := ClassFilter{TeacherID: &UUIDFilter{}}
		cf.TeacherID.Operator = UUIDOperator(OperatorTypeEq)
		cf.TeacherID.Value = UUID(v)
		classOr = append(classOr, cf)

		uf := UserFilter{}
		uf.UserID.Operator = UUIDOperator(OperatorTypeEq)
		uf.UserID.Value = UUID(v)
		userOr = append(userOr, uf)
	}

	query := `
query ($classOr:[ClassFilter!], $userOr:[UserFilter!], $classPageDirection: ConnectionDirection!, $classPageCursor: String, $studentPageDirection: ConnectionDirection!, $studentPageCursor: String, $teacherPageDirection:ConnectionDirection!, $teacherPageCursor: String){
	classesConnection(direction:$classPageDirection, directionArgs:{cursor: $classPageCursor,}  filter:{status: {operator: eq, value: "active"}, OR: $classOr}) {
    totalCount
    pageInfo {
      hasNextPage
      hasPreviousPage
      startCursor
      endCursor
    }
    edges{
      node{
        id
        name
        status
        studentsConnection(direction: $studentPageDirection, cursor:$studentPageCursor, filter:{userStatus: {operator:eq, value:"active"}}){
          totalCount
          pageInfo {
            hasNextPage
            hasPreviousPage
            startCursor
            endCursor
          }
          edges{
            node{
              id
              givenName
              familyName
              status
            }
          }
        }
        teachersConnection(direction:$teacherPageDirection, cursor: $teacherPageCursor,filter:{OR: $userOr}){
           pageInfo {
            hasNextPage
            hasPreviousPage
            startCursor
            endCursor
          }
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
`

	request := chlorine.NewRequest(query, chlorine.ReqToken(operator.Token))
	request.Var("classOr", classOr)
	request.Var("userOr", userOr)
	request.Var("classPageDirection", "FORWARD")
	request.Var("classPageCursor", "")
	request.Var("studentPageDirection", "FORWARD")
	request.Var("studentPageCursor", "")
	request.Var("teacherPageDirection", "FORWARD")
	request.Var("teacherPageCursor", "")

	var data ClassesConnection
	response := &chlorine.Response{
		Data: &struct {
			*ClassesConnection `json:"classesConnection"`
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

	for _, class := range data.Edges {
		studentsConnection := &(class.Node.StudentsConnection)
		var studentEdges []UserConnectionEdge
		studentEdges = append(studentEdges, studentsConnection.Edges...)
		for studentsConnection.HasNext() {
			result, err := studentsConnection.Next(ctx, func() (Iterator, error) {
				cursor := studentsConnection.PageInfo.ForwardCursor()
				studentQuery := `
						query ($classFilter:ClassFilter!,$classPageDirection: ConnectionDirection!, $studentPageDirection: ConnectionDirection!, $studentPageCursor: String){
							classesConnection(direction: $classPageDirection, filter: $classFilter) {
							totalCount
							edges{
							  node{
								id
								name
								status
								studentsConnection(direction: $studentPageDirection, cursor:$studentPageCursor, filter:{userStatus: {operator:eq, value:"active"}}){
											totalCount
								  pageInfo {
									hasNextPage
									hasPreviousPage
									startCursor
									endCursor
								  }
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
						`
				request := chlorine.NewRequest(studentQuery, chlorine.ReqToken(operator.Token))
				request.Var("classFilter", ClassFilter{ID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(class.Node.ID)}})
				request.Var("classPageDirection", Forward)
				request.Var("studentPageDirection", Forward)
				request.Var("studentPageCursor", cursor)
				var data ClassesConnection
				response := &chlorine.Response{
					Data: &struct {
						*ClassesConnection `json:"classesConnection"`
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
				if data.Edges == nil || len(data.Edges) == 0 {
					err = constant.ErrRecordNotFound
					log.Error(ctx, "BatchGetClassWithStudent: class not found", log.Err(err), log.Any("data", data), log.Strings("teacher_ids", ids))
					return nil, err

				}
				return &(data.Edges[0].Node.StudentsConnection), nil
			})
			if err != nil {
				return nil, err
			}
			edges, ok := result.([]UserConnectionEdge)
			if !ok {
				err = constant.ErrAssertFailed
				log.Error(ctx, "BatchGetClassWithStudent: assert failed", log.Err(err), log.Any("result", result), log.Strings("teacher_ids", ids))
				return nil, err
			}
			studentEdges = append(studentEdges, edges...)
		}

		classStudents := ClassStudents{ClassID: class.Node.ID}
		classStudents.Students = make([]*StudentInClass, 0, len(studentEdges))
		for _, edge := range studentEdges {
			classStudents.Students = append(classStudents.Students, &StudentInClass{UserID: edge.Node.ID})
		}

		for _, teacher := range class.Node.TeachersConnection.Edges {
			if _, ok := result[teacher.Node.ID]; !ok {
				result[teacher.Node.ID] = &TeacherClassWithStudent{
					UserID: teacher.Node.ID,
				}
			}
			result[teacher.Node.ID].ClassesTeaching = append(result[teacher.Node.ID].ClassesTeaching, &classStudents)
		}
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
