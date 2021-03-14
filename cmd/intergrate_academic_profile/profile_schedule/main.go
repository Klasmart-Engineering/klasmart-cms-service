package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/cmd/intergrate_academic_profile"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var (
	mapper   intergrate_academic_profile.Mapper
	token    = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTY5MjQyMiwiaXNzIjoia2lkc2xvb3AifQ.B4J7eDGwOca2pCZj2Ch2Z20PfQkhAUAq9R37BcFvOSyjaJ-EEGgwrbSO9N9JuGc2vBFQAYCmxwIrDi9UNSJOOBovzz-JuwJmvKEaGnKPvDHqE4qYsrcDpRQTHTxmTgSvXBXr4grR8-2tWI1ZRBriXTb2Fm7yWXUBnRg-MFlLJ_90rRyoa7G-BSgokW2yOty8psqfnis3UEzPFJkztzOcn2w0RxynEpq_Sdz-_kDiWk6TuL3aNENp9hUTeZfFvfLnsuiBabakApfBjCGKLk2NidiSeeeFZo3TeNCAmjhpst37_aKxoH1LNJI582dd_Jk1nuRcIhTTKvCRQxeTWmYWYA"
	operator = &entity.Operator{
		UserID: "14494c07-0d4f-5141-9db2-15799993f448", // PJ
		OrgID:  "10f38ce9-5152-4049-b4e7-6d2e2ba884e6", // Badanamu HQ
		Token:  token,
	}
	isExec = false
)

func main() {
	for {
		initArgs()
		// err := loadSchedule()
		// if err != nil {
		// 	log.Println("迁移失败", err)
		// 	return
		// }
		log.Println("迁移完成")
	}
}

func initArgs() {
	dsn := ""   //"admin:LH1MCuL3V0Ib3254@tcp(migration-test2.c2gspglsifnp.rds.cn-north-1.amazonaws.com.cn:28344)/kidsloop2?parseTime=true&charset=utf8mb4"
	point := "" //"https://api.kidsloop.net/user/"

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
		strings.TrimSpace(operator.Token) == "" ||
		strings.TrimSpace(operator.OrgID) == "" {
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

func loadSchedule() error {
	ctx := context.Background()
	condition := &da.ScheduleCondition{
		OrgID: sql.NullString{
			String: operator.OrgID,
			Valid:  true,
		},
	}
	//condition.Pager = dbo.Pager{Page: 1, PageSize: 10}
	var schedules []*entity.Schedule
	err := da.GetScheduleDA().Query(ctx, condition, &schedules)
	if err != nil {
		log.Println("get data from db error", err)
		return err
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
	count := 0
	for key, scheduleList := range programIDMap {
		amsProgramID, err := mapper.Program(ctx, operator.OrgID, key)
		if err != nil {
			log.Fatalf("ourProgramID:%s, amsProgramID:%s,error:%v \n", key, amsProgramID, err)
			return err
		}
		length := len(scheduleList)
		count += length
		//log.Printf("ourProgramID:%s, amsProgramID:%s, scheduleLen:%d \n", key, amsProgramID, length)
	}
	log.Printf("orgID:%s,查询到的schedule总数:%d, 符合迁移条件的schedule总数：%d \n", operator.OrgID, len(schedules), count)

	fmt.Println("Enter to continue....")
	inputReader := bufio.NewReader(os.Stdin)
	inputReader.ReadString('\n')

	if isExec {
		err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
			for key, _ := range programIDMap {
				newProgramID, err := mapper.Program(ctx, operator.OrgID, key)
				if err != nil {
					log.Printf("mapper program error,  orgID:%s,newProgramID:%s, oldProgramID：%s, error:%v \n", operator.OrgID, newProgramID, key, err)
					return err
				}
				err = da.GetScheduleDA().UpdateProgram(ctx, tx, operator, key, newProgramID)
				if err != nil {
					return err
				}
			}
			return nil
		})
		return err
	}
	return nil
}
