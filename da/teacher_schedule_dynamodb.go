package da

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	dbclient "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/dynamodbhelper"
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
	items[constant.TableNameTeacherSchedule] = itemsWriteRequest
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: items,
	}
	_, err := dbclient.GetClient().BatchWriteItem(input)
	return err
}

func (t *teacherScheduleDA) Update(ctx context.Context, data *entity.TeacherSchedule) error {
	panic("implement me")
}

func (t *teacherScheduleDA) BatchUpdate(ctx context.Context, data []*entity.TeacherSchedule) error {
	panic("implement me")
}

func (t *teacherScheduleDA) Delete(ctx context.Context, id string) error {
	panic("implement me")
}

func (t *teacherScheduleDA) BatchDelete(ctx context.Context, id []string) error {
	panic("implement me")
}

func (t *teacherScheduleDA) Page(ctx context.Context, condition dynamodbhelper.Condition) ([]*entity.TeacherSchedule, error) {
	keyCond := condition.KeyBuilder()
	startKey, limit := condition.PagerBuilder()
	//proj := expression.NamesList(expression.Name("title"), expression.Name("class_id"), expression.Name("teacher_ids"))
	expr, _ := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	input := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ExclusiveStartKey:         startKey,
		Limit:                     limit,
		IndexName:                 aws.String(condition.IndexName),
		TableName:                 aws.String(constant.TableNameTeacherSchedule),
	}
	result, err := dbclient.GetClient().Query(input)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var data []*entity.TeacherSchedule
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &data)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return data, nil
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
