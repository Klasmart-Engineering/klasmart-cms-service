package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
	"strconv"
)

func (s *Server) createCategory(c *gin.Context) {
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusBadRequest, responseMsg("operate not exist"))
		return
	}
	data := new(entity.CategoryObject)
	err := c.ShouldBind(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}

	category, err := model.GetCategoryModel().CreateCategory(c.Request.Context(), op, *data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, category)
}
func (s *Server) updateCategory(c *gin.Context) {
	id := c.Param("id")

	data := new(entity.CategoryObject)
	err := c.ShouldBind(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}
	data.ID = id

	err = model.GetCategoryModel().UpdateCategory(c.Request.Context(), nil, *data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, data)
}
func (s *Server) deleteCategory(c *gin.Context) {
	id := c.Param("id")
	err := model.GetCategoryModel().DeleteCategory(c.Request.Context(), nil, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, responseMsg("success"))
}

func (s *Server) getCategoryByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, "id illegal")
		return
	}
	category, err := model.GetCategoryModel().GetCategoryByID(c.Request.Context(), nil, id)
	if err != nil && err != constant.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	} else if err == constant.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, responseMsg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, category)
}
func (s *Server) searchCategories(c *gin.Context) {
	data := buildCategorySearchCondition(c)
	total, categories, err := model.GetCategoryModel().PageCategories(c.Request.Context(), nil, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"list":  categories,
	})
}

func buildCategorySearchCondition(c *gin.Context) *entity.SearchCategoryCondition {
	ids := c.QueryArray("ids")
	names := c.QueryArray("names")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))

	data := &entity.SearchCategoryCondition{
		IDs:      entity.NullStrings{Strings: ids, Valid: len(ids) > 0},
		Names:    entity.NullStrings{Strings: names, Valid: len(names) > 0},
		PageSize: int64(pageSize),
		Page:     int64(page),
		//OrderBy: "",
	}

	return data
}
