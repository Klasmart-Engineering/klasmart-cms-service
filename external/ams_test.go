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
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTYzMDE0NiwiaXNzIjoia2lkc2xvb3AifQ.mJ44ARXKv743_iRe6AUg0jywFkTt5lTdWan-2g_F6J9tsNPLIspSnv9b6AK3uGoIheTC_uLYqBv0uLRmtIfugOARTXN1Ze-SYzbe9KSA_eJ2lrn__JvL2tQwTYwKwrzH8abk_waSRPpf2ZNrfdOrze7dRpIeyS151L3FOxxtMl87Uavu_z508097Ic3MuOZIgNjnKfOI4CAwtHFSEJTODQJgEkf8ku5CqObFsxAFTidaUbCX5DnPGw5RpNTeCJEI9lWUHamZVtUJGndo3Cv8vL3H8_swOMPm7MO3rkFF04FlrNwBshavRxBnkgFEa5zUyB9nhQvgpfalf_qUWK8Mjw",
	}

	os.Exit(m.Run())
}
