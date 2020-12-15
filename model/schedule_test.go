package model

import (
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
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
