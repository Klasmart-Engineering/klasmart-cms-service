package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) addTag(c *gin.Context){
	ctx:=c.Request.Context()
	data:=new(entity.TagAddView)
	err:=c.ShouldBindJSON(data)
	if err!=nil{
		c.JSON(http.StatusBadRequest,err.Error())
		log.Error(ctx,"bind json data error",log.Err(err))
		return
	}
	if strings.TrimSpace(data.Name)==""{
		c.JSON(http.StatusBadRequest,errors.New("tag name is empty"))
		log.Error(ctx,"tag name is empty")
		return
	}

	status:=http.StatusOK
	id,err:=model.GetTagModel().Add(ctx,data)
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
	ctx:=c.Request.Context()
	err:=model.GetTagModel().Delete(ctx,c.Param("id"))
	if err!=nil{
		c.JSON(http.StatusInternalServerError,err.Error())
		return
	}
	c.JSON(http.StatusOK,nil)
}

func (s Server) updateTag(c *gin.Context) {
	ctx:=c.Request.Context()
	id:=c.Param("id")
	data:=new(entity.TagUpdateView)
	err:=c.ShouldBindJSON(data)
	if err!=nil{
		c.JSON(http.StatusBadRequest,err.Error())
		log.Error(ctx,"bind json data error",log.Err(err))
		return
	}
	data.ID = id
	if strings.TrimSpace(data.Name)==""{
		c.JSON(http.StatusBadRequest,errors.New("tag name is empty"))
		log.Error(ctx,"tag name is empty")
		return
	}
	if data.States!=constant.Enable || data.States!=constant.Disabled{
		c.JSON(http.StatusBadRequest,errors.New("tag states is invalid"))
		log.Error(ctx,"tag states is invalid")
		return
	}
	err=model.GetTagModel().Update(ctx,data)

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
	ctx:=c.Request.Context()
	result,err:=model.GetTagModel().GetByID(ctx,c.Param("id"))
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
		condition.PageSize = constant.DefaultPageSize
	}
	pageIndex, err := strconv.ParseInt(c.Query("page_size"), 10, 64)
	if err!=nil{
		condition.Page = constant.DefaultPageIndex
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
