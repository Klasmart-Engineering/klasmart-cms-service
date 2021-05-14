package main

import (
	"context"
	"flag"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"log"
)

func main() {
	cfg := &config.Config{
		DBConfig: config.DBConfig{
			ConnectionString: "",
			ShowLog:          true,
			ShowSQL:          true,
		},
	}

	flag.StringVar(&cfg.DBConfig.ConnectionString, "db", "root:Bada123@tcp(127.0.0.1:3307)/kidsloop2?parseTime=true&charset=utf8mb4", `db connection string,required`)
	flag.Parse()

	if cfg.DBConfig.ConnectionString == "" {
		fmt.Println("Please enter params, --help")
		return
	}
	fmt.Println("db:", cfg.DBConfig.ConnectionString)

	config.Set(cfg)
	option := dbo.WithConnectionString(cfg.DBConfig.ConnectionString)
	newDBO, err := dbo.NewWithConfig(option)
	if err != nil {
		log.Println("connection mysql db error:", err)
		return
	}
	dbo.ReplaceGlobal(newDBO)
	ctx := context.TODO()
	_, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, &entity.Operator{}, dbo.MustGetDB(ctx), &da.OutcomeCondition{
		PublishStatus: dbo.NullStrings{Strings: []string{entity.OutcomeStatusPublished}, Valid: true},
	})
	if err != nil {
		log.Panic(err)
	}
	orgIDsMap := make(map[string]bool)
	for i := range outcomes {
		if !orgIDsMap[outcomes[i].OrganizationID] {
			orgIDsMap[outcomes[i].OrganizationID] = true
		}
	}
	milestoneModel := model.GetMilestoneModel()
	for k, _ := range orgIDsMap {
		_, err = milestoneModel.CreateGeneral(ctx, &entity.Operator{OrgID: k}, dbo.MustGetDB(ctx), "")
		if err != nil {
			log.Panic(k, err)
		}
	}

	ancestorIDs := make(map[string]bool)
	for i := range outcomes {
		if !ancestorIDs[outcomes[i].AncestorID] {
			err = milestoneModel.BindToGeneral(ctx, &entity.Operator{}, dbo.MustGetDB(ctx), outcomes[i])
			if err != nil {
				log.Panic(*outcomes[i], err)
			}
			ancestorIDs[outcomes[i].AncestorID] = true
		}
	}
}
