package da

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	db "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
	"time"
)

var(
	ErrContentNotFound = errors.New("content not found")
	ErrUnmarshalContentFailed = errors.New("unmarshal content failed")
)

type IDyContentDA interface {
	CreateContent(ctx context.Context, co entity.Content) (string, error)
	UpdateContent(ctx context.Context, cid string, co entity.UpdateDyContent) error
	DeleteContent(ctx context.Context, cid string) error
	GetContentById(ctx context.Context, cid string) (*entity.Content, error)

	SearchContent(ctx context.Context, condition DyContentCondition) (string, []*entity.Content, error)
}

type DyContentDA struct {
}

func (d *DyContentDA) CreateContent(ctx context.Context, co entity.Content) (string, error) {
	now := time.Now()
	co.ID = utils.NewID()
	co.UpdatedAt = &now
	co.CreatedAt = &now
	dyMap, err := dynamodbattribute.MarshalMap(co)
	if err != nil{
		return "", err
	}
	_, err = db.GetClient().PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("content"),
		Item:      dyMap,
	})
	if err != nil {
		return "", err
	}
	return co.ID, nil
}

func (d *DyContentDA) UpdateContent(ctx context.Context, cid string, co entity.UpdateDyContent) error {
	now := time.Now()
	co.UpdatedAt = &now
	err := d.getContentForUpdateContent(ctx, cid, &co)
	if err != nil{
		return err
	}
	dyMap, err := dynamodbattribute.MarshalMap(co)
	if err != nil{
		return err
	}
	key, err := dynamodbattribute.Marshal(cid)
	if err != nil{
		return err
	}
	_, err = db.GetClient().UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("content"),
		ExpressionAttributeValues: dyMap,
		UpdateExpression: aws.String(entity.Content{}.UpdateExpress()),
		Key: map[string]*dynamodb.AttributeValue{"content_id": key},
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *DyContentDA) DeleteContent(ctx context.Context, cid string) error {
	key, err := dynamodbattribute.Marshal(cid)
	if err != nil{
		return err
	}
	_, err = db.GetClient().DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("content"),
		Key: map[string]*dynamodb.AttributeValue{"content_id": key},
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *DyContentDA) GetContentById(ctx context.Context, cid string) (*entity.Content, error) {
	key, err := dynamodbattribute.Marshal(cid)
	if err != nil{
		return nil, err
	}
	result, err := db.GetClient().GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("content"),
		Key:                      map[string]*dynamodb.AttributeValue{"content_id": key},
	})
	if err != nil{
		return nil, err
	}
	if result.Item == nil {
		return nil, ErrContentNotFound
	}
	content := new(entity.Content)

	err = dynamodbattribute.UnmarshalMap(result.Item, content)
	if err != nil {
		return nil, ErrUnmarshalContentFailed
	}
	return content, nil
}

func (d *DyContentDA) SearchContent(ctx context.Context, condition DyContentCondition) (string, []*entity.Content, error) {
	expr, err := expression.NewBuilder().WithFilter(condition.GetConditions()).Build()
	if err != nil{
		return "", nil, err
	}
	if condition.PageSize < 1 {
		condition.PageSize = 10000
	}
	input := &dynamodb.ScanInput{
		TableName: aws.String("content"),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		Limit: 			aws.Int64(condition.PageSize),
	}

	if condition.LastKey != "" {
		key := &entity.ContentID{ID: condition.LastKey}
		startKey, err := dynamodbattribute.MarshalMap(key)
		if err != nil{
			return "", nil, err
		}
		input.ExclusiveStartKey = startKey
	}

	result, err := db.GetClient().Scan(input)
	if err != nil{
		return "", nil, err
	}
	contentList := make([]*entity.Content, 0)
	for _, v := range result.Items {
		content := new(entity.Content)
		err = dynamodbattribute.UnmarshalMap(v, content)
		if err != nil{
			return "", nil, err
		}
		contentList = append(contentList, content)
	}
	key := new(entity.ContentID)
	dynamodbattribute.UnmarshalMap(result.LastEvaluatedKey, key)
	return key.ID, contentList ,nil
}

func (d *DyContentDA) getContentForUpdateContent(ctx context.Context, cid string, co *entity.UpdateDyContent) error{
	content, err := d.GetContentById(ctx, cid)
	if err != nil{
		return err
	}
	if co.ContentType == 0 {
		co.ContentType = content.ContentType
	}
	if co.Name == "" {
		co.Name = content.Name
	}
	if co.Program == "" {
		co.Program = content.Program
	}
	if co.Subject == "" {
		co.Subject = content.Subject
	}
	if co.Developmental == "" {
		co.Developmental = content.Developmental
	}
	if co.Skills == "" {
		co.Skills = content.Skills
	}
	if co.Age == "" {
		co.Age = content.Age
	}
	if co.Keywords == ""{
		co.Keywords = content.Keywords
	}
	if co.Description == "" {
		co.Description = content.Description
	}
	if co.Thumbnail == "" {
		co.Thumbnail = content.Thumbnail
	}
	if co.Data == "" {
		co.Data = content.Data
	}
	if co.Extra == "" {
		co.Extra = content.Extra
	}
	if co.Author == "" {
		co.Author = content.Author
	}
	if co.AuthorName == "" {
		co.AuthorName = content.AuthorName
	}
	if co.Org == "" {
		co.Org = content.Org
	}
	if co.PublishScope == "" {
		co.PublishScope = content.PublishScope
	}
	if co.PublishStatus == "" {
		co.PublishStatus = content.PublishStatus
	}
	if co.RejectReason == "" {
		co.RejectReason = content.RejectReason
	}
	if co.Version == 0 {
		co.Version = content.Version
	}
	if co.CreatedAt == nil {
		co.CreatedAt = content.CreatedAt
	}
	if co.UpdatedAt == nil {
		co.UpdatedAt = content.UpdatedAt
	}
	if co.DeletedAt == nil {
		co.DeletedAt = content.DeletedAt
	}
	fmt.Printf("content: %#v\n", co)
	return nil
}
type DyContentCondition struct {
	IDS          []string `json:"ids"`
	Name         string `json:"name"`
	ContentType  []int `json:"content_type"`
	Scope        []string `json:"scope"`
	PublishStatus []string `json:"publish_status"`
	Author      string `json:"author"`
	Org    		string `json:"org"`

	LastKey string `json:"last_key"`
	PageSize int64 `json:"page_size"`
	//OrderBy  ContentOrderBy `json:"order_by"`
}

func (d *DyContentCondition)GetConditions() expression.ConditionBuilder{
	conditions := make([]expression.ConditionBuilder, 0)
	if len(d.IDS) > 0 {
		var op []expression.OperandBuilder
		for i := range d.IDS{
			op = append(op, expression.Value(d.IDS[i]))
		}
		condition := expression.In(expression.Name("content_id"), op[0], op...)
		conditions = append(conditions, condition)
	}
	if d.Name != "" {
		condition := expression.Name("name").Equal(expression.Value(d.Name))
		conditions = append(conditions, condition)
	}
	if len(d.ContentType) > 0 {
		var op []expression.OperandBuilder
		for i := range d.ContentType{
			op = append(op, expression.Value(d.ContentType[i]))
		}
		condition := expression.In(expression.Name("content_type"), op[0], op...)
		conditions = append(conditions, condition)
	}
	if len(d.Scope) > 0 {
		var op []expression.OperandBuilder
		for i := range d.Scope{
			op = append(op, expression.Value(d.Scope[i]))
		}
		condition := expression.In(expression.Name("publish_scope"), op[0], op...)
		conditions = append(conditions, condition)
	}
	if len(d.PublishStatus) > 0 {
		var op []expression.OperandBuilder
		for i := range d.PublishStatus{
			op = append(op, expression.Value(d.PublishStatus[i]))
		}
		condition := expression.In(expression.Name("publish_status"), op[0], op...)
		conditions = append(conditions, condition)
	}
	if d.Author != "" {
		condition := expression.Name("author").Equal(expression.Value(d.Author))
		conditions = append(conditions, condition)
	}
	if d.Org != "" {
		condition := expression.Name("org").Equal(expression.Value(d.Org))
		conditions = append(conditions, condition)
	}
	var builder expression.ConditionBuilder
	for i := range conditions {
		if i == 0{
			builder = conditions[i]
			continue
		}
		builder = builder.And(conditions[i])
	}
	return builder
}

var(
	_dyContentDA IDyContentDA
	_dyContentDAOnce sync.Once
)

func GetDyContentDA() IDyContentDA {
	_dyContentDAOnce.Do(func() {
		_dyContentDA = new(DyContentDA)
	})
	return _dyContentDA
}
