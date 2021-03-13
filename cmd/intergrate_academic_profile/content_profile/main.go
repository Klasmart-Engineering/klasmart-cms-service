package main

import (
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
)

const connDBStr = "root:Badanamu123456@tcp(192.168.1.234:3310)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local"

func initDB(str string) error {
	dboHandler, err := dbo.NewWithConfig(func(c *dbo.Config) {
		c.ShowLog = true
		c.ShowSQL = true
		c.MaxIdleConns = 2
		c.MaxOpenConns = 4
		c.ConnectionString = str
	})
	if err != nil {
		log.Error(context.TODO(), "create dbo failed", log.Err(err))
		return err
	}
	config.Set(&config.Config{
		RedisConfig: config.RedisConfig{
			OpenCache: false,
			Host:      "",
			Port:      0,
			Password:  "",
		},
	})
	dbo.ReplaceGlobal(dboHandler)
}

func loadContents() {
	ctx := context.Background()
	total, contentList, err := da.GetContentDA().SearchContent(ctx, dbo.MustGet(ctx), da.ContentCondition{})
}

func main() {
	//打开数据库
	err := initDB(connDBStr)
	if err != nil {
		fmt.Println("Can't open database, err:", err)
		return
	}
	//读取contents记录

	//mapper

	//更新contents记录
}
