package external

import (
	"context"
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

func makeClassConnectionRequest(ids []string) (string, map[string]interface{}, chlorine.Response) {
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

	classQuery := `
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
	classVariables := make(map[string]interface{})
	classVariables["classOr"] = classOr
	classVariables["userOr"] = userOr
	classVariables["classPageDirection"] = Forward
	classVariables["studentPageDirection"] = Forward
	classVariables["studentPageCursor"] = ""
	classVariables["teacherPageDirection"] = Forward
	classVariables["teacherPageCursor"] = ""
	classResponse := chlorine.Response{
		Data: &struct {
			ClassesConnection `json:"classesConnection"`
		}{},
	}
	return classQuery, classVariables, classResponse
}

func makeStudentsConnectionRequest(classID string) (string, map[string]interface{}, chlorine.Response) {
	studentsQuery := `
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
	studentVariables := make(map[string]interface{})

	studentVariables["classFilter"] = ClassFilter{ID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(classID)}}
	studentVariables["classPageDirection"] = Forward
	studentVariables["studentPageDirection"] = Forward

	studentResponse := chlorine.Response{
		Data: &struct {
			ClassesConnection `json:"classesConnection"`
		}{},
	}
	return studentsQuery, studentVariables, studentResponse
}

func makeTeacherConnectionRequest(classID string, teacherIDs []string) (string, map[string]interface{}, chlorine.Response) {
	userOr := make([]UserFilter, 0, len(teacherIDs))
	for _, v := range teacherIDs {
		uf := UserFilter{}
		uf.UserID.Operator = UUIDOperator(OperatorTypeEq)
		uf.UserID.Value = UUID(v)
		userOr = append(userOr, uf)
	}
	teachersQuery := `
				query ($classFilter:ClassFilter!, $userOr:[UserFilter!], $classPageDirection: ConnectionDirection! $teacherPageDirection:ConnectionDirection!, $teacherPageCursor: String){
					classesConnection(direction: $classPageDirection, filter: $classFilter) {
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
	teacherVariables := make(map[string]interface{})

	teacherVariables["classFilter"] = ClassFilter{ID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(classID)}}
	teacherVariables["classPageDirection"] = Forward
	teacherVariables["teacherPageDirection"] = Forward
	teacherVariables["userOr"] = userOr

	teacherResponse := chlorine.Response{
		Data: &struct {
			ClassesConnection `json:"classesConnection"`
		}{},
	}
	return teachersQuery, teacherVariables, teacherResponse
}
func (t *AmsTeacherLoadService) BatchGetActiveClassWithStudent(ctx context.Context, operator *entity.Operator, teacherIDs []string) (map[string]*TeacherClassWithStudent, error) {

	ids := utils.SliceDeduplicationExcludeEmpty(teacherIDs)
	if len(ids) == 0 {
		return map[string]*TeacherClassWithStudent{}, nil
	}
	result := make(map[string]*TeacherClassWithStudent)

	classQuery, classVariables, classResponse := makeClassConnectionRequest(ids)

	var classData ClassesConnection
	cIterator := &classData
	for cIterator.HasNext() {
		classVariables["classPageCursor"] = cIterator.PageInfo.ForwardCursor()
		cEdge, err := cIterator.Next(ctx, operator, classQuery, classVariables, classResponse, func(response *chlorine.Response) (Iterator, error) {
			classesConnection, ok := response.Data.(*struct {
				ClassesConnection `json:"classesConnection"`
			})
			if !ok {
				err := constant.ErrAssertFailed
				log.Error(ctx, "Next: assert failed",
					log.Err(err),
					log.Any("data", response))
				return nil, err
			}
			return &classesConnection.ClassesConnection, nil
		})
		if err != nil {
			log.Error(ctx, "BatchGetClassWithStudent: cIterator next failed", log.Err(err), log.Any("result", result), log.Strings("teacher_ids", ids))
			return nil, err
		}
		classEdges, ok := cEdge.([]ClassesConnectionEdge)
		if !ok {
			err = constant.ErrAssertFailed
			log.Error(ctx, "BatchGetClassWithStudent: assert failed", log.Err(err), log.Any("result", result), log.Strings("teacher_ids", ids))
			return nil, err
		}

		for _, class := range classEdges {
			sIterator := &(class.Node.StudentsConnection)
			studentsQuery, studentVariables, studentResponse := makeStudentsConnectionRequest(class.Node.ID)

			var studentEdges []UserConnectionEdge
			studentEdges = append(studentEdges, sIterator.Edges...)
			for sIterator.HasNext() {
				studentVariables["studentPageCursor"] = sIterator.PageInfo.ForwardCursor()
				sEdge, err := sIterator.Next(ctx, operator, studentsQuery, studentVariables, studentResponse, func(response *chlorine.Response) (Iterator, error) {
					data, ok := response.Data.(*struct {
						ClassesConnection `json:"classesConnection"`
					})
					if !ok {
						err = constant.ErrAssertFailed
						log.Error(ctx, "Next: assert failed",
							log.Err(err),
							log.Any("data", response))
						return nil, err
					}
					if len(data.Edges) == 0 {
						err = constant.ErrAmsDataFailed
						log.Error(ctx, "Next: data failed",
							log.Err(err),
							log.Any("data", response))
						return nil, err
					}
					return &(data.Edges[0].Node.StudentsConnection), nil
				})
				if err != nil {
					return nil, err
				}
				edges, ok := sEdge.([]UserConnectionEdge)
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

			tIterator := &(class.Node.TeachersConnection)
			teachersQuery, teacherVariables, teacherResponse := makeTeacherConnectionRequest(class.Node.ID, ids)

			var teacherEdges []UserConnectionEdge
			teacherEdges = append(teacherEdges, tIterator.Edges...)
			for tIterator.HasNext() {
				teacherVariables["teacherPageCursor"] = tIterator.PageInfo.ForwardCursor()
				sEdge, err := sIterator.Next(ctx, operator, teachersQuery, teacherVariables, teacherResponse, func(response *chlorine.Response) (Iterator, error) {
					data, ok := response.Data.(*struct {
						ClassesConnection `json:"classesConnection"`
					})
					if !ok {
						err = constant.ErrAssertFailed
						log.Error(ctx, "Next: assert failed",
							log.Err(err),
							log.Any("data", response))
						return nil, err
					}
					if len(data.Edges) == 0 {
						err = constant.ErrAmsDataFailed
						log.Error(ctx, "Next: data failed",
							log.Err(err),
							log.Any("data", response))
						return nil, err
					}
					return &(data.Edges[0].Node.TeachersConnection), nil
				})
				if err != nil {
					return nil, err
				}
				edges, ok := sEdge.([]UserConnectionEdge)
				if !ok {
					err = constant.ErrAssertFailed
					log.Error(ctx, "BatchGetClassWithStudent: assert failed", log.Err(err), log.Any("result", result), log.Strings("teacher_ids", ids))
					return nil, err
				}
				teacherEdges = append(studentEdges, edges...)
			}

			for _, teacher := range teacherEdges {
				if _, ok := result[teacher.Node.ID]; !ok {
					result[teacher.Node.ID] = &TeacherClassWithStudent{
						UserID: teacher.Node.ID,
					}
				}
				result[teacher.Node.ID].ClassesTeaching = append(result[teacher.Node.ID].ClassesTeaching, &classStudents)
			}
		}
	}

	return result, nil
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
