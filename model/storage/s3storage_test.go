package storage

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/aws/aws-sdk-go/service/cloudfront/sign"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
)

func InitEnv() {
	os.Setenv("cloud_env", "aws")
	os.Setenv("storage_bucket", "kidsloop-global-resources-test")
	os.Setenv("storage_region", "ap-northeast-2")
	os.Setenv("AWS_ACCESS_KEY_ID", "AKIAXGKUAYT2P2IJ2KX7")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "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")
	os.Setenv("storage_accelerate", "true")
	os.Setenv("db_env", "mysql")
	os.Setenv("cdn_open", "true")
	os.Setenv("cdn_mode", "service")
	os.Setenv("cdn_path", "d2sl4gnftnfbyu.cloudfront.net")

	os.Setenv("cdn_service_path", "https://raven.dev.badanamu.net/cloudfront")
	os.Setenv("cdn_service_token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkb25naHVuLmNob2lAY2FsbWlkLmNvbSIsImF1ZCI6IkNhbG1Jc2xhbmQgQ2hpbmEiLCJzdWIiOiJyYXZlbiIsImlhdCI6MTUxNjIzOTAyMn0.oFThEoapYtp1BQQH8m-MQozOuFQeCZMNor3_jI3fNQo")
	os.Setenv("connection_string", "root:Badanamu123456@tcp(172.22.20.171:3306)/kidsloop2?parseTime=true&charset=utf8mb4")
	os.Setenv("max_open_conns", "8")
	os.Setenv("max_idle_conns", "4")
	os.Setenv("show_log", "true")
	os.Setenv("show_sql", "true")

	os.Setenv("open_cache", "true")
	os.Setenv("redis_host", "172.22.20.171")
	os.Setenv("redis_port", "6379")
}
func TestS3Storage_ExitsFile(t *testing.T) {
	InitEnv()
	config.LoadEnvConfig()
	storage := DefaultStorage()

	_, ret := storage.ExistFile(context.Background(), "thumbnail", "5f48815e9be1b049508ccb2c.jpg")
	t.Log(ret)
	_, ret = storage.ExistFile(context.Background(), "thumbnail", "5f48815e9be1b049508ccb2b.jpg")
	t.Log(ret)
}

func TestCDNSignature(t *testing.T) {
	InitEnv()
	config.LoadEnvConfig()
	storage := DefaultStorage()
	path, err := storage.GetFileTempPath(context.Background(), ThumbnailStoragePartition, "abby_cadabra_720p.mp4")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(path)
}

func TestS3Storage_GetUploadFileTempPath(t *testing.T) {
	InitEnv()
	config.LoadEnvConfig()
	storage := DefaultStorage()
	partition := ThumbnailStoragePartition
	path, err := storage.GetUploadFileTempPath(context.Background(), partition, "timg0.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(path)
}

func TestCDNSignature2(t *testing.T) {
	cdnConf := config.CDNConfig{
		CDNRestrictedViewer:           true,
		CDNPath:           "https://res-kl2-dev-cms.kidsloop.net",
		CDNPrivateKeyPath: "./rsa_private_key.pem",
		CDNKeyId: "K3PUGKGK3R1NHM",
	}
	ctx := context.Background()
	partition := "thumbnail"
	filePath := "5f4f5c380ba7e6e86a53b85f.jpg"
	path := fmt.Sprintf("%s/%s/%s", cdnConf.CDNPath, partition, filePath)
	keyID := cdnConf.CDNKeyId

	privateKeyPEM, err := ioutil.ReadFile(cdnConf.CDNPrivateKeyPath)
	if err != nil {
		log.Error(ctx, "read cdn path failed",
			log.String("cdn_key_path", cdnConf.CDNPrivateKeyPath),
			log.String("partition", string(partition)),
			log.String("file_path", filePath),
			log.Err(err),
		)
		return
	}
	block, _ := pem.Decode(privateKeyPEM)
	if block.Type != "RSA PRIVATE KEY" {
		log.Error(ctx, "parse key pem failed",
			log.String("cdn_key_path", cdnConf.CDNPrivateKeyPath),
			log.String("partition", string(partition)),
			log.String("file_path", filePath),
			log.String("pem", string(privateKeyPEM)),
		)
		return
	}
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Error(ctx, "parse public key failed",
			log.String("cdn_key_path", cdnConf.CDNPrivateKeyPath),
			log.String("partition", string(partition)),
			log.String("file_path", filePath),
			log.String("pem", string(privateKeyPEM)),
			log.Err(err),
		)
		return
	}

	signer := sign.NewURLSigner(keyID, privKey)
	signedURL, err := signer.Sign(path, time.Now().Add(constant.PresignDurationMinutes))
	if err != nil {
		log.Error(ctx, "Get presigned url failed",
			log.String("cdn_key_path", cdnConf.CDNPrivateKeyPath),
			log.String("partition", string(partition)),
			log.String("file_path", filePath),
			log.String("pem", string(privateKeyPEM)),
			log.Err(err),
		)
		return
	}
	t.Log(signedURL)
}
