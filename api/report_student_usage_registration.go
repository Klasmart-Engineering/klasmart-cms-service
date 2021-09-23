package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary get student usage of organization registration report
// @Description get student usage of organization registration report
// @Tags reports/studentUsage
// @ID getStudentUsageOrganizationRegistrationReport
// @Accept json
// @Produce json
// @Param organization_id query string true "organization id"
// @Success 200 {object} entity.StudentUsageRegistration
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/organization_registration [get]
func (s *Server) getStudentUsageOrganizationRegistration(c *gin.Context) {
	// TODO
	c.JSON(http.StatusOK, nil)
}

// @Summary get student usage of class registration report
// @Description get student usage of class registration report
// @Tags reports/studentUsage
// @ID getStudentUsageClassRegistrationReport
// @Accept json
// @Produce json
// @Param organization_id query string true "organization id"
// @Param class_id_list query []string false "class id list"
// @Param time_range_list query []string false "time range list"
// @Success 200 {array} entity.StudentUsageRegistration
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/class_registration [get]
func (s *Server) getStudentUsageClassRegistration(c *gin.Context) {
	// TODO
	c.JSON(http.StatusOK, nil)
}
