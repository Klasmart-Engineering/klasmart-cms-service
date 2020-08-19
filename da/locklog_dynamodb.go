package da

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"strconv"
	"time"

	dynamodbclient "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type lockLogDA struct{}

func (da *lockLogDA) Insert(ctx context.Context, l *entity.LockLog) error {
	item, err := dynamodbattribute.MarshalMap(l)
	if err != nil {
		log.Error(ctx, "insert lock log: lock log marshal map failed", log.Err(err))
		return err
	}
	in := dynamodb.PutItemInput{
		TableName: aws.String(entity.LockLog{}.TableName()),
		Item:      item,
	}
	if _, err = dynamodbclient.GetClient().PutItem(&in); err != nil {
		log.Error(ctx, "insert lock log: put item to dynamodb failed", log.Err(err))
		return err
	}
	return nil
}

func (da *lockLogDA) GetByID(ctx context.Context, id string) (*entity.LockLog, error) {
	in := dynamodb.GetItemInput{
		Key:       map[string]*dynamodb.AttributeValue{"id": {S: aws.String(id)}},
		TableName: aws.String(entity.LockLog{}.TableName()),
	}
	out, err := dynamodbclient.GetClient().GetItem(&in)
	if err != nil {
		log.Error(ctx, "get lock log by id: get item failed from dynamodb", log.Err(err))
		return nil, err
	}
	if out.Item == nil {
		log.Debug(ctx, "get lock log by id: item is nil")
		return nil, constant.ErrRecordNotFound
	}
	var item entity.LockLog
	if err := dynamodbattribute.UnmarshalMap(out.Item, &item); err != nil {
		log.Error(ctx, "get lock log by id: unmarshal map failed", log.Err(err))
		return nil, err
	}
	if item.DeletedAt > 0 {
		log.Debug(ctx, "get lock log by id: item has deleted")
		return nil, constant.ErrRecordNotFound
	}
	return &item, nil
}

func (da *lockLogDA) GetByRecordID(ctx context.Context, recordID string) (*entity.LockLog, error) {
	in := dynamodb.QueryInput{
		IndexName:                 aws.String(entity.LockLog{}.IndexNameWithRecordIDAndCreatedAt()),
		KeyConditionExpression:    aws.String("record_id = :record_id"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{":record_id": {S: aws.String(recordID)}},
		ScanIndexForward:          aws.Bool(false),
		TableName:                 aws.String(entity.LockLog{}.TableName()),
	}
	out, err := dynamodbclient.GetClient().Query(&in)
	if err != nil {
		log.Error(ctx, "get lock log by record id: query failed from dynamodb", log.Err(err))
		return nil, err
	}
	if len(out.Items) == 0 {
		log.Debug(ctx, "get lock log by record id: items length is 0")
		return nil, constant.ErrRecordNotFound
	}
	var items []*entity.LockLog
	if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		err := errors.New("items length should greater than 0")
		log.Error(ctx, "get lock log by record id: unmarshal list of maps failed", log.Err(err))
		return nil, err
	}
	var result *entity.LockLog
	for _, item := range items {
		if item.DeletedAt == 0 {
			result = item
			break
		}
	}
	if result == nil {
		log.Debug(ctx, fmt.Sprintf("get lock log by record id: %d items have deleted", len(items)))
		return nil, constant.ErrRecordNotFound
	}
	return result, nil
}

func (da *lockLogDA) SoftDeleteByID(ctx context.Context, id string) error {
	in := dynamodb.UpdateItemInput{
		TableName:        aws.String(entity.LockLog{}.TableName()),
		Key:              map[string]*dynamodb.AttributeValue{"id": {S: aws.String(id)}},
		UpdateExpression: aws.String("set deleted_at = :deleted_at"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":deleted_at": {N: aws.String(strconv.FormatInt(time.Now().Unix(), 10))},
		},
	}
	if _, err := dynamodbclient.GetClient().UpdateItem(&in); err != nil {
		log.Error(ctx, "soft delete lock log by id: update item from dynamodb failed", log.Err(err))
		return err
	}
	return nil
}

func (da *lockLogDA) SoftDeleteByRecordID(ctx context.Context, recordID string) error {
	in := dynamodb.QueryInput{
		IndexName:              aws.String(entity.LockLog{}.IndexNameWithRecordIDAndCreatedAt()),
		KeyConditionExpression: aws.String("record_id = :record_id"),
		FilterExpression:       aws.String("deleted_at = :empty_deleted_at"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":record_id":        {S: aws.String(recordID)},
			":empty_deleted_at": {N: aws.String("0")},
		},
		ScanIndexForward: aws.Bool(false),
		TableName:        aws.String(entity.LockLog{}.TableName()),
	}
	out, err := dynamodbclient.GetClient().Query(&in)
	if err != nil {
		log.Error(ctx, "soft delete lock log by record id: query failed from dynamodb", log.Err(err))
		return err
	}
	if len(out.Items) == 0 {
		log.Debug(ctx, "soft delete lock log by record id: items length is 0")
		return nil
	}
	var items []*entity.LockLog
	if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &items); err != nil {
		log.Error(ctx, "soft delete lock log by record id: unmarshal list of maps failed", log.Err(err))
		return err
	}
	for _, item := range items {
		in := dynamodb.UpdateItemInput{
			TableName:        aws.String(entity.LockLog{}.TableName()),
			Key:              map[string]*dynamodb.AttributeValue{"id": {S: aws.String(item.ID)}},
			UpdateExpression: aws.String("set deleted_at = :deleted_at"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":deleted_at": {N: aws.String(strconv.FormatInt(time.Now().Unix(), 10))},
			},
		}
		if _, err := dynamodbclient.GetClient().UpdateItem(&in); err != nil {
			log.Error(ctx, "soft delete lock log by record id: update item from dynamodb failed", log.Err(err))
			return err
		}
	}
	return nil
}
