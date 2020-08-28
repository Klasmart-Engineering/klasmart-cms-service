module gitlab.badanamu.com.cn/calmisland/kidsloop2

go 1.14

require (
	github.com/aws/aws-sdk-go v1.33.17
	github.com/gin-gonic/gin v1.6.3
	github.com/go-playground/validator/v10 v10.3.0 // indirect
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/jinzhu/gorm v1.9.16
	github.com/onsi/ginkgo v1.14.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/tencentyun/cos-go-sdk-v5 v0.7.7 // indirect
	github.com/tencentyun/scf-go-lib v0.0.0-20200624065115-ba679e2ec9c9 // indirect
	github.com/tidwall/pretty v1.0.1 // indirect
	gitlab.badanamu.com.cn/calmisland/common-cn v0.15.0
	gitlab.badanamu.com.cn/calmisland/common-log v0.1.3
	gitlab.badanamu.com.cn/calmisland/dbo v0.1.7
	gitlab.badanamu.com.cn/calmisland/distributed_lock v0.1.13
	gitlab.badanamu.com.cn/calmisland/ro v0.0.0-20200819092854-7b96095e0678
	go.mongodb.org/mongo-driver v1.4.0
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/net v0.0.0-20200813134508-3edf25e44fcc // indirect
	golang.org/x/sys v0.0.0-20200812155832-6a926be9bd1d // indirect
	google.golang.org/genproto v0.0.0-20200814021100-8c09557e8a18 // indirect
	google.golang.org/grpc v1.31.0 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0

)
