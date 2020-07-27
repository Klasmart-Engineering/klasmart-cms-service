package storage

import (
	"bytes"
	"calmisland/kidsloop2/log"
	"context"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

type CosStorageConfig struct {
	Bucket    string
	Region    string
	SecretId  string
	SecretKey string
}

type CosStorage struct {
	client    *cos.Client
	bucket    string
	region    string
	secretId  string
	secretKey string
}

func (c *CosStorage) OpenStorage(ctx context.Context) error {
	u, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", c.bucket, c.region))
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		//设置超时时间
		Timeout: 100 * time.Second,
		Transport: &cos.AuthorizationTransport{
			//如实填写账号和密钥，也可以设置为环境变量
			SecretID:  c.secretId,
			SecretKey: c.secretKey,
		},
	})

	log.Get().Infof("Open cos storage, bucket: %v, region: %v", c.bucket, c.region)
	c.client = client
	return nil
}

func (c *CosStorage) CloseStorage(ctx context.Context) {
}

func (c *CosStorage) UploadFileBytes(ctx context.Context, partition int, filePath string, fileStream *bytes.Buffer) error{
	path := fmt.Sprintf("%d/%s", partition, filePath)
	_, err := c.client.Object.Put(ctx, path, fileStream, nil)
	if err != nil {
		return err
	}
	return nil
}
func (c *CosStorage) UploadFile(ctx context.Context, partition int, filePath string, fileStream multipart.File) error {
	path := fmt.Sprintf("%d/%s", partition, filePath)
	_, err := c.client.Object.Put(ctx, path, fileStream, nil)
	if err != nil {
		return err
	}
	return nil
}
func (c *CosStorage) UploadFileLAN(ctx context.Context, partition int, filePath string, contentType string, f io.Reader) error {
	path := fmt.Sprintf("%d/%s", partition, filePath)
	_, err := c.client.Object.Put(ctx, path, f, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *CosStorage) DownloadFile(ctx context.Context, partition int, filePath string) (io.Reader, error) {
	path := fmt.Sprintf("%d/%s", partition, filePath)
	log.Get().Debugf("Get object, time: %v", time.Now())
	resp, err := c.client.Object.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	log.Get().Debugf("Object got, time: %v", time.Now())
	return resp.Body, nil
}

func (c *CosStorage) ExitsFile(ctx context.Context, partition int, filePath string) bool {
	path := fmt.Sprintf("%d/%s", partition, filePath)
	resp, err := c.client.Object.Get(ctx, path, nil)
	if err != nil {
		return false
	}
	return resp.Body == nil
}

func (c *CosStorage) CopyFile(ctx context.Context, source, target string) error{
	sourcePath := fmt.Sprintf("%s.cos.%s.myqcloud.com/%v", c.bucket, c.region, source)
	//targetPath := fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%v", c.bucket, c.region, target)
	_, _, err := c.client.Object.Copy(ctx, target, sourcePath, nil)
	return err
}

func (c *CosStorage) GetFilePath(ctx context.Context, partition int) string {
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%d/", c.bucket, c.region, partition)
}

func (c *CosStorage) GetFileTempPath(ctx context.Context, partition int, filePath string) (string, error) {
	path := fmt.Sprintf("%d/%s", partition, filePath)
	url, err := c.client.Object.GetPresignedURL(ctx, http.MethodGet, path, c.secretId, c.secretKey, time.Minute*PRESIGN_DURATION_MINUTES, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (c *CosStorage) GetUploadFileTempPath(ctx context.Context, partition int, fileName string) (string ,error) {
	path := fmt.Sprintf("%d/%s", partition, fileName)
	url, err := c.client.Object.GetPresignedURL(ctx, http.MethodPut, path, c.secretId, c.secretKey, time.Minute*PRESIGN_UPLOAD_DURATION_MINUTES, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (c *CosStorage) GetUploadFileTempRawPath(ctx context.Context, tempPath string, fileName string) (string ,error) {
	path := fmt.Sprintf("%s/%s", tempPath, fileName)
	url, err := c.client.Object.GetPresignedURL(ctx, http.MethodPut, path, c.secretId, c.secretKey, time.Minute*PRESIGN_UPLOAD_DURATION_MINUTES, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func newCosStorage(c CosStorageConfig) IStorage {
	return &CosStorage{
		bucket:    c.Bucket,
		region:    c.Region,
		secretId:  c.SecretId,
		secretKey: c.SecretKey,
	}
}
