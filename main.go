package kidsloop2

import (
	"calmisland/kidsloop2/api"
	"calmisland/kidsloop2/dynamodb"
	"calmisland/kidsloop2/storage"

	"gitlab.badanamu.com.cn/calmisland/common-cn/common"
)

func main() {
	// init dynamodb connection
	dynamodb.GetClient()
	storage.DefaultStorage()

	common.Setenv(common.EnvLAMBDA)
	common.RunWithHTTPHandler(api.NewServer(), "")
}
