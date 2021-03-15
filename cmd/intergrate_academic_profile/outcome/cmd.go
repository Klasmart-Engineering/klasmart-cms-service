package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/cmd/intergrate_academic_profile"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

//var token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTc5NTUyOCwiaXNzIjoia2lkc2xvb3AifQ.mLLPk3Qasoeow4Q5c-FWUBrhP7gURz35wlWUcq9ycjjOPTzIr8HN0lDjFwH6-3F-Tic6Skt1JSdJ-sS9jV4ZA4IsPHllIkvDXCSOGXKTXyUC0fLv_5GOS-5FHvjt5ekMtjhMRHtXKaimQ0xz_a2gPqxBWDzSs3Sy8svM4-wytshP53Cr31bUwLkGAnuG0h1StPMp8LYznIkAe9K7vjnwQJTvrCy-GbPLOA7bhXOVpkNut24-dIZEpPH_4KDjBUHsUvmzr5pHo-di1PhH0GmJBu1oBpn24Q1WG1CNcxRwJi7wuMZnfJdEBSwXhSf-YyvkdDV-jqBlS0-8rHOdR2bQcQ"
var orgID = "10f38ce9-5152-4049-b4e7-6d2e2ba884e6"

func isRecordNotFoundErr(err error) bool {
	if err.Error() == "record not found" {
		return true
	}
	return false
}

func requestToken() string {
	res, err := http.Get("http://192.168.1.233:10210/ll?email=pj.williams@calmid.com")
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	access := struct {
		Hit struct {
			Access string `json:"access"`
		} `json:"hit"`
	}{}
	err = json.Unmarshal(data, &access)
	if err != nil {
		panic(err)
	}
	return access.Hit.Access
}
func main() {
	setupConfig()
	withPage := flag.Bool("page", false, "true: page")
	tokenPtr := flag.String("token", "", "token")
	flag.Parse()
	if tokenPtr == nil {
		panic("need token")
	}
	token := *tokenPtr
	if token == "" {
		token = requestToken()
	}
	works := make(chan struct{}, 50)
	var wg sync.WaitGroup
	ctx := context.TODO()
	tx := dbo.MustGetDB(ctx)

	validID := regexp.MustCompile(`^([a-z]|[0-9])+-([a-z]|[0-9])+-([a-z]|[0-9])+-([a-z]|[0-9])+-([a-z]|[0-9])+$`)
	//valid := validID.MatchString("5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71")
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

	wg.Add(len(outcomes))
	for index := range outcomes {
		go func(i int) {
			works <- struct{}{}
			defer func() {
				<-works
				wg.Done()
			}()
			//ctx := context.TODO()
			//var programs, subjects, categories, subCategories, ages, grades []string
			programs := distinctSliceFromStr(outcomes[i].Program)
			subjects := distinctSliceFromStr(outcomes[i].Subject)
			categories := distinctSliceFromStr(outcomes[i].Developmental)
			subCategories := distinctSliceFromStr(outcomes[i].Skills)
			ages := distinctSliceFromStr(outcomes[i].Age)
			grades := distinctSliceFromStr(outcomes[i].Grade)
			outcomes[i].Program = ""
			outcomes[i].Subject = ""
			outcomes[i].Developmental = ""
			outcomes[i].Skills = ""
			outcomes[i].Age = ""
			outcomes[i].Grade = ""
			needUpdate := false

			if len(programs) == 0 {
				programs = append(programs, "")
			}
			if len(subjects) == 0 {
				subjects = append(subjects, "")
			}

			for p := range programs {
				if validID.MatchString(programs[p]) {
					log.Warn(ctx, "migrate:valid:program",
						log.String("outcome", outcomes[i].ID),
						log.String("program", programs[p]))
					continue
				}

				if programs[p] == "" {
					subjects = []string{""}
					categories = []string{""}
					subCategories = []string{""}
					ages = []string{""}
					grades = []string{""}
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
				needUpdate = true
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
					//if subjects[s] == ""{
					//	continue
					//}
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
						continue
					}
					log.Info(ctx, "migrate:info:subject",
						log.String("id", outcomes[i].ID),
						log.String("program", programs[p]),
						log.String("old subject", subjects[s]),
						log.String("new subject", sid))
					outcomes[i].Subject = outcomes[i].Subject + sid + ","
					break
				}

				for c := range categories {
					//if categories[c] == "" || programs[p] == "" {
					//	continue
					//}
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
						//if subCategories[sc] == "" || programs[p] == "" {
						//	continue
						//}
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

				for a := range ages {
					//if ages[a] == "" || programs[p] == "" {
					//	continue
					//}
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
						continue
					}
					log.Info(ctx, "migrate:info:age",
						log.String("id", outcomes[i].ID),
						log.String("program", programs[p]),
						log.String("old age", ages[a]),
						log.String("new age", aid))
					outcomes[i].Age = outcomes[i].Age + aid + ","
				}

				for g := range grades {
					if grades[g] == "" || programs[p] == "" {
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
						continue
					}
					log.Info(ctx, "migrate:info:grade",
						log.String("id", outcomes[i].ID),
						log.String("program", programs[p]),
						log.String("old grade", grades[g]),
						log.String("new grade", gid))
					outcomes[i].Grade = outcomes[i].Grade + gid + ","
				}
			}

			if needUpdate {
				outcomes[i].Program = strings.Join(distinctSliceFromStr(outcomes[i].Program), ",")
				outcomes[i].Subject = strings.Join(distinctSliceFromStr(outcomes[i].Subject), ",")
				outcomes[i].Developmental = strings.Join(distinctSliceFromStr(outcomes[i].Developmental), ",")
				outcomes[i].Skills = strings.Join(distinctSliceFromStr(outcomes[i].Skills), ",")
				outcomes[i].Age = strings.Join(distinctSliceFromStr(outcomes[i].Age), ",")
				outcomes[i].Grade = strings.Join(distinctSliceFromStr(outcomes[i].Grade), ",")

				log.Info(ctx, "migrate:finished",
					log.Int("index", i),
					log.Any("outcome", outcomes[i]))

				err = da.GetOutcomeDA().UpdateOutcome(ctx, dbo.MustGetDB(ctx), outcomes[i])
				if err != nil {
					log.Error(ctx, "update program failed",
						log.Any("outcome", outcomes[i]),
						log.Err(err))
				}
			} else {
				log.Info(ctx, "migrate:needn't",
					log.Int("index", i),
					log.Any("outcome", outcomes[i]))
			}
		}(index)
	}
	wg.Wait()
}

func distinctSliceFromStr(src string) []string {
	var result []string
	if src != "" && strings.Trim(src, ",") != "" {
		result = strings.Split(strings.Trim(src, ","), ",")
	}
	if len(result) > 1 {
		disMap := make(map[string]struct{})
		for i := range result {
			if result[i] != "" {
				disMap[result[i]] = struct{}{}
			}
		}

		var disResult []string
		for i := range result {
			if _, ok := disMap[result[i]]; ok {
				disResult = append(disResult, result[i])
			}
		}
		return disResult
	}
	return result
}

func setupConfig() {
	config.Set(&config.Config{
		DBConfig: config.DBConfig{
			ConnectionString: os.Getenv("connection_string"),
			//ConnectionString: os.Getenv("connection_string_local"),
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
