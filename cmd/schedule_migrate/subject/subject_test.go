package main

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"testing"
)

func Test_GetScheduleAboutOrgIDs(t *testing.T) {
	dsn := "root:Badanamu123456@tcp(192.168.1.234:3306)/kidsloop2?parseTime=true&charset=utf8mb4"
	option := dbo.WithConnectionString(dsn)
	newDBO, err := dbo.NewWithConfig(option)
	if err != nil {
		t.Log(err)
		return
	}
	dbo.ReplaceGlobal(newDBO)
	orgIDs, _ := GetScheduleAboutOrgIDs(context.Background())
	t.Log(orgIDs)
}

func Test_StartMigrateByOrg(t *testing.T) {
	ctx := context.Background()
	dsn := "root:Badanamu123456@tcp(192.168.1.234:3306)/kidsloop2?parseTime=true&charset=utf8mb4"
	option := dbo.WithConnectionString(dsn)
	newDBO, err := dbo.NewWithConfig(option)
	if err != nil {
		t.Log(err)
		return
	}
	dbo.ReplaceGlobal(newDBO)
	orgIDs, _ := GetScheduleAboutOrgIDs(ctx)
	orgID := orgIDs[0]
	t.Log(orgID)
	row, err := StartMigrateByOrg(ctx, orgID)

	t.Log(row)
}
