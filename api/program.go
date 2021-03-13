package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary getProgram
// @ID getProgram
// @Description get program
// @Accept json
// @Produce json
// @Tags program
// @Success 200 {array} entity.Program
// @Failure 500 {object} InternalServerErrorResponse
// @Router /programs [get]
func (s *Server) getProgram(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	result, err := model.GetProgramModel().GetByOrganization(ctx, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusOK, result)
}
