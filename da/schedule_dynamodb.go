package da

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/dynamodbhelper"
	"strings"
	"sync"

	dbclient "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
)

type scheduleDynamoDA struct{}

func (s *scheduleDynamoDA) Insert(ctx context.Context, schedule *entity.Schedule) error {
	return s.BatchInsert(ctx, []*entity.Schedule{schedule})
}

func (s *scheduleDynamoDA) BatchInsert(ctx context.Context, schedules []*entity.Schedule) error {
	items := make(map[string][]*dynamodb.WriteRequest)
	itemsWriteRequest := make([]*dynamodb.WriteRequest, len(schedules))
	for i, item := range schedules {
		attributeValue, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			return err
		}
		request := &dynamodb.WriteRequest{
			PutRequest: &dynamodb.PutRequest{
				Item: attributeValue,
			},
		}
		itemsWriteRequest[i] = request
	}
	items[constant.TableNameSchedule] = itemsWriteRequest
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: items,
	}
	_, err := dbclient.GetClient().BatchWriteItem(input)
	return err
}

func (s *scheduleDynamoDA) Update(ctx context.Context, schedule *entity.Schedule) error {
	key := make(map[string]*dynamodb.AttributeValue)
	key["id"] = &dynamodb.AttributeValue{
		S: aws.String(schedule.ID),
	}
	// expr
	updateBuilder, err := dynamodbhelper.GetUpdateBuilder(schedule)
	if err != nil {
		return err
	}
	expr, err := expression.NewBuilder().WithUpdate(updateBuilder).Build()
	if err != nil {
		return err
	}
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Key:                       key,
		ReturnValues:              aws.String("ALL_NEW"),
		TableName:                 aws.String(constant.TableNameSchedule),
		UpdateExpression:          expr.Update(),
	}
	_, err = dbclient.GetClient().UpdateItem(input)
	return err
}

func (s *scheduleDynamoDA) Query(ctx context.Context, condition *dynamodbhelper.Condition) ([]*entity.Schedule, error) {
	keyCond := condition.KeyBuilder()
	//proj := expression.NamesList(expression.Name("title"), expression.Name("class_id"), expression.Name("teacher_ids"))
	expr, _ := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	input := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		//FilterExpression:          expr.Filter(),
		KeyConditionExpression: expr.KeyCondition(),
		TableName:              aws.String(constant.TableNameSchedule),
		IndexName:              aws.String(condition.IndexName),
	}
	result, err := dbclient.GetClient().Query(input)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var data []*entity.Schedule
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &data)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return data, nil
}

func (s *scheduleDynamoDA) Page(ctx context.Context, condition *ScheduleCondition) (string, []*entity.Schedule, error) {
	keyCond := condition.KeyBuilder()
	startKey, limit := condition.PageBuilder(constant.GSI_Schedule_OrgIDAndStartAt)
	expr, _ := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	input := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ExclusiveStartKey:         startKey,
		Limit:                     limit,
		IndexName:                 aws.String(condition.IndexName),
		TableName:                 aws.String(constant.TableNameSchedule),
	}
	result, err := dbclient.GetClient().Query(input)
	if err != nil {
		fmt.Println(err)
		return "", nil, err
	}

	var data []*entity.Schedule
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &data)
	if err != nil {
		fmt.Println(err)
		return "", nil, err
	}
	var lastkey string
	if len(data) > 0 {
		lastkey = s.getLastKey(constant.GSI_Schedule_OrgIDAndStartAt, data[len(data)-1])
	}

	return lastkey, data, nil
}

// TODO：Refactor
func (s *scheduleDynamoDA) getLastKey(indexType constant.GSIName, lastData *entity.Schedule) string {
	var key string
	pk := lastData.ID
	switch indexType {
	case constant.GSI_Schedule_OrgIDAndStartAt:
		key = fmt.Sprintf("%s,%s,%d", pk, lastData.OrgID, lastData.StartAt)
	}
	return key
}

func (s *scheduleDynamoDA) GetByID(ctx context.Context, id string) (*entity.Schedule, error) {
	key := make(map[string]*dynamodb.AttributeValue)
	key["id"] = &dynamodb.AttributeValue{
		S: aws.String(id),
	}
	input := &dynamodb.GetItemInput{
		Key:       key,
		TableName: aws.String(constant.TableNameSchedule),
	}
	result, err := dbclient.GetClient().GetItem(input)
	if err != nil {
		log.Error(ctx, "update schedule error", log.Err(err), log.String("id", id))
		return nil, err
	}
	schedule := new(entity.Schedule)
	err = dynamodbattribute.UnmarshalMap(result.Item, schedule)
	if err != nil {
		log.Error(ctx, "dynamodb unmarshalmap error", log.Err(err), log.String("id", id))
		return nil, err
	}
	return schedule, nil
}

func (s *scheduleDynamoDA) Delete(ctx context.Context, id string) error {
	in := dynamodb.DeleteItemInput{
		TableName: aws.String(entity.Schedule{}.TableName()),
		Key:       map[string]*dynamodb.AttributeValue{"id": {S: aws.String(id)}},
	}
	if _, err := dbclient.GetClient().DeleteItem(&in); err != nil {
		log.Error(ctx, "delete schedule: delete item failed", log.String("id", id))
		return err
	}
	return nil
}

func (s *scheduleDynamoDA) BatchDelete(ctx context.Context, ids []string) error {
	tableName := entity.Schedule{}.TableName()
	in := dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]*dynamodb.WriteRequest{},
	}
	var requestItems []*dynamodb.WriteRequest
	for _, id := range ids {
		requestItems = append(requestItems, &dynamodb.WriteRequest{
			PutRequest: &dynamodb.PutRequest{Item: map[string]*dynamodb.AttributeValue{"id": {S: aws.String(id)}}},
		})
	}
	for i := 0; i < len(requestItems); i++ {
		if i != 0 && i%25 == 0 {
			in.RequestItems = map[string][]*dynamodb.WriteRequest{tableName: requestItems[:25]}
			if _, err := dbclient.GetClient().BatchWriteItem(&in); err != nil {
				log.Error(ctx, "batch delete schedule: batch write item failed", log.Strings("ids", ids))
				return err
			}
			requestItems = requestItems[25:]
		}
	}
	if len(requestItems) > 0 {
		in.RequestItems = map[string][]*dynamodb.WriteRequest{tableName: requestItems}
		if _, err := dbclient.GetClient().BatchWriteItem(&in); err != nil {
			log.Error(ctx, "batch delete schedule: batch write item failed", log.Strings("ids", ids))
			return err
		}
	}
	return nil
}

var (
	_scheduleOnce sync.Once
	_scheduleDA   IScheduleDA
)

func GetScheduleDA() IScheduleDA {
	_scheduleOnce.Do(func() {
		_scheduleDA = &scheduleDynamoDA{}
	})
	return _scheduleDA
}

type ScheduleCondition struct {
	dynamodbhelper.Condition
}

func (s ScheduleCondition) PageBuilder(indexType constant.GSIName) (map[string]*dynamodb.AttributeValue, *int64) {
	limit := s.Pager.PageSize
	if limit <= 0 {
		limit = dynamodbhelper.DefaultPageSize
	}
	if strings.TrimSpace(s.Pager.LastKey) == "" {
		return nil, aws.Int64(limit)
	}
	var lastEvaluatedKey map[string]*dynamodb.AttributeValue
	keys := strings.Split(s.Pager.LastKey, ",")
	if len(keys) < 1 {
		return nil, aws.Int64(limit)
	}
	lastEvaluatedKey = map[string]*dynamodb.AttributeValue{
		"id": &dynamodb.AttributeValue{
			S: aws.String(keys[0]),
		},
	}
	switch indexType {
	//case constant.GSI_TeacherSchedule_TeacherAndStartAt:
	//	if len(keys) >= 4 {
	//		lastEvaluatedKey[s.PrimaryKey.Key] = &dynamodb.AttributeValue{
	//			S: aws.String(keys[2]),
	//		}
	//		lastEvaluatedKey[s.SortKey.Key] = &dynamodb.AttributeValue{
	//			N: aws.String(keys[3]),
	//		}
	//	}
	}
	return lastEvaluatedKey, aws.Int64(s.Pager.PageSize)
}
