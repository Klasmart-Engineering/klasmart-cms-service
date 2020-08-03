package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
	"strconv"
)

func (s *Server) createCategory(c *gin.Context) {
	data := new(entity.CategoryObject)
	err := c.ShouldBind(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}

	category, err := model.GetCategoryModel().CreateCategory(c.Request.Context(), *data)
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

	err = model.GetCategoryModel().UpdateCategory(c.Request.Context(), *data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, data)
}
func (s *Server) deleteCategory(c *gin.Context) {
	id := c.Param("id")
	err := model.GetCategoryModel().DeleteCategory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, responseMsg("success"))
}

func (s *Server) getCategoryByID(c *gin.Context) {
	id := c.Param("id")
	category, err := model.GetCategoryModel().GetCategoryById(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, category)
}
func (s *Server) searchCategories(c *gin.Context) {
	data := buildCategorySearchCondition(c)
	total, categories, err := model.GetCategoryModel().PageCategories(c.Request.Context(), data)
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
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))

	// TODO: get ids names
	data := &entity.SearchCategoryCondition{
		IDs:      []string{},
		Names:    []string{},
		PageSize: int64(pageSize),
		Page:     int64(page),
		//OrderBy: "",
	}

	return data
}
