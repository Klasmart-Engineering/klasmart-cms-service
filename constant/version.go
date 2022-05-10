package constant

var (
	// set by go build -o deploy/handler -ldflags "-X github.com/KL-Engineering/kidsloop-cms-service/constant.GitHash=$(git rev-list -1 HEAD) -X github.com/KL-Engineering/kidsloop-cms-service/constant.BuildTimestamp=$(date +%s) -X github.com/KL-Engineering/kidsloop-cms-service/constant.LatestMigrate=$(ls schema/migrate | tail -1)"
	GitHash        = "undefined"
	BuildTimestamp = "undefined"
	LatestMigrate  = "undefined"
)
