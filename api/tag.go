package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"net/http"
	"strings"
)

func (s Server) addTag(c *gin.Context) {
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusBadRequest, responseMsg("operate not exist"))
		return
	}
	ctx := c.Request.Context()
	data := new(entity.TagAddView)
	err := c.ShouldBindJSON(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		log.Info(ctx, "bind json data error", log.Err(err))
		return
	}
	if strings.TrimSpace(data.Name) == "" {
		c.JSON(http.StatusBadRequest, errors.New("tag name is empty"))
		log.Info(ctx, "tag name is empty")
		return
	}

	id, err := model.GetTagModel().Add(ctx, op, data)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"id": id,
		})
		return
	}
	if err == constant.ErrDuplicateRecord {
		c.JSON(http.StatusConflict, err.Error())
		return
	}

	c.JSON(http.StatusInternalServerError, err.Error())
}

func (s Server) delTag(c *gin.Context) {
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusBadRequest, responseMsg("operate not exist"))
		return
	}
	ctx := c.Request.Context()
	err := model.GetTagModel().DeleteSoft(ctx, op, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, nil)
}

func (s Server) updateTag(c *gin.Context) {
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusBadRequest, responseMsg("operate not exist"))
		return
	}
	ctx := c.Request.Context()
	ID := c.Param("id")
	data := new(entity.TagUpdateView)
	err := c.ShouldBindJSON(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		log.Info(ctx, "bind json data error", log.Err(err))
		return
	}
	data.ID = ID
	if strings.TrimSpace(data.Name) == "" {
		c.JSON(http.StatusBadRequest, errors.New("tag name is empty"))
		log.Info(ctx, "tag name is empty")
		return
	}
	if data.States != constant.Enable || data.States != constant.Disabled {
		c.JSON(http.StatusBadRequest, errors.New("tag states is invalid"))
		log.Info(ctx, "tag states is invalid")
		return
	}
	err = model.GetTagModel().Update(ctx, op, data)

	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"id": data.ID,
		})
		return
	}
	if err == constant.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}
	if err == constant.ErrDuplicateRecord {
		c.JSON(http.StatusConflict, err.Error())
		return
	}

	c.JSON(http.StatusInternalServerError, err.Error())
}

func (s Server) getTagByID(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := model.GetTagModel().GetByID(ctx, c.Param("id"))

	if err == nil {
		c.JSON(http.StatusOK, result)
		return
	}
	if err == constant.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusInternalServerError, err.Error())
}

func (s Server) queryTag(c *gin.Context) {
	ctx := c.Request.Context()
	condition := new(da.TagCondition)
	condition.Pager = utils.GetPager(c.Query("page"), c.Query("page_size"))
	name := c.Query("name")
	condition.Name = entity.NullString{
		String: name,
		Valid:  len(name) != 0,
	}
	var (
		total  int64
		result []*entity.TagView
		err    error
	)
	if condition.Pager.Page == 0 || condition.Pager.PageSize == 0 {
		result, err = model.GetTagModel().Query(ctx, condition)
		total = int64(len(result))
	} else {
		total, result, err = model.GetTagModel().Page(ctx, condition)
	}

	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"total": total,
			"data":  result,
		})
		return
	}
	if err == constant.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusInternalServerError, err.Error())
}
