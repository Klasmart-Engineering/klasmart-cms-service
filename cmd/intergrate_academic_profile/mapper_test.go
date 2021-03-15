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
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTc3NDM2MSwiaXNzIjoia2lkc2xvb3AifQ.s6-mFHj0IG3vqHI_QM67bKh8iqdHwctU46LPywkk9tyjYmNu4emTJ_lqNkhwHIUWsc9iyN-2pjYk66yH6xZwg8BCDVtaXHfOxeD-cpDOGur_D98Cdy7TDwnO1eGz3KeaIGzVixAswjNezZRw5X7w6pAkDeCA85hqzIX5HWcYbVkFFoOIupRA4Wx-Ro4XIZLydyyGbVdCLe-5VjwE8vfDBsnL1MRrx8Nj0rpF6DMvKEjeY9njeoIuTLzSem7HjiJXaqMIJJboj-3ngA87me_aYqc0nSf3i5dCPWMVtLhVPAfoVXUHegWfYU5GI0pDxx5FtpOUd21kIYhspdIdy2EVnA; _gid=GA1.2.426692570.1615773475",
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
