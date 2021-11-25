module gitlab.badanamu.com.cn/calmisland/kidsloop2

go 1.14

require (
	github.com/aws/aws-sdk-go v1.35.23
	github.com/coreos/etcd v3.3.25+incompatible // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-contrib/pprof v1.3.0
	github.com/gin-gonic/gin v1.6.3
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-playground/assert/v2 v2.0.1
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hgfischer/go-otp v1.0.0
	github.com/jinzhu/gorm v1.9.16
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/newrelic/go-agent v3.15.0+incompatible
	github.com/onsi/ginkgo v1.14.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/swaggo/swag v1.7.4
	github.com/tencentcloud/tencentcloud-sdk-go v1.0.62
	github.com/tencentyun/cos-go-sdk-v5 v0.7.7 // indirect
	github.com/tencentyun/scf-go-lib v0.0.0-20200624065115-ba679e2ec9c9 // indirect
	github.com/tidwall/gjson v1.6.0
	github.com/tidwall/pretty v1.0.1 // indirect
	github.com/ugorji/go v1.2.2 // indirect
	github.com/urfave/cli/v2 v2.3.0
	gitlab.badanamu.com.cn/calmisland/chlorine v0.2.0
	gitlab.badanamu.com.cn/calmisland/common-cn v0.17.0
	gitlab.badanamu.com.cn/calmisland/common-log v0.1.6
	gitlab.badanamu.com.cn/calmisland/dbo v0.3.0
	gitlab.badanamu.com.cn/calmisland/distributed_lock v0.1.13
	gitlab.badanamu.com.cn/calmisland/imq v0.2.18
	gitlab.badanamu.com.cn/calmisland/kidsloop-cache v0.0.14
	gitlab.badanamu.com.cn/calmisland/ro v0.0.0-20210813055601-f0a5d22461a0
	go.etcd.io/etcd v3.3.25+incompatible // indirect
	go.mongodb.org/mongo-driver v1.7.2
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/net v0.0.0-20210924054057-cf34111cab4d
	golang.org/x/sys v0.0.0-20210923061019-b8560ed6a9b7 // indirect
	google.golang.org/genproto v0.0.0-20210924002016-3dee208752a0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/gorm v1.21.15 // indirect
)

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0

)
