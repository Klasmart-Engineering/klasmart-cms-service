package da

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"log"
	"strings"
	"testing"
)

func Test_GetLessonPlanIDsByCondition_Sql(t *testing.T) {
	c := ScheduleCondition{
		Status: sql.NullString{
			String: string(entity.ScheduleStatusClosed),
			Valid:  true,
		},
		// ClassID: sql.NullString{
		// 	String: "Class_1",
		// 	Valid:  true,
		// },
	}
	wheres, parameters := c.GetConditions()
	whereSql := strings.Join(wheres, " and ")
	t.Log(whereSql)
	t.Log(parameters)
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
	condition := NewScheduleTeachLoadCondition(input)

	GetScheduleDA().GetTeachLoadByCondition(ctx, dbo.MustGetDB(ctx), condition)
}
