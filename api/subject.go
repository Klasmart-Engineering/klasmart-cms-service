package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

// @Summary getSubject
// @ID getSubject
// @Description get subjects
// @Accept json
// @Produce json
// @Param program_id query string false "program id"
// @Tags subject
// @Success 200 {array} external.Subject
// @Failure 500 {object} InternalServerErrorResponse
// @Router /subjects [get]
func (s *Server) getSubject(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	var result []*external.Subject
	var err error

	programID := c.Query("program_id")

	if programID == "" {
		result, err = external.GetSubjectServiceProvider().GetByOrganization(ctx, operator)
	} else {
		result, err = external.GetSubjectServiceProvider().GetByProgram(ctx, operator, programID)
	}

	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
