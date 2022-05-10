package api

import (
	"net/http"

	"github.com/KL-Engineering/kidsloop-cms-service/external"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/model"
	"github.com/gin-gonic/gin"
)

// @Summary get Classes&Assignments Report
// @Description get Classes&Assignments overview
// @Tags reports/studentUsage
// @ID getClassesAssignmentsOverview
// @Accept json
// @Produce json
// @Param overview body entity.ClassesAssignmentOverViewRequest true "overview"
// @Success 200 {object} []entity.ClassesAssignmentOverView
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/classes_assignments_overview [post]
func (s *Server) getClassesAssignmentsOverview(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	var request entity.ClassesAssignmentOverViewRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		log.Error(ctx, "getClassesAssignmentsOverview: ShouldBindQuery failed",
			log.Err(err),
			log.Any("request", request))
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
// @Param classes_assignments body entity.ClassesAssignmentsViewRequest true "classAssignments"
// @Success 200 {object} []entity.ClassesAssignmentsView
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/classes_assignments [post]
func (s *Server) getClassesAssignments(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	var request entity.ClassesAssignmentsViewRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		log.Error(ctx, "getClassesAssignments: ShouldBindQuery failed",
			log.Err(err),
			log.Any("request", request))
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
// @Param unattended body entity.ClassesAssignmentsUnattendedViewRequest true "unattended"
// @Success 200 {object} []entity.ClassesAssignmentsUnattendedStudentsView
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/classes_assignments/{class_id}/unattended [post]
func (s *Server) getClassesAssignmentsUnattended(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	var request entity.ClassesAssignmentsUnattendedViewRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		log.Error(ctx, "getClassesAssignmentsUnattended: ShouldBindQuery failed",
			log.Err(err),
			log.Any("request", request))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	request.ClassID = c.Param("class_id")
	result, err := model.GetClassesAssignmentsModel().GetUnattended(ctx, op, &request)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary get learner usage Report
// @Description get learner usage Report
// @Tags reports/learnerUsage
// @ID getLearnerUsageOverview
// @Accept json
// @Produce json
// @Param classes_assignments body entity.LearnerUsageRequest true "classAssignments"
// @Success 200 {object} entity.LearnerUsageResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/learner_usage/overview [post]
func (s *Server) getLearnerUsageOverview(c *gin.Context) {

	ctx := c.Request.Context()
	op := s.getOperator(c)

	var request entity.LearnerUsageRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		log.Error(ctx, "getLearnerUsageOverview: ShouldBindJSON failed",
			log.Err(err),
			log.Any("request", request))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	permissions, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, []external.PermissionName{
		external.ReportOrganizationStudentUsage,
		external.ReportSchoolStudentUsage,
		external.ReportTeacherStudentUsage,
	})

	result, err := model.GetReportModel().GetLearnerUsageOverview(ctx, op, permissions, &request)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}
