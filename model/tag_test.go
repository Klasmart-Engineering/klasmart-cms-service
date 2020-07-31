package model

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)
func TestTagModel_Add(t *testing.T) {
	id,err:=GetTagModel().Add(context.Background(),&entity.TagAddView{Name: "history"})
	if err!=nil{
		fmt.Println(err)
		return
	}
	fmt.Println(id)
}

func TestTagModel_Update(t *testing.T) {
	err:=GetTagModel().Update(context.Background(),&entity.TagUpdateView{
		ID:     "3357bafe-7463-4254-8e48-b37fa3cc23ee",
		Name:   "video",
		States: 2,
	})
	fmt.Println(err)
}

func TestTagModel_Query(t *testing.T) {
	resut,err:=GetTagModel().Query(context.Background(),&da.TagCondition{
		Name:     "history",
		PageSize: 0,
		Page:     0,
		DeleteAt: 0,
	})
	if err!=nil{
		fmt.Println(err)
		return
	}
	for _,item:=range resut{
		fmt.Println(*item)
	}
}
