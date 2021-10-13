package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary getProgram
// @ID getProgram
// @Description get program
// @Accept json
// @Produce json
// @Tags program
// @Success 200 {array} external.Program
// @Failure 500 {object} InternalServerErrorResponse
// @Router /programs [get]
func (s *Server) getProgram(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.ViewProgram20111)
	if err != nil {
		log.Error(ctx, "getProgram: HasOrganizationPermission failed",
			log.Any("op", op),
			log.String("perm", string(external.ViewProgram20111)),
			log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "getProgram: HasOrganizationPermission failed",
			log.Any("op", op),
			log.String("perm", string(external.ViewProgram20111)))
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
		return
	}

	result, err := model.GetProgramModel().GetByOrganization(ctx, op)
	if err != nil {
		s.defaultErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
