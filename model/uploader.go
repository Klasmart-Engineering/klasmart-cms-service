package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/storage"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"strings"
	"sync"
)

type IResourceUploaderModel interface {
	GetResourceUploadPath(ctx context.Context, partition string, extension string) (string, string, error)
	GetResourcePath(ctx context.Context, resourceId string) (string, error)
}

type ResourceUploaderModel struct {
}

func (r *ResourceUploaderModel) GetResourceUploadPath(ctx context.Context, partition string, extension string) (string, string, error) {
	fileName := utils.NewID() + "." + extension
	pat, err := storage.NewStoragePartition(ctx, partition, extension)
	if err != nil {
		log.Error(ctx, "invalid partition", log.Err(err), log.String("partition", partition), log.String("extension", extension))
		return "", "", err
	}
	path, err := storage.DefaultStorage().GetUploadFileTempPath(ctx, pat, fileName)
	if err != nil {
		log.Error(ctx, "get upload file temp path failed", log.Err(err), log.String("partition", partition), log.String("extension", extension))
		return "", "", err
	}
	return partition + "-" + fileName, path, nil
}

func (r *ResourceUploaderModel) GetResourcePath(ctx context.Context, resourceId string) (string, error) {
	parts := strings.Split(resourceId, "-")
	if len(parts) != 2 {
		log.Error(ctx, "invalid resource id", log.String("resourceId", resourceId))
		return "", ErrInvalidResourceID
	}
	extensionPairs := strings.Split(parts[1], ".")
	if len(extensionPairs) != 2 {
		log.Error(ctx, "invalid extension", log.String("resourceId", resourceId))
		return "", ErrInvalidResourceID
	}

	pat, err := storage.NewStoragePartition(ctx, parts[0], extensionPairs[1])
	if err != nil {
		log.Error(ctx, "invalid partition", log.Err(err), log.String("resourceId", resourceId), log.Strings("parts", parts))
		return "", err
	}
	return storage.DefaultStorage().GetFileTempPath(ctx, pat, parts[1])
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
