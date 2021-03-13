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
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTYxODQxMCwiaXNzIjoia2lkc2xvb3AifQ.E9eoZt95ZYxjgR_8PaQ2c5P-tCyQ0QevZkM-h4degpmBFgMD9XttmVDi5Tnq8DqOh75YxZgZSCIUPsEoOwQ-R1Uzc9SJM0hvFxmgftwQYLj98hnBXLqtYrQCC7CVN6sqP9vA-Va6yURm5qzn4STGdg5TbH0Bk_tczGH4FCmyJBAzoeodb_A-ho_yZPlJvPi3Raxzs7A85b9-PcfTdQoB2SQwnUt9ljRdE9TYB8wxC5OvsUeeFEOwWQO-lYrEoEYIcWCea6rRiGU6O6vkQWkA4Lo9U63jXAQ5tZFaiaCPJAuZoybEUo-xxYmPV9hOyfzkzOkZjDpn-oRI5Hm-b6QHJw",
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
