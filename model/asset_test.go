package model

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	client "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"os"
	"testing"
)

func InitEnv() {
	//os.Setenv("cloud_env", "aws")
	//os.Setenv("storage_bucket", "kidsloop-global-resources-dev")
	//os.Setenv("storage_region", "ap-northeast-2")
	//os.Setenv("secret_id", "AKIAXGKUAYT2P2IJ2KX7")
	//os.Setenv("secret_key", "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")
	//os.Setenv("storage_accelerate", "true")
	//os.Setenv("cdn_open", "false")
}

func TestCreateTable(t *testing.T) {
	//_, err := client.GetClient().DeleteTable(&dynamodb.DeleteTableInput{TableName: aws.String("assets")})
	//t.Log(err)

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

	_, err := client.GetClient().CreateTable(input)
	if err != nil {
		fmt.Println("Got error calling CreateTable:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Created the table", "assets")

	out, err := client.GetClient().ListTables(&dynamodb.ListTablesInput{})
	if err != nil {
		panic(err)
	}
	t.Logf("%#v", out)
}
func TestAssetModel_CreateAsset(t *testing.T) {
	InitEnv()
	config.LoadEnvConfig()

	out, err := client.GetClient().ListTables(&dynamodb.ListTablesInput{})
	if err != nil {
		panic(err)
	}
	t.Logf("%#v", out)

	assetModel := GetAssetModel()
	id, err := assetModel.CreateAsset(context.Background(), entity.AssetObject{
		Name:     "Hello.mp3",
		Category: "HelloCategory",
		Size:     180,
		Tags:     []string{"Hello"},
		//URL:      "http://www.baidu.com",
		Uploader: "123",
	})
	if err != nil {
		panic(err)
	}
	t.Log(id)

	asset, err := assetModel.GetAssetByID(context.Background(), id)
	if err != nil {
		panic(err)
	}
	t.Logf("Asset: %#v", asset)
}

func TestAssetModel_GetAsset(t *testing.T) {
	InitEnv()
	config.LoadEnvConfig()

	assetModel := GetAssetModel()
	asset, err := assetModel.GetAssetByID(context.Background(), "269fdbaba6d4f1b4")
	if err != nil {
		panic(err)
	}
	t.Logf("Asset: %#v", asset)
}

func TestAssetModel_SearchAssets(t *testing.T) {
	InitEnv()
	config.LoadEnvConfig()

	assetModel := GetAssetModel()
	count, res, err := assetModel.SearchAssets(context.Background(), &entity.SearchAssetCondition{
		ID:   "5f22462699878cfe177a4101",
		Name: "Hello.mp3",
	})
	if err != nil {
		panic(err)
	}
	t.Log(count)
	for i := range res {
		t.Logf("%#v", res[i])
	}
}

func TestAssetModel_UpdateAssets(t *testing.T) {
	InitEnv()
	config.LoadEnvConfig()

	assetModel := GetAssetModel()
	err := assetModel.UpdateAsset(context.Background(), entity.UpdateAssetRequest{
		ID:       "ea2feb2590199523",
		Category: "123123123aabbcc",
	})
	if err != nil {
		panic(err)
	}

	asset, err := assetModel.GetAssetByID(context.Background(), "ea2feb2590199523")
	if err != nil {
		panic(err)
	}
	t.Logf("Asset: %#v", asset)
}

func TestAssetModel_DeleteAssets(t *testing.T) {
	InitEnv()
	config.LoadEnvConfig()

	assetModel := GetAssetModel()
	err := assetModel.DeleteAsset(context.Background(), "ea2feb2590199523")
	if err != nil {
		panic(err)
	}

	asset, err := assetModel.GetAssetByID(context.Background(), "ea2feb2590199523")
	if err != nil {
		panic(err)
	}
	t.Logf("Asset: %#v", asset)
}

func TestAssetModel_GetAssetUploadPath(t *testing.T) {
	InitEnv()
	config.LoadEnvConfig()

	assetModel := GetAssetModel()

	path, err := assetModel.GetAssetUploadPath(context.Background(), "jpg")
	if err != nil {
		panic(err)
	}
	t.Logf("path: %#v", path)
}

func TestAssetModel_GetAssetResourcePath(t *testing.T) {
	InitEnv()
	config.LoadEnvConfig()

	assetModel := GetAssetModel()

	path, err := assetModel.GetAssetResourcePath(context.Background(), "5f225eeee763b300cf63cb90.jpg")
	if err != nil {
		panic(err)
	}
	t.Logf("path: %#v", path)
}

func TestGetAssetJSON(t *testing.T) {
	obj := entity.AssetObject{

		Name:     "Hello.mp3",
		Category: "HelloCategory",
		Size:     180,
		Tags:     []string{"Hello"},
		//URL:      "http://www.baidu.com",
		Uploader: "123",
	}
	data, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	t.Log(string(data))
}
