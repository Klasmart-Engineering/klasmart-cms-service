package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary getClassType
// @ID getClassType
// @Description get class type
// @Accept json
// @Produce json
// @Tags classType
// @Success 200 {array} entity.ClassType
// @Failure 500 {object} InternalServerErrorResponse
// @Router /class_types [get]
func (s *Server) getClassType(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := model.GetClassTypeModel().Query(ctx, &da.ClassTypeCondition{})
	if err != nil {
		s.jsonInternalServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary getClassTypeByID
// @ID getClassTypeByID
// @Description get classType by id
// @Accept json
// @Produce json
// @Param id path string true "classType id"
// @Tags classType
// @Success 200 {object} entity.ClassType
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /class_types/{id} [get]
func (s *Server) getClassTypeByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	result, err := model.GetClassTypeModel().GetByID(ctx, id)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.jsonInternalServerError(c, err)
	}
}
