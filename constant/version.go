package constant

var (
	// set by go build -o deploy/handler -ldflags "-X gitlab.badanamu.com.cn/calmisland/kidsloop2/constant.GitHash=$(git rev-list -1 HEAD) -X gitlab.badanamu.com.cn/calmisland/kidsloop2/constant.BuildTimestamp=$(date +%s) -X gitlab.badanamu.com.cn/calmisland/kidsloop2/constant.LatestMigrate=$(ls schema/migrate | tail -1)"
	GitHash        = "undefined"
	BuildTimestamp = "undefined"
	LatestMigrate  = "undefined"
)
