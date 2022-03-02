package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type Iterator interface {
	HasNext() bool
	Next(ctx context.Context, do func() (Iterator, error)) (interface{}, error)
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
	OrganizationId *UUIDFilter   `json:"organizationId,omitempty"`
	TeacherID      *UUIDFilter   `json:"teacherId,omitempty"`
}

type PageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

func (pageInfo PageInfo) HasNext() bool {
	return pageInfo.HasNextPage
}

func (pageInfo PageInfo) ForwardCursor() string {
	if pageInfo.HasNextPage {
		return pageInfo.EndCursor
	}
	return ""
}

func (pageInfo PageInfo) BackwardCursor() string {
	if pageInfo.HasPreviousPage {
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

type StudentsConnection struct {
	PageInfo PageInfo             `json:"pageInfo"`
	Edges    []UserConnectionEdge `json:"edges"`
}

func (sc *StudentsConnection) HasNext() bool {
	return sc.PageInfo.HasNext()
}

func (sc *StudentsConnection) Next(ctx context.Context, do func() (Iterator, error)) (interface{}, error) {
	it, err := do()
	if err != nil {
		return nil, err
	}
	studentsConnection, ok := it.(*StudentsConnection)
	if !ok {
		err = constant.ErrAssertFailed
		log.Error(ctx, "BatchGetClassWithStudent: assert failed", log.Err(err), log.Any("it", it))
		return nil, err
	}
	sc.PageInfo = studentsConnection.PageInfo
	sc.Edges = studentsConnection.Edges
	return sc.Edges, err
}

type ClassesConnection struct {
	PageInfo PageInfo `json:"pageInfo"`
	Edges    []struct {
		Node struct {
			ID                 string             `json:"id"`
			Name               string             `json:"name"`
			Status             APStatus           `json:"status"`
			StudentsConnection StudentsConnection `json:"studentsConnection"`
			TeachersConnection TeachersConnection `json:"teachersConnection"`
		} `json:"node"`
	} `json:"edges"`
}
