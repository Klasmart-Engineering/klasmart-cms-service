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
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTYyNzQzNywiaXNzIjoia2lkc2xvb3AifQ.VksHh6vcKfM0uDPI8B1--UQ4jvaTGqWFAthElFLLfzU-z8dC_wWTanncOZnERUXwy-Li9xaJDP0aCMTJLutmOFues074BKl4pMVbhU5Dx3fRlrWTaV-E4X7y5PTlx38kgYQRfZYE9rznXfPtrbiA2zFjr1SMlqbPHAQDeRXwYuvqZgjXHnJzk5QttEc8RAT7nDwi61Y5aczOp9zqFZMzTDwA5ndGWQO1nTfpMBfv2EmgCItAfOz63ovPMME2I5fZ2-FKZk9Bi4QXrsF2uLxxcl1-k8n9Goe0hI8Ajble1ANBFegw0Vjl89IC11cbwmAwHcidcICLmS3Ipb3Hq5hwqg",
	}

	os.Exit(m.Run())
}
