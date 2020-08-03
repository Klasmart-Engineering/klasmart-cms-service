package entity

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"strconv"
	"strings"
	"time"
)

type CategoryObject struct {
	ID       string `json:"id" dynamodbav:"id"`
	Name     string `json:"name" dynamodbav:"name"`
	ParentID string `json:"parent_id" dynamodbav:"parent_id"`

	CreatedAt int64 `json:"-" dynamodbav:"created_at"`
	UpdatedAt int64 `json:"-" dynamodbav:"updated_at"`
	DeletedAt int64 `json:"-" dynamodbav:"deleted_at"`
}

func (CategoryObject) TableName() string {
	return "Categories"
}

func (co CategoryObject) ToKey() map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		"id": {S: aws.String(co.ID)},
	}
}

func (co CategoryObject) ToUpdateParam() (upExpr string, exprAttrNames map[string]*string, exprAttrValues map[string]*dynamodb.AttributeValue) {
	var upExprs []string
	exprAttrNames = make(map[string]*string)
	exprAttrValues = make(map[string]*dynamodb.AttributeValue)
	if co.ID == co.ParentID {
		upExprs = append(upExprs, "parent_id = :pid")
		exprAttrValues[":pid"] = &dynamodb.AttributeValue{S: aws.String("")}
	} else if co.ParentID != "" {
		upExprs = append(upExprs, "parent_id = :pid")
		exprAttrValues[":pid"] = &dynamodb.AttributeValue{S: aws.String(co.ParentID)}
	}

	if co.Name != "" {
		upExprs = append(upExprs, "#cat_name = :nm")
		exprAttrNames["#cat_name"] = aws.String("name")
		exprAttrValues[":nm"] = &dynamodb.AttributeValue{S: aws.String(co.Name)}
	}
	upExprs = append(upExprs, "updated_at = :uat")
	exprAttrValues[":uat"] = &dynamodb.AttributeValue{S: aws.String(strconv.FormatInt(time.Now().Unix(), 10))}

	upExpr = "set " + strings.Join(upExprs, ",")
	return
}

type SearchCategoryCondition struct {
	IDs   []string `json:"ids"`
	Names []string `json:"names"`

	PageSize int64  `json:"page_size"`
	Page     int64  `json:"page"`
	OrderBy  string `json:"order_by"`
}

func (s *SearchCategoryCondition) ToExpr() (expression.Expression, error) {
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

	if s.PageSize == 0 {
		s.PageSize = 10
	}
	expr, err := expression.NewBuilder().WithFilter(condition).Build()
	return expr, err
}
