package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"net/http"
)

type Iterator interface {
	HasNext() bool
	Next(ctx context.Context, operator *entity.Operator, query string, variables map[string]interface{}, res chlorine.Response, f func(*chlorine.Response) (Iterator, error)) (interface{}, error)
}

type Direction string

var (
	Forward  Direction = "FORWARD"
	BackWard Direction = "BACKWARD"
)

type UUID string
type UUIDOperator OperatorType
type StringOperator OperatorType

type UUIDFilter struct {
	Operator UUIDOperator `json:"operator"`
	Value    UUID         `json:"value"`
}

type UserFilter struct {
	UserID UUIDFilter `json:"userId"`
}

type StringFilter struct {
	Operator        StringOperator `json:"operator"`
	Value           string         `json:"value"`
	CaseInsensitive bool           `json:"caseInsensitive"`
}

type ClassFilter struct {
	ID             *UUIDFilter   `json:"id,omitempty"`
	Name           *StringFilter `json:"name,omitempty"`
	Status         *StringFilter `json:"status,omitempty"`
	OrganizationID *UUIDFilter   `json:"organizationId,omitempty"`
	TeacherID      *UUIDFilter   `json:"teacherId,omitempty"`
}

type PageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

func (pageInfo *PageInfo) HasNext() bool {
	if pageInfo == nil {
		// the beginning
		return true
	}
	return pageInfo.HasNextPage
}

func (pageInfo *PageInfo) ForwardCursor() string {
	if pageInfo != nil && pageInfo.HasNextPage {
		return pageInfo.EndCursor
	}
	return ""
}

func (pageInfo *PageInfo) BackwardCursor() string {
	if pageInfo != nil && pageInfo.HasPreviousPage {
		return pageInfo.StartCursor
	}
	return ""
}

type UserConnectionNode struct {
	ID         string   `json:"id"`
	GivenName  string   `json:"givenName"`
	FamilyName string   `json:"familyName"`
	Status     APStatus `json:"status"`
}

type UserConnectionEdge struct {
	Node UserConnectionNode `json:"node"`
}

type TeachersConnection struct {
	PageInfo PageInfo             `json:"pageInfo"`
	Edges    []UserConnectionEdge `json:"edges"`
}

func (tc *TeachersConnection) HasNext() bool {
	if tc == nil {
		return false
	}
	return tc.PageInfo.HasNext()
}

func (tc *TeachersConnection) Next(ctx context.Context, operator *entity.Operator, query string, variables map[string]interface{}, res chlorine.Response, f func(response *chlorine.Response) (Iterator, error)) (interface{}, error) {
	//var res chlorine.Response
	err := do(ctx, operator, query, variables, &res)
	if err != nil {
		return nil, err
	}
	it, err := f(&res)
	if err != nil {
		log.Error(ctx, "Next: f failed",
			log.Any("data", res))
		return nil, err
	}
	teacherConnection, ok := it.(*TeachersConnection)
	if !ok {
		err = constant.ErrAssertFailed
		log.Error(ctx, "Next: assert failed",
			log.Err(err),
			log.Any("data", res))
		return nil, err
	}
	log.Debug(ctx, "Next: success",
		log.String("query", query),
		log.Any("variables", variables),
		log.Any("classesConnection", teacherConnection))
	tc.PageInfo = teacherConnection.PageInfo
	tc.Edges = teacherConnection.Edges
	return tc.Edges, err
}

type StudentsConnection struct {
	PageInfo PageInfo             `json:"pageInfo"`
	Edges    []UserConnectionEdge `json:"edges"`
}

func (sc *StudentsConnection) HasNext() bool {
	if sc == nil {
		return false
	}
	return sc.PageInfo.HasNext()
}

func (sc *StudentsConnection) Next(ctx context.Context, operator *entity.Operator, query string, variables map[string]interface{}, res chlorine.Response, f func(response *chlorine.Response) (Iterator, error)) (interface{}, error) {
	//var res chlorine.Response
	err := do(ctx, operator, query, variables, &res)
	if err != nil {
		return nil, err
	}
	it, err := f(&res)
	if err != nil {
		log.Error(ctx, "Next: f failed",
			log.Any("data", res))
		return nil, err
	}
	studentsConnection, ok := it.(*StudentsConnection)
	if !ok {
		err = constant.ErrAssertFailed
		log.Error(ctx, "Next: assert failed",
			log.Err(err),
			log.Any("data", res))
		return nil, err
	}
	log.Debug(ctx, "Next: success",
		log.String("query", query),
		log.Any("variables", variables),
		log.Any("classesConnection", studentsConnection))
	sc.PageInfo = studentsConnection.PageInfo
	sc.Edges = studentsConnection.Edges
	return sc.Edges, err
}

type ClassesConnectionNode struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	Status             APStatus           `json:"status"`
	StudentsConnection StudentsConnection `json:"studentsConnection"`
	TeachersConnection TeachersConnection `json:"teachersConnection"`
}

type ClassesConnectionEdge struct {
	Node ClassesConnectionNode `json:"node"`
}

type ClassesConnection struct {
	PageInfo *PageInfo               `json:"pageInfo"`
	Edges    []ClassesConnectionEdge `json:"edges"`
}

func (cc *ClassesConnection) HasNext() bool {
	return cc.PageInfo.HasNext()
}

func (cc *ClassesConnection) Next(ctx context.Context, operator *entity.Operator, query string, variables map[string]interface{}, res chlorine.Response, f func(response *chlorine.Response) (Iterator, error)) (interface{}, error) {
	//var res chlorine.Response
	err := do(ctx, operator, query, variables, &res)
	if err != nil {
		return nil, err
	}

	it, err := f(&res)
	if err != nil {
		log.Error(ctx, "Next: f failed",
			log.Any("data", res))
		return nil, err
	}
	classesConnection, ok := it.(*ClassesConnection)
	if !ok {
		err = constant.ErrAssertFailed
		log.Error(ctx, "Next: assert failed",
			log.Err(err),
			log.Any("data", res))
		return nil, err
	}

	log.Debug(ctx, "Next: success",
		log.String("query", query),
		log.Any("variables", variables),
		log.Any("classesConnection", classesConnection))

	cc.PageInfo = classesConnection.PageInfo
	cc.Edges = classesConnection.Edges
	return cc.Edges, err
}

func do(ctx context.Context, operator *entity.Operator, query string, variables map[string]interface{}, response *chlorine.Response) error {
	request := chlorine.NewRequest(query, chlorine.ReqToken(operator.Token))
	for k, v := range variables {
		request.Var(k, v)
	}

	statusCode, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "do: run failed",
			log.Err(err),
			log.String("query", query),
			log.Any("variables", variables))
		return err
	}
	if statusCode != http.StatusOK {
		err = &entity.ExternalError{
			Err:  constant.ErrAmsHttpFailed,
			Type: constant.InternalErrorTypeAms,
		}
		log.Warn(ctx, "do: run failed",
			log.Int("status_code", statusCode),
			log.String("query", query),
			log.Any("variables", variables))
		return err
	}

	log.Debug(ctx, "do success",
		log.String("query", query),
		log.Any("variables", variables),
		log.Any("response", response))

	return nil
}