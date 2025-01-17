package api

import (
	"net/http"

	"github.com/KL-Engineering/kidsloop-cms-service/external"
	"github.com/gin-gonic/gin"
)

// @Summary getGrade
// @ID getGrade
// @Description get grade
// @Accept json
// @Produce json
// @Param program_id query string false "program id"
// @Tags grade
// @Success 200 {array} external.Grade
// @Failure 500 {object} InternalServerErrorResponse
// @Router /grades [get]
func (s *Server) getGrade(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	var result []*external.Grade
	var err error

	programID := c.Query("program_id")

	if programID == "" {
		result, err = external.GetGradeServiceProvider().GetByOrganization(ctx, operator)
	} else {
		result, err = external.GetGradeServiceProvider().GetByProgram(ctx, operator, programID)
	}

	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}
