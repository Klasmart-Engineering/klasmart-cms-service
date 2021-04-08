package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func readKey(path string) string {
	var keyPath = "D:\\workbench\\auth_private_key.pem"
	b, err := ReadAll(keyPath)
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
	privateKey := readKey("D:\\workbench\\demo_private_key.pem")
	publicKey := readKey("D:\\workbench\\demo_public_key.pem")

	cfg := &config.Config{
		DBConfig: config.DBConfig{
			//root:Passw0rd@tcp(127.0.0.1:3306)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local
			ConnectionString: "root:Badanamu123456@tcp(192.168.1.234:3306)/kidsloop2?parseTime=true&charset=utf8mb4",
		},
		Schedule: config.ScheduleConfig{
			MaxRepeatYear:    2,
			CacheExpiration:  3 * time.Minute,
			ClassEventSecret: publicKey,
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

	w := httptest.NewRecorder()
	data := &entity.ScheduleClassEvent{
		Action:  entity.ClassActionEventAdd,
		ClassID: "0c01504d-d6ae-4c40-9862-68566bff0767",
		Users: []*entity.ScheduleClassUserEvent{
			{
				ID:       "4f614ccc-0867-5e5c-91f2-b71b895d2c48",
				RoleType: entity.ClassUserRoleTypeEventStudent,
			},
		},
	}

	body := entity.ScheduleEventBody{Token: token}
	r, _ := http.NewRequest("POST", "/v1/class_user_edit_to_schedule", nil)

	server.engine.ServeHTTP(w, r)

	resp := w.Result()
	//body, _ := ioutil.ReadAll(resp.Body)
	t.Log(resp.StatusCode)
}
