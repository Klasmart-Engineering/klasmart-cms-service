package model

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"log"
	"testing"
	"time"
)

func TestScheduleModel_Add(t *testing.T) {

}

func TestScheduleModel_GetByID(t *testing.T) {
	tt := time.Now().Add(1 * time.Hour).Unix()
	t.Log(tt)
	tt2 := time.Now().Add(2 * time.Hour).Unix()
	t.Log(tt2)
}

func TestTemp(t *testing.T) {
	diff := constant.ScheduleAllowGoLiveTime
	loc, _ := time.LoadLocation("Africa/Algiers")
	now := time.Now().In(loc)
	temp := now.Add(diff)
	fmt.Println(temp.Unix())
	now = time.Now()
	temp = now.Add(-diff)
	fmt.Println(temp)
}

func TestGetTeachLoadByCondition(t *testing.T) {
	cfg := &config.Config{
		DBConfig: config.DBConfig{
			ConnectionString: "root:Badanamu123456@tcp(192.168.1.234:3306)/kidsloop2?parseTime=true&charset=utf8mb4",
		},
	}
	config.Set(cfg)
	option := dbo.WithConnectionString(cfg.DBConfig.ConnectionString)
	newDBO, err := dbo.NewWithConfig(option)
	if err != nil {
		log.Println("connection mysql db error:", err)
		return
	}
	dbo.ReplaceGlobal(newDBO)
	ctx := context.Background()
	input := &entity.ScheduleTeachingLoadInput{
		OrgID:      "72e47ef0-92bf-4429-a06f-2014e3d3df4b",
		ClassIDs:   []string{"5751555a-cc18-4662-9ae5-a5ad90569f79", "49b3be6a-d139-4f82-9b77-0acc89525d3f"},
		TeacherIDs: []string{"4fde6e1b-8efe-58e9-a404-51fb98ebf9b8", "42098862-28b1-5417-9800-3b89e557a2b9"},
		TimeRanges: []*entity.ScheduleTimeRange{
			&entity.ScheduleTimeRange{
				StartAt: 1605456000,
				EndAt:   1605544740,
			},
			&entity.ScheduleTimeRange{
				StartAt: 1605456000,
				EndAt:   1605544740,
			},
			&entity.ScheduleTimeRange{
				StartAt: 1605456000,
				EndAt:   1605544740,
			},
		},
	}
	result, _ := GetScheduleModel().GetTeachingLoad(ctx, input)
	for _, item := range result {
		t.Log(item.TeacherID, ":", item.ClassType, item.Durations)
	}
}
