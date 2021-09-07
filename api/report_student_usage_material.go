package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

// @Summary get student usage of material report
// @Description get student usage of material report
// @Tags reports/studentUsage
// @ID getStudentUsageMaterialReport
// @Accept json
// @Produce json
// @Param class_id_list query []string false "class id list"
// @Param content_type_list query []string false "content type list"
// @Param time_range_list query []string false "time range list"
// @Success 200 {object} entity.StudentUsageMaterialReportResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/material [get]
func (s *Server) getStudentUsageMaterialReport(c *gin.Context) {
	_ = entity.StudentUsageMaterialReportRequest{}
}

// @Summary get student usage of material report
// @Description get student usage of material report
// @Tags reports/studentUsage
// @ID getStudentUsageMaterialViewCountReport
// @Accept json
// @Produce json
// @Param time_range_list query []string false "time range list"
// @Param class_id_list query []string false "class id list"
// @Param content_type_list query []string false "content type list"
// @Success 200 {object} entity.StudentUsageMaterialViewCountReportResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/material_view_count [get]
func (s *Server) getStudentUsageMaterialViewCountReport(c *gin.Context) {
	_ = entity.StudentUsageMaterialViewCountReportRequest{}
}
