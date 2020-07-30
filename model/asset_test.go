package model

import (
	"calmisland/kidsloop2/conf"
	client "calmisland/kidsloop2/dynamodb"
	"calmisland/kidsloop2/entity"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"os"
	"testing"
)
func InitEnv(){
	os.Setenv("cloud_env", "aws")
	os.Setenv("storage_bucket", "kidsloop-global-resources-dev")
	os.Setenv("storage_region", "ap-northeast-2")
	os.Setenv("secret_id", "AKIAXGKUAYT2P2IJ2KX7")
	os.Setenv("secret_key", "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")
	os.Setenv("storage_accelerate", "true")
	os.Setenv("cdn_open", "false")
}

func TestCreateTable(t *testing.T){
	_, err := client.GetClient().DeleteTable(&dynamodb.DeleteTableInput{TableName: aws.String("assets")})
	t.Log(err)


	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String("assets"),
	}

	_, err = client.GetClient().CreateTable(input)
	if err != nil {
		fmt.Println("Got error calling CreateTable:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Created the table", "assets")

	out, err := client.GetClient().ListTables(&dynamodb.ListTablesInput{})
	if err != nil{
		panic(err)
	}
	t.Logf("%#v", out)
}
func TestAssetModel_CreateAsset(t *testing.T) {
	InitEnv()
	conf.LoadEnvConfig()

	assetModel := GetAssetModel()
	id, err := assetModel.CreateAsset(context.Background(), entity.AssetObject{
		Name:      "Hello.mp3",
		Category:  "HelloCategory",
		Size:      180,
		Tags:      []string{"Hello"},
		URL:       "http://www.baidu.com",
		Uploader:  "123",
	})
	if err != nil{
		panic(err)
	}
	t.Log(id)


	asset, err := assetModel.GetAssetById(context.Background(), id)
	if err != nil{
		panic(err)
	}
	t.Logf("Asset: %#v", asset)
}


func TestAssetModel_GetAsset(t *testing.T) {
	InitEnv()
	conf.LoadEnvConfig()

	assetModel := GetAssetModel()
	asset, err := assetModel.GetAssetById(context.Background(), "269fdbaba6d4f1b4")
	if err != nil{
		panic(err)
	}
	t.Logf("Asset: %#v", asset)
}

func TestAssetModel_SearchAssets(t *testing.T) {
	InitEnv()
	conf.LoadEnvConfig()

	assetModel := GetAssetModel()
	res, err := assetModel.SearchAssets(context.Background(), &SearchAssetCondition{
		Id:        "ea2feb2590199523",
		Name:      "Hello.mp3",
		//Categories: nil,
		//SizeMin:    0,
		//SizeMax:    0,
		//Tags:       nil,
		//PageSize:   10,
	})
	if err != nil{
		panic(err)
	}
	for i := range res{
		t.Logf("%#v", res[i])
	}
}