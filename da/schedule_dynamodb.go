package da

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"

	dbclient "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
)

type scheduleDynamoDA struct{}

func (s scheduleDynamoDA) Insert(ctx context.Context, schedule *entity.Schedule) error {
	return s.BatchInsert(ctx, []*entity.Schedule{schedule})
}

func (s scheduleDynamoDA) BatchInsert(ctx context.Context, schedules []*entity.Schedule) error {
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

func (s scheduleDynamoDA) Update(ctx context.Context, schedule *entity.Schedule) error {
	key := make(map[string]*dynamodb.AttributeValue)
	key["id"] = &dynamodb.AttributeValue{
		S: aws.String(schedule.ID),
	}
	// expr
	updateBuilder, err := utils.GetUpdateBuilder(schedule)
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
		TableName:                 aws.String("Music"),
		UpdateExpression:          expr.Update()}
	fmt.Println(input)
	//result, err := svc.UpdateItem(input)
	return nil
}

//func (s scheduleDynamoDA) BatchUpdate(ctx context.Context, schedule []*entity.Schedule) error {
//	panic("implement me")
//}

func (s scheduleDynamoDA) Query(ctx context.Context, condition *ScheduleCondition) ([]*entity.Schedule, error) {
	panic("implement me")
}

func (s scheduleDynamoDA) Page(ctx context.Context, condition *ScheduleCondition) (int64, []*entity.Schedule, error) {
	panic("implement me")
}

func (s scheduleDynamoDA) GetByID(ctx context.Context, id string) (*entity.Schedule, error) {
	panic("implement me")
}

func (s scheduleDynamoDA) SoftDelete(ctx context.Context, id string) error {
	panic("implement me")
}

func (s scheduleDynamoDA) BatchSoftDelete(ctx context.Context, op *entity.Operator, condition *ScheduleCondition) error {
	panic("implement me")
}

type ScheduleCondition struct {
	Pager utils.Pager

	DeleteAt entity.NullInt
}

func (s ScheduleCondition) GetCondition() (expression.Expression, error) {
	return expression.Expression{}, nil
}

var (
	_scheduleOnce sync.Once
	_scheduleDA   IScheduleDA
)

func GetScheduleDA() IScheduleDA {
	_scheduleOnce.Do(func() {
		_scheduleDA = scheduleDynamoDA{}
	})
	return _scheduleDA
}
