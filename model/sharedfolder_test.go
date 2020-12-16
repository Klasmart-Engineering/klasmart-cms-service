package model

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func createFolders(t *testing.T) ([]string, error) {
	ctx := context.Background()
	fid1, err := GetFolderModel().CreateFolder(ctx, entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentID:  "/",
		Name:      "ShareFolder1",
		Partition: entity.FolderPartitionMaterialAndPlans,
	}, fakeOperator())
	assert.NoError(t, err)
	if err != nil {
		return nil, err
	}
	t.Log(fid1)

	fid2, err := GetFolderModel().CreateFolder(ctx, entity.CreateFolderRequest{
		OwnerType: entity.OwnerTypeOrganization,
		ParentID:  "/",
		Name:      "ShareFolder2",
		Partition: entity.FolderPartitionMaterialAndPlans,
	}, fakeOperator())
	assert.NoError(t, err)
	if err != nil {
		return nil, err
	}

	return []string{fid1, fid2}, nil
}

func fakeContentsID(t *testing.T) []string {
	return []string{
		"5f6c2e1c2caa9a07b6fc3496",
		"5f6c3528cbf5a918df9ad98d",
		"5f6c688b71536559d91f2344",
		"5f713e4462762537b1ed9e1b",
		"5f7153feefe8f4edb8199b9d",
		"5f7175f557337f28cf6d1968",
		"5f71763557337f28cf6d1989",
		"5f7294c906b1527fd862a8b2",
		"5f729b53a4d66d963495c667",
		"5f72d378badf2cf49654be4e",
		"5f72d435badf2cf49654be62",
		"5f72d97e7ab4e552b5e72f14",
		"5f72daf021595497e2a90ffd",
		"5f72e7a69a2e366350a9b854",
		"5f7abd3234729859a144c564",
		"5f7ff91a970387d696c09366"}
}

func TestShareFolderProcess(t *testing.T) {
	folderIDs, err := createFolders(t)
	if err != nil {
		return
	}
	contentIDs := fakeContentsID(t)

	folderItemInfos := make([]entity.FolderIdWithFileType, len(contentIDs))
	for i := range contentIDs {
		folderItemInfos[i] = entity.FolderIdWithFileType{
			ID:             contentIDs[i],
			FolderFileType: entity.FolderFileTypeContent,
		}
	}

	err = GetFolderModel().MoveItemBulk(context.Background(), entity.MoveFolderIDBulkRequest{
		FolderInfo: folderItemInfos,
		OwnerType:  entity.OwnerTypeOrganization,
		Dist:       folderIDs[0],
		Partition:  entity.FolderPartitionMaterialAndPlans,
	}, fakeOperator())

	assert.NoError(t, err)
	if err != nil {
		return
	}

	err = GetFolderModel().ShareFolders(context.Background(), entity.ShareFoldersRequest{
		FolderIDs: folderIDs,
		OrgIDs:    []string{"1", "2"},
	}, fakeOperator())
	assert.NoError(t, err)
	if err != nil {
		return
	}
}
