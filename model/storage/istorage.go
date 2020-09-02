package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
)

var (
	ErrInvalidUploadPartition = errors.New("unknown storage partition")
)

const (
	AssetStoragePartition        StoragePartition      = "asset"
	ThumbnailStoragePartition      StoragePartition    = "thumbnail"
	ScheduleAttachmentStoragePartition StoragePartition = "schedule_attachment"

)
type StoragePartition string

func (s StoragePartition) SizeLimit() int64{
	switch s {
	case AssetStoragePartition:
		return 1024 * 1000
	case ThumbnailStoragePartition:
		return 1024 * 1024 * 5
	case ScheduleAttachmentStoragePartition:
		return 1024 * 100
	}
	return 0
}

func NewStoragePartition(partition string) (StoragePartition, error){
	switch partition {
	case string(AssetStoragePartition):
		return AssetStoragePartition, nil
	case string(ThumbnailStoragePartition):
		return ThumbnailStoragePartition, nil
	case string(ScheduleAttachmentStoragePartition):
		return ScheduleAttachmentStoragePartition, nil
	}
	return "", ErrInvalidUploadPartition
}

var (
	doOnce         sync.Once
	defaultStorage IStorage
)

var (
	ErrInvalidCDNSignatureServiceResponse = errors.New("invalid cdn signature service response")
)

type IStorage interface {
	OpenStorage(ctx context.Context) error
	CloseStorage(ctx context.Context)
	UploadFile(ctx context.Context, partition StoragePartition, filePath string, fileStream multipart.File) error
	DownloadFile(ctx context.Context, partition StoragePartition, filePath string) (io.Reader, error)
	ExistFile(ctx context.Context, partition StoragePartition, filePath string) (int64, bool)

	GetFilePath(ctx context.Context, partition StoragePartition) string
	GetFileTempPath(ctx context.Context, partition StoragePartition, filePath string) (string, error)

	GetUploadFileTempPath(ctx context.Context, partition StoragePartition, fileName string) (string, error)
	GetUploadFileTempRawPath(ctx context.Context, tempPath string, fileName string) (string, error)

	UploadFileBytes(ctx context.Context, partition StoragePartition, filePath string, fileStream *bytes.Buffer) error
	UploadFileLAN(ctx context.Context, partition StoragePartition, filePath string, contentType string, r io.Reader) error
	CopyFile(ctx context.Context, source, target string) error
}

func assertGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("Environment %v is nil", key))
	}
	return value
}

//根据环境变量创建存储对象
func createStorageByEnv() {
	conf := config.Get()

	switch conf.StorageConfig.CloudEnv {
	case "aws":
		defaultStorage = newS3Storage(S3StorageConfig{
			Bucket:    conf.StorageConfig.StorageBucket,
			Region:    conf.StorageConfig.StorageRegion,
			ArnBucket: os.Getenv("storage_arn_bucket"),
		})
		defaultStorage.OpenStorage(context.TODO())
	default:
		panic("Environment CLOUD_ENV is nil")
	}
}
func DefaultStorage() IStorage {
	doOnce.Do(func() {
		if defaultStorage == nil {
			createStorageByEnv()
		}
	})
	return defaultStorage
}
