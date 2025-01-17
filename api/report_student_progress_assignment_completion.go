package api

import (
	"net/http"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/model"
	"github.com/gin-gonic/gin"
)

// @Summary  getAssignmentsCompletion
// @Tags reports/studentProgress
// @ID getAssignmentsCompletion
// @Accept json
// @Produce json
// @Param request body entity.AssignmentRequest true "request "
// @Success 200 {object} entity.AssignmentResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_progress/assignment_completion [post]
func (s *Server) getAssignmentsCompletion(c *gin.Context) {

	ctx := c.Request.Context()
	op := s.getOperator(c)

	var request entity.AssignmentRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		log.Error(ctx, "getAssignmentsCompletion: ShouldBindQuery failed",
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
	result, err := model.GetReportModel().GetAssignmentCompletion(ctx, op, &request)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}
