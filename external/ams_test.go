package external

import (
	"os"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var (
	testOperator *entity.Operator
)

func TestMain(m *testing.M) {
	config.Set(&config.Config{
		AMS: config.AMSConfig{
			EndPoint: "https://api.beta.kidsloop.net/user/",
		},
	})

	testOperator = &entity.Operator{
		UserID: "14494c07-0d4f-5141-9db2-15799993f448",
		OrgID:  "10f38ce9-5152-4049-b4e7-6d2e2ba884e6", // Badanamu HQ
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTU2NzUzMiwiaXNzIjoia2lkc2xvb3AifQ.BjuiamDCwQAJzO2BIXlWqrI4gPItCIxWzG8i-XLanF0LfoWN2juDs_N4ip78ctxb30Su9XWp6Bmw8YBnMwhRQ5sHc5wnfOXiVcbqe2EagzbO0TgzVXCKneMNuiVkbYYVtK7d8zsQ0tbeiYduYo1aIO2AVWD8_6lu1J1P2YM_5p81BYMZCaLR38N60XyLQWFDJpS2kNuvxmkK73CmS-QZeMFFYTEXCIgbYebLokaF0lpfXsBrO-obAY3OvXpaRrVWWegVEOZxMz3NTkrk-QuCfehaqgxBGe_Uci78sTGo1f3uO_0RM2Jyvj-S1fwC0mnIt_JP8J-jqUvLBIlaCae_9g",
	}

	os.Exit(m.Run())
}
