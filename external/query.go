package external

import (
	"bytes"
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"net/http"
	"sync"
	"text/template"
)

func pageQuery[ResType ConnectionResponse](ctx context.Context, operator *entity.Operator, filter ConnectionFilter, result *[]ResType) error {
	qString, err := queryString(ctx, string(filter.FilterName()), string(filter.ConnectionName()), result)
	if err != nil {
		log.Error(ctx, "query: string failed",
			log.Err(err),
			log.Any("filter", filter),
			log.Any("operator", operator))
		return err
	}
	var pageInfo *ConnectionPageInfo
	for pageInfo.HasNext(Forward) {
		res := GraphQLResponse[ResType]{
			Data: map[string]ResType{},
		}
		err := fetch(ctx, operator, filter, pageInfo.Pager(Forward, PageDefaultCount), qString, &res)
		if err != nil {
			log.Error(ctx, "query: fetch failed",
				log.Err(err),
				log.Any("filter", filter),
				log.Any("pageInfo", pageInfo),
				log.String("query", qString),
				log.Any("operator", operator))
			return err
		}

		log.Debug(ctx, "query: fetch success",
			log.Any("response", res),
			log.Any("filter", filter),
			log.Any("pageInfo", pageInfo),
			log.String("query", qString),
			log.Any("operator", operator))

		connections := res.Data[string(filter.ConnectionName())]
		pageInfo = connections.GetPageInfo()
		*result = append(*result, connections)
	}
	return nil
}

func fetch[ResType ConnectionResponse](ctx context.Context, operator *entity.Operator, filter interface{}, pager map[string]interface{}, query string, res *GraphQLResponse[ResType]) error {
	req := NewRequest(query, RequestToken(operator.Token))
	req.Var("filter", filter)
	for k, v := range pager {
		req.Var(k, v)
	}

	statusCode, err := GetAmsConnection().Run(ctx, req, res)

	if statusCode != http.StatusOK {
		log.Error(ctx, "fetch: http is not ok",
			log.Err(res.Errors),
			log.Int("status_code", statusCode),
			log.Any("pager", pager),
			log.String("query", query),
			log.Any("operator", operator))
		return err
	}

	if err != nil {
		log.Error(ctx, "fetch: Run failed",
			log.Err(err),
			log.Any("pager", pager),
			log.String("query", query),
			log.Any("operator", operator))
		return err
	}
	if res.Errors != nil {
		log.Error(ctx, "fetch: data has error",
			log.Err(res.Errors),
			log.Any("pager", pager),
			log.String("query", query),
			log.Any("operator", operator))
		return res.Errors
	}
	return nil
}

var graphQLString = `
query(
	$direction:ConnectionDirection!
	{{if .DirectionArgs}},$directionArgs:ConnectionsDirectionArgs{{end}}
	{{if .Filter}},$filter:{{.FilterName}}{{end}}
	{{if .Sort}},$sort:{{.SortName}}{{end}}
){
    {{.ConnectionName}}(
        direction:$direction,
        {{if .DirectionArgs}}directionArgs:$directionArgs,{{end}}
        {{if .Filter}}filter:$filter,{{end}}
        {{if .Sort}},sort:$sort{{end}}
	){
        {{.ResultString}}
    }
}
`
var temp *template.Template
var tempOnce sync.Once

func queryString(ctx context.Context, filterName string, connectionName string, res interface{}) (string, error) {
	nodeFields, err := marshalFiled(ctx, res)

	if err != nil {
		log.Error(ctx, "queryString2: marshalFiled failed",
			log.Err(err),
			log.String("connection_name", connectionName),
			log.String("filter_name", filterName))
		return "", err
	}

	connectionsArguments := TemplateArguments{
		DirectionArgs: true,
		Filter:        true,
		//Sort:           true,
		FilterName:     filterName,
		ConnectionName: connectionName,
		ResultString:   nodeFields,
	}

	tempOnce.Do(func() {
		temp, err = template.New("UseConnection").Parse(graphQLString)
	})
	if temp == nil {
		log.Error(ctx, "string: template parse failed",
			log.Err(err),
			log.String("connection_name", connectionName),
			log.String("filter_name", filterName),
			log.String("result", nodeFields),
			log.String("connection", graphQLString))
		return "", err
	}

	buf := bytes.Buffer{}
	err = temp.Execute(&buf, connectionsArguments)
	if err != nil {
		log.Error(ctx, "string: template execute failed",
			log.Err(err),
			log.String("connection_name", connectionName),
			log.String("filter_name", filterName),
			log.Any("params", connectionsArguments),
			log.String("connection", graphQLString))
		return "", err
	}
	return buf.String(), nil
}
