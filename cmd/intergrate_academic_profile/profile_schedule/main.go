package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/cmd/intergrate_academic_profile"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	mapper   intergrate_academic_profile.Mapper
	token    = "" //"eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTY5MjQyMiwiaXNzIjoia2lkc2xvb3AifQ.B4J7eDGwOca2pCZj2Ch2Z20PfQkhAUAq9R37BcFvOSyjaJ-EEGgwrbSO9N9JuGc2vBFQAYCmxwIrDi9UNSJOOBovzz-JuwJmvKEaGnKPvDHqE4qYsrcDpRQTHTxmTgSvXBXr4grR8-2tWI1ZRBriXTb2Fm7yWXUBnRg-MFlLJ_90rRyoa7G-BSgokW2yOty8psqfnis3UEzPFJkztzOcn2w0RxynEpq_Sdz-_kDiWk6TuL3aNENp9hUTeZfFvfLnsuiBabakApfBjCGKLk2NidiSeeeFZo3TeNCAmjhpst37_aKxoH1LNJI582dd_Jk1nuRcIhTTKvCRQxeTWmYWYA"
	operator = &entity.Operator{
		UserID: "",                                     //"14494c07-0d4f-5141-9db2-15799993f448", // PJ
		OrgID:  "10f38ce9-5152-4049-b4e7-6d2e2ba884e6", // Badanamu HQ
		Token:  token,
	}
	isExec = false
)

func main() {
	//operator.Token = requestToken()

	initArgs()
	err := loadSchedule()
	if err != nil {
		log.Println("迁移失败", err)
		return
	}
	log.Println("迁移完成")

	//operator = &entity.Operator{
	//	UserID: "",
	//	OrgID:  "",
	//	Token:  "",
	//}
	//fmt.Println("Enter to continue mapper....")
	//inputReader := bufio.NewReader(os.Stdin)
	//inputReader.ReadString('\n')
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

func initArgs() {
	//root:Badanamu123456@tcp(192.168.1.234:3306)/kidsloop2?parseTime=true&charset=utf8mb4
	dsn := ""   //"root:Badanamu123456@tcp(192.168.1.234:3306)/kidsloop2?parseTime=true&charset=utf8mb4" //"admin:LH1MCuL3V0Ib3254@tcp(migration-test2.c2gspglsifnp.rds.cn-north-1.amazonaws.com.cn:28344)/kidsloop2?parseTime=true&charset=utf8mb4"
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

	fmt.Println("初始化数据完成，执行？: yes-执行，no-预执行 ")
	var exec = ""
	fmt.Scanln(&exec)
	if exec == "yes" {
		isExec = true
	} else {
		isExec = false
	}
}

func getSubjectMapKey(programID, subjectID string) string {
	return programID + ":" + subjectID
}

func loadSchedule() error {
	ctx := context.Background()
	condition := &da.ScheduleCondition{}
	//condition.Pager = dbo.Pager{Page: 1, PageSize: 1}
	var schedules []*entity.Schedule
	err := da.GetScheduleDA().Query(ctx, condition, &schedules)
	if err != nil {
		log.Println("get data from db error", err)
		return err
	}
	fmt.Println("schedule count", len(schedules))
	condition = &da.ScheduleCondition{
		DeleteAt: sql.NullInt64{
			Int64: 1,
			Valid: true,
		},
	}
	var schedulesDelete []*entity.Schedule
	err = da.GetScheduleDA().Query(ctx, condition, &schedulesDelete)
	if err != nil {
		log.Println("get data from db error", err)
		return err
	}
	schedules = append(schedules, schedulesDelete...)

	// programID->schedule array
	//programIDMap := make(map[string][]*entity.Schedule)
	// programID:subjectID->subjectID
	//subjectIDMap := make(map[string]string)
	programMap := make(map[string]string)
	subjectMap := make(map[string]string)

	for _, scheduleItem := range schedules {
		if scheduleItem.ProgramID == "" && scheduleItem.SubjectID == "" {
			continue
		}

		amsProgramID, err := mapper.Program(ctx, scheduleItem.OrgID, scheduleItem.ProgramID)
		if err != nil {
			log.Fatalf("ourProgramID:%s, amsProgramID:%s,error:%v \n", scheduleItem.ProgramID, amsProgramID, err)

			return err
		}

		newSubjectID, err := mapper.Subject(ctx, scheduleItem.OrgID, scheduleItem.ProgramID, scheduleItem.SubjectID)
		if err != nil {
			log.Fatalf("mapper Subject error,error:%v \n", scheduleItem.ProgramID, amsProgramID, err)

			return err
		}
		programMap[scheduleItem.OrgID+":"+scheduleItem.ProgramID] = amsProgramID
		subjectMap[scheduleItem.OrgID+":"+amsProgramID+":"+scheduleItem.SubjectID] = newSubjectID
	}
	fmt.Println("Enter to continue start update database....")
	inputReader := bufio.NewReader(os.Stdin)
	inputReader.ReadString('\n')

	tx, err := dbo.MustGetDB(ctx).DB.DB().Begin()
	for key, val := range programMap {
		keyArr := strings.Split(key, ":")
		if len(keyArr) < 2 {
			log.Println("")
			tx.Rollback()
			return constant.ErrInvalidArgs
		}
		orgID := keyArr[0]
		oldProgramID := keyArr[1]
		err = da.GetScheduleDA().UpdateProgram(ctx, tx, orgID, oldProgramID, val)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	for key, val := range subjectMap {
		keyArr := strings.Split(key, ":")
		if len(keyArr) < 3 {
			log.Println("")
			tx.Rollback()
			return constant.ErrInvalidArgs
		}
		oldID := keyArr[0]
		newProgramID := keyArr[1]
		oldSubjectID := keyArr[2]
		err = da.GetScheduleDA().UpdateSubject(ctx, tx, oldID, oldSubjectID, newProgramID, val)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	return err
	//count := 0
	//for key, scheduleList := range programIDMap {
	//	amsProgramID, err := mapper.Program(ctx, operator.OrgID, key)
	//	if err != nil {
	//		log.Fatalf("ourProgramID:%s, amsProgramID:%s,error:%v \n", key, amsProgramID, err)
	//		return err
	//	}
	//	length := len(scheduleList)
	//	count += length
	//	//log.Printf("ourProgramID:%s, amsProgramID:%s, scheduleLen:%d \n", key, amsProgramID, length)
	//}
	//log.Printf("orgID:%s,查询到的schedule总数:%d, 符合迁移条件的schedule总数：%d \n", operator.OrgID, len(schedules), count)
	//
	//programMapper := make(map[string]string)
	//
	//for key, _ := range programIDMap {
	//	newProgramID, err := mapper.Program(ctx, operator.OrgID, key)
	//	if err != nil {
	//		log.Printf("mapper program error,  orgID:%s,newProgramID:%s, oldProgramID：%s, error:%v \n", operator.OrgID, newProgramID, key, err)
	//		return err
	//	}
	//	log.Printf("orgID:%s,newProgramID:%s, oldProgramID：%s, \n", operator.OrgID, newProgramID, key)
	//	programMapper[key] = newProgramID
	//}
	//
	//subjectMapper := make(map[string]string)
	//for key, _ := range subjectIDMap {
	//	keyArr := strings.Split(key, ":")
	//	if len(keyArr) < 2 {
	//		log.Println("")
	//		return constant.ErrInvalidArgs
	//	}
	//	oldProgramID := keyArr[0]
	//	oldSubjectID := keyArr[1]
	//	newSubjectID, err := mapper.Subject(ctx, operator.OrgID, oldProgramID, oldSubjectID)
	//	newProgramID, err := mapper.Program(ctx, operator.OrgID, oldProgramID)
	//	if err != nil {
	//		log.Printf("mapper program error,  orgID:%s,newSubjectID:%s, key：%s, error:%v \n", operator.OrgID, newSubjectID, key, err)
	//		return err
	//	}
	//	log.Printf("orgID:%s,newSubjectID:%s, key：%s \n", operator.OrgID, newSubjectID, key)
	//	newKey := getSubjectMapKey(newProgramID, oldSubjectID)
	//	subjectMapper[newKey] = newSubjectID
	//}
	//fmt.Println("Enter to continue start update database....")
	//inputReader := bufio.NewReader(os.Stdin)
	//inputReader.ReadString('\n')
	//for key, val := range subjectMapper {
	//	fmt.Println(key, val)
	//}
	//
	//if isExec {
	//	tx, err := dbo.MustGetDB(ctx).DB.DB().Begin()
	//	for key, val := range programMapper {
	//		err = da.GetScheduleDA().UpdateProgram(ctx, tx, operator, key, val)
	//		if err != nil {
	//			tx.Rollback()
	//			return err
	//		}
	//	}
	//	for key, val := range subjectMapper {
	//		keyArr := strings.Split(key, ":")
	//		if len(keyArr) < 2 {
	//			log.Println("keyArr:", keyArr)
	//			return constant.ErrInvalidArgs
	//		}
	//		newProgramID := keyArr[0]
	//		oldSubjectID := keyArr[1]
	//		err = da.GetScheduleDA().UpdateSubject(ctx, tx, operator, oldSubjectID, newProgramID, val)
	//		if err != nil {
	//			tx.Rollback()
	//			return err
	//		}
	//	}
	//	//err = constant.ErrInvalidArgs
	//	//if err != nil {
	//	//	tx.Rollback()
	//	//	return err
	//	//}
	//
	//	err = tx.Commit()
	//	return err
	//	//dbo.MustGetDB(ctx).DB.DB().Begin()
	//	//err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
	//	//	for key, val := range programMapper {
	//	//		err = da.GetScheduleDA().UpdateProgram(ctx, tx, operator, key, val)
	//	//		if err != nil {
	//	//			return err
	//	//		}
	//	//	}
	//	//	for key, val := range subjectMapper {
	//	//		err = da.GetScheduleDA().UpdateSubject(ctx, tx, operator, key, val)
	//	//		if err != nil {
	//	//			return err
	//	//		}
	//	//	}
	//	//	return nil
	//	//})
	//	//return err
	//}
	//return nil
}
