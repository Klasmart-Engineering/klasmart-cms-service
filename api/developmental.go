package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/basicdata"
	"net/http"
)

// @Summary getDevelopmental
// @ID getDevelopmental
// @Description get developmental
// @Accept json
// @Produce json
// @Tags developmental
// @Success 200 {array} entity.Developmental
// @Failure 500 {object} InternalServerErrorResponse
// @Router /developmentals [get]
func (s *Server) getDevelopmental(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := basicdata.GetDevelopmentalModel().Query(ctx, &da.DevelopmentalCondition{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(Unknown))
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary getDevelopmentalByID
// @ID getDevelopmentalByID
// @Description get developmental by id
// @Accept json
// @Produce json
// @Param id path string true "developmental id"
// @Tags developmental
// @Success 200 {object} entity.Developmental
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /developmentals/{id} [get]
func (s *Server) getDevelopmentalByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	result, err := basicdata.GetDevelopmentalModel().GetByID(ctx, id)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(Unknown))
	}
}
