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
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTY5MTUxOCwiaXNzIjoia2lkc2xvb3AifQ.ZocoKugAdJy9nePic1FshT6vz7SFkglf7dZNhYP9Gz8Su1vELmqADSNxHX5zDxWQ_nnigfINBZlEKOJkXlDI5cSa9WPtZiPRPOu-shyS97BTNdO3M37vNcXcXB5NM8dFb44qBaYU_MbjoM1wxIneoTafBkbIApinTMVqFJ8U7MZVsVvT8cCoZgSEtEba5O3u5gSElPP8q22Tyq5flqOyZjvB9cijURra9e5CTXNM7SxJ8re-ePfbuIcypaWhDBKc9wPPiD8BC2bllvQIZJxSMsYlLareI9UmL642hNJfa1rn0sZaj7XHAPCvDECQndFQ1pjwLQME6pnxp3dMFwJiEg",
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
