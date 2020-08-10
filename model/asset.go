package model

import (
	"context"
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/storage"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

const(
	Asset_Storage_Partition = "asset"
	Thumbnail_Storage_Partition = "thumbnail"
)

var(
	ErrNoSuchURL = errors.New("no such url")
	ErrRequestItemIsNil = errors.New("request item is nil")
	ErrNoAuth = errors.New("no auth to operate")
)

type IAssetModel interface {
	CreateAsset(ctx context.Context, data entity.AssetObject, operator entity.Operator) (int64, error)
	UpdateAsset(ctx context.Context, data entity.UpdateAssetRequest, operator entity.Operator) error
	DeleteAsset(ctx context.Context, id int64, operator entity.Operator) error

	GetAssetByID(ctx context.Context, id int64, operator entity.Operator) (*entity.AssetObject, error)
	SearchAssets(ctx context.Context, condition *entity.SearchAssetCondition, operator entity.Operator) (int64, []*entity.AssetObject, error)

	GetAssetUploadPath(ctx context.Context, extension string, operator entity.Operator) (*entity.ResourcePath, error)
	GetAssetResourcePath(ctx context.Context, name string, operator entity.Operator) (string ,error)
}

type AssetModel struct{}

type AssetEntity struct {
}
type AssetSource struct {
	assetSource string
	thumbnailSource string
}

func (am AssetModel) checkResource(ctx context.Context, data AssetSource, must bool)(int64, error){
	if must && (data.assetSource == "" || data.thumbnailSource == "") {
		return -1, ErrRequestItemIsNil
	}
	size := int64(0)
	if data.assetSource != "" {
		tempSize, exist := storage.DefaultStorage().ExitsFile(ctx, Asset_Storage_Partition, data.assetSource)
		if !exist {
			return -1, ErrNoSuchURL
		}
		size = tempSize
	}

	if data.assetSource != "" {
		_, exist := storage.DefaultStorage().ExitsFile(ctx, Thumbnail_Storage_Partition, data.thumbnailSource)
		if !exist {
			return -1, ErrNoSuchURL
		}
	}
	return size, nil
}

func (am AssetModel) checkEntity(ctx context.Context, entity AssetEntity, must bool) error {
	return nil
}

func (am *AssetModel) CreateAsset(ctx context.Context, data entity.AssetObject, operator entity.Operator) (int64, error) {
	err := am.checkEntity(ctx, AssetEntity{}, true)
	if err != nil {
		return -1, err
	}

	data.Size = 0
	size, err := am.checkResource(ctx, AssetSource{
		assetSource:     data.Resource,
		thumbnailSource: data.Thumbnail,
	}, true)
	if err != nil {
		return -1, err
	}
	data.Size = size

	//TODO: get user name
	data.Org = operator.OrgID
	data.Author = operator.UserID
	data.AuthorName = operator.UserID

	return da.GetAssetDA().CreateAsset(ctx, data)
}

func (am *AssetModel) UpdateAsset(ctx context.Context, data entity.UpdateAssetRequest, operator entity.Operator) error {
	assets, err := am.GetAssetByID(ctx, data.ID, operator)
	if err != nil{
		return err
	}
	if assets.Author != operator.UserID {
		return ErrNoAuth
	}

	err = am.checkEntity(ctx, AssetEntity{}, false)
	if err != nil{
		return err
	}

	return da.GetAssetDA().UpdateAsset(ctx, data)
}

func (am *AssetModel) DeleteAsset(ctx context.Context, id int64, operator entity.Operator) error {
	assets, err := am.GetAssetByID(ctx, id, operator)
	if err != nil{
		return err
	}
	if assets.Author != operator.UserID {
		return ErrNoAuth
	}

	return da.GetAssetDA().DeleteAsset(ctx, id)
}

func (am *AssetModel) GetAssetByID(ctx context.Context, id int64, operator entity.Operator) (*entity.AssetObject, error) {
	return da.GetAssetDA().GetAssetByID(ctx, id)
}

func (am *AssetModel) SearchAssets(ctx context.Context, condition *entity.SearchAssetCondition, operator entity.Operator) (int64, []*entity.AssetObject, error) {
	cd := &da.SearchAssetCondition{
		ID:          condition.ID,
		SearchWords: condition.SearchWords,
		OrgID:  	 operator.OrgID,
		OrderBy:     da.NewAssetsOrderBy(condition.OrderBy),
		PageSize:    condition.PageSize,
		Page:        condition.Page,
	}
	if condition.IsSelf {
		cd.Author = []string{operator.UserID}
	}
	return da.GetAssetDA().SearchAssets(ctx, cd)
}

func (am *AssetModel) GetAssetUploadPath(ctx context.Context, extension string, operator entity.Operator) (*entity.ResourcePath, error) {
	storage := storage.DefaultStorage()
	name := fmt.Sprintf("%s.%s", utils.NewID(), extension)

	path, err := storage.GetUploadFileTempPath(ctx, Asset_Storage_Partition, name)
	if err != nil{
		return nil, err
	}
	return &entity.ResourcePath{
		Path: path,
		Name: name,
	}, nil
}

func (am *AssetModel) GetAssetResourcePath(ctx context.Context, name string, operator entity.Operator) (string ,error){
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
