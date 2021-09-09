package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary get Classes&Assignments Report
// @Description get Classes&Assignments overview
// @Tags reports/studentUsage
// @ID getClassesAssignmentsOverview
// @Accept json
// @Produce json
// @Param class_ids query []string false "class id list"
// @Param durations query []string false "time durations, for example: [startTime1-endTime1, startTime2-endTime2]"
// @Success 200 {object} []entity.ClassesAssignmentOverView
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/classes_assignments_overview [get]
func (s *Server) getClassesAssignmentsOverview(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	var request entity.ClassesAssignmentOverViewRequest
	err := c.ShouldBindQuery(&request)
	if err != nil {
		log.Error(ctx, "getClassesAssignmentsOverview: ShouldBindQuery failed", log.Any("request", request))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	result, err := model.GetClassesAssignmentsModel().GetOverview(ctx, op, &request)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary get Classes&Assignments Report
// @Description get Classes&Assignments Report
// @Tags reports/studentUsage
// @ID getClassesAssignments
// @Accept json
// @Produce json
// @Param page query integer false "page"
// @Param page_size query integer false "page size"
// @Param class_ids query []string false "class id list"
// @Param durations query []string false "time durations, for example: [startTime1-endTime1, startTime2-endTime2]"
// @Success 200 {object} []entity.ClassesAssignmentsView
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/classes_assignments [get]
func (s *Server) getClassesAssignments(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	var request entity.ClassesAssignmentsViewRequest
	err := c.ShouldBindQuery(&request)
	if err != nil {
		log.Error(ctx, "getClassesAssignments: ShouldBindQuery failed", log.Any("request", request))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	result, err := model.GetClassesAssignmentsModel().GetStatistic(ctx, op, &request)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary get Classes&Assignments Report
// @Description get Classes&Assignments unattended
// @Tags reports/studentUsage
// @ID getClassesAssignmentsUnattended
// @Accept json
// @Produce json
// @Param page query integer false "page"
// @Param page_size query integer false "page size"
// @Param durations query []string false "time durations, for example: [startTime1-endTime1, startTime2-endTime2]"
// @Success 200 {object} []entity.ClassesAssignmentsUnattendedStudentsView
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/classes_assignments/{class_id}/unattended [get]
func (s *Server) getClassesAssignmentsUnattended(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	var request entity.ClassesAssignmentsUnattendedViewRequest
	err := c.ShouldBindQuery(&request)
	if err != nil {
		log.Error(ctx, "getClassesAssignmentsUnattended: ShouldBindQuery failed", log.Any("request", request))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	result, err := model.GetClassesAssignmentsModel().GetUnattended(ctx, op, &request)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}
