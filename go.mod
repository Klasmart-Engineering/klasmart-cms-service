module gitlab.badanamu.com.cn/calmisland/kidsloop2

go 1.14

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/aws/aws-sdk-go v1.35.23
	github.com/coreos/etcd v3.3.25+incompatible // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-contrib/pprof v1.3.0
	github.com/gin-gonic/gin v1.6.3
	github.com/go-playground/assert/v2 v2.0.1
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/hgfischer/go-otp v1.0.0
	github.com/jinzhu/gorm v1.9.16
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.14.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.6.1
	github.com/swaggo/swag v1.7.0
	github.com/tencentcloud/tencentcloud-sdk-go v1.0.62
	github.com/tencentyun/cos-go-sdk-v5 v0.7.7 // indirect
	github.com/tencentyun/scf-go-lib v0.0.0-20200624065115-ba679e2ec9c9 // indirect
	github.com/tidwall/gjson v1.6.0
	github.com/tidwall/pretty v1.0.1 // indirect
	github.com/ugorji/go v1.2.2 // indirect
	gitlab.badanamu.com.cn/calmisland/chlorine v0.1.6
	gitlab.badanamu.com.cn/calmisland/common-cn v0.16.0
	gitlab.badanamu.com.cn/calmisland/common-log v0.1.4
	gitlab.badanamu.com.cn/calmisland/dbo v0.1.9
	gitlab.badanamu.com.cn/calmisland/distributed_lock v0.1.13
	gitlab.badanamu.com.cn/calmisland/ro v0.0.0-20200819092854-7b96095e0678
	go.etcd.io/etcd v3.3.25+incompatible // indirect
	go.mongodb.org/mongo-driver v1.4.0
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
	golang.org/x/sys v0.0.0-20210314195730-07df6a141424 // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/genproto v0.0.0-20210312152112-fc591d9ea70f // indirect
	google.golang.org/grpc v1.36.0 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0

)
