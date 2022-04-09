package connections

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"regexp"
	"strings"
	"text/template"
)

type ConnectionFilter interface {
	ProgramFilter | OrganizationFilter
	FilterType() FilterOfType
}

type ConnectionResponse interface {
	ProgramsConnectionResponse | OrganizationsConnectionResponse
	GetPageInfo() *ConnectionPageInfo
}

type ConnectionEdge interface {
	ProgramsConnectionEdge | OrganizationsConnectionEdge
}

type ConnectionNode interface {
	ProgramConnectionNode | OrganizationConnectionNode
}

//go:embed template.txt
var connection string

type EdgesFunc func(ctx context.Context, res interface{}) error

func Query[FilterType ConnectionFilter, ResType ConnectionResponse](
	ctx context.Context, operator *entity.Operator, filter FilterType, edgesFunc EdgesFunc) error {
	qString, err := queryString(ctx, filter)
	if err != nil {
		log.Error(ctx, "query: string failed",
			log.Any("filter", filter),
			log.Any("operator", operator))
		return err
	}
	var pageInfo *ConnectionPageInfo
	for pageInfo.HasNext(FORWARD) {
		res, err := do[ResType](ctx, operator, pageInfo.Pager(FORWARD, PageDefaultCount), qString)
		if err != nil {
			log.Error(ctx, "query: do failed",
				log.Any("filter", filter),
				log.Any("pageInfo", pageInfo),
				log.String("query", qString),
				log.Any("operator", operator))
			return err
		}
		connectionData := res[string(filter.FilterType())]
		pageInfo = connectionData.GetPageInfo()
		err = edgesFunc(ctx, connectionData)
		if err != nil {
			log.Error(ctx, "query: edgesFunc failed",
				log.Err(err),
				log.Any("filter", filter),
				log.Any("pageInfo", pageInfo),
				log.String("query", qString),
				log.Any("connection", connectionData),
				log.Any("operator", operator))
			return err
		}
	}
	return nil
}

func do[ResType ConnectionResponse](
	ctx context.Context,
	operator *entity.Operator,
	pager map[string]interface{},
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
		log.Error(ctx, "do: Run failed",
			log.Err(err),
			log.Any("pager", pager),
			log.String("query", query),
			log.Any("operator", operator))
		return nil, err
	}

	return res.Data, nil
}

func nodeFieldsString[NodeType ConnectionNode](ctx context.Context, node NodeType) (string, error) {
	data, err := json.Marshal(node)
	if err != nil {
		log.Error(ctx, "fieldString: marshal failed",
			log.Err(err),
			log.Any("node", node))
		return "", err
	}

	reg := regexp.MustCompile(`[{ ,]"([^:]*)":`)
	value := reg.FindAllStringSubmatch(string(data), -1)
	fields := make([]string, 0, len(value))
	for _, v := range value {
		if len(v) < 2 {
			log.Error(ctx, "fieldString: marshal failed",
				log.Err(ErrMatchFailed),
				log.Strings("v", v),
				log.Any("matched", value),
				log.Any("node", node))
			return "", ErrMatchFailed
		}
		fields = append(fields, v[1])
	}
	return strings.Join(fields, " "), nil
}

type argument struct {
	ConnectionName string
	FilterString   string
	NodeString     string
}

func queryArgument[FilterType ConnectionFilter](ctx context.Context, filter FilterType) (*argument, error) {
	data, err := json.Marshal(filter)
	if err != nil {
		log.Error(ctx, "argument: marshal failed",
			log.Err(err),
			log.Any("filter", filter))
		return nil, err
	}
	reg := regexp.MustCompile(`"__([^_]*)__"`)
	filterString := reg.ReplaceAllString(string(data), "${1}")
	filterString = strings.TrimSpace(filterString)
	if filterString == "{}" {
		filterString = ""
	}
	var connectionName, nodeFields string
	switch filter.FilterType() {
	case ProgramsConnectionType:
		connectionName = string(ProgramsConnectionType)
		nodeFields, err = nodeFieldsString(ctx, ProgramConnectionNode{})
	case OrganizationsConnectionType:
		connectionName = string(OrganizationsConnectionType)
		nodeFields, err = nodeFieldsString(ctx, OrganizationConnectionNode{})
	default:
		err = ErrUnsupported
	}

	if err != nil {
		log.Error(ctx, "argument: failed",
			log.Err(err),
			log.Any("filter", filter))
		return nil, err
	}
	return &argument{connectionName, filterString, nodeFields}, nil
}

func queryString[FilterType ConnectionFilter](ctx context.Context, filter FilterType) (string, error) {
	params, err := queryArgument(ctx, filter)
	if err != nil {
		log.Error(ctx, "string: failed", log.Any("filter", filter))
		return "", err
	}
	temp, err := template.New("ProgramsConnection").Parse(connection)
	if err != nil {
		log.Error(ctx, "string: template parse failed",
			log.Err(err),
			log.Any("filter", filter),
			log.Any("params", params),
			log.String("connection", connection))
		return "", err
	}

	buf := bytes.Buffer{}
	err = temp.Execute(&buf, params)
	if err != nil {
		log.Error(ctx, "string: template execute failed",
			log.Err(err),
			log.Any("filter", filter),
			log.Any("params", params),
			log.String("connection", connection))
		return "", err
	}
	return buf.String(), nil
}
