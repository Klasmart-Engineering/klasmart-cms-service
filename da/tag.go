package da

import (
	"context"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"sync"
)

type ITagDA interface{
	//Insert(context.Context, interface{}) (interface{}, error)
	//Update(context.Context, interface{}) (int64, error)
	//Page(context.Context, Conditions, interface{}) (int, error)
	//Query(context.Context, Conditions, interface{}) error
}

type tagDA struct{}

type TagCondition struct {
	Name     string
	PageSize int64
	Page     int64

	DeleteAt int
}
//aws dynamodb put-item \
//--table-name tags  \
//--item \
//'{"id": {"N": "2"}, "name": {"S": "Call Me Today"}, "states": {"N": "1"}, "createdAt": {"N": "0"},"updated_at": {"N": "0"},"deletedAt": {"N": "0"}}'
func (tagDA) Insert(ctx context.Context)(interface{},error){
	return nil,nil
}
//aws dynamodb create-table \
//--table-name tags \
//--attribute-definitions \
//AttributeName=id,AttributeType=S \
//--key-schema \
//AttributeName=id,KeyType=HASH \
//--provisioned-  \
//ReadCapacityUnits=10,WriteCapacityUnits=5

//aws dynamodb update-table \
//--table-name tags \
//--attribute-definitions AttributeName=name,AttributeType=S \
//--global-secondary-index-updates \
//"[{\"Create\":{\"IndexName\": \"name-index\",\"KeySchema\":[{\"AttributeName\":\"name\",\"KeyType\":\"HASH\"}], \
//\"ProvisionedThroughput\": {\"ReadCapacityUnits\": 10, \"WriteCapacityUnits\": 5      },\"Projection\":{\"ProjectionType\":\"ALL\"}}}]"

func (t TagCondition) GetConditions()[]expression.ConditionBuilder{
	//conditions := make([]expression.ConditionBuilder, 0)
	//if t.Name==""{
	//	condition := expression.
	//}
	return nil
}


var (
	_tagOnce sync.Once
	_tagDA ITagDA
)

func GetTagDA() ITagDA{
	_tagOnce.Do(func(){
		_tagDA = tagDA{}
	})
	return _tagDA
}

