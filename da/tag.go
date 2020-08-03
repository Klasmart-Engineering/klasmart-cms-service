package da

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	dbclient "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ITagDA interface {
	Insert(ctx context.Context, tag entity.Tag) error
	Update(ctx context.Context, tag entity.Tag) error
	//Page(context.Context, Conditions, interface{}) (int, error)
	Query(ctx context.Context, condition TagCondition) ([]*entity.Tag, error)
	GetByID(ctx context.Context, id string) (*entity.Tag, error)
	GetByIDs(ctx context.Context,ids []string)([]*entity.Tag,error)
	Delete(ctx context.Context, id string) error
	Page(ctx context.Context, condition TagCondition) (int64,[]*entity.Tag, error)
}

type tagDA struct{}

type TagCondition struct {
	Name     string
	PageSize int64
	Page     int64

	DeleteAt int
}

func (tagDA) Insert(ctx context.Context, tag entity.Tag) error {
	svc := dbclient.GetClient()
	item, err := dynamodbattribute.MarshalMap(tag)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(constant.TableNameTag),
	}

	_, err = svc.PutItem(input)
	return nil
}

func (tagDA) Query(ctx context.Context, condition TagCondition) ([]*entity.Tag, error) {
	expr, err := condition.GetCondition()
	if err != nil {
		return nil, err
	}
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(constant.TableNameTag),
		//Limit:                     nil,
		//ReturnConsumedCapacity:    aws.String("TOTAL"),
		//ScanFilter:                nil,
		//Segment:                   nil,
		//Select:                    nil,
		//TotalSegments:             nil,
	}
	scanResult, err := dbclient.GetClient().Scan(params)

	if err != nil {
		return nil, err
	}

	result := make([]*entity.Tag, 0)
	for _, i := range scanResult.Items {
		tagItem := &entity.Tag{}
		err = dynamodbattribute.UnmarshalMap(i, &tagItem)
		if err != nil {
			return nil, err
		}
		result = append(result, tagItem)
	}

	return result, nil
}

func (tagDA) Page(ctx context.Context, condition TagCondition) (int64,[]*entity.Tag, error) {
	expr, err := condition.GetCondition()
	if err != nil {
		return 0,nil, err
	}
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(constant.TableNameTag),
		Limit:                     aws.Int64(condition.PageSize),
		ReturnConsumedCapacity:    aws.String("TOTAL"),
		//ReturnConsumedCapacity:    nil,
		//ScanFilter:                nil,
		//Segment:                   nil,
		//Select:                    nil,
		//TotalSegments:             nil,
	}
	result := make([]*entity.Tag, 0)

	var pageNum int64
	var count int64
	err = dbclient.GetClient().ScanPages(params, func(page *dynamodb.ScanOutput, lastPage bool) bool {
		count += aws.Int64Value(page.Count)

		if pageNum == condition.Page {
			for _, i := range page.Items {
				tagItem := &entity.Tag{}
				err = dynamodbattribute.UnmarshalMap(i, &tagItem)
				if err != nil {
					return false
				}
				result = append(result, tagItem)
			}
			return false
		}
		if lastPage {
			return false
		}
		pageNum++
		return true
	})
	return count,result, err
}

func (tagDA) Update(ctx context.Context, tag entity.Tag) error {
	key := make(map[string]*dynamodb.AttributeValue)
	key["id"] = &dynamodb.AttributeValue{
		S: aws.String(tag.ID),
	}
	// expr
	params := make(map[string]*dynamodb.AttributeValue)
	params[":n"] = &dynamodb.AttributeValue{
		S: aws.String(tag.Name),
	}
	statesStr := strconv.Itoa(tag.States)
	params[":s"] = &dynamodb.AttributeValue{
		N: aws.String(statesStr),
	}
	updateTimeStr := strconv.FormatInt(time.Now().Unix(), 10)
	params[":up"] = &dynamodb.AttributeValue{
		N: aws.String(updateTimeStr),
	}

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#n":  aws.String("name"),
			"#s":  aws.String("states"),
			"#up": aws.String("updated_at"),
		},
		ExpressionAttributeValues: params,
		Key:                       key,
		ReturnValues:              aws.String("UPDATED_NEW"),
		TableName:                 aws.String(constant.TableNameTag),
		UpdateExpression:          aws.String("set #n = :n, #s = :s, #up = :up"),
	}

	_, err := dbclient.GetClient().UpdateItem(input)

	return err
}

func (tagDA) GetByID(ctx context.Context, id string) (*entity.Tag, error) {
	key := make(map[string]*dynamodb.AttributeValue)
	key["id"] = &dynamodb.AttributeValue{
		S: aws.String(id),
	}
	input := &dynamodb.GetItemInput{
		Key:       key,
		TableName: aws.String(constant.TableNameTag),
	}
	result, err := dbclient.GetClient().GetItem(input)
	if err != nil {
		return nil, err
	}
	tag := new(entity.Tag)
	err = dynamodbattribute.UnmarshalMap(result.Item, tag)
	if err != nil {
		return nil, err
	}
	return tag, nil
}
func (tagDA)GetByIDs(ctx context.Context,ids []string)([]*entity.Tag,error){
	keys:=make([]map[string]*dynamodb.AttributeValue,0)
	for _,id:=range ids{
		if strings.TrimSpace(id)!=""{
			keymap:=make(map[string]*dynamodb.AttributeValue)
			keymap["id"] = &dynamodb.AttributeValue{
				S: aws.String(id),
			}
			keys = append(keys,keymap)
		}
	}

	 attributes := map[string]*dynamodb.KeysAndAttributes{
		constant.TableNameTag:{
			Keys:keys,
		},
	}
	input := &dynamodb.BatchGetItemInput{
		RequestItems: attributes,
	}
	result, err := dbclient.GetClient().BatchGetItem(input)
	if err !=nil{
		return nil,err
	}
	tagList:=result.Responses[constant.TableNameTag]
	tags := make([]*entity.Tag,0)
	err=dynamodbattribute.UnmarshalListOfMaps(tagList,&tags)
	if err !=nil{
		return nil,err
	}

	return tags, nil
}

func (tagDA) Delete(ctx context.Context, id string) error {
	key := make(map[string]*dynamodb.AttributeValue)
	key["id"] = &dynamodb.AttributeValue{
		S: aws.String(id),
	}
	input := &dynamodb.DeleteItemInput{
		Key:       key,
		TableName: aws.String(constant.TableNameTag),
	}
	_, err := dbclient.GetClient().DeleteItem(input)
	return err
}

func (t TagCondition) GetCondition() (expression.Expression, error) {
	var filt expression.ConditionBuilder
	if t.DeleteAt != 0 {
		filt = expression.Name("deleted_at").NotEqual(expression.Value(0))
	} else {
		filt = expression.Name("deleted_at").Equal(expression.Value(0))
	}
	if len(t.Name) != 0 {
		filt = filt.And(expression.Name("name").Equal(expression.Value(t.Name)))
	}
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return expression.Expression{}, err
	}
	return expr, nil
}

var (
	_tagOnce sync.Once
	_tagDA   ITagDA
)

func GetTagDA() ITagDA {
	_tagOnce.Do(func() {
		_tagDA = tagDA{}
	})
	return _tagDA
}

//aws dynamodb create-table \
//--endpoint-url http://192.168.1.234:18000 \
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

//aws dynamodb put-item \
//--table-name tags  \
//--item \
//'{"id": {"N": "2"}, "name": {"S": "Call Me Today"}, "states": {"N": "1"}, "createdAt": {"N": "0"},"updated_at": {"N": "0"},"deletedAt": {"N": "0"}}'
