package model

import (
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

}
