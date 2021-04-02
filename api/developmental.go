package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

// @Summary getDevelopmental
// @ID getDevelopmental
// @Description get developmental
// @Accept json
// @Produce json
// @Param program_id query string false "program id"
// @Param subject_ids query string false "subject ids,separated by comma"
// @Tags developmental
// @Success 200 {array} external.Category
// @Failure 500 {object} InternalServerErrorResponse
// @Router /developmentals [get]
func (s *Server) getDevelopmental(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	var result []*external.Category
	var err error

	programID := c.Query("program_id")
	subjectIDQuery := c.Query("subject_ids")

	if subjectIDQuery == "" {
		if programID == "" {
			result, err = external.GetCategoryServiceProvider().GetByOrganization(ctx, operator)
		} else {
			result, err = external.GetCategoryServiceProvider().GetByProgram(ctx, operator, programID)
		}
	} else {
		subjectIDs := strings.Split(subjectIDQuery, ",")
		result, err = external.GetCategoryServiceProvider().GetBySubjects(ctx, operator, subjectIDs)
	}

	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
