package model

import (
	client "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/log"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"strconv"
	"sync"
	"time"
)

type ICategoryModel interface {
	CreateCategory(ctx context.Context, data entity.CategoryObject) (string, error)
	UpdateCategory(ctx context.Context, data entity.CategoryObject) error
	DeleteCategory(ctx context.Context, id string) error
	GetCategoryById(ctx context.Context, id string) (*entity.CategoryObject, error)
	
	SearchCategories(ctx context.Context, condition *SearchCategoryCondition) (int64, []*entity.CategoryObject, error)
}

type CategoryModel struct{}


// Repeated insertion with the same primary key will overwrite non-primary key data
func (cm *CategoryModel) CreateCategory(ctx context.Context, data entity.CategoryObject) (string, error) {
	now := time.Now().Unix()
	data.ID = "id_test2" // TODO: change to real id
	data.CreatedAt = now
	data.UpdatedAt = now

	av, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		log.Get().Errorf("CreateCategory marshal failed: %v", err)
		return "", err
	}

	_, err = client.GetClient().PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(data.TableName()),
		Item:      av,
	})
	if err != nil {
		log.Get().Errorf("CreateCategory put item failed: %v", err)
		return "", err
	}
	return data.ID, nil
}

func (cm *CategoryModel) UpdateCategory(ctx context.Context, data entity.CategoryObject) error {
	upExpr, exprAttrName, exprAttrValue := data.ToUpdateParam()
	fmt.Println("upExpr: ", upExpr)
	output, err := client.GetClient().UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(entity.CategoryObject{}.TableName()),
		Key: data.ToKey(),
		ReturnValues: aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String(upExpr),
		ExpressionAttributeNames: exprAttrName,
		ExpressionAttributeValues: exprAttrValue,
	})
	if err != nil {
		log.Get().Errorf("UpdateCategory get item failed: %v", err)
		return err
	}
	fmt.Printf("%+v\n", output)
	return nil
}

func (cm *CategoryModel) DeleteCategory(ctx context.Context, id string) error {
	//output, err := client.GetClient().DeleteItem(&dynamodb.DeleteItemInput{
	//	TableName: aws.String("niobium"),
	//	Key: map[string]*dynamodb.AttributeValue{
	//		"id": {S: aws.String("id")},
	//	},
	//})
	_, err := client.GetClient().UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(entity.CategoryObject{}.TableName()),
		Key: map[string]*dynamodb.AttributeValue{"id": {S: aws.String(id)}},
		ReturnValues: aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String("set deleted_at = :del_at"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":del_at": {N: aws.String(strconv.FormatInt(time.Now().Unix(), 10))},
		},
	})
	if err != nil {
		log.Get().Errorf("DeleteCategory failed: %v", err)
		return err
	}
	return nil
}

func (cm *CategoryModel) GetCategoryById(ctx context.Context, id string) (*entity.CategoryObject, error) {
	output, err := client.GetClient().GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(entity.CategoryObject{}.TableName()),
		Key: map[string]*dynamodb.AttributeValue{"id": {S: aws.String(id)}},
	})
	if err != nil {
		log.Get().Errorf("GetCategoryById get item failed: %v", err)
		return nil, err
	}
	var category entity.CategoryObject
	err = dynamodbattribute.UnmarshalMap(output.Item, &category)
	if err != nil {
		log.Get().Errorf("GetCategoryById unmarshal failed: %v", err)
		return nil, err
	}
	return &category, nil
}

func (cm *CategoryModel) SearchCategories(ctx context.Context, condition *SearchCategoryCondition) (int64, []*entity.CategoryObject, error) {

	expr, err := condition.toExpr()
	if err != nil {
		log.Get().Errorf("SearchCategories build expression failed: %v", err)
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
	fmt.Printf("%+v", output)
	if err != nil {
		log.Get().Errorf("SearchCategories scan failed: %v", err)
		return 0, nil, err
	}
	var categories []*entity.CategoryObject
	for _, i := range output.Items {
		var item entity.CategoryObject
		err = dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			log.Get().Errorf("SearchCategories unmarshal failed: %v", err)
			return 0, nil, err
		}
		categories = append(categories, &item)
	}
	return *output.ScannedCount, categories, nil
}

var categoryModel *CategoryModel
var _categoryOnce sync.Once

func GetCategoryModel() ICategoryModel {
	_categoryOnce.Do(func() {
		categoryModel = new(CategoryModel)
	})
	return categoryModel
}

type SearchCategoryCondition struct {
	IDs        []string `json:"ids"`
	Names      []string `json:"names"`

	PageSize int64 `json:"page_size"`
	Page     int64 `json:"page"`
	OrderBy	 string `json:"order_by"`
}

func (s *SearchCategoryCondition) toExpr() (expression.Expression, error) {
	condition := expression.Name("deleted_at").Equal(expression.Value(0))
	if len(s.IDs) > 0 {
		var exprValues []expression.OperandBuilder
		for _, id := range s.IDs {
			exprValues = append(exprValues, expression.Value(id))
		}
		condition = condition.And(expression.Name("id").In(exprValues[0], exprValues...))
	}
	if len(s.Names) > 0 {
		var exprValues []expression.OperandBuilder
		for _, name := range s.Names {
			exprValues = append(exprValues, expression.Value(name))
		}
		condition = condition.And(expression.Name("name").In(exprValues[0], exprValues...))
	}

	expr, err := expression.NewBuilder().WithFilter(condition).Build()
	return expr, err
}