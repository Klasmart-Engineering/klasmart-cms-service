package model

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func initDB() {
	dboHandler, err := dbo.NewWithConfig(func(c *dbo.Config) {
		c.ShowLog = true
		c.ShowSQL = true
		c.MaxIdleConns = 2
		c.MaxOpenConns = 4
		c.ConnectionString = "root:Badanamu123456@tcp(192.168.1.234:3310)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local"
	})
	if err != nil {
		log.Error(context.TODO(), "create dbo failed", log.Err(err))
		panic(err)
	}
	config.Set(&config.Config{
		RedisConfig:     config.RedisConfig{
			OpenCache: false,
			Host:      "",
			Port:      0,
			Password:  "",
		},
	})
	dbo.ReplaceGlobal(dboHandler)
}

func TestMain(m *testing.M) {
	fmt.Println("begin test")
	initDB()
	m.Run()
	fmt.Println("end test")
}
func fakeOperator() *entity.Operator{
	return &entity.Operator{
		UserID: "1",
		Role:   "teacher",
		OrgID:  "1",
	}
}

func TestCreateFolder(t *testing.T) {
	parentFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  "",
		Name:      "TestFolder1",
		Thumbnail: "thumbnail-001",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log("parentFolder:", parentFolderId)

	subFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  parentFolderId,
		Name:      "TestSubFolder1",
		Thumbnail: "thumbnail-002",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId)

	subFolderId2, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  parentFolderId,
		Name:      "TestSubFolder2",
		Thumbnail: "thumbnail-003",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId2)

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
		ParentId:  "",
		Name:      RandName("TestFolder"),
		Thumbnail: "thumbnail-001",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log("parentFolder:", parentFolderId)

	subFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  parentFolderId,
		Name:      RandName("TestSubFolder"),
		Thumbnail: "thumbnail-002",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId)

	subFolderId2, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  parentFolderId,
		Name:      RandName("TestSubFolder"),
		Thumbnail: "thumbnail-003",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId2)

	itemID1, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
		FolderID:  subFolderId2,
		Link:      "content-5f6c0389bff85a26a0e86585",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(itemID1)

	itemID2, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
		FolderID:  subFolderId2,
		Link:      "content-5f6c04712ac27c637384d9cb",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(itemID2)

	items, err := GetFolderModel().ListItems(context.Background(), subFolderId2, entity.FolderItemTypeAll, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	for i := range items {
		t.Log("ITEM:", items[i].ID)
	}

	err = GetFolderModel().UpdateFolder(context.Background(), itemID2, entity.UpdateFolderRequest{
		Name:      "updated name",
		Thumbnail: "",
	}, fakeOperator())
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
	parentFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  "",
		Name:      RandName("TestFolder"),
		Thumbnail: "thumbnail-001",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log("parentFolder:", parentFolderId)

	subFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  parentFolderId,
		Name:      RandName("TestSubFolder"),
		Thumbnail: "thumbnail-002",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId)

	subFolderId2, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  parentFolderId,
		Name:      RandName("TestSubFolder"),
		Thumbnail: "thumbnail-003",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId2)

	itemID1, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
		FolderID:  subFolderId2,
		Link:      "content-5f6c0389bff85a26a0e86585",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(itemID1)

	itemID2, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
		FolderID:  subFolderId2,
		Link:      "content-5f6c04712ac27c637384d9cb",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(itemID2)

	err = GetFolderModel().MoveItem(context.Background(), itemID2, subFolderId, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}

	subFolderItems, err := GetFolderModel().ListItems(context.Background(), subFolderId, entity.FolderItemTypeAll, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 1, len(subFolderItems))
	assert.Equal(t, itemID2, subFolderItems[0].ID)
}

func TestMoveItemAndFolder(t *testing.T) {
	parentFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  "",
		Name:      RandName("TestFolder"),
		Thumbnail: "thumbnail-001",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log("parentFolder:", parentFolderId)

	subFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  parentFolderId,
		Name:      RandName("TestSubFolder"),
		Thumbnail: "thumbnail-002",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId)

	subFolderId2, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  parentFolderId,
		Name:      RandName("TestSubFolder"),
		Thumbnail: "thumbnail-003",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId2)

	itemID1, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
		FolderID:  subFolderId2,
		Link:      "content-5f6c0389bff85a26a0e86585",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(itemID1)

	itemID2, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
		FolderID:  subFolderId2,
		Link:      "content-5f6c04712ac27c637384d9cb",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(itemID2)

	err = GetFolderModel().MoveItem(context.Background(), itemID2, subFolderId, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}

	subFolderItems, err := GetFolderModel().ListItems(context.Background(), subFolderId, entity.FolderItemTypeAll, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 1, len(subFolderItems))
	assert.Equal(t, itemID2, subFolderItems[0].ID)

	err = GetFolderModel().MoveItem(context.Background(), subFolderId2, subFolderId, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	subFolderItems, err = GetFolderModel().ListItems(context.Background(), subFolderId, entity.FolderItemTypeAll, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 2, len(subFolderItems))
	for i := range subFolderItems {
		t.Log("ITEM:", subFolderItems[i].ID)
	}
}

func TestRemoveItemAndFolder(t *testing.T) {
	parentFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  "",
		Name:      RandName("TestFolder"),
		Thumbnail: "thumbnail-001",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log("parentFolder:", parentFolderId)

	subFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  parentFolderId,
		Name:      RandName("TestSubFolder"),
		Thumbnail: "thumbnail-002",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId)

	subFolderId2, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  parentFolderId,
		Name:      RandName("TestSubFolder"),
		Thumbnail: "thumbnail-003",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId2)

	itemID1, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
		FolderID:  subFolderId2,
		Link:      "content-5f6c0389bff85a26a0e86585",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(itemID1)

	itemID2, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
		FolderID:  subFolderId2,
		Link:      "content-5f6c04712ac27c637384d9cb",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(itemID2)

	err = GetFolderModel().RemoveItem(context.Background(), subFolderId2, fakeOperator())
	if !assert.Error(t, err) {
		return
	}

	err = GetFolderModel().RemoveItem(context.Background(), itemID2, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	subFolderItems, err := GetFolderModel().ListItems(context.Background(), subFolderId2, entity.FolderItemTypeAll, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 1, len(subFolderItems))


	err = GetFolderModel().RemoveItem(context.Background(), itemID1, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}

	err = GetFolderModel().RemoveItem(context.Background(), subFolderId2, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
}

func TestSearchFolder(t *testing.T) {
	parentFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  "",
		Name:      RandName("TestFolder"),
		Thumbnail: "thumbnail-001",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log("parentFolder:", parentFolderId)

	subFolderId, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  parentFolderId,
		Name:      RandName("TestSubFolder"),
		Thumbnail: "thumbnail-002",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId)

	subFolderId2, err := GetFolderModel().CreateFolder(context.Background(), entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentId:  parentFolderId,
		Name:      RandName("TestSubFolder"),
		Thumbnail: "thumbnail-003",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(subFolderId2)

	itemID1, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
		FolderID:  subFolderId2,
		Link:      "content-5f6c0389bff85a26a0e86585",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(itemID1)

	itemID2, err := GetFolderModel().AddItem(context.Background(), entity.CreateFolderItemRequest{
		FolderID:  subFolderId2,
		Link:      "content-5f6c04712ac27c637384d9cb",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(itemID2)

	total, folders, err := GetFolderModel().SearchOrgFolder(context.Background(), entity.SearchFolderCondition{
		Path:      "/",
	}, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Log(total)
	for i := range folders {
		t.Log(folders[i].ID)
	}

	folderInfo, err := GetFolderModel().GetFolderByID(context.Background(), folders[0].ID, fakeOperator())
	if !assert.NoError(t, err) {
		return
	}
	t.Logf("%#v\n", folderInfo)
}