package entity

import (
	"io/ioutil"
	"os"

	"github.com/dgrijalva/jwt-go"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
)

func Setup() {
	privateKeyPath := os.Getenv("kidsloop_cn_login_private_key_path")
	content, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		panic(err)
	}
	prv, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(content))
	if err != nil {
		panic(err)
	}

	publicKeyPath := os.Getenv("kidsloop_cn_login_public_key_path")
	content, err = ioutil.ReadFile(publicKeyPath)
	if err != nil {
		panic(err)
	}
	pub, err := jwt.ParseRSAPublicKeyFromPEM(content)
	if err != nil {
		panic(err)
	}

	config.Set(&config.Config{
		KidsloopCNLoginConfig: config.KidsloopCNLoginConfig{
			PrivateKey: prv,
			PublicKey:  pub,
		},
	})
}

// func TestUser_Token(t *testing.T) {
// 	Setup()
// 	user := User{
// 		UserID: "1",
// 	}
// 	token, err := user.Token()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Println(token)
// }

// //var token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJLaWRzbG9vcCIsImV4cCI6MTYwOTM4NDk2MCwiaWF0IjoxNjA2NzkyOTMwLCJpc3MiOiJLaWRzbG9vcENuIiwic3ViIjoiYXV0aG9yaXphdGlvbiJ9.lfiTxqqVuCxvLaKf_yp2igLnWy261NKfqY9mMffXxFDI7vmY5BldsyM9Ui7A_A1OQqtcMPZyMQL9MkXEtJsEDYlc30T8zaS-W3US_WZSBcAUBvokceamXA85xpbwtqWrK7cim3em2JhOcqqxGYma9LeMZQstjJ8XYG1UtclEociYiDqCa8p4R5SCEbl_kOvr_siKPMYBqmWG-M9gLjJsLIBZbkYpekMgMXHA7W2xvqSZQOq2rVdlTTkR57oEer49CstFwLC09qphEnqSQPJBowQI1Bg8jRhc44MS1rpnU_Xv8q-SMObWll1ZArz9Ws33lmB83TjEUsnuzeR6MoB_tQ"
// var token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJLaWRzbG9vcCIsImV4cCI6MTYwOTM4NTMyMSwianRpIjoiMSIsImlhdCI6MTYwNjc5MzI5MSwiaXNzIjoiS2lkc2xvb3BDbiIsInN1YiI6ImF1dGhvcml6YXRpb24ifQ.NDdnhm8KY8-jxpGycxIGjSzYy075yr94MOENowPlEpNEWd2TVhJpA10gsrgBTdKLhG4awsCErvZmXeUTNAF1Hyd-pfwdwwIYSj1VUxVz0IMaWbNNCgwDDhAhA1TpF95NT8UHh8GGeCMVP5puxuUoBv0t4XQYMcPEaCRnEWY3CtEAIz_y5Fz8blxvBTkYFy7EsrAIUiosMBRwDx8LBwGdGpuFXYLHkYVReag129O5VEWaBq3kVaSDww0Tgto8j_VRbS8KUFig4WwlAYGijcO6tNGYHYos7lWVwh2cdV9udeQdqA0ZnROVIkCcY76XZXuv5PoAJGZlofUYjYAJIDR4_g"

// func TestNewUserFromToken(t *testing.T) {
// 	Setup()
// 	user, err := NewUserFromToken(token)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Println(user)
// }
