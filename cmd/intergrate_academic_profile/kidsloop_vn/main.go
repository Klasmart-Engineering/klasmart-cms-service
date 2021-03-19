package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	iap "gitlab.badanamu.com.cn/calmisland/kidsloop2/cmd/intergrate_academic_profile"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var (
	mysqlConn    = flag.String("mysql", "root:Badanamu123456@tcp(192.168.1.234:3309)/kidsloop2?parseTime=true&charset=utf8mb4", "mysql conn")
	amsEndpoint  = flag.String("ams", "https://api.kidsloop.net/user/", "ams endpoint")
	specialOrgID = flag.String("org", "9d42af2a-d943-4bb7-84d8-9e2e28b0e290", "special organization id")
	token        = flag.String("token", "", "ams token")
)

const (
	BatchSize = 500
)

func main() {
	flag.Parse()

	if *token == "" {
		fmt.Printf("Usage: %s -token {ams token} -mysql \"conn\" -ams \"{ams endpoint}\n", os.Args[0])
		return
	}

	config.Set(&config.Config{
		DBConfig: config.DBConfig{
			ConnectionString: *mysqlConn,
		},
		AMS: config.AMSConfig{
			EndPoint: *amsEndpoint,
		},
	})

	ctx := context.Background()

	db, err := dbo.NewWithConfig(func(c *dbo.Config) {
		c.ShowLog = false
		c.ShowSQL = false
		c.MaxIdleConns = 4
		c.MaxOpenConns = 64
		c.ConnectionString = config.Get().DBConfig.ConnectionString
	})
	if err != nil {
		log.Fatal(ctx, "create dbo failed", log.Err(err))
	}
	dbo.ReplaceGlobal(db)

	operator := &entity.Operator{
		UserID: "14494c07-0d4f-5141-9db2-15799993f448", // PJ
		OrgID:  "10f38ce9-5152-4049-b4e7-6d2e2ba884e6", // Badanamu HQ
		Token:  *token,
	}

	mapper := iap.NewMapper(operator)

	var commands []string

	_commands, err := contentMigrate(ctx, operator, mapper, *specialOrgID)
	if err != nil {
		log.Fatal(ctx, "content migrate failed", log.Err(err))
	}
	commands = append(commands, _commands...)

	_commands, err = scheduleMigrate(ctx, operator, mapper, *specialOrgID)
	if err != nil {
		log.Fatal(ctx, "schedule migrate failed", log.Err(err))
	}
	commands = append(commands, _commands...)

	filename := fmt.Sprintf("kl2_migrate_%s.sql", time.Now().Format("200601021504"))
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(ctx, "open file for write failed", log.Err(err))
	}
	defer file.Close()

	for _, command := range commands {
		fmt.Fprintln(file, command)
	}

	log.Info(ctx, "Finished!")
}
