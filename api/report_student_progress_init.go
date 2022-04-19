package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

// @Summary  getProgressInit
// @Tags reports/getProgressInit
// @ID getProgressInit
// @Accept json
// @Produce json
// @Param request body entity.ProgressInitRequest true "request "
// @Success 200 {object} entity.ProgressInitResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_progress/class_attendance [post]
func (s *Server) getProgressInit(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var request entity.ProgressInitRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		log.Error(ctx, "getProgressInit: ShouldBindQuery failed",
			log.Err(err),
			log.Any("request", request))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	err = s.checkPermissionForReportStudentProgress(ctx, op, request.ClassID, request.StudentID)
	if err != nil {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}

	var result entity.ProgressInitResponse
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}
