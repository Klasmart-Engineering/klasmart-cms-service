package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/storage"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"strings"
	"sync"
)

type IResourceUploaderModel interface{
	GetResourceUploadPath(ctx context.Context, partition string, extension string) (string, string, error)
	GetResourcePath(ctx context.Context, resourceId string) (string, error)
}

type ResourceUploaderModel struct {

}

func (r *ResourceUploaderModel) GetResourceUploadPath(ctx context.Context, partition string, extension string) (string, string, error) {
	fileName := utils.NewID() + "." + extension
	path, err := storage.DefaultStorage().GetUploadFileTempPath(ctx, partition, fileName)
	if err != nil{
		return "", "", err
	}
	return partition + "-" + fileName, path, nil
}

func (r *ResourceUploaderModel) GetResourcePath(ctx context.Context, resourceId string) (string, error) {
	parts := strings.Split(resourceId, "-")
	if len(parts) != 2 {
		return "", ErrInvalidResourceId
	}
	return storage.DefaultStorage().GetFileTempPath(ctx, parts[0], parts[1])
}
var (
	_uploaderModel     IResourceUploaderModel
	_uploaderModelOnce sync.Once
)

func GetResourceUploaderModel() IResourceUploaderModel {
	_uploaderModelOnce.Do(func() {
		_uploaderModel = new(ResourceUploaderModel)
	})
	return _uploaderModel
}

