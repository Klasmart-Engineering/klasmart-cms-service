package connections

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"regexp"
	"strings"
	"text/template"
)

type ConnectionResponse interface {
	ProgramsConnectionResponse | OrganizationsConnectionResponse
}

type ConnectionEdge interface {
	ProgramsConnectionEdge | OrganizationsConnectionEdge
}

type ConnectionNode interface {
	ProgramConnectionNode | OrganizationConnectionNode
}
type ProgramFilter struct {
	ID             *UUIDFilter         `json:"__id__,omitempty"`
	Name           *StringFilter       `json:"__name__,omitempty"`
	Status         *StringFilter       `json:"__status__,omitempty"`
	System         *BooleanFilter      `json:"__system__,omitempty"`
	OrganizationID *UUIDFilter         `json:"__organizationId__,omitempty"`
	GradeID        *UUIDFilter         `json:"__gradeId__,omitempty"`
	AgeRangeFrom   *AgeRangeTypeFilter `json:"__ageRangeFrom__,omitempty"`
	AgeRangeTo     *AgeRangeTypeFilter `json:"__ageRangeTo__,omitempty"`
	SubjectID      *UUIDFilter         `json:"__subjectId__,omitempty"`
	SchoolID       *UUIDFilter         `json:"__schoolId,omitempty"`
	ClassID        *UUIDFilter         `json:"__classId__,omitempty"`
	AND            []ProgramFilter     `json:"__AND__,omitempty"`
	OR             []ProgramFilter     `json:"__OR__,omitempty"`
}

func (ProgramFilter) FilterType() FilterOfType {
	return ProgramsConnectionType
}

//go:embed template.txt
var connection string

func nodeFieldsString[NodeType ConnectionNode](node NodeType) (string, error) {
	data, err := json.Marshal(node)
	if err != nil {
		// TODO
		return "", err
	}

	reg := regexp.MustCompile(`[{ ,]"([^:]*)":`)
	value := reg.FindAllStringSubmatch(string(data), -1)
	fields := make([]string, 0, len(value))
	for _, v := range value {
		fields = append(fields, v[1])
	}
	return strings.Join(fields, " "), nil
}

type argument struct {
	ConnectionName string
	FilterString   string
	NodeString     string
}

func queryArgument[FilterType ConnectionFilter](filter FilterType) (*argument, error) {
	data, err := json.Marshal(filter)
	if err != nil {
		// TODO
		return nil, err
	}
	reg := regexp.MustCompile(`"__([^_]*)__"`)
	filterString := reg.ReplaceAllString(string(data), "${1}")

	var connectionName, nodeFields string
	switch filter.FilterType() {
	case ProgramsConnectionType:
		connectionName = string(ProgramsConnectionType)
		nodeFields, err = nodeFieldsString(ProgramConnectionNode{})
	case OrganizationsConnectionType:
		connectionName = string(OrganizationsConnectionType)
		nodeFields, err = nodeFieldsString(OrganizationConnectionNode{})
	default:
		err = errors.New("unsupported")
	}

	if err != nil {
		// TODO
		return nil, err
	}
	return &argument{connectionName, filterString, nodeFields}, nil
}

func queryString[FilterType ConnectionFilter](filter FilterType) (string, error) {
	params, err := queryArgument(filter)
	if err != nil {
		// TODO
		return "", err
	}
	temp, err := template.New("ProgramsConnection").Parse(connection)
	if err != nil {
		// TODO
	}

	buf := bytes.Buffer{}
	err = temp.Execute(&buf, params)
	if err != nil {
		// TODO
	}
	return buf.String(), nil
}

func Query[FilterType ConnectionFilter, ResType ConnectionResponse, EdgeType ConnectionEdge](
	ctx context.Context, operator *entity.Operator, filter FilterType) ([]EdgeType, error) {
	queryString, err := queryString(filter)
	if err != nil {
		// TODO
	}
	pageInfo := ConnectionPageInfo{
		HasNextPage:     true,
		HasPreviousPage: true,
		StartCursor:     "",
		EndCursor:       "",
	}
	for pageInfo.HasNextPage {
		do[ResType](ctx, operator, pageInfo.Pager(FORWARD, 50), queryString)
		if err != nil {
			// TODO
		}
		//if response != nil {
		//	pageInfo = response.PageInfo
		//}
		//fmt.Println(response.Edges)
	}
	return nil, nil
}

func do[ResType ConnectionResponse](
	ctx context.Context,
	operator *entity.Operator,
	pager Pager,
	query string) (map[string]ResType, error) {
	res := GraphQLResponse[ResType]{
		Data: map[string]ResType{},
	}
	req := NewRequest(query, ReqToken(operator.Token))
	for k, v := range pager {
		req.Var(k, v)
	}
	_, err := Run(ctx, GetAmsProvider(), req, &res)
	if err != nil {
		// TODO
		return nil, err
	}

	return res.Data, nil
}
