package model

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
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
func TestTagModel_BatchAdd(t *testing.T) {
	for i:=0;i<10;i++{
		name:=fmt.Sprintf("tag-%s",utils.NewId())
		id,err:=GetTagModel().Add(context.Background(),&entity.TagAddView{Name: name})
		if err!=nil{
			fmt.Println(err)
			return
		}
		fmt.Println(id)
	}
}

func TestTagModel_Update(t *testing.T) {
	err:=GetTagModel().Update(context.Background(),&entity.TagUpdateView{
		ID:     "1",
		Name:   "video",
		States: 2,
	})
	fmt.Println(err)
}
func TestTagModel_Delete(t *testing.T) {
	err:=GetTagModel().Delete(context.Background(),"3357bafe-7463-4254-8e48-b37fa3cc23ee")
	fmt.Println(err)
}

func TestTagModel_Query(t *testing.T) {
	resut,err:=GetTagModel().Query(context.Background(),da.TagCondition{
		Name:     "",
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
func TestTagModel_Page(t *testing.T) {
	resut,err:=GetTagModel().Page(context.Background(),da.TagCondition{
		Name:     "",
		PageSize: 0,
		Page:     1,
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

func TestTagModel_GetByIDs(t *testing.T) {
	result,_:=GetTagModel().GetByIDs(context.Background(),[]string{
		"6235f3c66cb63d43","351d1c9472be37e3",
	})
	for _,item:=range result{
		fmt.Println(item)
	}
}
