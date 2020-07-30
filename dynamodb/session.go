package dynamodb

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var (
	client *dynamodb.DynamoDB
	_once  sync.Once
)

func GetClient() *dynamodb.DynamoDB {
	_once.Do(func() {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			Config: aws.Config{
				Region: aws.String("ap-northeast-2"),
				Endpoint: aws.String("http://192.168.1.234:18000"),
			},
			SharedConfigState: session.SharedConfigEnable,
		}))

		// Create DynamoDB client
		client = dynamodb.New(sess)
	})

	return client
}
