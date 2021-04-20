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

func Test_classMembersEvent(t *testing.T) {
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
	server.engine.DELETE("/v1/classes_members", server.classDeleteMembersEvent)

	stdClaims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
		IssuedAt:  time.Now().Unix(),
	}
	type ScheduleClassEvent2 struct {
		ClassID string                `json:"class_id"`
		Members []*entity.ClassMember `json:"members"`
	}
	data := &ScheduleClassEvent2{
		ClassID: "dc33ce61-6ac7-4a12-8592-ca857e6eb395",
		Members: []*entity.ClassMember{
			{
				ID:       "11bb492e-ff80-5928-ac86-8b639a0c1a44",
				RoleType: entity.ClassUserRoleTypeEventTeacher,
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
	if err != nil {
		t.Fatal(err)
	}
	body := entity.ClassEventBody{Token: token}
	jsonStr, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/v1/classes_members", bytes.NewBufferString(string(jsonStr)))
	req.Header.Add("content-type", "application/json")
	server.ServeHTTP(w, req)
	res := w.Result()

	str, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(str))
}
