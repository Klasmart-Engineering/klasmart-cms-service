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
	token    = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTYzMjM4NCwiaXNzIjoia2lkc2xvb3AifQ.iOglxJSKo9f8IW14UTZahj2xVlYwu1XBsx-w7um639Bw-pTvfOBDyn2Hfip0x2nb8qC8z8lRiRiPB1bXoXy7bI6WcRynyldJY-RivLLjyc4IGBVysSo1T6VBNT9fc3ynk4iHwQ6CwwJPF_nH0AR6G8hqaWQ05P3qzDeoYVHYN0XosxP63zKAiZ0MHaudwXSoVO-C2VNMSrISPggxzjfHoEm_mSIqVuHzJAHFCyH68Ql03Odcu39P91vqSLcJpp77ga1FIJ-o8SGIeARyXoQrugknnLksdkay_B82eJA1p9OKK9yBYDO0OYRnyjYrVklBXnJSEKrCiHSba-eOwIVcEQ"
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
	point := "https://api.beta.kidsloop.net/user/"

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
	//
	programIDMap := make(map[string]bool)
	for _, item := range schedules {
		programIDMap[item.ProgramID] = true
	}
	subjectIDMap := make(map[string]bool)
	for _, item := range schedules {
		subjectIDMap[item.SubjectID] = true
	}
	fmt.Println(len(programIDMap), len(subjectIDMap))

	// scheduleProgramMap := make(map[string]string)
	// for _, item := range programIDs {
	// 	amsProgramID, err := mapper.Program(ctx, operator.OrgID, item)
	// 	if err != nil {
	// 		log.Println(err)
	// 		return
	// 	}
	// }
}
