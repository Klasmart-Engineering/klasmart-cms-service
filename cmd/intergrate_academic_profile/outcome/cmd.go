package main

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/cmd/intergrate_academic_profile"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"os"
	"strings"
)

func main() {
	setupConfig()
	ctx := context.TODO()
	tx := dbo.MustGetDB(ctx)
	mapper := intergrate_academic_profile.NewMapper(&entity.Operator{Token: ""})
	_, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, tx, &da.OutcomeCondition{
		Pager: dbo.Pager{
			Page: 1,
			PageSize: 100,
		},
		OrderBy: da.OrderByCreatedAt,
		IncludeDeleted: true,
	})
	if err != nil {
		log.Error(ctx, "outcome migrate failed",
			log.String("connect", config.Get().DBConfig.ConnectionString),
			log.Err(err))
		return
	}
	for i := range outcomes {
		programs := strings.Split(outcomes[i].Program, ",")
		subjects := strings.Split(outcomes[i].Subject, ",")
		categories := strings.Split(outcomes[i].Developmental, ",")
		subCategories := strings.Split(outcomes[i].Skills, ",")
		ages := strings.Split(outcomes[i].Age, ",")
		grades := strings.Split(outcomes[i].Grade, ",")
		outcomes[i].Program = ""
		for p := range programs {
			if programs[p] == "" {
				continue
			}
			pid, err := mapper.Program(ctx, outcomes[i].OrganizationID, programs[p])
			if err != nil {
				log.Error(ctx, "map program failed",
					log.Any("outcome", outcomes[i]),
					log.String("old program", programs[p]),
					log.Err(err))
				return
			}
			if pid == "" {
				log.Warn(ctx, "program warn",
					log.Any("outcome", outcomes[i]),
					log.String("old program", programs[p]),
					log.Err(err))
				continue
			}
			log.Info(ctx, "program",
				log.String("id", outcomes[i].ID),
				log.String("old program", programs[p]),
				log.String("new program", pid))
			outcomes[i].Program = outcomes[i].Program + pid + ","

			outcomes[i].Subject = ""
			for s := range subjects {
				if subjects[s] == "" {
					continue
				}
				sid, err := mapper.Subject(ctx, outcomes[i].OrganizationID, programs[p], subjects[s])
				if err != nil {
					log.Error(ctx, "map subject failed",
						log.Any("outcome", outcomes[i]),
						log.String("old program", programs[p]),
						log.String("new program", pid),
						log.String("old subject", subjects[s]),
						log.Err(err))
					return
				}
				if sid == "" {
					log.Warn(ctx, "subject warn",
						log.Any("outcome", outcomes[i]),
						log.String("old program", programs[p]),
						log.String("new program", pid),
						log.String("old subject", subjects[s]))
					continue
				}
				log.Info(ctx, "subject",
					log.String("id", outcomes[i].ID),
					log.String("program", programs[p]),
					log.String("old subject", subjects[s]),
					log.String("new subject", sid))
				outcomes[i].Subject = outcomes[i].Subject + sid + ","
			}
			strings.TrimSuffix(outcomes[i].Subject, ",")

			outcomes[i].Developmental = ""
			outcomes[i].Skills = ""
			for c := range categories {
				if categories[c] == "" {
					continue
				}
				cid , err := mapper.Category(ctx, outcomes[i].OrganizationID, programs[p], categories[c])
				if err != nil {
					log.Error(ctx, "map category failed",
						log.Any("outcome", outcomes[i]),
						log.String("old program", programs[p]),
						log.String("new program", pid),
						log.String("old category", categories[c]),
						log.Err(err))
					return
				}
				if cid == "" {
					log.Warn(ctx, "category warn",
						log.Any("outcome", outcomes[i]),
						log.String("old program", programs[p]),
						log.String("new program", pid),
						log.String("old category", categories[c]))
					continue
				}
				log.Info(ctx, "category",
					log.String("id", outcomes[i].ID),
					log.String("program", programs[p]),
					log.String("old category", categories[c]),
					log.String("new category", cid))
				outcomes[i].Developmental = outcomes[i].Developmental + cid + ","

				for sc := range subCategories {
					if subCategories[sc] == "" {
						continue
					}
					scid, err := mapper.SubCategory(ctx, outcomes[i].OrganizationID, programs[p], categories[c], subCategories[sc])
					if err != nil {
						log.Error(ctx, "map sub category failed",
							log.Any("outcome", outcomes[i]),
							log.String("old program", programs[p]),
							log.String("new program", pid),
							log.String("old category", categories[c]),
							log.String("new category", cid),
							log.Err(err))
						return
					}
					if scid == "" {
						log.Warn(ctx, "sub category warn",
							log.Any("outcome", outcomes[i]),
							log.String("old program", programs[p]),
							log.String("new program", pid),
							log.String("old category", categories[c]),
							log.String("new category", cid))
						continue
					}
					log.Info(ctx, "sub category",
						log.String("id", outcomes[i].ID),
						log.String("program", programs[p]),
						log.String("category", categories[c]),
						log.String("old sub category", subCategories[sc]),
						log.String("new sub category", scid))
					outcomes[i].Skills = outcomes[i].Skills + scid + ","
				}
			}
			strings.TrimSuffix(outcomes[i].Skills, ",")
			strings.TrimSuffix(outcomes[i].Developmental, ",")

			outcomes[i].Age = ""
			for a := range ages {
				if ages[a] == "" {
					continue
				}
				aid, err := mapper.Age(ctx, outcomes[i].OrganizationID, programs[p], ages[a])
				if err != nil {
					log.Error(ctx, "map age failed",
						log.Any("outcome", outcomes[i]),
						log.String("old program", programs[p]),
						log.String("new program", pid),
						log.String("old age", ages[a]),
						log.Err(err))
					return
				}
				if aid == "" {
					log.Warn(ctx, "age warn",
						log.Any("outcome", outcomes[i]),
						log.String("old program", programs[p]),
						log.String("new program", pid),
						log.String("old age", ages[a]))
					continue
				}
				log.Info(ctx, "age",
					log.String("id", outcomes[i].ID),
					log.String("program", programs[p]),
					log.String("old age", ages[a]),
					log.String("new age", aid))
				outcomes[i].Age = outcomes[i].Age + aid + ","
			}
			strings.TrimSuffix(outcomes[i].Age, ",")

			outcomes[i].Grade = ""
			for g := range grades {
				if grades[g] == "" {
					continue
				}
				gid, err := mapper.Grade(ctx, outcomes[i].OrganizationID, programs[p], grades[g])
				if err != nil {
					log.Error(ctx, "map grade failed",
						log.Any("outcome", outcomes[i]),
						log.String("old program", programs[p]),
						log.String("new program", pid),
						log.String("old grade", grades[g]),
						log.Err(err))
					return
				}
				if gid == "" {
					log.Warn(ctx, "age warn",
						log.Any("outcome", outcomes[i]),
						log.String("old program", programs[p]),
						log.String("new program", pid),
						log.String("old grade", grades[g]))
					continue
				}
				log.Info(ctx, "grade",
					log.String("id", outcomes[i].ID),
					log.String("program", programs[p]),
					log.String("old grade", grades[g]),
					log.String("new grade", gid))
				outcomes[i].Grade = outcomes[i].Grade + gid + ","
			}
			strings.TrimSuffix(outcomes[i].Grade, ",")
		}
		strings.TrimSuffix(outcomes[i].Program, ",")


		// TODO
		//err = da.GetOutcomeDA().UpdateOutcome(ctx, tx, outcomes[i])
		//if err != nil {
		//	log.Error(ctx, "update program failed",
		//		log.Any("outcome", outcomes[i]),
		//		log.Err(err))
		//}
	}
}

func setupConfig() {
	config.Set(&config.Config{
		DBConfig: config.DBConfig{
			ConnectionString: os.Getenv("connection_string"),
		},
		RedisConfig: config.RedisConfig{OpenCache: false},
		AMS: config.AMSConfig{
			EndPoint: os.Getenv("ams_endpoint"),
		},
	})
	dboHandler, err := dbo.NewWithConfig(func(c *dbo.Config) {
		dbConf := config.Get().DBConfig
		c.ShowLog = dbConf.ShowLog
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
