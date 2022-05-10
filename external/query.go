package external

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"text/template"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

var graphQLString = `
query($direction:ConnectionDirection!, $cursor:String!, $count:PageSize){
    {{.ConnectionName}}(
        direction:$direction,
        directionArgs:{cursor:$cursor, count: $count},
        {{if .FilterString}}filter:{{.FilterString}},{{end}}
        sort:{field:id,order:ASC}) {
        {{.NodeString}}
    }
}
`

func pageQuery[ResType ConnectionResponse](ctx context.Context, operator *entity.Operator, key FilterOfType, filter interface{}, result *[]ResType) error {
	qString, err := queryString(ctx, key, filter, result)
	if err != nil {
		log.Error(ctx, "query: string failed",
			log.Err(err),
			log.String("key", string(key)),
			log.Any("filter", filter),
			log.Any("operator", operator))
		return err
	}
	var pageInfo *ConnectionPageInfo
	for pageInfo.HasNext(Forward) {
		res := GraphQLResponse[ResType]{
			Data: map[string]ResType{},
		}
		err := fetch(ctx, operator, pageInfo.Pager(Forward, PageDefaultCount), qString, &res)
		if err != nil {
			log.Error(ctx, "query: do failed",
				log.Err(err),
				log.String("key", string(key)),
				log.Any("filter", filter),
				log.Any("pageInfo", pageInfo),
				log.String("query", qString),
				log.Any("operator", operator))
			return err
		}

		log.Debug(ctx, "query: do failed",
			log.Any("response", res),
			log.String("key", string(key)),
			log.Any("filter", filter),
			log.Any("pageInfo", pageInfo),
			log.String("query", qString),
			log.Any("operator", operator))

		connections := res.Data[string(key)]
		pageInfo = connections.GetPageInfo()
		*result = append(*result, connections)
	}
	return nil
}

func fetch[ResType ConnectionResponse](ctx context.Context, operator *entity.Operator, pager map[string]interface{}, query string, res *GraphQLResponse[ResType]) error {
	req := NewRequest(query, RequestToken(operator.Token))
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
			log.String("key", string(key)),
			log.Any("filter", filter))
		return nil, err
	}
	nodeFields, err := marshalFiled(ctx, res)

	if err != nil {
		log.Error(ctx, "argument: failed",
			log.Err(err),
			log.String("key", string(key)),
			log.Any("filter", filter))
		return nil, err
	}
	return &argument{string(key), filterString, nodeFields}, nil
}

var temp *template.Template
var tempOnce sync.Once

func queryString(ctx context.Context, key FilterOfType, filter interface{}, res interface{}) (string, error) {
	params, err := toArgument(ctx, key, filter, res)

	if err != nil {
		log.Error(ctx, "argument: failed",
			log.Err(err),
			log.String("key", string(key)),
			log.Any("filter", filter))
		return "", err
	}

	tempOnce.Do(func() {
		temp, err = template.New("UseConnection").Parse(graphQLString)
	})
	if temp == nil {
		log.Error(ctx, "string: template parse failed",
			log.Err(err),
			log.String("key", string(key)),
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
			log.String("key", string(key)),
			log.Any("filter", filter),
			log.Any("params", params),
			log.String("connection", graphQLString))
		return "", err
	}
	return buf.String(), nil
}
