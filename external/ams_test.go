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
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNjU2MDE0MiwiaXNzIjoia2lkc2xvb3AifQ.xJHpajJWKqWsLtQ2wCqwbTFFlmdKNRNhREnDuoRvXOmoDCa3jET--dnAcNhYiwevFEMJ403OCzdQS_97LgtJFP6an8ffKlCS-Ug99F0Tm7mrugo7jSfepasAbdwh1Soa0xYrmBUa8yhY4CY1JphBYtfoMdJRPVsixeldtiI3rq2KpH-QFSil2k9YSd5hNY82UZH5c11QRbuBQNQ58gjjlF_fJ_JE3Kst_6vBwJGgwocWz-vqmusWMgs83NikMtaCWL3SXSjBQZQu-q5YUBnO-jX1-WaocXIQHD8_1IFw6ERp7OOmEQBXVPcFE_naIsiAeZzpmcLcFahCvR9LV6rYmw",
	}

	os.Exit(m.Run())
}
