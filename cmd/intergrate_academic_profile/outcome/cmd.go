package main

import (
	"context"
	"flag"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/cmd/intergrate_academic_profile"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"os"
	"regexp"
	"strings"
	"sync"
)

var token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTc4MDQyMCwiaXNzIjoia2lkc2xvb3AifQ.Jw0Nl8w3JbQF1XvHDyJekn4rnU3faRZHQdYd9R0XjU73v3pqf--SIUK_M59dYd_bBO17zBKbO2paEbu54LcVAE5KhrRJcd8SZwG1wWYgWzDIBNfgqrUJFCNv4a2zJWaZSwYEfka2E4bX6aHjoix7fwObVWOs3ewbIVyIRuty_WPEAaVpXcgPowkT_L-NfRn6ZWLRdZg_YlmLhEA6x2XCViXprLa_M9GLuYQBSDffVkmqBLhLJRr-CX4uPLhjeqZBFbMcMNFfhUZ0bPJZsEqZxiroqtyCQgjZj1b34jzeyBWYs6xvSuB-RAAGMr4Oqe27M7Lh0oHaQOtdOW442zzVig"

func isRecordNotFoundErr(err error) bool {
	if err.Error() == "record not found" {
		return true
	}
	return false
}

var orgID = "10f38ce9-5152-4049-b4e7-6d2e2ba884e6"

func main() {
	setupConfig()

	works := make(chan struct{}, 20)
	var wg sync.WaitGroup
	ctx := context.TODO()
	tx := dbo.MustGetDB(ctx)
	validID := regexp.MustCompile(`^([a-z]|[0-9])+-([a-z]|[0-9])+-([a-z]|[0-9])+-([a-z]|[0-9])+-([a-z]|[0-9])+$`)
	//valid := validID.MatchString("5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71")
	withPage := flag.Bool("page", false, "true: page")
	flag.Parse()
	var outcomes []*entity.Outcome
	var err error
	if !(*withPage) {
		_, outcomes, err = da.GetOutcomeDA().SearchOutcome(ctx, tx, &da.OutcomeCondition{
			IncludeDeleted: true,
		})
		if err != nil {
			log.Error(ctx, "outcome migrate failed",
				log.String("connect", config.Get().DBConfig.ConnectionString),
				log.Err(err))
			return
		}
	} else {
		page := 1
		pageSize := 200
		for n := 200; n == pageSize; page++ {

			fmt.Printf("migrate:page:%d---------------------\n", page)
			var oc []*entity.Outcome
			_, oc, err = da.GetOutcomeDA().SearchOutcome(ctx, tx, &da.OutcomeCondition{
				Pager: dbo.Pager{
					Page:     page,
					PageSize: pageSize,
				},
				OrderBy:        da.OrderByCreatedAt,
				IncludeDeleted: true,
			})
			if err != nil {
				log.Error(ctx, "outcome migrate failed",
					log.String("connect", config.Get().DBConfig.ConnectionString),
					log.Int("page", page),
					log.Err(err))
				return
			}
			n = len(oc)
			outcomes = append(outcomes, oc...)
		}
	}

	mapper := intergrate_academic_profile.NewMapper(&entity.Operator{
		Token:  token,
		UserID: "14494c07-0d4f-5141-9db2-15799993f448",
		OrgID:  "10f38ce9-5152-4049-b4e7-6d2e2ba884e6",
	})

	for i := range outcomes {
		works <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-works
				wg.Done()
			}()
			programs := strings.Split(outcomes[i].Program, ",")
			subjects := strings.Split(outcomes[i].Subject, ",")
			categories := strings.Split(outcomes[i].Developmental, ",")
			subCategories := strings.Split(outcomes[i].Skills, ",")
			ages := strings.Split(outcomes[i].Age, ",")
			grades := strings.Split(outcomes[i].Grade, ",")
			outcomes[i].Program = ""
			outcomes[i].Subject = ""
			outcomes[i].Developmental = ""
			outcomes[i].Skills = ""
			outcomes[i].Age = ""
			outcomes[i].Grade = ""
			for p := range programs {
				if programs[p] == "" || validID.MatchString(programs[p]) {
					log.Warn(ctx, "migrate:valid:program",
						log.String("outcome", outcomes[i].ID),
						log.String("program", programs[p]))
					continue
				}
				org := orgID
				pid, err := mapper.Program(ctx, org, programs[p])
				if err != nil && !isRecordNotFoundErr(err) {
					log.Error(ctx, "map program failed",
						log.String("outcome", outcomes[i].ID),
						log.String("old program", programs[p]),
						log.Err(err))
					return
				}
				if pid == "" {
					log.Warn(ctx, "migrate:warn:program",
						log.String("outcome", outcomes[i].ID),
						log.String("old program", programs[p]))
					//fmt.Printf("Migrate: can't match program %s\n", programs[p])
					continue
				}
				log.Info(ctx, "migrate:info:program",
					log.String("id", outcomes[i].ID),
					log.String("old program", programs[p]),
					log.String("new program", pid))
				outcomes[i].Program = outcomes[i].Program + pid + ","

				for s := range subjects {
					if subjects[s] == "" {
						continue
					}
					sid, err := mapper.Subject(ctx, org, programs[p], subjects[s])
					if err != nil && !isRecordNotFoundErr(err) {
						log.Error(ctx, "map subject failed",
							log.String("outcome", outcomes[i].ID),
							log.String("old program", programs[p]),
							log.String("new program", pid),
							log.String("old subject", subjects[s]),
							log.Err(err))
						return
					}
					if sid == "" {
						log.Warn(ctx, "migrate:warn:subject",
							log.String("outcome", outcomes[i].ID),
							log.String("old program", programs[p]),
							log.String("new program", pid),
							log.String("old subject", subjects[s]))
						//fmt.Printf("Migrate: program %s can't match subject %s\n",
						//	programs[p], subjects[s])
						continue
					}
					log.Info(ctx, "migrate:info:subject",
						log.String("id", outcomes[i].ID),
						log.String("program", programs[p]),
						log.String("old subject", subjects[s]),
						log.String("new subject", sid))
					outcomes[i].Subject = outcomes[i].Subject + sid + ","
				}
				outcomes[i].Subject = strings.TrimSuffix(outcomes[i].Subject, ",")

				for c := range categories {
					if categories[c] == "" {
						continue
					}
					cid, err := mapper.Category(ctx, org, programs[p], categories[c])
					if err != nil && !isRecordNotFoundErr(err) {
						log.Error(ctx, "map category failed",
							log.String("outcome", outcomes[i].ID),
							log.String("old program", programs[p]),
							log.String("new program", pid),
							log.String("old category", categories[c]),
							log.Err(err))
						return
					}
					if cid == "" {
						log.Warn(ctx, "migrate:warn:category",
							log.String("outcome", outcomes[i].ID),
							log.String("old program", programs[p]),
							log.String("new program", pid),
							log.String("old category", categories[c]))
						//fmt.Printf("Migrate: program %s can't match category %s\n",
						//	programs[p], categories[c])
						continue
					}
					log.Info(ctx, "migrate:info:category",
						log.String("id", outcomes[i].ID),
						log.String("program", programs[p]),
						log.String("old category", categories[c]),
						log.String("new category", cid))
					outcomes[i].Developmental = outcomes[i].Developmental + cid + ","

					for sc := range subCategories {
						if subCategories[sc] == "" {
							continue
						}
						scid, err := mapper.SubCategory(ctx, org, programs[p], categories[c], subCategories[sc])
						if err != nil && !isRecordNotFoundErr(err) {
							log.Error(ctx, "map sub category failed",
								log.String("outcome", outcomes[i].ID),
								log.String("old program", programs[p]),
								log.String("new program", pid),
								log.String("old category", categories[c]),
								log.String("new category", cid),
								log.Err(err))
							return
						}
						if scid == "" {
							log.Warn(ctx, "migrate:warn:sub-category",
								log.String("outcome", outcomes[i].ID),
								log.String("old program", programs[p]),
								log.String("new program", pid),
								log.String("old category", categories[c]),
								log.String("new category", cid),
								log.String("old sub-category", subCategories[sc]))
							//fmt.Printf("Migrate: program %s category %s can't match subCategory %s\n",
							//	programs[p], categories[c], subCategories[sc])
							continue
						}
						log.Info(ctx, "migrate:info:sub-category",
							log.String("id", outcomes[i].ID),
							log.String("program", programs[p]),
							log.String("category", categories[c]),
							log.String("old sub category", subCategories[sc]),
							log.String("new sub category", scid))
						outcomes[i].Skills = outcomes[i].Skills + scid + ","
					}
				}
				if outcomes[i].Skills != "" {
					outcomes[i].Skills = strings.TrimSuffix(outcomes[i].Skills, ",")
				}
				if outcomes[i].Developmental != "" {
					outcomes[i].Developmental = strings.TrimSuffix(outcomes[i].Developmental, ",")
				}

				for a := range ages {
					if ages[a] == "" {
						continue
					}
					aid, err := mapper.Age(ctx, org, programs[p], ages[a])
					if err != nil && !isRecordNotFoundErr(err) {
						log.Error(ctx, "map age failed",
							log.String("outcome", outcomes[i].ID),
							log.String("old program", programs[p]),
							log.String("new program", pid),
							log.String("old age", ages[a]),
							log.Err(err))
						return
					}
					if aid == "" {
						log.Warn(ctx, "migrate:warn:age",
							log.String("outcome", outcomes[i].ID),
							log.String("old program", programs[p]),
							log.String("new program", pid),
							log.String("old age", ages[a]))
						//fmt.Printf("Migrate: program %s can't match age %s\n",
						//	programs[p], ages[a])
						continue
					}
					log.Info(ctx, "migrate:info:age",
						log.String("id", outcomes[i].ID),
						log.String("program", programs[p]),
						log.String("old age", ages[a]),
						log.String("new age", aid))
					outcomes[i].Age = outcomes[i].Age + aid + ","
				}
				if outcomes[i].Age != "" {
					outcomes[i].Age = strings.TrimSuffix(outcomes[i].Age, ",")
				}

				for g := range grades {
					if grades[g] == "" {
						continue
					}
					gid, err := mapper.Grade(ctx, org, programs[p], grades[g])
					if err != nil && !isRecordNotFoundErr(err) {
						log.Error(ctx, "map grade failed",
							log.String("outcome", outcomes[i].ID),
							log.String("old program", programs[p]),
							log.String("new program", pid),
							log.String("old grade", grades[g]),
							log.Err(err))
						return
					}
					if gid == "" {
						log.Warn(ctx, "migrate:warn:grade",
							log.String("outcome", outcomes[i].ID),
							log.String("old program", programs[p]),
							log.String("new program", pid),
							log.String("old grade", grades[g]))
						//fmt.Printf("Migrate: program %s can't match grade %s\n",
						//	programs[p], grades[g])
						continue
					}
					log.Info(ctx, "migrate:info:grade",
						log.String("id", outcomes[i].ID),
						log.String("program", programs[p]),
						log.String("old grade", grades[g]),
						log.String("new grade", gid))
					outcomes[i].Grade = outcomes[i].Grade + gid + ","
				}
				if outcomes[i].Grade != "" {
					outcomes[i].Grade = strings.TrimSuffix(outcomes[i].Grade, ",")
				}
			}
			if outcomes[i].Program != "" {
				outcomes[i].Program = strings.TrimSuffix(outcomes[i].Program, ",")
			}

			fmt.Printf("migrate:finished:%d:-------------------\n", i)
			//err = da.GetOutcomeDA().UpdateOutcome(ctx, tx, outcomes[i])
			//if err != nil {
			//	log.Error(ctx, "update program failed",
			//		log.Any("outcome", outcomes[i]),
			//		log.Err(err))
			//}
		}()
	}
	wg.Wait()
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
