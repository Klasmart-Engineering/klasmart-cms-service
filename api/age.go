package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

// @Summary getAge
// @ID getAge
// @Description get age
// @Accept json
// @Produce json
// @Param program_id query string false "program id"
// @Tags age
// @Success 200 {array} external.Age
// @Failure 500 {object} InternalServerErrorResponse
// @Router /ages [get]
func (s *Server) getAge(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	var result []*external.Age
	var err error

	programID := c.Query("program_id")

	if programID != "" {
		result, err = external.GetAgeServiceProvider().GetByOrganization(ctx, operator)
	} else {
		result, err = external.GetAgeServiceProvider().GetByProgram(ctx, operator, programID)
	}

	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
