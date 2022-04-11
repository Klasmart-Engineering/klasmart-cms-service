package gql

import (
	"bytes"
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"text/template"
)

var graphQLString = `
query($direction:ConnectionDirection!, $cursor:String!, $count:PageSize){
    {{.ConnectionName}}(
        direction:$direction,
        directionArgs:{cursor:$cursor, count: $count},
        {{if .FilterString}}filter:{{.FilterString}},{{end}}
        sort:{field:[id],order:ASC}) {
        {{.NodeString}}
    }
}
`

func Query[ResType ConnectionResponse](ctx context.Context, operator *entity.Operator, key FilterOfType, filter interface{}, result *[]ResType) error {
	qString, err := queryString(ctx, key, filter, result)
	if err != nil {
		log.Error(ctx, "query: string failed",
			log.Any("filter", filter),
			log.Any("operator", operator))
		return err
	}
	var pageInfo *ConnectionPageInfo
	for pageInfo.HasNext(FORWARD) {
		res := GraphQLResponse[ResType]{
			Data: map[string]ResType{},
		}
		err := do(ctx, operator, pageInfo.Pager(FORWARD, PageDefaultCount), qString, &res)
		if err != nil {
			log.Error(ctx, "query: do failed",
				log.Any("filter", filter),
				log.Any("pageInfo", pageInfo),
				log.String("query", qString),
				log.Any("operator", operator))
			return err
		}
		connections := res.Data[string(key)]
		pageInfo = connections.GetPageInfo()
		*result = append(*result, connections)
	}
	return nil
}

func do[ResType ConnectionResponse](ctx context.Context, operator *entity.Operator, pager map[string]interface{}, query string, res *GraphQLResponse[ResType]) error {
	req := NewRequest(query, ReqToken(operator.Token))
	for k, v := range pager {
		req.Var(k, v)
	}
	_, err := Run(ctx, GetAmsProvider(), req, res)
	if err != nil {
		log.Error(ctx, "do: Run failed",
			log.Err(err),
			log.Any("pager", pager),
			log.String("query", query),
			log.Any("operator", operator))
		return err
	}

	return nil
}

type argument struct {
	ConnectionName string
	FilterString   string
	NodeString     string
}

func toArgument(ctx context.Context, key FilterOfType, filter interface{}, res interface{}) (*argument, error) {
	filterString, err := marshalFilter(filter)
	if err != nil {
		log.Error(ctx, "argument: marshal failed",
			log.Err(err),
			log.Any("filter", filter))
		return nil, err
	}
	nodeFields, err := marshalFiled(ctx, res)

	if err != nil {
		log.Error(ctx, "argument: failed",
			log.Err(err),
			log.Any("filter", filter))
		return nil, err
	}
	return &argument{string(key), filterString, nodeFields}, nil
}

func queryString(ctx context.Context, key FilterOfType, filter interface{}, res interface{}) (string, error) {
	params, err := toArgument(ctx, key, filter, res)

	if err != nil {
		log.Error(ctx, "argument: failed",
			log.Err(err),
			log.Any("filter", filter))
		return "", err
	}

	temp, err := template.New("ProgramsConnection").Parse(graphQLString)
	if err != nil {
		log.Error(ctx, "string: template parse failed",
			log.Err(err),
			log.Any("filter", filter),
			log.Any("params", params),
			log.String("connection", graphQLString))
		return "", err
	}

	buf := bytes.Buffer{}
	err = temp.Execute(&buf, params)
	if err != nil {
		log.Error(ctx, "string: template execute failed",
			log.Err(err),
			log.Any("filter", filter),
			log.Any("params", params),
			log.String("connection", graphQLString))
		return "", err
	}
	return buf.String(), nil
}
