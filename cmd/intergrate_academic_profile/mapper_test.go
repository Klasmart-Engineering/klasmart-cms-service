package intergrate_academic_profile

import (
	"log"
	"os"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var (
	testMapper   Mapper
	testOperator = &entity.Operator{
		UserID: "14494c07-0d4f-5141-9db2-15799993f448", // PJ
		OrgID:  "10f38ce9-5152-4049-b4e7-6d2e2ba884e6", // Badanamu HQ
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTYzMDU4MywiaXNzIjoia2lkc2xvb3AifQ.sgwnCUP_TvEV2htOoGLAW5_5po02llspImeq2Lg2I_ubnJcbboWJQ9Ky3pxbULd7-cvtF_m-Q1-D6-747Canboi9aqLdEuKGdn7Y6zhIHf0ZGlTAxwkm93Pq26CsVq7XiRJKEwqYktuuBT63QQ60Ecsa7yqx6tafUSvNsUBbCI349LnCG0DigRmrbxB0aESPH6_f1vlod2JAy0KCBAWYBixQPNbN4vF1bQ8X9HUwhPKS211c5kt1dRsK7_jtH4jtIiGopL9dYsDkqO8_93w0c5mBD8Rt89UfIaKQQrlRKKN4aZ_ol0g1y3l1jA6rwbpphBb1SJUlT1iZHb5dRbC_Iw",
	}
)

func TestMain(m *testing.M) {
	config.Set(&config.Config{
		AMS: config.AMSConfig{
			EndPoint: "https://api.beta.kidsloop.net/user/",
		},
	})

	dsn := "admin:LH1MCuL3V0Ib3254@tcp(migration-test2.c2gspglsifnp.rds.cn-north-1.amazonaws.com.cn:28344)/kidsloop2?parseTime=true&charset=utf8mb4"
	option := dbo.WithConnectionString(dsn)
	newDBO, err := dbo.NewWithConfig(option)
	if err != nil {
		log.Println(err)
		return
	}
	dbo.ReplaceGlobal(newDBO)

	testMapper = NewMapper(testOperator)
	os.Exit(m.Run())
}
