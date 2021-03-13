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
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTY0NzQ0NywiaXNzIjoia2lkc2xvb3AifQ.Cmc1xoTfCDAJyHq842HkNJ59119suzawUP7BtUtEcZJjXgfQv4IMqJ22ilQx3hDoOAjtSq0ipSDtpYLR_WCsHc0L_gSKKMm3T1KD6NpoCHfgGgT1zSAm6bvbEjroH7_Wtkyqf2TowyCdNDMV_Fm_okYJn53p2Ha9arpKZ1mA2GdZiE44N0jg9uBxEAZVe79cTpf9y41GHCcmAXy3IhVADaq5v_gWN3EnPoYeaJAOBwYGd93wMxCFI5Njn9znVCX5Cn-n6Yj_3L9eggPxwv29YeLHwEbZ2lJCKUADdiVQfbDtYnYJJddkzFZDd8EQN09O35AAKTG3H8Jmig7EAPLgZA",
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
