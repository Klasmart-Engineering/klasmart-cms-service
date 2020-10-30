package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
)

// @Summary getGrade
// @ID getGrade
// @Description get grade
// @Accept json
// @Produce json
// @Tags grade
// @Success 200 {array} entity.Grade
// @Failure 500 {object} InternalServerErrorResponse
// @Router /grades [get]
func (s *Server) getGrade(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := model.GetGradeModel().Query(ctx, &da.GradeCondition{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(Unknown))
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary getGradeByID
// @ID getGradeByID
// @Description get grade by id
// @Accept json
// @Produce json
// @Param id path string true "grade id"
// @Tags grade
// @Success 200 {object} entity.Grade
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /grades/{id} [get]
func (s *Server) getGradeByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	result, err := model.GetGradeModel().GetByID(ctx, id)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(Unknown))
	}
}
