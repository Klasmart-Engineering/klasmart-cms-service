module gitlab.badanamu.com.cn/calmisland/kidsloop2

go 1.14

require (
	github.com/aws/aws-sdk-go v1.33.17
	github.com/gin-gonic/gin v1.6.3
	github.com/go-playground/validator/v10 v10.3.0 // indirect
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/jinzhu/gorm v1.9.15 // indirect
	github.com/onsi/ginkgo v1.14.0 // indirect

	github.com/tencentyun/cos-go-sdk-v5 v0.7.7 // indirect
	github.com/tencentyun/scf-go-lib v0.0.0-20200624065115-ba679e2ec9c9 // indirect
	github.com/tidwall/pretty v1.0.1 // indirect
	gitlab.badanamu.com.cn/calmisland/common-cn v0.15.0
	gitlab.badanamu.com.cn/calmisland/common-log v0.1.3
	gitlab.badanamu.com.cn/calmisland/dbo v0.1.6
	gitlab.badanamu.com.cn/calmisland/distributed_lock v0.1.13
	gitlab.badanamu.com.cn/calmisland/ro v0.0.0-20200819092854-7b96095e0678
	go.mongodb.org/mongo-driver v1.4.0
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/sys v0.0.0-20200810151505-1b9f1253b3ed // indirect
	google.golang.org/genproto v0.0.0-20200808173500-a06252235341 // indirect
	google.golang.org/grpc v1.31.0 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0

)
