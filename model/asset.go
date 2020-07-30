package model

import (
	client "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/storage"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"net/http"
	"strings"
	"sync"
	"time"
)

var(
	ErrNoSuchURL = errors.New("no such url")
	ErrRequestItemIsNil = errors.New("request item is nil")
)

type IAssetModel interface {
	CreateAsset(ctx context.Context, data entity.AssetObject) (string, error)
	UpdateAsset(ctx context.Context, data entity.UpdateAssetRequest) error
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

func (u *UpdateParams) key() string {
	return strings.Join(u.keys, ",")
}

func (am AssetModel) checkEntity(ctx context.Context, entity AssetEntity, must bool) error {
	if must && (entity.URL == "" || entity.Category == "") {
		return ErrRequestItemIsNil
	}

	//TODO:Check if url is exists
	if entity.URL != "" {
		err := checkURL(entity.URL)
		if err != nil{
			return err
		}
	}
	//TODO:Check tag & category entity

	return nil
}

func checkURL(url string) error {
	resp, err := http.Get(url)
	if err != nil{
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return ErrNoSuchURL
	}
	return nil

}

func (am *AssetModel) CreateAsset(ctx context.Context, data entity.AssetObject) (string, error) {
	err := am.checkEntity(ctx, AssetEntity{
		Category: data.Category,
		Tag:      data.Tags,
		URL:      data.URL,
	}, true)

	if err != nil {
		return "", err
	}
	return am.doCreateAsset(ctx, data)
}

func (am *AssetModel) doCreateAsset(ctx context.Context, data entity.AssetObject) (string, error) {
	now := time.Now()
	data.CreatedAt = &now
	data.UpdatedAt = &now
	data.Id = utils.NewId()
	m, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		log.Get().Errorf("marshal asset failed, error: %v", err)
	}
	_, err = client.GetClient().PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("assets"),
		Item:      m,
	})

	if err != nil {
		log.Get().Errorf("insert assets failed, error: %v", err)
		return "", err
	}
	return data.Id, nil
}

func (am *AssetModel) buildUpdateParams(ctx context.Context, data entity.UpdateAssetRequest) (*UpdateParams, error) {
	updateStr := make([]string, 0)
	updateValues := make(map[string]*dynamodb.AttributeValue)

	if data.Category != "" {
		updateStr = append(updateStr, "set category = :c")
		updateValues[":c"] = &dynamodb.AttributeValue{
			S: aws.String(data.Category),
		}
	}
	if data.Name != "" {
		updateStr = append(updateStr, "set name = :n")
		updateValues[":n"] = &dynamodb.AttributeValue{
			S: aws.String(data.Name),
		}
	}

	if data.URL != "" {
		updateStr = append(updateStr, "set name = :u")
		updateValues[":u"] = &dynamodb.AttributeValue{
			S: aws.String(data.URL),
		}
	}

	if data.Tag != nil {
		updateStr = append(updateStr, "set tag = :t")
		updateValues[":t"] = &dynamodb.AttributeValue{
			SS: aws.StringSlice(data.Tag),
		}
	}

	//TODO:Updated_at
	updateStr = append(updateStr, "set updated_at = :ud")
	updateValues[":ud"] = &dynamodb.AttributeValue{
		S:    aws.String(time.Now().String()),
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

	av := make(map[string]*dynamodb.AttributeValue)
	av["id"] = &dynamodb.AttributeValue{
		S:    aws.String(data.Id),
	}

	params, err := am.buildUpdateParams(ctx, data)
	if err != nil {
		log.Get().Errorf("build params failed, error: %v", err)
		return err
	}

	_, err = client.GetClient().UpdateItem(&dynamodb.UpdateItemInput{
		ExpressionAttributeValues: params.values,
		Key:                       av,
		TableName:                 aws.String(entity.AssetObject{}.TableName()),
		ReturnValues:              aws.String("UPDATED_NEW"),
		UpdateExpression:          aws.String(params.key()),
	})
	if err != nil {
		log.Get().Errorf("update asset failed, error: %v", err)
		return err
	}
	return nil
}

func (am *AssetModel) DeleteAsset(ctx context.Context, id string) error {
	av := make(map[string]*dynamodb.AttributeValue)
	av["id"] = &dynamodb.AttributeValue{
		S:    aws.String(id),
	}
	_, err := client.GetClient().DeleteItem(&dynamodb.DeleteItemInput{
		Key:       av,
		TableName: aws.String(entity.AssetObject{}.TableName()),
	})
	if err != nil {
		return err
	}
	return nil
}

func (am *AssetModel) GetAssetById(ctx context.Context, id string) (*entity.AssetObject, error) {
	av := make(map[string]*dynamodb.AttributeValue)
	av["id"] = &dynamodb.AttributeValue{
		S:    aws.String(id),
	}

	result, err := client.GetClient().GetItem(&dynamodb.GetItemInput{
		Key:       av,
		TableName: aws.String(entity.AssetObject{}.TableName()),
	})
	if err != nil {
		return nil, err
	}
	asset := new(entity.AssetObject)
	err = dynamodbattribute.UnmarshalMap(result.Item, &asset)
	if err != nil {
		return nil, err
	}

	return asset, nil
}

func (am *AssetModel) SearchAssets(ctx context.Context, condition *SearchAssetCondition) ([]*entity.AssetObject, error) {
	builder := expression.NewBuilder()
	builder = builder.WithFilter(condition.getConditions())
	expr, err := builder.Build()
	if err != nil {
		log.Get().Errorf("Got error building expression: %v", err)
		return nil, err
	}
	fmt.Println("Condition:", *expr.Filter())
	fmt.Println("NAMES:", expr.Values())

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(entity.AssetObject{}.TableName()),
		//Limit:                     aws.Int64(condition.PageSize),
		//Segment:                   aws.Int64(condition.Page),
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
	Id        string `json:"id"`
	Name      string `json:"name"`
	Category string `json:"category"`
	SizeMin    int      `json:"size_min"`
	SizeMax    int      `json:"size_max"`

	Tag 	string `json:"tag"`

	PageSize int `json:"page_size"`
	Page     int `json:"page"`
}

func (s *SearchAssetCondition) getConditions() expression.ConditionBuilder {
	conditions := make([]expression.ConditionBuilder, 0)
	if s.Id != "" {
		condition := expression.Name("id").Equal(expression.Value(s.Id))
		conditions = append(conditions, condition)
	}
	if s.Name != ""{
		condition := expression.Name("name").Equal(expression.Value(s.Name))
		conditions = append(conditions, condition)
	}
	if s.Category != "" {
		condition := expression.Name("category").Equal(expression.Value(s.Category))
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

	if s.Tag != "" {
		condition := expression.Name("tag").Equal(expression.Value(s.Tag))
		conditions = append(conditions, condition)
	}

	if len(conditions) > 0 {
		for i := range conditions {
			if i == 0 {
				continue
			}
			conditions[0] = conditions[0].And(conditions[i])
		}
	}

	return conditions[0]
}

var assetModel *AssetModel
var _assetOnce sync.Once

func GetAssetModel() *AssetModel {
	_assetOnce.Do(func() {
		assetModel = new(AssetModel)
	})
	return assetModel
}
