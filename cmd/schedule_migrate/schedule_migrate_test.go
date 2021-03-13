package main

import (
	"context"
	"fmt"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

func TestAMS(t *testing.T) {
	cfg := &config.Config{
		AMS: config.AMSConfig{
			EndPoint:       "https://api.kidsloop.net/user/",
			TokenVerifyKey: "",
		},
	}
	config.Set(cfg)
	result, err := external.GetClassServiceProvider().BatchGet(context.Background(), &entity.Operator{
		UserID: "",
		Role:   "",
		OrgID:  "",
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImQ5NGI2ZmNjLTk0MDgtNTRjMi1hYzI5LTEzYzc5NmY0Yjg2MSIsImVtYWlsIjoidGVhY2hjazAwMTExQHlvcG1haWwuY29tIiwiZXhwIjoxNjEwMzQwNzIzLCJpc3MiOiJraWRzbG9vcCJ9.RHQ8Xm4HnH_6nEEyCAKDkqskOkI_zsCz10IU8OkuT4tQLgtGc7a0PLTLPpngyG9K0a409Rq87VDVBkY4B8W7buXmjtrDcMzSiHlxK1i4gwV6-iR2KPVQniHiC1urqfEh8Pe2eXM38tE5SERgJRIeehwB8bh3TQeBUldEABh0IqHrwPVOSkgLieRbxQZYYP2h1GppSeaYVC-hEqHO88e0dgN3p6r0iheECstZqdKdioJFtjoVMOp-xlS23Rs-J3AF0nYDZZn6XN02EUYMcmi0LtGT30qyrIOl4ULyRlNDcG165F_n9kg_U6tSwlR7s2SiwOLiGabSNFAG4CtoOBT0Iw; _gat_gtag_UA_149920485_2=1",
	}, []string{"cec8deca-1107-4d08-8507-030accafc0e2"})
	if err != nil {
		t.Log(err)
		return
	}
	for _, item := range result {
		t.Log(item.Name)
	}

}

func TestDB(t *testing.T) {
	dsn := "admin:LH1MCuL3V0Ib3254@tcp(kl2-migration-test.copqnkcbdsts.ap-northeast-2.rds.amazonaws.com:28344)/kidsloop2?parseTime=true&charset=utf8mb4" //"admin:LH1MCuL3V0Ib3254@tcp(migration-test2.c2gspglsifnp.rds.cn-north-1.amazonaws.com.cn:28344)/kidsloop2?parseTime=true&charset=utf8mb4"
	option := dbo.WithConnectionString(dsn)
	newDBO, err := dbo.NewWithConfig(option)
	if err != nil {
		t.Log(err)
		return
	}
	dbo.ReplaceGlobal(newDBO)

	// var scheduleList []*entity.Schedule
	// err = da.GetScheduleDA().Query(context.Background(), &da.ScheduleCondition{}, &scheduleList)
	// if err != nil {
	// 	t.Log(err)
	// 	return
	// }
	ourAges, err := model.GetAgeModel().Query(context.Background(), &da.AgeCondition{})
	if err != nil {
		t.Log(err)
		return
	}

	t.Log(ourAges)
}

func TestGetAboutOrgInfo(t *testing.T) {
	cfg := &config.Config{
		AMS: config.AMSConfig{
			EndPoint:       "https://api.kidsloop.net/user/",
			TokenVerifyKey: "",
		},
	}
	config.Set(cfg)
	dsn := "root:Passw0rd@tcp(127.0.0.1:3306)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local"
	option := dbo.WithConnectionString(dsn)
	newDBO, err := dbo.NewWithConfig(option)
	if err != nil {
		t.Log(err)
		return
	}
	dbo.ReplaceGlobal(newDBO)
	op := &entity.Operator{
		UserID: "",
		Role:   "",
		OrgID:  "",
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImQ5NGI2ZmNjLTk0MDgtNTRjMi1hYzI5LTEzYzc5NmY0Yjg2MSIsImVtYWlsIjoidGVhY2hjazAwMTExQHlvcG1haWwuY29tIiwiZXhwIjoxNjEwMzYxMTE3LCJpc3MiOiJraWRzbG9vcCJ9.uxQuYcezJf27oamOcZTm8tXTrcdPfnie_u8MNuV1af7bVX40CDqMknG903ertkjoFjBHb8_RujeIsIiYPfsROCpKdUsMmQyUpSjxgaQPjG_n6-1MSvqoujXNsrb1nhcL0gV_e179UBYDn9wxv3zR7WhJDqK9UJnab14qRygJvJ8ElYB3EOYaKflRTZsReUBy53Wke4OCwatHijijOMXM4KOIWiRRx6-9zpAwmiJf0h75d3_VAZJ2msNSn6VBY4qMUbrXYa2vjCdI4_rMgnu0zvdG8NrImnQPWwlU8K8yNTbYVrY46hazWYhN7RY6ZXful1kSudx7yMOAUWu_P44C1w; _gat_gtag_UA_149920485_2=1",
	}
	data, err := GetScheduleAboutOrgList(context.Background(), op)
	if err != nil {
		t.Log(err)
		return
	}
	for _, item := range data {
		fmt.Println(*item)
	}
}
func InitData() {
	cfg := &config.Config{
		AMS: config.AMSConfig{
			EndPoint:       "https://api.kidsloop.net/user/",
			TokenVerifyKey: "",
		},
	}
	config.Set(cfg)
	dsn := "root:Passw0rd@tcp(127.0.0.1:3306)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local"
	option := dbo.WithConnectionString(dsn)
	newDBO, err := dbo.NewWithConfig(option)
	if err != nil {
		fmt.Println(err)
		return
	}
	dbo.ReplaceGlobal(newDBO)
}

func TestPrepareDataByOrg(t *testing.T) {
	InitData()
	op := &entity.Operator{
		UserID: "",
		Role:   "",
		OrgID:  "",
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImQ5NGI2ZmNjLTk0MDgtNTRjMi1hYzI5LTEzYzc5NmY0Yjg2MSIsImVtYWlsIjoidGVhY2hjazAwMTExQHlvcG1haWwuY29tIiwiZXhwIjoxNjEwMzYxMTE3LCJpc3MiOiJraWRzbG9vcCJ9.uxQuYcezJf27oamOcZTm8tXTrcdPfnie_u8MNuV1af7bVX40CDqMknG903ertkjoFjBHb8_RujeIsIiYPfsROCpKdUsMmQyUpSjxgaQPjG_n6-1MSvqoujXNsrb1nhcL0gV_e179UBYDn9wxv3zR7WhJDqK9UJnab14qRygJvJ8ElYB3EOYaKflRTZsReUBy53Wke4OCwatHijijOMXM4KOIWiRRx6-9zpAwmiJf0h75d3_VAZJ2msNSn6VBY4qMUbrXYa2vjCdI4_rMgnu0zvdG8NrImnQPWwlU8K8yNTbYVrY46hazWYhN7RY6ZXful1kSudx7yMOAUWu_P44C1w; _gat_gtag_UA_149920485_2=1",
	}
	StartMigrateByOrg(context.Background(), op, "10f38ce9-5152-4049-b4e7-6d2e2ba884e6")
}

func TestPage(t *testing.T) {
	strs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}

	total := len(strs)
	pageSize := 5
	pageCount := (total + pageSize - 1) / pageSize
	for i := 0; i < pageCount; i++ {
		start := i * pageSize
		end := (i + 1) * pageSize
		if end >= total {
			end = total
		}
		data := strs[start:end]
		t.Log(data)
	}
}
