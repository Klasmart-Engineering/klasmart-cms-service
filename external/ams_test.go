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
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNjU3NjI2MCwiaXNzIjoia2lkc2xvb3AifQ.Cfj1FioyjriW3Rhc-lShD5k6Ucn1mNXiFhfqDBz4jc726_R574zM0ztDruyk0htki7AJ-itITJADFasbR7NkXCzes6MzD7yHwWNNnX5o4Qr5xQTsdqThifwrKTUeyIwPs-V4JNJ0RsuEjuby9SQkfw1-VxsD4JQ8Pq6razI65jLhwPp4snjXU5TFCeilzULM80HCFMLYpN_S-LmejVEHdOGQJNhkObHsLC18vzWfaJIPckLKaUxn4EnDb_KUXiKbI-t9S6B8TtcID8Amskpc_8mjdgV1qm_WSNdPYnXjsGDjliyLdse9OATkTecWhQZWCz5ny7Ww2bReZl0HZNe8PA",
	}

	os.Exit(m.Run())
}
