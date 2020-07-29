package utils

import (
	"calmisland/kidsloop2/constant"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func ConvertDynamodbError(err error) error{
	if err!=nil{
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeResourceNotFoundException:
				err = constant.ErrRecordNotFound
			case dynamodb.ErrCodeRequestLimitExceeded:
				err = constant.ErrExceededLimit
			}
		}
	}

	return err
}
