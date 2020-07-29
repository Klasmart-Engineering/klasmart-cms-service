package model

import (
	"calmisland/kidsloop2/constant"
	dbclient "calmisland/kidsloop2/dynamodb"
	"calmisland/kidsloop2/entity"
	"calmisland/kidsloop2/utils"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"strconv"
	"sync"
	"time"
)

type ITagModel interface {
	Add(ctx context.Context, tag *entity.TagAddView) (string, error)
	Update(ctx context.Context, tag *entity.TagUpdateView) error
	Query(ctx context.Context, condition *entity.TagCondition) ([]*entity.TagView, error)
	GetByID(ctx context.Context, id string) (*entity.TagView, error)
	GetByName(ctx context.Context, name string) (*entity.TagView, error)
}

type tagModel struct{}

var (
	_tagOnce  sync.Once
	_tagModel ITagModel
)

func GetTagModel() ITagModel {
	_tagOnce.Do(func() {
		_tagModel = &tagModel{}
	})
	return _tagModel
}

func (t tagModel) Add(ctx context.Context, tag *entity.TagAddView) (string, error) {
	old, err := t.GetByName(ctx, tag.Name)
	if err != nil && err != constant.ErrRecordNotFound {
		return "", err
	}
	if old != nil {
		return "", constant.ErrDuplicateRecord
	}
	svc := dbclient.GetClient()
	in := entity.Tag{
		ID:        utils.NewId(),
		Name:      tag.Name,
		States:    constant.Enable,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: 0,
		DeletedAt: 0,
	}
	item, err := dynamodbattribute.MarshalMap(in)
	if err != nil {
		return "", err
	}
	input := &dynamodb.PutItemInput{
		Item:                   item,
		ReturnConsumedCapacity: aws.String("TOTAL"),
		TableName:              aws.String(constant.TableNameTag),
	}

	_, err = svc.PutItem(input)
	err = utils.ConvertDynamodbError(err)
	return in.ID, err
}

func (t tagModel) Update(ctx context.Context, tag *entity.TagUpdateView) error {
	// key
	tagKey := entity.Tag{
		ID: tag.ID,
	}
	key, err := dynamodbattribute.MarshalMap(tagKey)
	if err != nil {
		return err
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
		ExpressionAttributeValues: params,
		Key:                       key,
		ReturnValues:              aws.String("UPDATED_NEW"),
		TableName:                 aws.String(constant.TableNameTag),
		UpdateExpression:          aws.String("set name = :n, states = :s, updated_at = :up"),
	}

	_, err = dbclient.GetClient().UpdateItem(input)
	return utils.ConvertDynamodbError(err)
}
//func (t tagModel) getConditions()[]expression.ConditionBuilder{
//	conditions := make([]expression.ConditionBuilder, 0)
//
//}
func (t tagModel) Query(ctx context.Context, condition *entity.TagCondition) ([]*entity.TagView, error) {

	return nil, nil
}

func (t tagModel) GetByID(ctx context.Context, id string) (*entity.TagView, error) {
	in := entity.Tag{
		ID: id,
	}
	result, err := t.getItem(in)
	err = utils.ConvertDynamodbError(err)
	return result, err
}

func (t tagModel) GetByName(ctx context.Context, name string) (*entity.TagView, error) {
	in := entity.Tag{
		Name: name,
	}
	result, err := t.getItem(in)
	err = utils.ConvertDynamodbError(err)
	return result, err
}

func (t tagModel) getItem(in entity.Tag) (*entity.TagView, error) {
	key, err := dynamodbattribute.MarshalMap(in)
	if err != nil {
		return nil, err
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
	tagView := &entity.TagView{
		ID:       tag.ID,
		Name:     tag.Name,
		CreateAt: tag.CreatedAt,
	}
	return tagView, nil
}
