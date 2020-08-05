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
	os.Setenv("cdn_open", "false")
}
func TestS3Storage_ExitsFile(t *testing.T) {
	InitEnv()
	config.LoadEnvConfig()
	storage := DefaultStorage()

	ret := storage.ExitsFile(context.Background(), "asset", "5f225eeee763b300cf63cb901.jpg")
	t.Log(ret)
	ret = storage.ExitsFile(context.Background(), "asset", "5f225eeee763b300cf63cb90.jpg")
	t.Log(ret)
}
