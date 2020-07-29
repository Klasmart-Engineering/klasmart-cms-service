package kidsloop2

import(
	"calmisland/kidsloop2/dynamodb"
	"calmisland/kidsloop2/entity"
	"calmisland/kidsloop2/log"
	"calmisland/kidsloop2/route"
	"calmisland/kidsloop2/storage"
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_"calmisland/kidsloop2/conf"
	_ "calmisland/kidsloop2/route"
	"net/http"
)

func response(statusCode int, body interface{}) events.APIGatewayProxyResponse{
	data, err := json.Marshal(body)
	if err != nil{
		log.Get().Errorf("Can't marshal response body, err: %v", err)
	}
	return events.APIGatewayProxyResponse{
		StatusCode:        statusCode,
		Body:              string(data),
	}}

func doLambda(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	res, err := route.Route(ctx, request.PathParameters["action"], []byte(request.Body))
	if err != nil{
		return response(http.StatusInternalServerError, entity.ErrMsg{
			ErrMsg: "Internal server error",
		}), err
	}
	return response(res.StatusCode, res.StatusMsg), nil

	//return response(http.StatusNotFound, "Status not found"), nil
}

func main() {
	//获取数据库连接
	dynamodb.GetClient()
	storage.DefaultStorage()

	lambda.Start(doLambda)
}
