package model

import (
	"context"
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/storage"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

const (
	Asset_Storage_Partition = "asset"
)

var (
	ErrNoSuchURL        = errors.New("no such url")
	ErrRequestItemIsNil = errors.New("request item is nil")
)

type IAssetModel interface {
	CreateAsset(ctx context.Context, data entity.AssetObject) (string, error)
	UpdateAsset(ctx context.Context, data entity.UpdateAssetRequest) error
	DeleteAsset(ctx context.Context, id string) error

	GetAssetByID(ctx context.Context, id string) (*entity.AssetObject, error)
	SearchAssets(ctx context.Context, condition *entity.SearchAssetCondition) (int64, []*entity.AssetObject, error)

	GetAssetUploadPath(ctx context.Context, extension string) (*entity.ResourcePath, error)
	GetAssetResourcePath(ctx context.Context, name string) (string, error)
}

type AssetModel struct{}

type AssetEntity struct {
	Category string
	Tag      []string
}

func (am AssetModel) checkResource(ctx context.Context, url string, must bool) (int64, error) {
	if must && url == "" {
		return -1, ErrRequestItemIsNil
	}
	if url != "" {
		size, exist := storage.DefaultStorage().ExitsFile(ctx, Asset_Storage_Partition, url)
		if !exist {
			return -1, ErrNoSuchURL
		}
		return size, nil
	}
	return 0, nil
}

func (am AssetModel) checkEntity(ctx context.Context, entity AssetEntity, must bool) error {
	if must && entity.Category == "" {
		return ErrRequestItemIsNil
	}

	//TODO:Check tag & category entity
	_, err := GetCategoryModel().GetCategoryByID(ctx, entity.Category)
	if err != nil {
		log.Error(ctx, "Invalid category ", log.Err(err))
		return err
	}

	return nil
}

func (am *AssetModel) CreateAsset(ctx context.Context, data entity.AssetObject) (string, error) {
	err := am.checkEntity(ctx, AssetEntity{
		Category: data.Category,
		Tag:      data.Tags,
	}, true)
	if err != nil {
		return "", err
	}

	data.Size = 0
	size, err := am.checkResource(ctx, data.ResourceName, true)
	if err != nil {
		return "", err
	}
	data.Size = size

	return da.GetAssetDA().CreateAsset(ctx, data)
}

func (am *AssetModel) UpdateAsset(ctx context.Context, data entity.UpdateAssetRequest) error {
	err := am.checkEntity(ctx, AssetEntity{
		Category: data.Category,
		Tag:      data.Tag,
	}, false)
	if err != nil {
		return err
	}

	data.Size = 0
	if data.ResourceName != "" {
		size, err := am.checkResource(ctx, data.ResourceName, true)
		if err != nil {
			return err
		}
		data.Size = size
	}

	return da.GetAssetDA().UpdateAsset(ctx, data)
}

func (am *AssetModel) DeleteAsset(ctx context.Context, id string) error {
	return da.GetAssetDA().DeleteAsset(ctx, id)
}

func (am *AssetModel) GetAssetByID(ctx context.Context, id string) (*entity.AssetObject, error) {
	return da.GetAssetDA().GetAssetByID(ctx, id)
}

func (am *AssetModel) SearchAssets(ctx context.Context, condition *entity.SearchAssetCondition) (int64, []*entity.AssetObject, error) {
	return da.GetAssetDA().SearchAssets(ctx, (*da.SearchAssetCondition)(condition))
}

func (am *AssetModel) GetAssetUploadPath(ctx context.Context, extension string) (*entity.ResourcePath, error) {
	storage := storage.DefaultStorage()
	name := fmt.Sprintf("%s.%s", utils.NewID(), extension)

	path, err := storage.GetUploadFileTempPath(ctx, Asset_Storage_Partition, name)
	if err != nil {
		return nil, err
	}
	return &entity.ResourcePath{
		Path: path,
		Name: name,
	}, nil
}

func (am *AssetModel) GetAssetResourcePath(ctx context.Context, name string) (string, error) {
	storage := storage.DefaultStorage()
	return storage.GetFileTempPath(ctx, Asset_Storage_Partition, name)
}

var assetModel *AssetModel
var _assetOnce sync.Once

func GetAssetModel() *AssetModel {
	_assetOnce.Do(func() {
		assetModel = new(AssetModel)
	})
	return assetModel
}
