package model

import (
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
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
	var s1 []*external.Class
	s1 = append(s1, &external.Class{
		ID:   "1",
		Name: "aaaa",
	})
	s1 = append(s1, &external.Class{
		ID:   "2",
		Name: "bbbb",
	})
	var s2 []*external.Class
	s2 = s1
	s2 = append(s1, &external.Class{
		ID:   "3",
		Name: "cccc",
	})
	t.Log("s1:", len(s1))
	t.Log("s2:", len(s2))

	t.Log("s1:", s1[0].ID)
	t.Log("s2:", s2[0].ID)
	s1[0].ID = "10"
	t.Log("s1:", s1[0].ID)
	t.Log("s2:", s2[0].ID)
	s2[2].ID = "11"
	t.Log("s1:", s1[1].ID)
	t.Log("s2:", s2[2].ID)

}
