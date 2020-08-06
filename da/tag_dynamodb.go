package da

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	dbclient "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"strconv"
	"strings"
	"sync"
	"time"
)


type tagDynamoDA struct{}

func (tagDynamoDA) Insert(ctx context.Context, tag *entity.Tag) error {
	svc := dbclient.GetClient()
	item, err := dynamodbattribute.MarshalMap(tag)
	if err != nil {
		log.Error(ctx, "dynamodb marshalmap error", log.Err(err), log.String("tagID", tag.ID), log.String("tagName", tag.Name))
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(constant.TableNameTag),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		log.Error(ctx, "insert tag error", log.Err(err), log.String("tagID", tag.ID), log.String("tagName", tag.Name))
		return err
	}
	return nil
}

func (tagDynamoDA) Query(ctx context.Context, condition *TagCondition) ([]*entity.Tag, error) {
	expr, err := condition.GetCondition()
	if err != nil {
		log.Error(ctx, "get tag condition error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(constant.TableNameTag),
	}
	scanResult, err := dbclient.GetClient().Scan(params)

	if err != nil {
		log.Error(ctx, "scan tag by condition error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}

	result := make([]*entity.Tag, len(scanResult.Items))
	for i, item := range scanResult.Items {
		tagItem := &entity.Tag{}
		err = dynamodbattribute.UnmarshalMap(item, &tagItem)
		if err != nil {
			log.Error(ctx, "dynamodb unmarshalmap error", log.Err(err), log.Any("condition", condition))
			return nil, err
		}
		result[i] = tagItem
	}

	return result, nil
}

func (tagDynamoDA) Page(ctx context.Context, condition *TagCondition) (int64, []*entity.Tag, error) {
	expr, err := condition.GetCondition()
	if err != nil {
		log.Error(ctx, "get tag condition error", log.Err(err), log.Any("condition", condition))
		return 0, nil, err
	}

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(constant.TableNameTag),
		Limit:                     aws.Int64(condition.Pager.PageSize),
	}
	result := make([]*entity.Tag, 0)

	var pageIndex int64 = 1
	var count int64
	err = dbclient.GetClient().ScanPages(params, func(page *dynamodb.ScanOutput, lastPage bool) bool {
		if pageIndex == condition.Pager.PageIndex {
			for _, item := range page.Items {
				tagItem := &entity.Tag{}
				err = dynamodbattribute.UnmarshalMap(item, tagItem)
				if err != nil {
					log.Error(ctx, "dynamodb unmarshalmap error", log.Err(err))
					return false
				}
				result=append(result,tagItem)
			}
			count = *page.ScannedCount
		}

		pageIndex++
		if lastPage {
			return false
		}
		return true
	})

	return count, result, err
}

func (tagDynamoDA) Update(ctx context.Context, tag *entity.Tag) error {
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
	if err != nil {
		log.Error(ctx, "update tag error", log.Err(err), log.Any("taginfo", tag))
		return err
	}
	return nil
}

func (tagDynamoDA) GetByID(ctx context.Context, id string) (*entity.Tag, error) {
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
		log.Error(ctx, "update tag error", log.Err(err), log.String("id", id))
		return nil, err
	}
	tag := new(entity.Tag)
	err = dynamodbattribute.UnmarshalMap(result.Item, tag)
	if err != nil {
		log.Error(ctx, "dynamodb unmarshalmap error", log.Err(err), log.String("id", id))
		return nil, err
	}
	return tag, nil
}
func (tagDynamoDA) GetByIDs(ctx context.Context, ids []string) ([]*entity.Tag, error) {
	keys := make([]map[string]*dynamodb.AttributeValue, 0)
	for _, id := range ids {
		if strings.TrimSpace(id) != "" {
			keymap := make(map[string]*dynamodb.AttributeValue)
			keymap["id"] = &dynamodb.AttributeValue{
				S: aws.String(id),
			}
			keys = append(keys, keymap)
		}
	}

	attributes := map[string]*dynamodb.KeysAndAttributes{
		constant.TableNameTag: {
			Keys: keys,
		},
	}
	input := &dynamodb.BatchGetItemInput{
		RequestItems: attributes,
	}
	result, err := dbclient.GetClient().BatchGetItem(input)
	if err != nil {
		log.Error(ctx, "get tag by ids", log.Err(err), log.Strings("id", ids))
		return nil, err
	}
	tagList := result.Responses[constant.TableNameTag]
	tags := make([]*entity.Tag, len(tagList))
	err = dynamodbattribute.UnmarshalListOfMaps(tagList, &tags)
	if err != nil {
		log.Error(ctx, "dynamodb unmarshalmap error", log.Err(err))
		return nil, err
	}

	return tags, nil
}

func (tagDynamoDA) Delete(ctx context.Context, id string) error {
	key := make(map[string]*dynamodb.AttributeValue)
	key["id"] = &dynamodb.AttributeValue{
		S: aws.String(id),
	}
	input := &dynamodb.DeleteItemInput{
		Key:       key,
		TableName: aws.String(constant.TableNameTag),
	}
	_, err := dbclient.GetClient().DeleteItem(input)
	if err != nil {
		log.Error(ctx, "dynamodb delete tag error", log.Err(err), log.String("id", id))
		return err
	}
	return nil
}

var (
	_tagOnce sync.Once
	_tagDA   ITagDA
)

func GetTagDA() ITagDA {
	_tagOnce.Do(func() {
		_tagDA = tagDynamoDA{}
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
