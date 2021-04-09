package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"io/ioutil"
	"log"
	"net/http/httptest"

	"os"
	"testing"
	"time"
)

func readKey(path string) string {
	b, err := ReadAll(path)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	key := string(b)
	return key
}
func ReadAll(filePth string) ([]byte, error) {
	f, err := os.Open(filePth)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func Test_classUserEditEventToSchedule(t *testing.T) {
	privateKey := readKey("D:\\workbench\\auth_private_key.pem")
	publicKey := readKey("D:\\workbench\\auth_public_key.pem")
	privateKeyPB, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		fmt.Println("ParseRSAPrivateKeyFromPEM:", err.Error())
		return
	}
	publicKeyPB, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
	if err != nil {
		fmt.Println("ParseRSAPublicKeyFromPEM:", err.Error())
		return
	}
	cfg := &config.Config{
		DBConfig: config.DBConfig{
			//root:Passw0rd@tcp(127.0.0.1:3306)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local
			ConnectionString: "root:Badanamu123456@tcp(192.168.1.234:3306)/kidsloop2?parseTime=true&charset=utf8mb4",
		},
		Schedule: config.ScheduleConfig{
			MaxRepeatYear:    2,
			CacheExpiration:  3 * time.Minute,
			ClassEventSecret: publicKeyPB,
		},
	}
	config.Set(cfg)
	option := dbo.WithConnectionString(cfg.DBConfig.ConnectionString)
	newDBO, err := dbo.NewWithConfig(option)
	if err != nil {
		log.Println("connection mysql db error:", err)
		return
	}
	dbo.ReplaceGlobal(newDBO)

	gin.SetMode(gin.TestMode)
	server := &Server{
		engine: gin.New(),
	}
	server.engine.POST("/v1/class_user_edit_to_schedule", server.classUserEditEventToSchedule)

	//w := httptest.NewRecorder()

	stdClaims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
		IssuedAt:  time.Now().Unix(),
		Audience:  "kidsloop-live",

		Issuer:    "KidsLoopUser-live",
		NotBefore: 0,
		Subject:   "authorization",
	}
	type ScheduleClassEvent2 struct {
		Action  entity.ClassActionEvent          `json:"action" enums:"Add,Delete"`
		ClassID string                           `json:"class_id"`
		Users   []*entity.ScheduleClassUserEvent `json:"users"`
	}
	data := &ScheduleClassEvent2{
		Action:  entity.ClassActionEventDelete,
		ClassID: "0c01504d-d6ae-4c40-9862-68566bff0767",
		Users: []*entity.ScheduleClassUserEvent{
			{
				ID:       "4f614ccc-0867-5e5c-91f2-b71b895d2c48",
				RoleType: entity.ClassUserRoleTypeEventStudent,
			},
		},
	}
	type StandardClaims struct {
		*jwt.StandardClaims
		ScheduleClassEvent2
	}
	claims := &StandardClaims{
		StandardClaims:      stdClaims,
		ScheduleClassEvent2: *data,
	}

	token, err := utils.CreateJWT(context.Background(), claims, privateKeyPB)
	fmt.Println(err)
	fmt.Println(token)
	body := entity.ScheduleEventBody{Token: token}
	jsonStr, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/class_user_edit_to_schedule", bytes.NewBufferString(string(jsonStr)))
	req.Header.Add("content-type", "application/json")
	server.ServeHTTP(w, req)
	res := w.Result()

	str, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(str))
	//r, err := http.NewRequest("POST", "/v1/class_user_edit_to_schedule", bytes.NewBuffer(jsonStr))
	//
	//server.engine.ServeHTTP(w, r)
	//
	//resp := w.Result()
	//ioutil.ReadAll(resp.Body)
	//fmt.Println(resp)
	//fmt.Println(err)
}
