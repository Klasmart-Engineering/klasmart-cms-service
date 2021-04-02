package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"log"
	"os"
)

func main() {
	//a, err := parseArgs()
	a, err := loadLocalDevArgs()
	if err != nil {
		panic(err)
	}
	confirmArgs(a)
	if err := initConfig(a); err != nil {
		panic(err)
	}

	fmt.Println("=> Congratulation! All done!")
}

type args struct {
	DSN   string `json:"dsn"`
	AMS   string `json:"ams"`
	Token string `json:"token"`
}

func parseArgs() (*args, error) {
	// parse flag and set args
	a := args{}
	flag.StringVar(&a.DSN, "dsn", "", `db connection string, required`)
	flag.StringVar(&a.AMS, "ams", "", "ams endpoint")
	flag.StringVar(&a.Token, "token", "", "token")
	flag.Parse()

	// check args
	if a.DSN == "" {
		return nil, errors.New("require dsn argument")
	}

	fmt.Println("=> Parse args done!")

	return &a, nil
}

func loadLocalDevArgs() (*args, error) {
	a := args{
		DSN:   "root:admin@tcp(127.0.0.1:3306)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local",
		AMS:   "",
		Token: "",
	}

	fmt.Println("=> Load local dev args done!")

	return &a, nil
}

func confirmArgs(a *args) {
	bs, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	fmt.Println("Please check args:", string(bs))
	fmt.Print("Enter to continue ...")
	if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
		panic(err)
	}
}

func initConfig(a *args) error {
	// set config
	c := &config.Config{
		DBConfig: config.DBConfig{ConnectionString: a.DSN},
		AMS: config.AMSConfig{
			//https://api.kidsloop.net/user/
			EndPoint: a.AMS,
		},
	}
	config.Set(c)

	// replace dbo
	newDBO, err := dbo.NewWithConfig(dbo.WithConnectionString(c.DBConfig.ConnectionString))
	if err != nil {
		log.Println("connection mysql error:", err)
		return err
	}
	dbo.ReplaceGlobal(newDBO)

	fmt.Println("=> Init config done!")

	return nil
}
