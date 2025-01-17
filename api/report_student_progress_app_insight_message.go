package api

import (
	"net/http"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/model"
	"github.com/gin-gonic/gin"
)

// @Summary  getAppInsightMessage
// @Tags reports/studentProgress
// @ID getAppInsightMessage
// @Accept json
// @Produce json
// @Param class_id query string true "class_id"
// @Param student_id query string true "student_id"
// @Param org_id query string true "org_id"
// @Param end_time query int true "end_time" default(0)
// @Success 200 {object} entity.AppInsightMessageResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_progress/app/insight_message [get]
func (s *Server) getAppInsightMessage(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var request entity.AppInsightMessageRequest
	err := c.ShouldBindQuery(&request)
	if err != nil || request.ClassID == "" || request.StudentID == "" || request.OrgID == "" {
		log.Warn(ctx, "getAppInsightMessage: ShouldBindQuery failed",
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
	result, err := model.GetReportModel().GetAppInsightMessage(ctx, op, &request)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}
