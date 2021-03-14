package profile_schedule

import (
	"context"
	"fmt"
	"log"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/cmd/intergrate_academic_profile"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var (
	mapper   intergrate_academic_profile.Mapper
	token    = "yJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTY5MTUxOCwiaXNzIjoia2lkc2xvb3AifQ.ZocoKugAdJy9nePic1FshT6vz7SFkglf7dZNhYP9Gz8Su1vELmqADSNxHX5zDxWQ_nnigfINBZlEKOJkXlDI5cSa9WPtZiPRPOu-shyS97BTNdO3M37vNcXcXB5NM8dFb44qBaYU_MbjoM1wxIneoTafBkbIApinTMVqFJ8U7MZVsVvT8cCoZgSEtEba5O3u5gSElPP8q22Tyq5flqOyZjvB9cijURra9e5CTXNM7SxJ8re-ePfbuIcypaWhDBKc9wPPiD8BC2bllvQIZJxSMsYlLareI9UmL642hNJfa1rn0sZaj7XHAPCvDECQndFQ1pjwLQME6pnxp3dMFwJiEg"
	operator = &entity.Operator{
		UserID: "14494c07-0d4f-5141-9db2-15799993f448", // PJ
		OrgID:  "10f38ce9-5152-4049-b4e7-6d2e2ba884e6", // Badanamu HQ
		Token:  token,
	}
)

func main() {
	initArgs()
	loadSchedule()
}

func initArgs() {
	dsn := "admin:LH1MCuL3V0Ib3254@tcp(migration-test2.c2gspglsifnp.rds.cn-north-1.amazonaws.com.cn:28344)/kidsloop2?parseTime=true&charset=utf8mb4"
	point := "https://api.kidsloop.net/user/"

	cfg := &config.Config{
		DBConfig: config.DBConfig{
			ConnectionString: dsn,
		},
		AMS: config.AMSConfig{
			EndPoint: point,
		},
	}

	for strings.TrimSpace(cfg.DBConfig.ConnectionString) == "" ||
		strings.TrimSpace(cfg.AMS.EndPoint) == "" ||
		strings.TrimSpace(operator.Token) == "" {
		if strings.TrimSpace(cfg.DBConfig.ConnectionString) == "" {
			fmt.Println("Please enter mysql dsn: ")
			fmt.Scanln(&cfg.DBConfig.ConnectionString)
		}
		if strings.TrimSpace(cfg.AMS.EndPoint) == "" {
			fmt.Println("Please enter ams endPoint: ")
			fmt.Scanln(&cfg.AMS.EndPoint)
		}
		if strings.TrimSpace(operator.Token) == "" {
			fmt.Println("Please enter login operator token: ")
			fmt.Scanln(&operator.Token)
		}
		if strings.TrimSpace(operator.OrgID) == "" {
			fmt.Println("Please enter login operator org_id: ")
			fmt.Scanln(&operator.OrgID)
		}
	}
	config.Set(cfg)

	option := dbo.WithConnectionString(cfg.DBConfig.ConnectionString)
	newDBO, err := dbo.NewWithConfig(option)
	if err != nil {
		log.Println("connection mysql db error:", err)
		return
	}
	dbo.ReplaceGlobal(newDBO)

	mapper = intergrate_academic_profile.NewMapper(operator)
}

func loadSchedule() {
	ctx := context.Background()
	condition := &da.ScheduleCondition{}
	//condition.Pager = dbo.Pager{Page: 1, PageSize: 10}
	var schedules []*entity.Schedule
	err := da.GetScheduleDA().Query(ctx, condition, &schedules)
	if err != nil {
		log.Println(err)
		return
	}

	programIDMap := make(map[string][]*entity.Schedule)
	subjectIDMap := make(map[string][]*entity.Schedule)

	for _, scheduleItem := range schedules {
		if _, ok := programIDMap[scheduleItem.ProgramID]; !ok {
			programIDMap[scheduleItem.ProgramID] = make([]*entity.Schedule, 0)
		}
		programIDMap[scheduleItem.ProgramID] = append(programIDMap[scheduleItem.ProgramID], scheduleItem)

		if _, ok := subjectIDMap[scheduleItem.SubjectID]; !ok {
			subjectIDMap[scheduleItem.SubjectID] = make([]*entity.Schedule, 0)
		}
		subjectIDMap[scheduleItem.SubjectID] = append(subjectIDMap[scheduleItem.SubjectID], scheduleItem)
	}

	fmt.Println(len(programIDMap), len(subjectIDMap))

	for key, scheduleList := range programIDMap {
		amsProgramID, err := mapper.Program(ctx, operator.OrgID, key)
		if err != nil {
			log.Println(err)
			return
		}
		for _, item := range scheduleList {
			log.Printf("ourProgramID:%s, amsProgramID:%s, scheduleID:%s \n", key, amsProgramID, item.ID)
		}
		log.Println("..............")
	}
}
