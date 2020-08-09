package model

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"testing"
)
func TestTagModel_Add(t *testing.T) {
	id,err:=GetTagModel().Add(context.Background(),&entity.Operator{
		UserID: "1",
		Role:   "2",
	},&entity.TagAddView{Name: "history"})
	if err!=nil{
		fmt.Println(err)
		return
	}
	fmt.Println(id)
}
func TestTagModel_BatchAdd(t *testing.T) {
	for i:=0;i<10;i++{
		name:=fmt.Sprintf("tag-%s",utils.NewID())
		id,err:=GetTagModel().Add(context.Background(),&entity.Operator{
			UserID: "1",
			Role:   "2",
		},&entity.TagAddView{Name: name})
		if err!=nil{
			fmt.Println(err)
			return
		}
		fmt.Println(id)
	}
}
//&{5f2bc7213ed54eb1351e4501 history 1 1   1596704545 0 0}
func TestTagModel_Update(t *testing.T) {
	err:=GetTagModel().Update(context.Background(),&entity.Operator{
		UserID: "1",
		Role:   "2",
	},&entity.TagUpdateView{
		ID:     "5f2bc7213ed54eb1351e4501",
		Name:   "video",
		States: constant.Disabled,
	})
	fmt.Println(err)
}
func TestTagModel_DeleteSoft(t *testing.T) {
	err:=GetTagModel().DeleteSoft(context.Background(),&entity.Operator{
		UserID: "2",
		Role:   "2",
	},"5f2bcafe55a45129d40d0cc0")
	fmt.Println(err)
}
func TestTagModel_Delete(t *testing.T) {
	err:=GetTagModel().Delete(context.Background(),"5f2bc7213ed54eb1351e4501")
	fmt.Println(err)
}

func TestTagModel_Query(t *testing.T) {
	resut,err:=GetTagModel().Query(context.Background(),&da.TagCondition{
		Name:     entity.NullString{
			Strings: "",
			Valid:   false,
		},
		Pager: utils.Pager{
			PageIndex: 1,
			PageSize:  2,
		},
		DeleteAt: entity.NullInt{
			Int:   0,
			Valid: true,
		},
	})
	if err!=nil{
		fmt.Println(err)
		return
	}

	fmt.Println(len(resut))
}
func TestTagModel_Page(t *testing.T) {
	total,resut,err:=GetTagModel().Page(context.Background(),&da.TagCondition{
		Name:     entity.NullString{
			Strings: "",
			Valid:   false,
		},
		Pager: utils.Pager{
			PageIndex: 1,
			PageSize:  12,
		},
	})
	if err!=nil{
		fmt.Println(err)
		return
	}

	fmt.Println("total:",total)
	for _,item:=range resut{
		fmt.Println(*item)
	}
}

func TestTagModel_GetByIDs(t *testing.T) {
	result,_:=GetTagModel().GetByIDs(context.Background(),[]string{
		"5f2bcafe55a45129d40d0cc0","351d1c9472be37e3233",
	})
	for _,item:=range result{
		fmt.Println(item)
	}
}
