package da

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	db "gitlab.badanamu.com.cn/calmisland/kidsloop2/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"os"
	"testing"
)

func TestCreateTable(t *testing.T) {
	tableName := "content"
	db.GetClient().DeleteTable(&dynamodb.DeleteTableInput{TableName: aws.String("content")})

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("content_id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("org"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("author"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("publish_status"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("org_user_id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("content_name"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("author_name"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("description"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("keywords"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("content_id"),
				KeyType:       aws.String("HASH"),
			},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: 	aws.String("publish_status"),
				KeySchema:  []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("publish_status"),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String("org"),
						KeyType:       aws.String("RANGE"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType:   aws.String(dynamodb.ProjectionTypeAll),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(10),
					WriteCapacityUnits: aws.Int64(10),
				},
			},
			{
				IndexName: 	aws.String("author"),
				KeySchema:  []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("author"),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String("org"),
						KeyType:       aws.String("RANGE"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType:   aws.String(dynamodb.ProjectionTypeAll),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(10),
					WriteCapacityUnits: aws.Int64(10),
				},
			},
			{
				IndexName: 	aws.String("name"),
				KeySchema:  []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("content_name"),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String("org_user_id"),
						KeyType:       aws.String("RANGE"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType:   aws.String(dynamodb.ProjectionTypeAll),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(10),
					WriteCapacityUnits: aws.Int64(10),
				},
			},
			{
				IndexName: 	aws.String("author_name"),
				KeySchema:  []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("author_name"),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String("org_user_id"),
						KeyType:       aws.String("RANGE"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType:   aws.String(dynamodb.ProjectionTypeAll),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(10),
					WriteCapacityUnits: aws.Int64(10),
				},
			},
			{
				IndexName: 	aws.String("description"),
				KeySchema:  []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("description"),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String("org_user_id"),
						KeyType:       aws.String("RANGE"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType:   aws.String(dynamodb.ProjectionTypeAll),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(10),
					WriteCapacityUnits: aws.Int64(10),
				},
			},
			{
				IndexName: 	aws.String("keywords"),
				KeySchema:  []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("keywords"),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String("org_user_id"),
						KeyType:       aws.String("RANGE"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType:   aws.String(dynamodb.ProjectionTypeAll),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(10),
					WriteCapacityUnits: aws.Int64(10),
				},
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(tableName),
	}

	_, err := db.GetClient().CreateTable(input)
	if err != nil {
		fmt.Println("Got error calling CreateTable:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Created the table", tableName)
}

func TestCreateContent(t *testing.T) {
	id, err := GetDyContentDA().CreateContent(context.Background(), entity.Content{
		ContentType:   0,
		Name:          "TestContent",
		Program:       "Program1",
		Subject:       "Subject1",
		Developmental: "Developmental1",
		Skills:        "Skills1",
		Age:           "Age1",
		Keywords:      "Keywords1,Keywords2",
		Description:   "My Description",
		Thumbnail:     "/Thumbnail1.png",
		Data:          "{Source:\"source_data.png\"}",
		Extra:         "{}",
		Author:        "Author1",
		AuthorName:    "Author name",
		Org:           "org1",
		PublishScope:  "org0",
		PublishStatus: "draft",
		Version:       1,
	})
	if err != nil{
		panic(err)
	}
	fmt.Println("ID:", id)

	content, err := GetDyContentDA().GetContentById(context.Background(), id)
	if err != nil{
		panic(err)
	}
	fmt.Printf("%#v\n", content)
}

func TestGetContent(t *testing.T){
	id := "5f3b8a9159eb66924f94ad1c"
	content, err := GetDyContentDA().GetContentById(context.Background(), id)
	if err != nil{
		panic(err)
	}
	fmt.Printf("%#v\n", content)
}

func TestUpdateContent(t *testing.T) {
	id := "5f3b90b291fa8fed55d310a0"
	err := GetDyContentDA().UpdateContent(context.Background(), id, entity.Content{
		Org:           "org2",
		PublishScope:  "org1",
		PublishStatus: "draft",
		Version:       2,
	})
	if err != nil{
		panic(err)
	}
	content, err := GetDyContentDA().GetContentById(context.Background(), id)
	if err != nil{
		panic(err)
	}
	fmt.Printf("%#v\n", content)
}

func TestDeleteContent(t *testing.T) {
	id := "5f3b8adee5eef1ee75e97532"
	err := GetDyContentDA().DeleteContent(context.Background(), id)
	if err != nil{
		panic(err)
	}
	content, err := GetDyContentDA().GetContentById(context.Background(), id)
	if err != nil{
		panic(err)
	}
	fmt.Printf("%#v\n", content)
}

func TestSearchContent(t *testing.T) {
	id, err := GetDyContentDA().CreateContent(context.Background(), entity.Content{
		ContentType:   0,
		Name:          "TestContent000",
		Program:       "Program1",
		Subject:       "Subject1",
		Developmental: "Developmental1",
		Skills:        "Skills1",
		Age:           "Age1",
		Keywords:      "Keywords1,Keywords2",
		Description:   "My Description",
		Thumbnail:     "/Thumbnail1.png",
		Data:          "{Source:\"source_data.png\"}",
		Extra:         "{}",
		Author:        "Author1",
		AuthorName:    "Author name",
		Org:           "org1",
		PublishScope:  "org0",
		PublishStatus: "draft",
		Version:       1,
	})
	if err != nil{
		panic(err)
	}
	fmt.Println("ID:", id)
	sub, contents, err := GetDyContentDA().SearchContent(context.Background(), &DyContentCondition{
		//IDS:           []string{id},
		//Name:          "",
		//ContentType:   nil,
		//Scope:         nil,
		PublishStatus: []string{"draft"},
		LastKey: "5f3b90b291fa8fed55d310a0",
		//Author:        "",
		//Org:           "",
		PageSize: 1,
	})
	if err != nil{
		panic(err)
	}
	fmt.Printf("contents: %#v\n", contents[0])
	fmt.Printf("sub: %#v\n", sub)
}


func TestSearchContentKey(t *testing.T) {
	id, err := GetDyContentDA().CreateContent(context.Background(), entity.Content{
		ContentType:   0,
		Name:          "TestContent000",
		Program:       "Program1",
		Subject:       "Subject1",
		Developmental: "Developmental1",
		Skills:        "Skills1",
		Age:           "Age1",
		Keywords:      "Keywords1,Keywords2",
		Description:   "My Description",
		Thumbnail:     "/Thumbnail1.png",
		Data:          "{Source:\"source_data.png\"}",
		Extra:         "{}",
		Author:        "Author1",
		AuthorName:    "Author name",
		Org:           "org1",
		PublishScope:  "org0",
		PublishStatus: "draft",
		Version:       1,
	})
	if err != nil{
		panic(err)
	}
	fmt.Println("ID:", id)
	sub, contents, err := GetDyContentDA().SearchContentByKey(context.Background(), DyKeyContentCondition{
		PublishStatus:      "draft",
		Org: "org1",
	})
	if err != nil{
		panic(err)
	}
	fmt.Printf("contents: %#v\n", contents[0])
	fmt.Printf("sub: %#v\n", sub)

	sub, contents, err = GetDyContentDA().SearchContentByKey(context.Background(), DyKeyContentCondition{
		Description:        "My Description",
		OrgUserId: "org1" + id,
	})
	if err != nil{
		panic(err)
	}
	fmt.Printf("contents: %#v\n", contents[0])
	fmt.Printf("sub: %#v\n", sub)

}
