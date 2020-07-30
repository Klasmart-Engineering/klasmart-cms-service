package model

import (
	"context"
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/storage"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"net/http"
	"sync"
)

var(
	ErrNoSuchURL = errors.New("no such url")
	ErrRequestItemIsNil = errors.New("request item is nil")
)

type IAssetModel interface {
	CreateAsset(ctx context.Context, data entity.AssetObject) (string, error)
	UpdateAsset(ctx context.Context, data entity.UpdateAssetRequest) error
	DeleteAsset(ctx context.Context, id string) error

	GetAssetByID(ctx context.Context, id string) (*entity.AssetObject, error)
	SearchAssets(ctx context.Context, condition *entity.SearchAssetCondition) ([]*entity.AssetObject, error)

	GetAssetUploadPath(ctx context.Context, extension string) (string, error)
}

type AssetModel struct{}

type AssetEntity struct {
	Category string
	Tag      []string
	URL      string
}

func (am AssetModel) checkEntity(ctx context.Context, entity AssetEntity, must bool) error {
	if must && (entity.URL == "" || entity.Category == "") {
		return ErrRequestItemIsNil
	}

	//TODO:Check if url is exists
	if entity.URL != "" {
		err := checkURL(entity.URL)
		if err != nil{
			return err
		}
	}
	//TODO:Check tag & category entity

	return nil
}

func checkURL(url string) error {
	resp, err := http.Get(url)
	if err != nil{
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return ErrNoSuchURL
	}
	return nil

}

func (am *AssetModel) CreateAsset(ctx context.Context, data entity.AssetObject) (string, error) {
	err := am.checkEntity(ctx, AssetEntity{
		Category: data.Category,
		Tag:      data.Tags,
		URL:      data.URL,
	}, true)

	if err != nil {
		return "", err
	}
	return da.GetAssetDA().CreateAsset(ctx, data)
}

func (am *AssetModel) UpdateAsset(ctx context.Context, data entity.UpdateAssetRequest) error {
	err := am.checkEntity(ctx, AssetEntity{
		Category: data.Category,
		Tag:      data.Tag,
		URL:      data.URL,
	}, false)

	if err != nil{
		return err
	}
	return da.GetAssetDA().UpdateAsset(ctx, data)
}

func (am *AssetModel) DeleteAsset(ctx context.Context, id string) error {
	return da.GetAssetDA().DeleteAsset(ctx, id)
}

func (am *AssetModel) GetAssetByID(ctx context.Context, id string) (*entity.AssetObject, error) {
	return da.GetAssetDA().GetAssetByID(ctx, id)
}

func (am *AssetModel) SearchAssets(ctx context.Context, condition *entity.SearchAssetCondition) ([]*entity.AssetObject, error) {
	return da.GetAssetDA().SearchAssets(ctx, (*da.SearchAssetCondition)(condition))
}

func (am *AssetModel) GetAssetUploadPath(ctx context.Context, extension string) (string, error) {
	client := storage.DefaultStorage()
	name := fmt.Sprintf("%s.%s", utils.NewID(), extension)

	return client.GetUploadFileTempPath(ctx, "asset", name)
}

var assetModel *AssetModel
var _assetOnce sync.Once

func GetAssetModel() *AssetModel {
	_assetOnce.Do(func() {
		assetModel = new(AssetModel)
	})
	return assetModel
}
