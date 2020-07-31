package model

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	dbclient "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"strconv"
	"sync"
	"time"
)

type ITagModel interface {
	Add(ctx context.Context, tag *entity.TagAddView) (string, error)
	Update(ctx context.Context, tag *entity.TagUpdateView) error
	Query(ctx context.Context, condition *da.TagCondition) ([]*entity.TagView, error)
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
	key:=make(map[string]*dynamodb.AttributeValue)
	key["id"] = &dynamodb.AttributeValue{
		S:    aws.String(tag.ID),
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

	_, err := dbclient.GetClient().UpdateItem(input)
	return utils.ConvertDynamodbError(err)
}

func (t tagModel) Query(ctx context.Context, condition *da.TagCondition) ([]*entity.TagView, error) {
	var filt expression.ConditionBuilder
	if len(condition.Name)!=0{
		filt = expression.Name("name").Equal(expression.Value(condition.Name))
	}
	proj:=expression.NamesList(expression.Name("id"),expression.Name("name"),expression.Name("states"),expression.Name("created_at"))
	expr,err:=expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err!=nil{
		return nil,err
	}
	params:=&dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(constant.TableNameTag),
		//Limit:                     nil,
		//ReturnConsumedCapacity:    nil,
		//ScanFilter:                nil,
		//Segment:                   nil,
		//Select:                    nil,
		//TotalSegments:             nil,
	}
	scanResult, err := dbclient.GetClient().Scan(params)
	if err!=nil{
		return nil,err
	}
	result:=make([]*entity.TagView,0)
	for _,i:=range scanResult.Items{
		tagItem:=&entity.TagView{}
		err = dynamodbattribute.UnmarshalMap(i, &tagItem)
		if err!=nil{
			return nil,err
		}
		result = append(result,tagItem)
	}
	return result, nil
}

func (t tagModel) GetByID(ctx context.Context, id string) (*entity.TagView, error) {
	result, err := t.getItem(id)
	err = utils.ConvertDynamodbError(err)
	return result, err
}

func (t tagModel) GetByName(ctx context.Context, name string) (*entity.TagView, error) {
	queryItems,_:=t.Query(ctx,&da.TagCondition{
		Name:     name,
		DeleteAt: 0,
	})
	if len(queryItems)>0{
		return queryItems[0],nil
	}
	return nil,constant.ErrRecordNotFound
}

func (t tagModel) getItem(id string) (*entity.TagView, error) {
	key:=make(map[string]*dynamodb.AttributeValue)
	key["id"] = &dynamodb.AttributeValue{
		S:    aws.String(id),
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
