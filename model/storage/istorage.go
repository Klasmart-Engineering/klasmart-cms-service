package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"sync"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"

	"github.com/KL-Engineering/kidsloop-cms-service/config"
)

var (
	ErrInvalidUploadPartition          = errors.New("unknown storage partition")
	ErrInvalidPrivateKeyFile           = errors.New("invalid private key file")
	ErrInvalidExtensionInPartitionFile = errors.New("invalid extension in partition")
)

const (
	AssetStoragePartition              StoragePartition = "assets"
	ThumbnailStoragePartition          StoragePartition = "thumbnail"
	TeacherManualStoragePartition      StoragePartition = "teacher_manual"
	ScheduleAttachmentStoragePartition StoragePartition = "schedule_attachment"
	DrawingFeedbackStoragePartition    StoragePartition = "drawing_feedback"
)

type StoragePartition string

func (s StoragePartition) SizeLimit() int64 {
	switch s {
	case AssetStoragePartition:
		// 1 GB
		return 1024 * 1024 * 1024
	case ThumbnailStoragePartition:
		// 20 MB
		return 1024 * 1024 * 20
	case ScheduleAttachmentStoragePartition:
		// 1 GB
		return 1024 * 1024 * 1024
	case DrawingFeedbackStoragePartition:
		// 1 GB
		return 1024 * 1024 * 1024
	}
	// 500 MB
	return 1024 * 1024 * 500
}

func NewStoragePartition(ctx context.Context, partition, extension string) (StoragePartition, error) {
	extension = strings.ToLower(extension)
	switch partition {
	case string(AssetStoragePartition):
		ret := utils.CheckInStringArray(extension, constant.MaterialsExtension)
		if !ret {
			log.Warn(ctx, "Check partition extension failed",
				log.String("extension", extension),
				log.String("partition", partition),
				log.Strings("expected", constant.MaterialsExtension))
			return "", ErrInvalidExtensionInPartitionFile
		}
		return AssetStoragePartition, nil
	case string(ThumbnailStoragePartition):
		ret := utils.CheckInStringArray(extension, constant.AssetsImageExtension)
		if !ret {
			log.Warn(ctx, "Check partition extension failed",
				log.String("extension", extension),
				log.String("partition", partition),
				log.Strings("expected", constant.AssetsImageExtension))
			return "", ErrInvalidExtensionInPartitionFile
		}
		return ThumbnailStoragePartition, nil
	case string(ScheduleAttachmentStoragePartition):
		return ScheduleAttachmentStoragePartition, nil
	case string(DrawingFeedbackStoragePartition):
		return DrawingFeedbackStoragePartition, nil
	case string(TeacherManualStoragePartition):
		ret := utils.CheckInStringArray(extension, constant.TeacherManualExtension)
		if !ret {
			log.Warn(ctx, "Check partition extension failed",
				log.String("extension", extension),
				log.String("partition", partition),
				log.Strings("expected", constant.TeacherManualExtension))
			return "", ErrInvalidExtensionInPartitionFile
		}
		return TeacherManualStoragePartition, nil
	}
	log.Warn(ctx, "Invalid upload partition",
		log.String("extension", extension),
		log.String("partition", partition))
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

	UploadFileBytes(ctx context.Context, partition StoragePartition, filePath string, fileStream *bytes.Buffer) error
	UploadFileLAN(ctx context.Context, partition StoragePartition, filePath string, contentType string, r io.Reader) error
	CopyFile(ctx context.Context, source, target string) error
}

//根据环境变量创建存储对象
func createStorageByEnv() {
	conf := config.Get()

	switch conf.StorageConfig.StorageProtocol {
	case "s3":
		defaultStorage = newS3Storage(S3StorageConfig{
			Endpoint:   conf.StorageConfig.StorageEndPoint,
			Bucket:     conf.StorageConfig.StorageBucket,
			Region:     conf.StorageConfig.StorageRegion,
			ArnBucket:  os.Getenv("storage_arn_bucket"),
			Accelerate: config.Get().StorageConfig.Accelerate,
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
