package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
)

const (
	PRESIGN_DURATION_MINUTES        = 60 * 24
	PRESIGN_UPLOAD_DURATION_MINUTES = 60
)

var (
	doOnce         sync.Once
	defaultStorage IStorage
)

type IStorage interface {
	OpenStorage(ctx context.Context) error
	CloseStorage(ctx context.Context)
	UploadFile(ctx context.Context, partition string, filePath string, fileStream multipart.File) error
	DownloadFile(ctx context.Context, partition string, filePath string) (io.Reader, error)
	ExitsFile(ctx context.Context, partition string, filePath string) (int64, bool)

	GetFilePath(ctx context.Context, partition string) string
	GetFileTempPath(ctx context.Context, partition string, filePath string) (string, error)

	GetUploadFileTempPath(ctx context.Context, partition string, fileName string) (string, error)
	GetUploadFileTempRawPath(ctx context.Context, tempPath string, fileName string) (string, error)

	UploadFileBytes(ctx context.Context, partition string, filePath string, fileStream *bytes.Buffer) error
	UploadFileLAN(ctx context.Context, partition string, filePath string, contentType string, r io.Reader) error
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
		os.Setenv("AWS_ACCESS_KEY_ID", os.Getenv("secret_id"))
		os.Setenv("AWS_SECRET_ACCESS_KEY", os.Getenv("secret_key"))
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
