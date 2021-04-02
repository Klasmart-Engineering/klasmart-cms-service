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
		OrgID:  "2e922238-decb-438e-b960-a0e404e015a5", // Badanamu HQ
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImE5MmI5MzVmLTMyYjItNTA2NC05YTNjLTgzY2RlOTYxZTdmOSIsImVtYWlsIjoia2lkc2xvb3BAY2hyeXNhbGlzLndvcmxkIiwiZXhwIjoxNjE3MjQ3OTg2LCJpc3MiOiJraWRzbG9vcCJ9.R87rt217XUkQu_SoX4ZajoNbRnKtcHR8_PI77FYdMiBx2pKtIS7S_GNf--QxuMGoBbYrtFxJp_y2eNK3D9uj6AtXlyDIQI5zkk5HXl8raH5f9aLUF5eXVLmkwfwag10LqkrxIroFdbl0LmzNq8L7kMtSd40b26hGrxKSYqNAej3hXk5Zjlq7kIYa_VsiHOyCj5R1tSvsOmUEQ79u9lp_gkwoof1t84t1R9bvZmUa23IVVzG_G1DZAicnn3wmgDy-vTSaplnz4Z4-06KQb7BrHuPqa8mGVENaFIp54NPly1v-dknR6mf_ERZgPijx3134ToyDuyWJgpaio_4tdzGdCQ",
	}

	os.Exit(m.Run())
}
