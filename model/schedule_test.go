package model

import (
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

func TestScheduleModel_Add(t *testing.T) {
	s := entity.ScheduleDetailsView{
		ID:     "1",
		Title:  "是否",
		OrgID:  "2",
		Repeat: entity.RepeatOptions{},
		ScheduleBasic: entity.ScheduleBasic{
			Class: entity.ScheduleShortInfo{
				ID:   "1",
				Name: "班級",
			},
		},
	}
	b, _ := json.Marshal(s)
	t.Log(string(b))
}

func TestScheduleModel_GetByID(t *testing.T) {

}
