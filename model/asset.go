package model

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/storage"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

const (
	Asset_Storage_Partition              = "asset"
	Thumbnail_Storage_Partition          = "thumbnail"
	ScheduleAttachment_Storage_Partition = "schedule_attachment"
)

var (
	ErrNoSuchURL           = errors.New("no such url")
	ErrRequestItemIsNil    = errors.New("request item is nil")
	ErrNoAuth              = errors.New("no auth to operate")
	ErrCreateContentFailed = errors.New("create contentdata into data access failed")

	ErrNoContentData                 = errors.New("no content data")
	ErrNoContent                     = errors.New("no content")
	ErrContentAlreadyLocked          = errors.New("content is already locked")
	ErrInvalidPublishStatus          = errors.New("invalid publish status")
	ErrGetUnpublishedContent         = errors.New("unpublished content")
	ErrGetUnauthorizedContent        = errors.New("unauthorized content")
	ErrCloneContentFailed            = errors.New("clone content failed")
	ErrParseContentDataFailed        = errors.New("parse content data failed")
	ErrParseContentDataDetailsFailed = errors.New("parse content data details failed")
	ErrUpdateContentFailed           = errors.New("update contentdata into data access failed")
	ErrInvalidContentStatusToPublish = errors.New("content status is invalid to publish")
	ErrReadContentFailed             = errors.New("read content failed")
	ErrDeleteContentFailed           = errors.New("delete contentdata into data access failed")
)

type IAssetModel interface {
	CreateAsset(ctx context.Context, data entity.CreateAssetData, operator entity.Operator) (string, error)
	UpdateAsset(ctx context.Context, data entity.UpdateAssetRequest, operator entity.Operator) error
	DeleteAsset(ctx context.Context, id string, operator entity.Operator) error

	GetAssetByID(ctx context.Context, id string, operator entity.Operator) (*entity.AssetData, error)
	SearchAssets(ctx context.Context, condition *entity.SearchAssetCondition, operator entity.Operator) (int64, []*entity.AssetData, error)

	GetAssetUploadPath(ctx context.Context, extension string, operator entity.Operator) (*entity.ResourcePath, error)
	GetAssetResourcePath(ctx context.Context, name string, operator entity.Operator) (string, error)
}

type AssetModel struct{}

type AssetEntity struct {
}
type AssetSource struct {
	assetSource     string
	thumbnailSource string
}

func (am AssetModel) checkResource(ctx context.Context, data AssetSource, must bool) (int64, error) {
	if must && (data.assetSource == "" || data.thumbnailSource == "") {

		return -1, ErrRequestItemIsNil
	}
	size := int64(0)
	if data.assetSource != "" {
		tempSize, exist := storage.DefaultStorage().ExistFile(ctx, Asset_Storage_Partition, data.assetSource)
		if !exist {
			return -1, ErrNoSuchURL
		}
		size = tempSize
	}

	if data.assetSource != "" {
		_, exist := storage.DefaultStorage().ExistFile(ctx, Thumbnail_Storage_Partition, data.thumbnailSource)
		if !exist {
			return -1, ErrNoSuchURL
		}

	}
	return size, nil
}

func (am AssetModel) checkEntity(ctx context.Context, entity AssetEntity, must bool) error {
	return nil
}

func (am *AssetModel) CreateAsset(ctx context.Context, req entity.CreateAssetData, operator entity.Operator) (string, error) {
	err := am.checkEntity(ctx, AssetEntity{}, true)

	if err != nil {
		return "", err
	}

	size, err := am.checkResource(ctx, AssetSource{
		assetSource:     req.Resource,
		thumbnailSource: req.Thumbnail,
	}, true)
	if err != nil {
		return "", err
	}
	data := req.ToAssetObject()

	data.Size = size

	//TODO: get user name
	data.Org = operator.OrgID
	data.Author = operator.UserID
	data.AuthorName = operator.UserID

	return da.GetAssetDA().CreateAsset(ctx, *data)
}

func (am *AssetModel) UpdateAsset(ctx context.Context, data entity.UpdateAssetRequest, operator entity.Operator) error {
	assets, err := am.GetAssetByID(ctx, data.ID, operator)

	if err != nil {
		return err
	}
	if assets.Author != operator.UserID {
		return ErrNoAuth
	}

	err = am.checkEntity(ctx, AssetEntity{}, false)
	if err != nil {
		return err

	}

	return da.GetAssetDA().UpdateAsset(ctx, data)
}

func (am *AssetModel) DeleteAsset(ctx context.Context, id string, operator entity.Operator) error {
	assets, err := am.GetAssetByID(ctx, id, operator)
	if err != nil {
		return err
	}
	if assets.Author != operator.UserID {
		return ErrNoAuth
	}

	return da.GetAssetDA().DeleteAsset(ctx, id)
}

func (am *AssetModel) GetAssetByID(ctx context.Context, id string, operator entity.Operator) (*entity.AssetData, error) {
	res, err := da.GetAssetDA().GetAssetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return res.ToAssetData(), nil
}

func (am *AssetModel) SearchAssets(ctx context.Context, condition *entity.SearchAssetCondition, operator entity.Operator) (int64, []*entity.AssetData, error) {
	cd := &da.SearchAssetCondition{
		ID:          condition.ID,
		SearchWords: condition.SearchWords,
		FuzzyQuery:  condition.FuzzyQuery,
		OrgID:       operator.OrgID,
		OrderBy:     da.NewAssetsOrderBy(condition.OrderBy),
		PageSize:    condition.PageSize,
		Page:        condition.Page,
	}
	if condition.IsSelf {
		cd.Author = []string{operator.UserID}
	}
	count, res, err := da.GetAssetDA().SearchAssets(ctx, cd)
	if err != nil {
		return count, nil, err
	}
	data := make([]*entity.AssetData, len(res))
	for i := range res {
		data[i] = res[i].ToAssetData()
	}
	return count, data, nil
}

func (am *AssetModel) GetAssetUploadPath(ctx context.Context, extension string, operator entity.Operator) (*entity.ResourcePath, error) {
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

func (am *AssetModel) GetAssetResourcePath(ctx context.Context, name string, operator entity.Operator) (string, error) {
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
