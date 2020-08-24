package da

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	db "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	ErrContentNotFound        = errors.New("content not found")
	ErrUnmarshalContentFailed = errors.New("unmarshal content failed")
)

type IDyContentDA interface {
	CreateContent(ctx context.Context, co entity.Content) (string, error)
	UpdateContent(ctx context.Context, cid string, co entity.Content) error
	DeleteContent(ctx context.Context, cid string) error
	GetContentById(ctx context.Context, cid string) (*entity.Content, error)

	SearchContent(ctx context.Context, condition IDyCondition) (string, []*entity.Content, error)
	SearchContentByKey(ctx context.Context, condition DyKeyContentCondition) (string, []*entity.Content, error)
}

type DyContentDA struct {
}

func (d *DyContentDA) CreateContent(ctx context.Context, co entity.Content) (string, error) {
	now := time.Now()
	co.ID = utils.NewID()
	co.OrgUserId = co.Org + co.ID
	co.ContentTypeOrgIdPublishStatus = fmt.Sprintf("%v%v%v", co.ContentType, co.Org, co.PublishStatus)
	co.UpdatedAt = now.Unix()
	co.CreatedAt = now.Unix()
	dyMap, err := dynamodbattribute.MarshalMap(co)
	if err != nil {
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

func (d *DyContentDA) UpdateContent(ctx context.Context, cid string, co0 entity.Content) error {
	now := time.Now()
	co0.UpdatedAt = now.Unix()
	co, err := d.getContentForUpdateContent(ctx, cid, &co0)
	if err != nil {
		return err
	}
	dyMap, err := dynamodbattribute.MarshalMap(co)
	if err != nil {
		return err
	}
	key, err := dynamodbattribute.Marshal(cid)
	if err != nil {
		return err
	}
	_, err = db.GetClient().UpdateItem(&dynamodb.UpdateItemInput{
		TableName:                 aws.String("content"),
		ExpressionAttributeValues: dyMap,
		UpdateExpression:          aws.String(entity.Content{}.UpdateExpress()),
		Key:                       map[string]*dynamodb.AttributeValue{"content_id": key},
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *DyContentDA) DeleteContent(ctx context.Context, cid string) error {
	key, err := dynamodbattribute.Marshal(cid)
	if err != nil {
		return err
	}
	_, err = db.GetClient().DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("content"),
		Key:       map[string]*dynamodb.AttributeValue{"content_id": key},
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *DyContentDA) GetContentById(ctx context.Context, cid string) (*entity.Content, error) {
	key, err := dynamodbattribute.Marshal(cid)
	if err != nil {
		return nil, err
	}
	result, err := db.GetClient().GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("content"),
		Key:       map[string]*dynamodb.AttributeValue{"content_id": key},
	})
	if err != nil {
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

func (d *DyContentDA) SearchContent(ctx context.Context, condition IDyCondition) (string, []*entity.Content, error) {
	expr, err := expression.NewBuilder().WithFilter(condition.GetConditions()).Build()
	if err != nil {
		return "", nil, err
	}
	pageSize := condition.GetPageSize()
	if pageSize < 1 {
		pageSize = 10000
	}

	input := &dynamodb.ScanInput{
		TableName:                 aws.String("content"),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		Limit:                     aws.Int64(pageSize),
	}

	if condition.GetLastKey() != "" {
		key := &entity.ContentID{ID: condition.GetLastKey()}
		startKey, err := dynamodbattribute.MarshalMap(key)
		if err != nil {
			return "", nil, err
		}
		input.ExclusiveStartKey = startKey
	}
	result, err := db.GetClient().Scan(input)
	if err != nil {
		return "", nil, err
	}
	contentList := make([]*entity.Content, 0)
	for _, v := range result.Items {
		content := new(entity.Content)
		err = dynamodbattribute.UnmarshalMap(v, content)
		if err != nil {
			return "", nil, err
		}
		contentList = append(contentList, content)
	}
	key := new(entity.ContentID)
	dynamodbattribute.UnmarshalMap(result.LastEvaluatedKey, key)
	return key.ID, contentList, nil
}

func (d *DyContentDA) SearchContentByKey(ctx context.Context, condition DyKeyContentCondition) (string, []*entity.Content, error) {
	index, cond := condition.GetConditions()
	expr, err := expression.NewBuilder().WithKeyCondition(cond).Build()
	if err != nil {
		return "", nil, err
	}
	pageSize := condition.PageSize
	if pageSize < 1 {
		pageSize = 10000
	}
	input := &dynamodb.QueryInput{
		TableName:                 aws.String("content"),
		IndexName:                 aws.String(index),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      expr.Projection(),
		KeyConditionExpression:    expr.KeyCondition(),
		Limit:                     aws.Int64(pageSize),
	}
	if condition.LastKey != "" {
		key := &entity.ContentID{ID: condition.LastKey}
		startKey, err := dynamodbattribute.MarshalMap(key)
		if err != nil {
			return "", nil, err
		}
		input.ExclusiveStartKey = startKey
	}
	result, err := db.GetClient().Query(input)
	if err != nil {
		return "", nil, err
	}
	contentList := make([]*entity.Content, 0)
	for _, v := range result.Items {
		content := new(entity.Content)
		err = dynamodbattribute.UnmarshalMap(v, content)
		if err != nil {
			return "", nil, err
		}
		contentList = append(contentList, content)
	}
	key := new(entity.ContentID)
	dynamodbattribute.UnmarshalMap(result.LastEvaluatedKey, key)
	return key.ID, contentList, nil
}

func (d *DyContentDA) getContentForUpdateContent(ctx context.Context, cid string, co *entity.Content) (*entity.UpdateDyContent, error) {
	content, err := d.GetContentById(ctx, cid)
	if err != nil {
		return nil, err
	}
	co0 := &entity.UpdateDyContent{
		Name:          co.Name,
		Program:       co.Program,
		Subject:       co.Subject,
		Developmental: co.Developmental,
		Skills:        co.Skills,
		Age:           co.Age,
		Keywords:      co.Keywords,
		Description:   co.Description,
		Thumbnail:     co.Thumbnail,
		Data:          co.Data,
		Extra:         co.Extra,
		Author:        co.Author,
		AuthorName:    co.AuthorName,
		Org:           co.Org,
		PublishScope:  co.PublishScope,
		PublishStatus: co.PublishStatus,
		RejectReason:  co.RejectReason,
		SourceId:      co.SourceId,
		LatestId:      co.LatestId,
		LockedBy:      co.LockedBy,
		Version:       co.Version,
		CreatedAt:     co.CreatedAt,
		UpdatedAt:     co.UpdatedAt,
		DeletedAt:     co.DeletedAt,
	}
	co0.OrgUserId = co.Org + co.ID
	co0.ContentTypeOrgIdPublishStatus = fmt.Sprintf("%v%v%v", co.ContentType, co.Org, co.PublishStatus)
	if co.ContentType == 0 {
		co0.ContentType = content.ContentType
	}
	if co.Name == "" {
		co0.Name = content.Name
	}
	if co.Program == "" {
		co0.Program = content.Program
	}
	if co.Subject == "" {
		co0.Subject = content.Subject
	}
	if co.Developmental == "" {
		co0.Developmental = content.Developmental
	}
	if co.Skills == "" {
		co0.Skills = content.Skills
	}
	if co.Age == "" {
		co0.Age = content.Age
	}
	if co.Keywords == "" {
		co0.Keywords = content.Keywords
	}
	if co.Description == "" {
		co0.Description = content.Description
	}
	if co.Thumbnail == "" {
		co0.Thumbnail = content.Thumbnail
	}
	if co.Data == "" {
		co0.Data = content.Data
	}
	if co.Extra == "" {
		co0.Extra = content.Extra
	}
	if co.Author == "" {
		co0.Author = content.Author
	}
	if co.AuthorName == "" {
		co0.AuthorName = content.AuthorName
	}
	if co.Org == "" {
		co0.Org = content.Org
	}
	if co.PublishScope == "" {
		co0.PublishScope = content.PublishScope
	}
	if co.PublishStatus == "" {
		co0.PublishStatus = content.PublishStatus
	}
	if co.RejectReason == "" {
		co0.RejectReason = content.RejectReason
	}
	if co.SourceId == "" {
		co0.SourceId = content.SourceId
	}
	if co.LatestId == "" {
		co0.LatestId = content.LatestId
	}
	if co.LockedBy == "" {
		co0.LockedBy = content.LockedBy
	}
	if co.Version == 0 {
		co0.Version = content.Version
	}
	if co.OrgUserId == "" {
		co0.OrgUserId = content.OrgUserId
	}
	if co.CreatedAt == 0 {
		co0.CreatedAt = content.CreatedAt
	}
	if co.UpdatedAt == 0 {
		co0.UpdatedAt = content.UpdatedAt
	}
	if co.DeletedAt == 0 {
		co0.DeletedAt = content.DeletedAt
	}
	fmt.Printf("content: %#v\n", co)
	return co0, nil
}

type DyContentCondition struct {
	IDS           []string `json:"ids"`
	Name          string   `json:"name"`
	ContentType   []int    `json:"content_type"`
	Scope         []string `json:"scope"`
	PublishStatus []string `json:"publish_status"`
	Author        string   `json:"author"`
	Org           string   `json:"org"`

	LastKey  string `json:"last_key"`
	PageSize int64  `json:"page_size"`
	//OrderBy  ContentOrderBy `json:"order_by"`
}

func (d *DyContentCondition) GetConditions() expression.ConditionBuilder {
	conditions := make([]expression.ConditionBuilder, 0)
	if len(d.IDS) > 0 {
		var op []expression.OperandBuilder
		for i := range d.IDS {
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
		for i := range d.ContentType {
			op = append(op, expression.Value(d.ContentType[i]))
		}
		condition := expression.In(expression.Name("content_type"), op[0], op...)
		conditions = append(conditions, condition)
	}
	if len(d.Scope) > 0 {
		var op []expression.OperandBuilder
		for i := range d.Scope {
			op = append(op, expression.Value(d.Scope[i]))
		}
		condition := expression.In(expression.Name("publish_scope"), op[0], op...)
		conditions = append(conditions, condition)
	}
	if len(d.PublishStatus) > 0 {
		var op []expression.OperandBuilder
		for i := range d.PublishStatus {
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
	builder := expression.Name("deleted_at").Equal(expression.Value(nil))
	for i := range conditions {
		builder = builder.And(conditions[i])
	}
	return builder
}
func (d *DyContentCondition) GetPageSize() int64 {
	return d.PageSize
}
func (d *DyContentCondition) GetLastKey() string {
	return d.LastKey
}

type DyCombineContentCondition struct {
	Condition1 IDyCondition
	Condition2 IDyCondition
	PageSize   int64
	LastKey    string
}

func (d *DyCombineContentCondition) GetConditions() expression.ConditionBuilder {
	var builder expression.ConditionBuilder
	builder = d.Condition1.GetConditions().Or(d.Condition2.GetConditions())
	return builder
}
func (d *DyCombineContentCondition) GetPageSize() int64 {
	return d.PageSize
}
func (d *DyCombineContentCondition) GetLastKey() string {
	return d.LastKey
}

type DyKeyContentCondition struct {
	Name        string `json:"name"`
	AuthorName  string `json:"author_name"`
	Description string `json:"description"`
	KeyWords    string `json:"key_words"`

	PublishStatus string `json:"publish_status"`
	Author        string `json:"author"`
	Org           string `json:"org"`

	OrgUserId string `json:"org_user_id"`

	ContentTypeOrgIdPublishStatus string `json:"content_type_org_id_publish_status"`

	LastKey  string `json:"last_key"`
	PageSize int64  `json:"page_size"`
}

func (d *DyKeyContentCondition) GetConditions() (string, expression.KeyConditionBuilder) {
	var builder expression.KeyConditionBuilder
	var index string
	if d.PublishStatus != "" {
		builder = expression.KeyEqual(expression.Key("publish_status"), expression.Value(d.PublishStatus))
		if d.Org != "" {
			condition := expression.KeyEqual(expression.Key("org"), expression.Value(d.Org))
			builder = builder.And(condition)
		}
		index = "publish_status"
	}
	if d.Author != "" {
		builder = expression.KeyEqual(expression.Key("author"), expression.Value(d.Author))
		if d.Org != "" {
			condition := expression.KeyEqual(expression.Key("org"), expression.Value(d.Org))
			builder = builder.And(condition)
		}
		index = "author"
	}

	if d.Name != "" {
		builder = expression.KeyEqual(expression.Key("content_name"), expression.Value(d.Name))
		if d.OrgUserId != "" {
			fmt.Println(d.OrgUserId)
			condition := expression.KeyEqual(expression.Key("org_user_id"), expression.Value(d.OrgUserId))
			builder = builder.And(condition)
		}
		index = "name"
	}

	if d.AuthorName != "" {
		builder = expression.KeyEqual(expression.Key("author_name"), expression.Value(d.AuthorName))
		if d.OrgUserId != "" {
			fmt.Println(d.OrgUserId)
			condition := expression.KeyEqual(expression.Key("org_user_id"), expression.Value(d.OrgUserId))
			builder = builder.And(condition)
		}
		index = "author_name"
	}

	if d.Description != "" {
		builder = expression.KeyEqual(expression.Key("description"), expression.Value(d.Description))
		if d.OrgUserId != "" {
			fmt.Println(d.OrgUserId)
			condition := expression.KeyEqual(expression.Key("org_user_id"), expression.Value(d.OrgUserId))
			builder = builder.And(condition)
		}
		index = "description"
	}

	if d.KeyWords != "" {
		builder = expression.KeyEqual(expression.Key("keywords"), expression.Value(d.KeyWords))
		if d.OrgUserId != "" {
			condition := expression.KeyEqual(expression.Key("org_user_id"), expression.Value(d.OrgUserId))
			builder = builder.And(condition)
		}
		index = "keywords"
	}

	if d.ContentTypeOrgIdPublishStatus != "" {
		builder = expression.KeyEqual(expression.Key("ctoips"), expression.Value(d.ContentTypeOrgIdPublishStatus))
		if d.Name != "" {
			condition := expression.KeyEqual(expression.Key("content_name"), expression.Value(d.Name))
			builder = builder.And(condition)
		}
		index = "ctoips"
	}

	return index, builder
}

type IDyCondition interface {
	GetConditions() expression.ConditionBuilder
	GetPageSize() int64
	GetLastKey() string
}

var (
	_dyContentDA     IDyContentDA
	_dyContentDAOnce sync.Once
)

func GetDyContentDA() IDyContentDA {
	_dyContentDAOnce.Do(func() {
		_dyContentDA = new(DyContentDA)
	})
	return _dyContentDA
}
