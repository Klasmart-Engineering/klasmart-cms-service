package da

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	client "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"strings"
	"sync"
	"time"
)

var(
	ErrRecordNotFound = errors.New("record not found")
	ErrPageOutOfRange = errors.New("page out of range")
)

type SearchCategoryCondition struct {
	IDs        []string `json:"ids"`
	Names      []string `json:"names"`

	PageSize int64 `json:"page_size"`
	Page     int64 `json:"page"`
	OrderBy	 string `json:"order_by"`
}

type IAssetDA interface {
	CreateAsset(ctx context.Context, data entity.AssetObject) (string, error)
	UpdateAsset(ctx context.Context, data entity.UpdateAssetRequest) error
	DeleteAsset(ctx context.Context, id string) error

	GetAssetByID(ctx context.Context, id string) (*entity.AssetObject, error)
	SearchAssets(ctx context.Context, condition *SearchAssetCondition) (int64, []*entity.AssetObject, error)
}
type UpdateParams struct {
	keys   []string
	values map[string]*dynamodb.AttributeValue
}

func (u *UpdateParams) key() string {
	return "set " + strings.Join(u.keys, ",")
}

type DynamoDBAssetDA struct{}

func (DynamoDBAssetDA) CreateAsset(ctx context.Context, data entity.AssetObject) (string, error) {
	now := time.Now()
	data.CreatedAt = &now
	data.UpdatedAt = &now
	data.ID = utils.NewID()
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
	return data.ID, nil
}

func (am *DynamoDBAssetDA) UpdateAsset(ctx context.Context, data entity.UpdateAssetRequest) error {
	av := make(map[string]*dynamodb.AttributeValue)
	av["id"] = &dynamodb.AttributeValue{
		S:    aws.String(data.ID),
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

func (DynamoDBAssetDA) DeleteAsset(ctx context.Context, id string) error {
	av := make(map[string]*dynamodb.AttributeValue)
	av["id"] = &dynamodb.AttributeValue{
		S:    aws.String(id),
	}
	_, err := client.GetClient().DeleteItem(&dynamodb.DeleteItemInput{
		Key:       av,
		TableName: aws.String(entity.AssetObject{}.TableName()),
	})
	if err != nil{
		return err
	}
	return nil
}

func (DynamoDBAssetDA) GetAssetByID(ctx context.Context, id string) (*entity.AssetObject, error) {
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

	if len(result.Item) < 1 {
		return nil, ErrRecordNotFound
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &asset)
	if err != nil {
		return nil, err
	}
	return asset, nil
}

func (DynamoDBAssetDA) SearchAssets(ctx context.Context, condition *SearchAssetCondition) (int64, []*entity.AssetObject, error) {
	builder := expression.NewBuilder()
	builder = builder.WithFilter(condition.getConditions())
	expr, err := builder.Build()
	if err != nil {
		log.Get().Errorf("Got error building expression: %v", err)
		return 0, nil, err
	}

	if condition.PageSize < 1 {
		condition.PageSize = 10000000
	}
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(entity.AssetObject{}.TableName()),
		Limit:						aws.Int64(int64(condition.PageSize)),
	}
	//result, err := client.GetClient().Scan(params)
	result, err := scanPages(params, condition.Page)
	if err != nil {
		log.Get().Errorf("Query API call failed:", err)
		return 0, nil, err
	}

	var ret []*entity.AssetObject
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &ret)
	if err != nil {
		log.Get().Errorf("Got error unmarshalling:", err)
		return 0, nil, err
	}

	return *result.Count, ret, nil
}

func (am *DynamoDBAssetDA) buildUpdateParams(ctx context.Context, data entity.UpdateAssetRequest) (*UpdateParams, error) {
	updateStr := make([]string, 0)
	updateValues := make(map[string]*dynamodb.AttributeValue)

	if data.Category != "" {
		updateStr = append(updateStr, "category = :c")
		updateValues[":c"] = &dynamodb.AttributeValue{
			S: aws.String(data.Category),
		}
	}
	if data.Name != "" {
		updateStr = append(updateStr, "name = :n")
		updateValues[":n"] = &dynamodb.AttributeValue{
			S: aws.String(data.Name),
		}
	}

	if data.URL != "" {
		updateStr = append(updateStr, "name = :u")
		updateValues[":u"] = &dynamodb.AttributeValue{
			S: aws.String(data.URL),
		}
	}

	if data.Tag != nil {
		updateStr = append(updateStr, "tag = :t")
		updateValues[":t"] = &dynamodb.AttributeValue{
			SS: aws.StringSlice(data.Tag),
		}
	}

	//TODO:Updated_at
	updateStr = append(updateStr, "updated_at = :ud")
	updateValues[":ud"] = &dynamodb.AttributeValue{
		S:    aws.String(time.Now().Format("2006-01-02T15:04:05Z07:00")),
	}
	return &UpdateParams{
		keys:   updateStr,
		values: updateValues,
	}, nil
}

type SearchAssetCondition struct {
	ID        string `json:"id"`
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
	if s.ID != "" {
		condition := expression.Name("id").Equal(expression.Value(s.ID))
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

func scanPages(input *dynamodb.ScanInput, pageIndex int) (*dynamodb.ScanOutput, error){
	currentPage := 0
	var result *dynamodb.ScanOutput
	err := client.GetClient().ScanPages(input, func(output *dynamodb.ScanOutput, b bool) bool {
		if !b {
			return false
		}
		if currentPage == pageIndex {
			result = output
		}
		currentPage ++
		
		return currentPage < pageIndex
	})
	if err != nil{
		return nil, err
	}

	return result, nil
}

var _assetDA IAssetDA
var _assetDAOnce sync.Once


func GetAssetDA() IAssetDA{
	_assetDAOnce.Do(func() {
		_assetDA = new(DynamoDBAssetDA)
	})
	return _assetDA
}