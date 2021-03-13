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
	token    = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTYyNzg3MiwiaXNzIjoia2lkc2xvb3AifQ.EnJ8b_PRo10_GVbOyV3GlF0YNGF2RFDRcS9Lb9aijpnfoHGcye_Dya8gAeZgiSyYczyUJs37mB4buq8Oe5OlcXyBoZj7lEJ1bvjZqJ4GxkZA0DPHUlXSFiZI8aD64J0DiR_fOjiWkeqYCwpaJ9yod9GTF0UAp9SE90GTosaUWqkgSPlHMXjlmgR8opLoOpP9ZgrtuXo1CcZA9MO_78RnkJKO4D3iUB7ewCK3A_kCUIPhbOMHRFAzFtvuSIzkGfkACcrlOQGRRj0A4YCXVWOu7Psj4i-xox0c8TosOWhHukvE98Jchy6_smiPPJUkKWBLbPmJ3z9opBK8uRumwmgJBA"
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
	condition.Pager = dbo.Pager{Page: 1, PageSize: 10}
	var schedules []*entity.Schedule
	err := da.GetScheduleDA().Query(ctx, condition, &schedules)
	if err != nil {
		log.Println(err)
		return
	}
	programIDs := make([]string, len(schedules))
	for i, item := range schedules {
		programIDs[i] = item.ProgramID
	}
	subjectIDs := make([]string, len(schedules))
	for i, item := range schedules {
		subjectIDs[i] = item.SubjectID
	}
	id, err := mapper.Program(ctx, operator.OrgID, programIDs[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(id)
}
