package da

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	client "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strconv"
	"time"
)

type CategoryDynamoDA struct{}

func (c *CategoryDynamoDA) CreateCategory(ctx context.Context, data entity.CategoryObject) (*entity.CategoryObject, error) {
	av, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		log.Error(ctx, "CreateCategory marshal failed", log.Err(err))
		return nil, err
	}

	_, err = client.GetClient().PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(data.TableName()),
		Item:      av,
	})
	if err != nil {
		log.Error(ctx, "CreateCategory put item failed", log.Err(err))
		return nil, err
	}
	return &data, nil
}

func (c *CategoryDynamoDA) UpdateCategory(ctx context.Context, data entity.CategoryObject) error {
	upExpr, exprAttrName, exprAttrValue := data.ToUpdateParam()
	_, err := client.GetClient().UpdateItem(&dynamodb.UpdateItemInput{
		TableName:                 aws.String(entity.CategoryObject{}.TableName()),
		Key:                       data.ToKey(),
		ReturnValues:              aws.String("UPDATED_NEW"),
		UpdateExpression:          aws.String(upExpr),
		ExpressionAttributeNames:  exprAttrName,
		ExpressionAttributeValues: exprAttrValue,
	})
	if err != nil {
		log.Error(ctx, "UpdateCategory get item failed", log.Err(err))
		return err
	}
	return nil
}

func (c *CategoryDynamoDA) DeleteCategory(ctx context.Context, id string) error {
	category, err := c.GetCategoryByID(ctx, id)
	if err != nil && err != constant.ErrRecordNotFound {
		return err
	} else if err == constant.ErrRecordNotFound {
		return nil
	}

	if category.DeletedAt != 0 {
		return nil
	}

	_, err = client.GetClient().UpdateItem(&dynamodb.UpdateItemInput{
		TableName:        aws.String(entity.CategoryObject{}.TableName()),
		Key:              map[string]*dynamodb.AttributeValue{"id": {S: aws.String(id)}},
		ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String("set deleted_at = :del_at"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":del_at": {N: aws.String(strconv.FormatInt(time.Now().Unix(), 10))},
		},
	})
	if err != nil {
		log.Error(ctx, "DeleteCategory failed", log.Err(err))
		return err
	}
	return nil
}

func (c *CategoryDynamoDA) GetCategoryByID(ctx context.Context, id string) (*entity.CategoryObject, error) {
	output, err := client.GetClient().GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(entity.CategoryObject{}.TableName()),
		Key:       map[string]*dynamodb.AttributeValue{"id": {S: aws.String(id)}},
	})
	if err != nil {
		log.Error(ctx, "GetCategoryByID get item failed", log.Err(err))
		return nil, err
	}

	if output.Item == nil {
		log.Debug(ctx, "GetCategoryByID no record found")
		return nil, constant.ErrRecordNotFound
	}

	var category entity.CategoryObject
	err = dynamodbattribute.UnmarshalMap(output.Item, &category)
	if err != nil {
		log.Error(ctx, "GetCategoryByID unmarshal failed", log.Err(err))
		return nil, err
	}
	return &category, nil
}

func (c *CategoryDynamoDA) SearchCategories(ctx context.Context, condition *entity.SearchCategoryCondition) (int64, []*entity.CategoryObject, error) {
	expr, err := condition.ToExpr()
	if err != nil {
		log.Error(ctx, "SearchCategories build expression failed", log.Err(err))
		return 0, nil, err
	}
	input := &dynamodb.ScanInput{
		TableName:                 aws.String(entity.CategoryObject{}.TableName()),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		//ProjectionExpression:      expr.Projection(),
	}

	output, err := client.GetClient().Scan(input)
	if err != nil {
		log.Error(ctx, "SearchCategories scan failed", log.Err(err))
		return 0, nil, err
	}
	//var categories []*entity.CategoryObject
	categories := make([]*entity.CategoryObject, *output.Count)
	for k, i := range output.Items {
		var item entity.CategoryObject
		err = dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			log.Error(ctx, "SearchCategories unmarshal failed", log.Err(err))
			return 0, nil, err
		}
		categories[k] = &item
	}
	return *output.ScannedCount, categories, nil
}

func (c *CategoryDynamoDA) PageCategories(ctx context.Context, condition *entity.SearchCategoryCondition) (int64, []*entity.CategoryObject, error) {

	expr, err := condition.ToExpr()
	if err != nil {
		log.Error(ctx, "PageCategories build expression failed", log.Err(err))
		return 0, nil, err
	}

	input := &dynamodb.ScanInput{
		TableName:                 aws.String(entity.CategoryObject{}.TableName()),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		Limit:                     aws.Int64(condition.PageSize),
	}

	var page int64
	var count int64
	var categories []*entity.CategoryObject
	err = client.GetClient().ScanPages(input, func(output *dynamodb.ScanOutput, hasNoPage bool) bool {
		if page != condition.Page {
			page++
			return !hasNoPage
		}
		categories = make([]*entity.CategoryObject, *output.Count)
		for k, i := range output.Items {
			var item entity.CategoryObject
			err = dynamodbattribute.UnmarshalMap(i, &item)
			if err != nil {
				log.Error(ctx, "PageCategories unmarshal failed", log.Err(err))
			}
			categories[k] = &item
		}

		count = *output.ScannedCount
		return false
	})

	if err != nil {
		log.Error(ctx, "PageCategories failed", log.Err(err))
		return 0, nil, err
	}
	return count, categories, nil
}
