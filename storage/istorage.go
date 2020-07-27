package storage

import (
	"bytes"
	"calmisland/kidsloop2/conf"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"sync"
)

const(
	PRESIGN_DURATION_MINUTES = 60 * 24
	PRESIGN_UPLOAD_DURATION_MINUTES = 60
)

var (
	doOnce sync.Once
	defaultStorage IStorage
)

type IStorage interface{
	OpenStorage(ctx context.Context) error
	CloseStorage(ctx context.Context)
	UploadFile(ctx context.Context, partition int, filePath string, fileStream multipart.File) error
	DownloadFile(ctx context.Context, partition int, filePath string) (io.Reader, error)
	ExitsFile(ctx context.Context, partition int, filePath string) bool
	GetFilePath(ctx context.Context, partition int) string
	GetFileTempPath(ctx context.Context, partition int, filePath string) (string, error)

	GetUploadFileTempPath(ctx context.Context, partition int, fileName string) (string ,error)
	GetUploadFileTempRawPath(ctx context.Context, tempPath string, fileName string) (string ,error)

	UploadFileBytes(ctx context.Context, partition int, filePath string, fileStream *bytes.Buffer) error
	UploadFileLAN(ctx context.Context, partition int, filePath string, contentType string, r io.Reader) error
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
func createStorageByEnv(){
	config := conf.Get()

	switch config.StorageConfig.CloudEnv{
	case "aws":
		assertGetEnv("AWS_ACCESS_KEY_ID")
		assertGetEnv("AWS_SECRET_ACCESS_KEY")
		defaultStorage = newS3Storage(S3StorageConfig{
			Bucket: config.StorageConfig.StorageBucket,
			Region: config.StorageConfig.StorageRegion,
			ArnBucket: os.Getenv("storage_arn_bucket"),
		})
		defaultStorage.OpenStorage(context.TODO())
	default:
		panic("Environment CLOUD_ENV is nil")
	}
}
func DefaultStorage() IStorage{
	doOnce.Do(func() {
		if defaultStorage == nil {
			createStorageByEnv()
		}
	})
	return defaultStorage
}