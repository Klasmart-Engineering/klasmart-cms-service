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
	dbclient "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/dynamodbhelper"
	"strings"
	"sync"
)

type teacherScheduleDA struct{}

func (t *teacherScheduleDA) Add(ctx context.Context, data *entity.TeacherSchedule) error {
	return t.BatchAdd(ctx, []*entity.TeacherSchedule{data})
}

func (t *teacherScheduleDA) BatchAdd(ctx context.Context, datalist []*entity.TeacherSchedule) error {
	items := make(map[string][]*dynamodb.WriteRequest)
	itemsWriteRequest := make([]*dynamodb.WriteRequest, len(datalist))
	for i, item := range datalist {
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
	length := len(itemsWriteRequest)
	for i := 0; i < length; i++ {
		if i != 0 && i%25 == 0 {
			items[constant.TableNameTeacherSchedule] = itemsWriteRequest[:25]
			input := &dynamodb.BatchWriteItemInput{
				RequestItems: items,
			}
			_, err := dbclient.GetClient().BatchWriteItem(input)
			if err != nil {
				log.Error(ctx, "batch add teacher_schedule: batch write item failed", log.Any("items", items))
				return err
			}
			itemsWriteRequest = itemsWriteRequest[25:]
		}
	}
	if len(itemsWriteRequest) > 0 {
		items[constant.TableNameTeacherSchedule] = itemsWriteRequest
		input := &dynamodb.BatchWriteItemInput{
			RequestItems: items,
		}
		_, err := dbclient.GetClient().BatchWriteItem(input)
		if err != nil {
			log.Error(ctx, "batch add teacher_schedule: batch write item failed", log.Any("items", items))
			return err
		}
	}
	return nil
}

func (t *teacherScheduleDA) Delete(ctx context.Context, teacherID string, scheduleID string) error {
	in := dynamodb.DeleteItemInput{
		TableName: aws.String(entity.TeacherSchedule{}.TableName()),
		Key: map[string]*dynamodb.AttributeValue{
			"teacher_id":  {S: aws.String(teacherID)},
			"schedule_id": {S: aws.String(scheduleID)},
		},
	}
	if _, err := dbclient.GetClient().DeleteItem(&in); err != nil {
		log.Error(ctx, "delete teacher_schedule: delete item failed",
			log.String("teacher_id", teacherID),
			log.String("schedule_id", scheduleID),
		)
		return err
	}
	return nil
}

func (t *teacherScheduleDA) BatchDelete(ctx context.Context, pks [][2]string) error {
	tableName := entity.TeacherSchedule{}.TableName()
	in := dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]*dynamodb.WriteRequest{},
	}
	var requestItems []*dynamodb.WriteRequest
	for _, pk := range pks {
		requestItems = append(requestItems, &dynamodb.WriteRequest{
			DeleteRequest: &dynamodb.DeleteRequest{
				Key: map[string]*dynamodb.AttributeValue{
					"teacher_id":  {S: aws.String(pk[0])},
					"schedule_id": {S: aws.String(pk[1])},
				},
			},
		})
	}
	for i := 0; i < len(requestItems); i++ {
		if i != 0 && i%25 == 0 {
			in.RequestItems = map[string][]*dynamodb.WriteRequest{tableName: requestItems[:25]}
			if _, err := dbclient.GetClient().BatchWriteItem(&in); err != nil {
				log.Error(ctx, "batch delete teacher_schedule: batch write item failed", log.Any("pks", pks))
				return err
			}
			requestItems = requestItems[25:]
		}
	}
	if len(requestItems) > 0 {
		in.RequestItems = map[string][]*dynamodb.WriteRequest{tableName: requestItems}
		if _, err := dbclient.GetClient().BatchWriteItem(&in); err != nil {
			log.Error(ctx, "batch delete teacher_schedule: batch write item failed", log.Any("pks", pks))
			return err
		}
	}
	return nil
}

func (t *teacherScheduleDA) Page(ctx context.Context, condition TeacherScheduleCondition) (string, []*entity.TeacherSchedule, error) {
	keyCond := condition.QueryBuilder()
	startKey, limit := condition.PageBuilder()

	//proj := expression.NamesList(expression.Name("title"), expression.Name("class_id"), expression.Name("teacher_ids"))
	expr, _ := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	input := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ExclusiveStartKey:         startKey,
		Limit:                     limit,
		IndexName:                 aws.String(condition.IndexName.String()),
		TableName:                 aws.String(constant.TableNameTeacherSchedule),
	}

	result, err := dbclient.GetClient().Query(input)
	if err != nil {
		log.Error(ctx, "query error", log.Any("input", input))
		return "", nil, err
	}

	var data []*entity.TeacherSchedule
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &data)
	if err != nil {
		log.Error(ctx, "UnmarshalListOfMaps error", log.Any("data", data))
		return "", nil, err
	}
	var lastkey string
	if len(data) > 0 {
		lastkey = t.getLastKey(constant.GSI_TeacherSchedule_TeacherAndStartAt, data[len(data)-1])
	}

	return lastkey, data, nil
}

// TODO：Refactor
func (s *teacherScheduleDA) getLastKey(indexType constant.GSIName, lastData *entity.TeacherSchedule) string {
	var key string
	pk := fmt.Sprintf("%s,%s", lastData.TeacherID, lastData.ScheduleID)
	switch indexType {
	case constant.GSI_TeacherSchedule_TeacherAndStartAt:
		key = fmt.Sprintf("%s,%s,%d", pk, lastData.TeacherID, lastData.StartAt)
	}
	return key
}

var (
	_teacherScheduleOnce sync.Once
	_teacherScheduleDA   ITeacherScheduleDA
)

func GetTeacherScheduleDA() ITeacherScheduleDA {
	_teacherScheduleOnce.Do(func() {
		_teacherScheduleDA = &teacherScheduleDA{}
	})
	return _teacherScheduleDA
}

type TeacherScheduleCondition struct {
	TeacherID string
	StartAt   int64
	dynamodbhelper.Condition
}

func (c *TeacherScheduleCondition) Init(gsiName constant.GSIName, compareType dynamodbhelper.CompareType) {
	c.IndexName = gsiName
	c.CompareType = compareType

	switch c.IndexName {
	case constant.GSI_TeacherSchedule_TeacherAndStartAt:
		c.Condition.PrimaryKey = dynamodbhelper.KeyValue{
			Key:   "teacher_id",
			Value: c.TeacherID,
		}
		c.Condition.SortKey = dynamodbhelper.KeyValue{
			Key:   "start_at",
			Value: c.StartAt,
		}
	}
}

func (c *TeacherScheduleCondition) QueryBuilder() expression.KeyConditionBuilder {
	return c.KeyBuilder()
}

func (c *TeacherScheduleCondition) PageBuilder() (map[string]*dynamodb.AttributeValue, *int64) {
	limit := c.Condition.Pager.PageSize
	if limit <= 0 {
		limit = dynamodbhelper.DefaultPageSize
	}
	if strings.TrimSpace(c.Condition.Pager.LastKey) == "" {
		return nil, aws.Int64(limit)
	}

	var lastEvaluatedKey map[string]*dynamodb.AttributeValue
	keys := strings.Split(c.Condition.Pager.LastKey, ",")
	if len(keys) < 2 {
		return nil, aws.Int64(limit)
	}

	lastEvaluatedKey = map[string]*dynamodb.AttributeValue{
		"teacher_id": &dynamodb.AttributeValue{
			S: aws.String(keys[0]),
		},
		"schedule_id": &dynamodb.AttributeValue{
			S: aws.String(keys[1]),
		},
	}
	switch c.IndexName {
	case constant.GSI_TeacherSchedule_TeacherAndStartAt:
		if len(keys) >= 4 {
			lastEvaluatedKey[c.Condition.PrimaryKey.Key] = &dynamodb.AttributeValue{
				S: aws.String(keys[2]),
			}
			lastEvaluatedKey[c.Condition.SortKey.Key] = &dynamodb.AttributeValue{
				N: aws.String(keys[3]),
			}
		}
	}
	return lastEvaluatedKey, aws.Int64(c.Condition.Pager.PageSize)
}
