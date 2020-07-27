package dynamodb

import(
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"sync"
)

var (
	client *dynamodb.DynamoDB
	_once sync.Once
)

func GetClient() *dynamodb.DynamoDB{
	_once.Do(func() {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			Config: aws.Config{
				Region: aws.String("ap-northeast-2"),
			},
			SharedConfigState: session.SharedConfigEnable,
		}))

		// Create DynamoDB client
		client = dynamodb.New(sess)
	})

	return client
}
