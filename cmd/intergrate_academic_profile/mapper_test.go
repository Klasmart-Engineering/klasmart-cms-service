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
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTY1MTk2MCwiaXNzIjoia2lkc2xvb3AifQ.UWeymILO5Hw7rAsaqMF1KSZGcDs1vSzwg6XMTWuC9EjgjSC157pkf0bYnnBje-cp7yXykUYXCXeSzSqiPBqR-ZuYpuuFuBkTD9x6U6DQL04o0astz3PlL16vAvCg000j7mlpdIZeyBMrPfQlN3u7pUSfI4pZ1LZzqOkeADuw2t-bauAkZvMfe8pHfBeqYk_Gae-J0INURNZIb7mE0mno4F458Bu3sthO6kxEsz7bGO4x71eSQHhK3e5H1PrUYF3kqfGG_NAAdh0OjNC5LnemZKWgxpYOO2XJmCbcs1eweAysbNfbq5qnsqyIvXCCooFrhgJAzK9nUPL97JM_Z2bXkQ",
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
