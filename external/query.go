package external

import (
	"bytes"
	"context"
	"fmt"
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

var subQueryString = `
query(
{{range $k, $v := .VariableName}}
	${{$k}}:ID!
	${{$v}}:String
{{end}}
){
{{range $k, $v := .VariableName}}
  {{$k}}: {{$.NodeName}}(id:${{$k}}){
    {{$.ConnectionName}}(count:50, cursor: ${{$v}})
	{{$.FieldStructString}}
  }
{{end}}
}`

var subTemp *template.Template
var subTempOnce sync.Once

type SubTemplateArgument struct {
	VariableName      map[string]string
	NodeName          string
	ConnectionName    string
	FieldStructString string
}

func querySubString(ctx context.Context, templateArgument SubTemplateArgument) (string, error) {
	var err error
	subTempOnce.Do(func() {
		subTemp, err = template.New("UseNode").Parse(subQueryString)
	})
	if err != nil {
		log.Error(ctx, "string: template parse failed",
			log.Err(err),
			log.Any("params", templateArgument),
			log.String("connection", subQueryString))
		return "", err
	}

	buf := bytes.Buffer{}
	err = subTemp.Execute(&buf, templateArgument)
	if err != nil {
		log.Error(ctx, "string: template execute failed",
			log.Err(err),
			log.Any("params", templateArgument),
			log.String("connection", graphQLString))
		return "", err
	}
	return buf.String(), nil
}

func subFetch(ctx context.Context, operator *entity.Operator, argument SubTemplateArgument, IDVariable, CursorVariable map[string]string, res *GraphQLSubResponse) error {
	qString, err := querySubString(ctx, argument)
	if err != nil {
		log.Error(ctx, "subFetch: querySubString failed",
			log.Err(err),
			log.Any("ids", IDVariable),
			log.Any("cursors", CursorVariable))
		return err
	}
	req := NewRequest(qString, RequestToken(operator.Token))
	for k, v := range IDVariable {
		req.Var(k, v)
	}
	for k, v := range CursorVariable {
		req.Var(k, v)
	}

	statusCode, err := GetAmsConnection().Run(ctx, req, &res)

	if statusCode != http.StatusOK {
		log.Error(ctx, "subFetch: http is not ok",
			log.Err(res.Errors),
			log.Int("status_code", statusCode),
			log.Any("operator", operator))
		return err
	}

	if err != nil {
		log.Error(ctx, "subFetch: Run failed",
			log.Err(err),
			log.Any("operator", operator))
		return err
	}
	if res.Errors != nil {
		log.Error(ctx, "subFetch: data has error",
			log.Err(res.Errors),
			log.Any("operator", operator))
		return res.Errors
	}
	return nil
}

func subPageQuery[ResType ConnectionResponse](ctx context.Context, operator *entity.Operator, nodeName, connectionName string, IDs []string, result map[string][]ResType) error {
	nodeFields, err := marshalFiled(ctx, result)
	if err != nil {
		log.Error(ctx, "subPageQuery: marshalFiled failed",
			log.Err(err),
			log.Strings("ids", IDs),
			log.String("connection_name", connectionName),
			log.String("nodeName", nodeName))
		return err
	}
	templateArgument := SubTemplateArgument{
		VariableName:      make(map[string]string),
		NodeName:          nodeName,
		ConnectionName:    connectionName,
		FieldStructString: nodeFields,
	}

	IDVariable := make(map[string]string)
	CursorVariable := make(map[string]string)

	for k, v := range IDs {
		IDKey := fmt.Sprintf("key%d", k)
		CursorKey := fmt.Sprintf("cursor%d", k)
		templateArgument.VariableName[IDKey] = CursorKey

		IDVariable[IDKey] = v
		CursorVariable[CursorKey] = ""
	}

	for len(templateArgument.VariableName) > 0 {
		data := make(map[string]map[string]ResType)
		res := GraphQLSubResponse{
			Data: &data,
		}
		err = subFetch(ctx, operator, templateArgument, IDVariable, CursorVariable, &res)
		if err != nil {
			log.Error(ctx, "subPageQuery: subFetch failed",
				log.Err(err),
				log.Any("template", templateArgument),
				log.Any("id_variable", IDVariable),
				log.Any("cursor_variable", CursorVariable))
			return err
		}

		variableName := make(map[string]string)
		IDsMap := make(map[string]string)
		CursorsMap := make(map[string]string)
		for k, v := range data {
			if page, ok := v[connectionName]; ok {
				if page.GetPageInfo().HasNext(Forward) {
					cursor := page.GetPageInfo().EndCursor
					variableName[k] = templateArgument.VariableName[k]
					IDsMap[k] = IDVariable[k]
					CursorsMap[templateArgument.VariableName[k]] = cursor
				}
				keyID := IDVariable[k]
				result[keyID] = append(result[keyID], page)
			}
		}
		templateArgument.VariableName = variableName
		IDVariable = IDsMap
		CursorVariable = CursorsMap
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
