package da

import (
	"context"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	dbclient "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/dynamodbhelper"
	"sync"
)

type teacherScheduleDA struct{}

func (t teacherScheduleDA) Add(ctx context.Context, data *entity.TeacherSchedule) error {
	return t.BatchAdd(ctx, []*entity.TeacherSchedule{data})
}

func (t teacherScheduleDA) BatchAdd(ctx context.Context, datalist []*entity.TeacherSchedule) error {
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

func (t teacherScheduleDA) Update(ctx context.Context, data *entity.TeacherSchedule) error {

	panic("implement me")
}

func (t teacherScheduleDA) BatchUpdate(ctx context.Context, data []*entity.TeacherSchedule) error {
	panic("implement me")
}

func (t teacherScheduleDA) Delete(ctx context.Context, id string) error {

	panic("implement me")
}

func (t teacherScheduleDA) BatchDelete(ctx context.Context, id []string) error {
	panic("implement me")
}

func (t teacherScheduleDA) Page(ctx context.Context, condition dynamodbhelper.Condition) ([]*entity.TeacherSchedule, error) {
	panic("implement me")
}

var (
	_teacherScheduleOnce sync.Once
	_teacherScheduleDA   ITeacherScheduleDA
)

func GetTeacherScheduleDA() ITeacherScheduleDA {
	_teacherScheduleOnce.Do(func() {
		_teacherScheduleDA = teacherScheduleDA{}
	})
	return _teacherScheduleDA
}
