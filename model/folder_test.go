package model

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/stretchr/testify/assert"
)

func initDB() {
	dboHandler, err := dbo.NewWithConfig(func(c *dbo.Config) {
		c.ShowLog = true
		c.ShowSQL = true
		c.MaxIdleConns = 2
		c.MaxOpenConns = 4
		c.ConnectionString = os.Getenv("connection_string")
	})
	if err != nil {
		log.Error(context.TODO(), "create dbo failed", log.Err(err))
		panic(err)
	}
	config.Set(&config.Config{
		RedisConfig: config.RedisConfig{
			OpenCache: false,
			Host:      "",
			Port:      0,
			Password:  "",
		},
	})
	dbo.ReplaceGlobal(dboHandler)
}

func fakeOperator() *entity.Operator {
	return &entity.Operator{
		UserID: "1",
		OrgID:  "1",
	}
}

func TestCreateFolderCluster1(t *testing.T) {
	parentFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentID:  "",
		Name:      "Cluster1-1",
		Thumbnail: "thumbnail-001",
		Partition: entity.FolderPartitionMaterialAndPlans,
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log("parentFolder:", parentFolderId)

	subFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentID:  parentFolderId,
		Name:      "Cluster1-2",
		Thumbnail: "thumbnail-002",
		Partition: entity.FolderPartitionMaterialAndPlans,
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId)

	subSubFolderId1, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentID:  subFolderId,
		Name:      "Cluster1-3",
		Thumbnail: "thumbnail-003",
		Partition: entity.FolderPartitionMaterialAndPlans,
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subSubFolderId1)

	items, err := GetFolderModel().ListItems(context.Background(), parentFolderId, entity.FolderItemTypeAll, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	for i := range items {
		t.Log("ITEM:", items[i].ID)
	}
}
func TestCreateFolderCluster2(t *testing.T) {
	parentFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentID:  "",
		Name:      "Cluster2-1",
		Thumbnail: "thumbnail-001",
		Partition: entity.FolderPartitionMaterialAndPlans,
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log("parentFolder:", parentFolderId)

	subFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentID:  parentFolderId,
		Name:      "Cluster2-2",
		Thumbnail: "thumbnail-002",
		Partition: entity.FolderPartitionMaterialAndPlans,
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId)

	subSubFolderId1, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentID:  subFolderId,
		Name:      "Cluster2-3",
		Thumbnail: "thumbnail-003",
		Partition: entity.FolderPartitionMaterialAndPlans,
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subSubFolderId1)

	items, err := GetFolderModel().ListItems(context.Background(), parentFolderId, entity.FolderItemTypeAll, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	for i := range items {
		t.Log("ITEM:", items[i].ID)
	}
}
func RandName(prefix string) string {
	rand.Seed(time.Now().UnixNano())
	return prefix + "-" + strconv.Itoa(rand.Int())
}

func TestAddAndUpdateItem(t *testing.T) {
	parentFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentID:  "",
		Name:      RandName("TestFolder"),
		Thumbnail: "thumbnail-001",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log("parentFolder:", parentFolderId)

	subFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentID:  parentFolderId,
		Name:      RandName("TestSubFolder"),
		Thumbnail: "thumbnail-002",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId)

	subFolderId2, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentID:  parentFolderId,
		Name:      RandName("TestSubFolder"),
		Thumbnail: "thumbnail-003",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId2)

	//itemID1, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
	//	ParentFolderID: subFolderId2,
	//	Link:           "content-5f6c0389bff85a26a0e86585",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(itemID1)
	//
	//itemID2, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
	//	ParentFolderID: subFolderId2,
	//	Link:           "content-5f6c04712ac27c637384d9cb",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(itemID2)

	items, err := GetFolderModel().ListItems(context.Background(), subFolderId2, entity.FolderItemTypeAll, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	for i := range items {
		t.Log("ITEM:", items[i].ID)
	}

	//err = GetFolderModel().UpdateFolder(context.Background(), itemID2, entity.UpdateFolderRequest{
	//	Name:      "updated name",
	//	Thumbnail: "",
	//}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}

	err = GetFolderModel().UpdateFolder(context.Background(), subFolderId, entity.UpdateFolderRequest{
		Name:      "updated name222",
		Thumbnail: "",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
}

func TestMoveItem(t *testing.T) {
	//parentFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
	//	OwnerType: entity.OwnerTypeOrganization,
	//	ParentID:  "",
	//	Name:      RandName("TestFolder"),
	//	Thumbnail: "thumbnail-001",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log("parentFolder:", parentFolderId)
	//
	//subFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
	//	OwnerType: entity.OwnerTypeOrganization,
	//	ParentID:  parentFolderId,
	//	Name:      RandName("TestSubFolder"),
	//	Thumbnail: "thumbnail-002",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(subFolderId)
	//
	//subFolderId2, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
	//	OwnerType: entity.OwnerTypeOrganization,
	//	ParentID:  parentFolderId,
	//	Name:      RandName("TestSubFolder"),
	//	Thumbnail: "thumbnail-003",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(subFolderId2)

	//itemID1, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
	//	ParentFolderID: subFolderId2,
	//	Link:           "content-5f6c0389bff85a26a0e86585",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(itemID1)
	//
	//itemID2, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
	//	ParentFolderID: subFolderId2,
	//	Link:           "content-5f6c04712ac27c637384d9cb",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(itemID2)

	err := GetFolderModel().MoveItem(context.Background(), entity.MoveFolderRequest{
		ID:             "612859c7c313acc19cf064ee",
		OwnerType:      entity.OwnerTypeOrganization,
		Dist:           "612859e5d09f68c9445bac27",
		Partition:      entity.FolderPartitionMaterialAndPlans,
		FolderFileType: entity.FolderFileTypeFolder,
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}

	//subFolderItems, err := GetFolderModel().ListItems(context.Background(), subFolderId, entity.FolderItemTypeAll, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//assert.Equal(t, 1, len(subFolderItems))
	//assert.Equal(t, itemID2, subFolderItems[0].ID)
}

func TestBulkMoveItem(t *testing.T) {
	err := GetFolderModel().MoveItemBulk(context.Background(), entity.MoveFolderIDBulkRequest{
		FolderInfo: []entity.FolderIdWithFileType{
			{ID: "612883c70b505aef16d685d3", FolderFileType: entity.FolderFileTypeFolder},
		},
		OwnerType: entity.OwnerTypeOrganization,
		Dist:      "612883c17c5d9b40ac0b8951",
		Partition: entity.FolderPartitionMaterialAndPlans,
	}, fakeOperator())
	if err != nil {
		t.Fatal(err)
	}
}

func TestMoveItemAndFolder(t *testing.T) {
	parentFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentID:  "",
		Name:      RandName("TestFolder"),
		Thumbnail: "thumbnail-001",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log("parentFolder:", parentFolderId)

	subFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentID:  parentFolderId,
		Name:      RandName("TestSubFolder"),
		Thumbnail: "thumbnail-002",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId)

	subFolderId2, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentID:  parentFolderId,
		Name:      RandName("TestSubFolder"),
		Thumbnail: "thumbnail-003",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId2)

	//itemID1, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
	//	ParentFolderID: subFolderId2,
	//	Link:           "content-5f6c0389bff85a26a0e86585",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(itemID1)

	//itemID2, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
	//	ParentFolderID: subFolderId2,
	//	Link:           "content-5f6c04712ac27c637384d9cb",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(itemID2)

	//err = GetFolderModel().MoveItem(context.Background(), entity.MoveFolderRequest{
	//	ID:             itemID2,
	//	OwnerType:      entity.OwnerTypeOrganization,
	//	Dist:           subFolderId,
	//	Partition:      entity.FolderPartitionMaterialAndPlans,
	//	FolderFileType: entity.FolderFileTypeContent,
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//
	//subFolderItems, err := GetFolderModel().ListItems(context.Background(), subFolderId, entity.FolderItemTypeAll, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//assert.Equal(t, 1, len(subFolderItems))
	//assert.Equal(t, itemID2, subFolderItems[0].ID)
	//
	//err = GetFolderModel().MoveItem(context.Background(), entity.MoveFolderRequest{
	//	ID:             itemID2,
	//	OwnerType:      entity.OwnerTypeOrganization,
	//	Dist:           subFolderId2,
	//	Partition:      entity.FolderPartitionMaterialAndPlans,
	//	FolderFileType: entity.FolderFileTypeFolder,
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//subFolderItems, err = GetFolderModel().ListItems(context.Background(), subFolderId, entity.FolderItemTypeAll, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//assert.Equal(t, 2, len(subFolderItems))
	//for i := range subFolderItems {
	//	t.Log("ITEM:", subFolderItems[i].ID)
	//}
}

func TestRemoveItemAndFolder(t *testing.T) {
	//parentFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
	//	OwnerType: entity.OwnerTypeOrganization,
	//	ParentID:  "",
	//	Name:      RandName("TestFolder"),
	//	Thumbnail: "thumbnail-001",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log("parentFolder:", parentFolderId)
	//
	//subFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
	//	OwnerType: entity.OwnerTypeOrganization,
	//	ParentID:  parentFolderId,
	//	Name:      RandName("TestSubFolder"),
	//	Thumbnail: "thumbnail-002",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(subFolderId)
	//
	//subFolderId2, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
	//	OwnerType: entity.OwnerTypeOrganization,
	//	ParentID:  parentFolderId,
	//	Name:      RandName("TestSubFolder"),
	//	Thumbnail: "thumbnail-003",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(subFolderId2)
	//
	//itemID1, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
	//	ParentFolderID: subFolderId2,
	//	Link:           "content-5f6c0389bff85a26a0e86585",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(itemID1)
	//
	//itemID2, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
	//	ParentFolderID: subFolderId2,
	//	Link:           "content-5f6c04712ac27c637384d9cb",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(itemID2)
	//
	//err = GetFolderModel().RemoveItem(context.Background(), subFolderId2, fakeOperator())
	//if !assert.Error(t, err) {
	//	return
	//}
	//
	//err = GetFolderModel().RemoveItem(context.Background(), itemID2, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//subFolderItems, err := GetFolderModel().ListItems(context.Background(), subFolderId2, entity.FolderItemTypeAll, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//assert.Equal(t, 1, len(subFolderItems))
	//
	//err = GetFolderModel().RemoveItem(context.Background(), itemID1, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//
	//err = GetFolderModel().RemoveItem(context.Background(), subFolderId2, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
}

func TestSearchFolder(t *testing.T) {
	//parentFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
	//	OwnerType: entity.OwnerTypeOrganization,
	//	ParentID:  "",
	//	Name:      RandName("TestFolder"),
	//	Thumbnail: "thumbnail-001",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log("parentFolder:", parentFolderId)
	//
	//subFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
	//	OwnerType: entity.OwnerTypeOrganization,
	//	ParentID:  parentFolderId,
	//	Name:      RandName("TestSubFolder"),
	//	Thumbnail: "thumbnail-002",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(subFolderId)
	//
	//subFolderId2, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
	//	OwnerType: entity.OwnerTypeOrganization,
	//	ParentID:  parentFolderId,
	//	Name:      RandName("TestSubFolder"),
	//	Thumbnail: "thumbnail-003",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(subFolderId2)
	//
	//itemID1, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
	//	ParentFolderID: subFolderId2,
	//	Link:           "content-5f6c0389bff85a26a0e86585",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(itemID1)
	//
	//itemID2, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
	//	ParentFolderID: subFolderId2,
	//	Link:           "content-5f6c04712ac27c637384d9cb",
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(itemID2)
	//
	//total, folders, err := GetFolderModel().SearchOrgFolder(context.Background(), entity.SearchFolderCondition{
	//	Path: constant.FolderRootPath,
	//}, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Log(total)
	//for i := range folders {
	//	t.Log(folders[i].ID)
	//}
	//
	//folderInfo, err := GetFolderModel().GetFolderByID(context.Background(), folders[0].ID, fakeOperator())
	//if !assert.NoError(t, err) {
	//	return
	//}
	//t.Logf("%#v\n", folderInfo)
}

func TestSplit(t *testing.T) {
	t.Log(strings.Split("/", "/"))
}
