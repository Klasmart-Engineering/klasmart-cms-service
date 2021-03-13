package intergrate_academic_profile

import (
	"os"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var (
	testMapper   Mapper
	testOperator = &entity.Operator{
		UserID: "14494c07-0d4f-5141-9db2-15799993f448", // PJ
		OrgID:  "10f38ce9-5152-4049-b4e7-6d2e2ba884e6", // Badanamu HQ
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTYwNzU4MCwiaXNzIjoia2lkc2xvb3AifQ.ouFXlTwbafZhmL9edumKyQ1BRAEmcDe9OP5PA9k4_0gRXx5zf7GAUpv1JZ8kMJQobMXXcUYseST9VbrsLr6-_pGCbjCI2p0aMgkaHS3Qt1w1v1AS116Rfl-qe07liiwdrWJ89aMGQ08MVxTmBvNQ7uH4LdGVYSc0VxBM5sh1r7hLzGFEz34vKJ39LDbzvqMmxAhAUt8NpGeC1oCLSs3AVkOw8Qnfw4gVPYKT-QPajEuF3yKyTtulpgP072_XzD7fw_QhA9ZXRGE0vRTTXha6BGQRfBLCMVflbJ9SpR84jmqnsSP_uf782hruvbWSt0xcd4Lo8cvRNomaQxXn3gUKgA",
	}
)

func TestMain(m *testing.M) {
	config.Set(&config.Config{
		AMS: config.AMSConfig{
			EndPoint: "https://api.beta.kidsloop.net/user/",
		},
	})

	testMapper = NewMapper(testOperator)

	os.Exit(m.Run())
}
