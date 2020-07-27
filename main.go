package kidsloop2

import(
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_"calmisland/kidsloop2/conf"
)

func response(statusCode int, body string) events.APIGatewayProxyResponse{
	return events.APIGatewayProxyResponse{
		StatusCode:        statusCode,
		Body:              body,
	}
}

func doLambda(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return response(200, "success"), nil
}

func main() {
	lambda.Start(doLambda)
}
