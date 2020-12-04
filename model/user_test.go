package model

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

func SetupUser() {
	config.Set(&config.Config{
		DBConfig: config.DBConfig{
			ConnectionString: os.Getenv("connection_string"),
		},
	})
	dboHandler, err := dbo.NewWithConfig(func(c *dbo.Config) {
		dbConf := config.Get().DBConfig
		c.ShowLog = true
		c.ShowSQL = dbConf.ShowSQL
		c.MaxIdleConns = dbConf.MaxIdleConns
		c.MaxOpenConns = dbConf.MaxOpenConns
		c.ConnectionString = dbConf.ConnectionString
	})
	if err != nil {
		log.Error(context.TODO(), "create dbo failed", log.Err(err))
		panic(err)
	}
	dbo.ReplaceGlobal(dboHandler)

}
func TestUserModel_RegisterUser(t *testing.T) {
	SetupUser()
	user, err := GetUserModel().RegisterUser(context.Background(), "15221776376", "12345", constant.AccountPhone)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
}

func TestUserModel_GetUserByAccount(t *testing.T) {
	SetupUser()
	user, err := GetUserModel().GetUserByAccount(context.Background(), "15221776376")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
}

var userID = "5fc9a40c8d7ccd72bf7c84cd"

func TestUserModel_ResetUserPassword(t *testing.T) {
	SetupUser()
	err := GetUserModel().ResetUserPassword(context.Background(), userID, "Bada1234", "12345")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserModel_UpdateAccountPassword(t *testing.T) {
	SetupUser()
	u, err := GetUserModel().UpdateAccountPassword(context.Background(), "15221776376", "123456")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(u)
}
