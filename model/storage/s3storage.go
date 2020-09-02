package storage

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront/sign"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3StorageConfig struct {
	Bucket    string
	Region    string
	ArnBucket string
}

type S3Storage struct {
	session   *session.Session
	bucket    string
	region    string
	arnBucket string
}

type CDNServiceRequest struct {
	URL       string   `json:"domain"`
	Duration  int      `json:"durationSeconds"`
	FilePaths []string `json:"filePaths"`
}

type CDNServiceResponse struct {
	Result []CDNServiceResult `json:"result"`
}

type CDNServiceResult struct {
	SignedURL string `json:"signedUrl"`
}

func (s *S3Storage) OpenStorage(ctx context.Context) error {
	//在~/.aws/credentials文件中保存secretId和secretKey

	sess, err := session.NewSession(&aws.Config{
		Region:          aws.String(s.region),
		S3UseAccelerate: aws.Bool(config.Get().StorageConfig.Accelerate),
	})
	if err != nil {
		log.Error(ctx, "Session create failed", log.Err(err))
		return err
	}

	log.Info(ctx, "Open s3 storage", log.String("bucket", s.bucket), log.String("region", s.region))
	s.session = sess
	return nil
}
func (s *S3Storage) CloseStorage(ctx context.Context) {

}

func getContentType(fileStream multipart.File) string {
	data := make([]byte, 512)
	fileStream.Read(data)

	t := http.DetectContentType(data)
	fileStream.Seek(0, io.SeekStart)
	return t
}

func getContentTypeBytes(fileStream *bytes.Buffer) string {
	data := make([]byte, 512)
	fileStream.Read(data)

	t := http.DetectContentType(data)
	fileStream.Reset()
	return t
}

func (s *S3Storage) UploadFileBytes(ctx context.Context, partition StoragePartition, filePath string, fileStream *bytes.Buffer) error {
	path := fmt.Sprintf("%s/%s", partition, filePath)
	uploader := s3manager.NewUploader(s.session)

	contentType := getContentTypeBytes(fileStream)

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(path),
		Body:        fileStream,
		ContentType: &contentType,
	})

	if err != nil {
		log.Error(ctx, "Object upload failed", log.Err(err))
		return err
	}
	return nil
}
func (s *S3Storage) UploadFile(ctx context.Context, partition StoragePartition, filePath string, fileStream multipart.File) error {
	path := fmt.Sprintf("%s/%s", partition, filePath)
	uploader := s3manager.NewUploader(s.session)
	contentType := getContentType(fileStream)

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(path),
		Body:        fileStream,
		ContentType: &contentType,
	})
	if err != nil {
		log.Error(ctx, "Object upload failed", log.Err(err))
		return err
	}
	return nil
}

func (s *S3Storage) UploadFileLAN(ctx context.Context, partition StoragePartition, filePath string, contentType string, r io.Reader) error {
	//建立session
	sess, err := session.NewSession(&aws.Config{
		Region:          aws.String(s.region),
		S3UseAccelerate: aws.Bool(false),
	})
	if err != nil {
		log.Error(ctx, "Session create failed", log.Err(err))
		return err
	}

	path := fmt.Sprintf("%s/%s", partition, filePath)
	uploader := s3manager.NewUploader(sess)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(path),
		Body:        r,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		log.Error(ctx, "Object upload failed", log.Err(err))
		return err
	}
	return nil
}

func (s *S3Storage) DownloadFile(ctx context.Context, partition StoragePartition, filePath string) (io.Reader, error) {
	path := fmt.Sprintf("%s/%s", partition, filePath)
	downloader := s3manager.NewDownloader(s.session)
	data := make([]byte, 1024)
	writerAt := aws.NewWriteAtBuffer(data)

	numBytes, err := downloader.Download(writerAt, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})

	log.Info(ctx, "Object download", log.Int64("size", numBytes))
	if err != nil {
		log.Warn(ctx, "Object download failed", log.Err(err))
		return nil, err
	}

	buffer := bytes.NewReader(writerAt.Bytes())
	return buffer, nil
}

func (s *S3Storage) ExistFile(ctx context.Context, partition StoragePartition, filePath string) (int64, bool) {
	//_, err := s.DownloadFile(ctx, partition, filePath)
	path := fmt.Sprintf("%s/%s", partition, filePath)
	svc := s3.New(s.session)
	res, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})

	if err != nil {
		return -1, false
	}
	return *res.ContentLength, true
}

func (s *S3Storage) GetFilePath(ctx context.Context, partition StoragePartition) string {
	return fmt.Sprintf("http://%s.s3-website-%s.amazonaws.com/%s/", s.bucket, s.region, partition)
}

func (s *S3Storage) CopyFile(ctx context.Context, source, target string) error {
	svc := s3.New(s.session)
	_, err := svc.CopyObject(&s3.CopyObjectInput{
		CopySource: aws.String(s.bucket + "/" + source),
		Key:        aws.String(target),
		Bucket:     aws.String(s.bucket),
	})
	if err != nil {
		return err
	}

	// Wait to see if the item got copied
	err = svc.WaitUntilObjectExists(&s3.HeadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(target)})
	if err != nil {
		return err
	}

	return nil
}

func (s *S3Storage) GetUploadFileTempRawPath(ctx context.Context, tempPath string, fileName string) (string, error) {
	path := fmt.Sprintf("%s/%s", tempPath, fileName)
	svc := s3.New(s.session)

	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})

	urlStr, err := req.Presign(PresignUploadDurationMinutes * time.Minute)

	if err != nil {
		log.Error(ctx, "Get presigned url failed", log.Err(err))
		return "", err
	}

	return urlStr, nil
}

func (s *S3Storage) GetUploadFileTempPath(ctx context.Context, partition StoragePartition, fileName string) (string, error) {
	path := fmt.Sprintf("%s/%s", partition, fileName)
	svc := s3.New(s.session)

	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
		ContentLength: aws.Int64(partition.SizeLimit()),
	})

	urlStr, err := req.Presign(PresignUploadDurationMinutes * time.Minute)

	if err != nil {
		log.Error(ctx, "Get presigned url failed", log.Err(err))
		return "", err
	}
	return urlStr, nil
}

func (s *S3Storage) GetFileTempPath(ctx context.Context, partition StoragePartition, filePath string) (string, error) {
	log.Info(ctx, "Must Get CDN config", log.Any("config", config.Get().CDNConfig))
	if config.Get().CDNConfig.CDNOpen {
		switch config.Get().CDNConfig.CDNMode {
		case "service":
			return s.GetFileTempPathForCDNByService(ctx, partition, filePath)
		case "key":
			fallthrough
		default:
			return s.GetFileTempPathForCDN(ctx, partition, filePath)
		}
	}
	path := fmt.Sprintf("%s/%s", partition, filePath)
	svc := s3.New(s.session)

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})

	urlStr, err := req.Presign(PresignDurationMinutes * time.Minute)

	if err != nil {
		log.Error(ctx, "Get presigned url failed", log.Err(err))
		return "", err
	}

	return urlStr, nil
}

func (s *S3Storage) GetFileTempPathForCDN(ctx context.Context, partition StoragePartition, filePath string) (string, error) {
	cdnConf := config.Get().CDNConfig

	path := fmt.Sprintf("%s/%s/%s", cdnConf.CDNPath, partition, filePath)
	keyID := cdnConf.CDNKeyId

	privateKeyDer, err := base64.StdEncoding.DecodeString(cdnConf.CDNPrivateKey)
	if err != nil {
		return "", err
	}
	privKey, err := x509.ParsePKCS1PrivateKey(privateKeyDer)
	if err != nil {
		return "", err
	}

	signer := sign.NewURLSigner(keyID, privKey)
	signedURL, err := signer.Sign(path, time.Now().Add(PresignDurationMinutes*time.Minute))
	if err != nil {
		log.Error(ctx, "Get presigned url failed", log.Err(err))
		return "", err
	}

	return signedURL, nil
}

func (s *S3Storage) GetFileTempPathForCDNByService(ctx context.Context, partition StoragePartition, filePath string) (string, error) {
	cdnConf := config.Get().CDNConfig

	params := &CDNServiceRequest{
		URL:       cdnConf.CDNPath,
		Duration:  PresignDurationMinutes * 60,
		FilePaths: []string{fmt.Sprintf("%s/%s", partition, filePath)},
	}
	data, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequest("POST", cdnConf.CDNServicePath, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("charset", "utf-8")
	request.Header.Set("Authorization", "Bearer "+cdnConf.CDNServiceToken)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	respData := new(CDNServiceResponse)
	err = json.Unmarshal(respBody, respData)
	if err != nil {
		return "", err
	}
	if len(respData.Result) < 1 {
		return "", ErrInvalidCDNSignatureServiceResponse
	}

	return respData.Result[0].SignedURL, nil
}

func newS3Storage(c S3StorageConfig) IStorage {
	return &S3Storage{
		bucket:    c.Bucket,
		region:    c.Region,
		arnBucket: c.ArnBucket,
	}
}
