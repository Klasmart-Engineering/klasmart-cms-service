package model

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"os"
	"testing"
)

func setup() {
	config.Set(&config.Config{
		DBConfig: config.DBConfig{
			ConnectionString: os.Getenv("connection_string"),
		},
		RedisConfig: config.RedisConfig{
			OpenCache: true,
			Host:      os.Getenv("redis_host"),
			Port:      16379,
			Password:  "",
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
	ro.SetConfig(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", config.Get().RedisConfig.Host, config.Get().RedisConfig.Port),
		Password: config.Get().RedisConfig.Password,
	})
}

func TestOutcomeModel_GetLearningOutcomeByID(t *testing.T) {
	setup()
	ctx := context.TODO()
	outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, dbo.MustGetDB(ctx), "60616800e7c9026bb00d1d6c")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", outcome)
}

func TestGenerateShortcode(t *testing.T) {
	setup()
	ctx := context.TODO()
	shortcode, err := GetShortcodeModel().Generate(ctx, dbo.MustGetDB(ctx), entity.KindOutcome, "org-1", "")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(shortcode)
}

func TestSearchSetsByOutcome(t *testing.T) {
	setup()
	ctx := context.TODO()
	outcomesSets, err := da.GetOutcomeSetDA().SearchSetsByOutcome(ctx, dbo.MustGetDB(ctx), []string{"60616800e7c9026bb00d1d6c"})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", outcomesSets)
}

func TestOutcomeSetModel_CreateOutcomeSet(t *testing.T) {
	setup()
	ctx := context.TODO()
	set, err := GetOutcomeSetModel().CreateOutcomeSet(ctx, &entity.Operator{OrgID: "org-1"}, "math2")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", set)
}
