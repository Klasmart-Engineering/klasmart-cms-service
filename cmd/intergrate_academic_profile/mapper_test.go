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
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTY0MjAzMSwiaXNzIjoia2lkc2xvb3AifQ.QftgSOm3v2L8KMTYum30pf4b-QwTSaxbRIDcIDfGYMBHRWVPpxqX16vNjs6BgVICoEtKM_nmBFuCU-8kFuN5TeU-nV-wO3bbZ3ebRvzxUXX1HeyOqedwc-nRYNHYyAwIp_kWVrUweIfjI6pRur2P1Wuyzk4DBsGHSxrF1z0qtoPdsSdN2KiE2IYiKxq7MzcUmd-eqtKBhGA8lrQd21aU_mG1koCBIpCbHMwD608Khd2RHsAE-Gm31eBUd4Umi9LdhyRqgtPeKUP2Rf9uLE7JUUJ5eMklEDdAQTTIcaDwMx-UoXXxQ5tJVqWNifZabCDHfeweV6NgZoBzgp_OKUWxUQ",
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
