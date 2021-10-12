package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
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

	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.ViewSubjects20115)
	if err != nil {
		log.Error(ctx, "getDevelopmental: HasOrganizationPermission failed",
			log.Any("op", operator),
			log.String("perm", string(external.ViewSubjects20115)),
			log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "getDevelopmental: HasOrganizationPermission failed",
			log.Any("op", operator),
			log.String("perm", string(external.ViewSubjects20115)))
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
		return
	}

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
		s.defaultErrorHandler(c, err)
	}
}
