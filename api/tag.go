package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
	"strconv"
)

func (s Server) addTag(c *gin.Context){
	data:=new(entity.TagAddView)
	err:=c.ShouldBindJSON(data)
	if err!=nil{
		c.JSON(http.StatusBadRequest,err.Error())
		return
	}
	status:=http.StatusOK

	id,err:=model.GetTagModel().Add(c.Request.Context(),data)
	if err!=nil{
		status = http.StatusInternalServerError
		if err == constant.ErrDuplicateRecord{
			status = http.StatusConflict
		}
		c.JSON(status,err.Error())
		return
	}
	c.JSON(status,gin.H{
		"id":id,
	})
}

func (s Server) delTag(c *gin.Context){
	err:=model.GetTagModel().Delete(c.Request.Context(),c.Param("id"))
	if err!=nil{
		c.JSON(http.StatusInternalServerError,err.Error())
		return
	}
	c.JSON(http.StatusOK,nil)
}

func (s Server) updateTag(c *gin.Context) {
	id:=c.Param("id")
	data:=new(entity.TagUpdateView)
	err:=c.ShouldBindJSON(data)
	if err!=nil{
		c.JSON(http.StatusBadRequest,err.Error())
		return
	}
	data.ID = id
	err=model.GetTagModel().Update(c.Request.Context(),data)

	status:=http.StatusOK
	if err!=nil{
		status = http.StatusInternalServerError
		if err == constant.ErrRecordNotFound{
			status = http.StatusNotFound
		}
		if err == constant.ErrDuplicateRecord{
			status = http.StatusConflict
		}
		c.JSON(status,err.Error())
		return
	}

	c.JSON(status,gin.H{
		"id":data.ID,
	})
}

func (s Server) getTagByID(c *gin.Context){
	result,err:=model.GetTagModel().GetByID(c.Request.Context(),c.Param("id"))
	status:=http.StatusOK
	if err!=nil{
		status = http.StatusInternalServerError
		if err == constant.ErrRecordNotFound{
			status = http.StatusNotFound
		}
		c.JSON(status,err.Error())
		return
	}
	c.JSON(status,result)
}

func (s Server) queryTag(c *gin.Context){
	condition:=da.TagCondition{}
	pageSize, err := strconv.ParseInt(c.Query("page_size"), 10, 64)
	if err!=nil{

	}
	pageIndex, err := strconv.ParseInt(c.Query("page_size"), 10, 64)
	if err!=nil{

	}
	condition.PageSize = pageSize
	condition.Page = pageIndex
	condition.Name = c.Query("name")

	total,result,err:=model.GetTagModel().Page(c.Request.Context(),condition)
	status:=http.StatusOK
	if err!=nil{
		status = http.StatusInternalServerError
		if err == constant.ErrRecordNotFound{
			status = http.StatusNotFound
		}
		c.JSON(status,err.Error())
		return
	}
	c.JSON(status,gin.H{
		"total":total,
		"data":result,
	})
}
