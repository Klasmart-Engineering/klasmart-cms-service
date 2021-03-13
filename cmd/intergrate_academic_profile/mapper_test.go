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
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTYyNzg3MiwiaXNzIjoia2lkc2xvb3AifQ.EnJ8b_PRo10_GVbOyV3GlF0YNGF2RFDRcS9Lb9aijpnfoHGcye_Dya8gAeZgiSyYczyUJs37mB4buq8Oe5OlcXyBoZj7lEJ1bvjZqJ4GxkZA0DPHUlXSFiZI8aD64J0DiR_fOjiWkeqYCwpaJ9yod9GTF0UAp9SE90GTosaUWqkgSPlHMXjlmgR8opLoOpP9ZgrtuXo1CcZA9MO_78RnkJKO4D3iUB7ewCK3A_kCUIPhbOMHRFAzFtvuSIzkGfkACcrlOQGRRj0A4YCXVWOu7Psj4i-xox0c8TosOWhHukvE98Jchy6_smiPPJUkKWBLbPmJ3z9opBK8uRumwmgJBA",
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
