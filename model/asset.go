package model

import (
	client "calmisland/kidsloop2/dynamodb"
	"calmisland/kidsloop2/entity"
	"calmisland/kidsloop2/log"
	"calmisland/kidsloop2/storage"
	"calmisland/kidsloop2/utils"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"strings"
	"sync"
)

type IAssetModel interface {
	CreateAsset(ctx context.Context, data entity.AssetObject) (string, error)
	UpdateAsset(ctx context.Context, data entity.AssetObject) error
	DeleteAsset(ctx context.Context, id string) error

	GetAssetById(ctx context.Context, id string) (*entity.AssetObject, error)
	SearchAssets(ctx context.Context, condition *SearchAssetCondition) ([]*entity.AssetObject, error)

	GetAssetUploadPath(ctx context.Context, extension string) (string, error)
}

type AssetModel struct{}

type UpdateParams struct {
	keys   []string
	values map[string]*dynamodb.AttributeValue
}

type AssetEntity struct {
	Category string
	Tag      []string
	URL      string
}

func (u *UpdateParams) key()string{
	return strings.Join(u.keys, ",")
}

func (am AssetModel) checkEntity(ctx context.Context, entity AssetEntity, must bool) error {
	//TODO:Check if url is exists

	//TODO:Check tag & category entity

	return nil
}

func (am *AssetModel) CreateAsset(ctx context.Context, data entity.AssetObject) (string, error) {
	err := am.checkEntity(ctx, AssetEntity{
		Category: data.Category,
		Tag:      data.Tags,
		URL:      data.URL,
	}, true)

	if err != nil{
		return "", err
	}
	return am.doCreateAsset(ctx, data)
}

func (am *AssetModel) doCreateAsset(ctx context.Context, data entity.AssetObject) (string, error) {
	data.Id = utils.NewId()
	m, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		log.Get().Errorf("marshal asset failed, error: %v", err)
	}

	_, err = client.GetClient().PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("niobium"),
		Item:      m,
	})
	if err != nil {
		log.Get().Errorf("insert asset failed, error: %v", err)
	}
	return data.Id, nil
}

func (am *AssetModel) buildUpdateParams(ctx context.Context, data entity.UpdateAssetRequest)(*UpdateParams, error){
	updateStr := make([]string, 0)
	updateValues := make(map[string]*dynamodb.AttributeValue)

	if data.Category != "" {
		updateStr = append(updateStr, "set category = :c")
		updateValues[":c"] = &dynamodb.AttributeValue{
			S:    aws.String(data.Category),
		}
	}
	if data.Name != "" {
		updateStr = append(updateStr, "set name = :n")
		updateValues[":n"] = &dynamodb.AttributeValue{
			S:    aws.String(data.Name),
		}
	}

	if data.URL != "" {
		updateStr = append(updateStr, "set name = :u")
		updateValues[":u"] = &dynamodb.AttributeValue{
			S:    aws.String(data.URL),
		}
	}

	if data.Tag != nil {
		updateStr = append(updateStr, "set tag = :t")
		updateValues[":t"] = &dynamodb.AttributeValue{
			SS:    aws.StringSlice(data.Tag),
		}
	}
	return &UpdateParams{
		keys:   updateStr,
		values: updateValues,
	}, nil
}

func (am *AssetModel) UpdateAsset(ctx context.Context, data entity.UpdateAssetRequest) error {
	err := am.checkEntity(ctx, AssetEntity{
		Category: data.Category,
		Tag:      data.Tag,
		URL:      data.URL,
	}, false)

	co := entity.AssetObject{
		Id:            data.Id,
	}
	av, err := dynamodbattribute.MarshalMap(co)
	if err != nil{
		log.Get().Errorf("marshal asset failed, error: %v", err)
		return err
	}

	params, err := am.buildUpdateParams(ctx, data)
	if err != nil{
		log.Get().Errorf("build params failed, error: %v", err)
		return err
	}

	_, err = client.GetClient().UpdateItem(&dynamodb.UpdateItemInput{
		ExpressionAttributeValues:   params.values,
		Key:                         av,
		TableName:                   aws.String(co.TableName()),
		ReturnValues:     			 aws.String("UPDATED_NEW"),
		UpdateExpression:            aws.String(params.key()),
	})
	if err != nil{
		log.Get().Errorf("update asset failed, error: %v", err)
		return err
	}
	return nil
}

func (am *AssetModel) DeleteAsset(ctx context.Context, id string) error {
	co := entity.AssetObject{
		Id:            id,
	}
	av, err := dynamodbattribute.MarshalMap(co)
	if err != nil{
		return err
	}
	_, err = client.GetClient().DeleteItem(&dynamodb.DeleteItemInput{
		Key:                         av,
		TableName:                   aws.String(co.TableName()),
	})
	if err != nil{
		return err
	}
	return nil
}

func (am *AssetModel) GetAssetById(ctx context.Context, id string) (*entity.AssetObject, error) {
	co := entity.AssetObject{
		Id:            id,
	}
	av, err := dynamodbattribute.MarshalMap(co)
	if err != nil{
		return nil, err
	}

	result, err := client.GetClient().GetItem(&dynamodb.GetItemInput{
		Key:                      av,
		TableName:                aws.String(co.TableName()),
	})
	if err != nil{
		return nil, err
	}
	asset := new(entity.AssetObject)
	err = dynamodbattribute.UnmarshalMap(result.Item, &asset)
	if err != nil{
		return nil, err
	}

	return asset, nil
}

func (am *AssetModel) SearchAssets(ctx context.Context, condition *SearchAssetCondition) ([]*entity.AssetObject, error) {
	builder := expression.NewBuilder()
	conditions := condition.getConditions()
	for i := range conditions {
		builder = builder.WithFilter(conditions[i])
	}
	expr, err := builder.Build()
	if err != nil {
		log.Get().Errorf("Got error building expression: %v", err)
		return nil, err
	}

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(entity.AssetObject{}.TableName()),
	}
	result, err := client.GetClient().Scan(params)
	if err != nil {
		log.Get().Errorf("Query API call failed:", err)
		return nil, err
	}

	ret := make([]*entity.AssetObject, 0)
	for _, i := range result.Items {
		item := new(entity.AssetObject)

		err = dynamodbattribute.UnmarshalMap(i, item)
		if err != nil {
			log.Get().Errorf("Got error unmarshalling:", err)
			return nil, err
		}
		ret = append(ret, item)
	}

	return ret, nil
}

func (am *AssetModel) GetAssetUploadPath(ctx context.Context, extension string) (string, error) {
	client := storage.DefaultStorage()
	name := fmt.Sprintf("%s.%s", utils.NewId(), extension)

	return client.GetUploadFileTempPath(ctx, "asset", name)
}

type SearchAssetCondition struct {
	Ids        []string `json:"ids"`
	Names      []string `json:"names"`
	Categories []string `json:"categories"`
	SizeMin    int      `json:"size_min"`
	SizeMax    int      `json:"size_max"`

	Tags []string `json:"tag"`
}

func (s *SearchAssetCondition) getConditions() []expression.ConditionBuilder{
	conditions := make([]expression.ConditionBuilder, 0)
	if len(s.Ids) > 0 {
		condition := expression.Name("_id").In(expression.Value(s.Ids))
		conditions = append(conditions, condition)
	}
	if len(s.Names) > 0 {
		condition := expression.Name("name").In(expression.Value(s.Names))
		conditions = append(conditions, condition)
	}
	if len(s.Categories) > 0 {
		condition := expression.Name("category").In(expression.Value(s.Categories))
		conditions = append(conditions, condition)
	}
	if s.SizeMin > 0 {
		condition := expression.Name("size").GreaterThanEqual(expression.Value(s.SizeMin))
		conditions = append(conditions, condition)
	}
	if s.SizeMax > 0 {
		condition := expression.Name("size").LessThanEqual(expression.Value(s.SizeMax))
		conditions = append(conditions, condition)
	}

	if len(s.Tags) > 0 {
		condition := expression.Name("tag").In(expression.Value(s.Tags))
		conditions = append(conditions, condition)
	}

	return conditions
}

var assetModel *AssetModel
var _assetOnce sync.Once

func GetAssetModel()*AssetModel{
	_assetOnce.Do(func() {
		assetModel = new(AssetModel)
	})
	return assetModel
}