package profile_schedule

import (
	"log"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/cmd/intergrate_academic_profile"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var (
	testMapper   intergrate_academic_profile.Mapper
	testOperator = &entity.Operator{
		UserID: "14494c07-0d4f-5141-9db2-15799993f448", // PJ
		OrgID:  "10f38ce9-5152-4049-b4e7-6d2e2ba884e6", // Badanamu HQ
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTYyOTY4MSwiaXNzIjoia2lkc2xvb3AifQ.O-Tg88NifXicJGp5ttKqLfFAPM8HG9Ld6Jgjb5lCs_QMCEAfPm6LtZyGkOaSCOYB8FeoS8qr61xytY8d0oTSBVc-0H21x4WxzJaZ0D2Eb_vi5VDhob3P-6lydIJ2jNyf0vyhfvDttBxJKK3Keuf3q44Wgynez-1Dtsl-nWAQ741zAKfyZy-OctdOE7bEZLuyP-syeW-l_nmhx1vzBvZ_3MHDfkw_b_3IGN1nZ6kLk_JEdh0MWAmKH6_pCCl_51DLQeeVW7eXWBs1SZuy2It-d47JyLyRMDB4DRADlNcb5vme4mDMi4pcpdtNAxuDxmsv757ajXjmDIivsfg5Hhx5jw",
	}
)

func Test_loadSchedule(t *testing.T) {
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

	testMapper = intergrate_academic_profile.NewMapper(testOperator)
	//loadSchedule()
}
