package storage

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"os"
	"testing"
)

func InitEnv(){
	os.Setenv("cloud_env", "aws")
	os.Setenv("storage_bucket", "kidsloop-global-resources-dev")
	os.Setenv("storage_region", "ap-northeast-2")
	os.Setenv("secret_id", "AKIAXGKUAYT2P2IJ2KX7")
	os.Setenv("secret_key", "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")
	os.Setenv("storage_accelerate", "true")
	os.Setenv("cdn_open", "true")
	os.Setenv("cdn_mode", "service")
	os.Setenv("cdn_path", "d2sl4gnftnfbyu.cloudfront.net")

	os.Setenv("cdn_service_path", "https://raven.dev.badanamu.net/cloudfront")
	os.Setenv("cdn_service_token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkb25naHVuLmNob2lAY2FsbWlkLmNvbSIsImF1ZCI6IkNhbG1Jc2xhbmQgQ2hpbmEiLCJzdWIiOiJyYXZlbiIsImlhdCI6MTUxNjIzOTAyMn0.oFThEoapYtp1BQQH8m-MQozOuFQeCZMNor3_jI3fNQo")

}
func TestS3Storage_ExitsFile(t *testing.T) {
	InitEnv()
	config.LoadEnvConfig()
	storage := DefaultStorage()

	_, ret := storage.ExitsFile(context.Background(), "asset", "5f225eeee763b300cf63cb901.jpg")
	t.Log(ret)
	_, ret = storage.ExitsFile(context.Background(), "asset", "5f225eeee763b300cf63cb90.jpg")
	t.Log(ret)
}

func TestCDNSignature(t *testing.T){
	InitEnv()
	config.LoadEnvConfig()
	storage := DefaultStorage()
	path, err := storage.GetFileTempPath(context.Background(), "ESL/Songs", "abby_cadabra_720p.mp4")
	if err != nil{
		t.Error(err)
		return
	}
	t.Log(path)
}