package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

// @Summary get student usage of material report
// @Description get student usage of material report
// @Tags reports/studentUsage
// @ID getStudentUsageReport
// @Accept json
// @Produce json
// @Param page query integer false "page"
// @Param page_size query integer false "page size"
// @Param start_at query integer false "week start timestamp(unit: second)"
// @Param end_at query integer false "week end timestamp(unit: second)"
// @Param class_id_list query []string false "class id list"
// @Param content_type_list query []string false "content type list"
// @Success 200 {object} entity.StudentUsageMaterialReportResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/material [get]
func (s *Server) getStudentUsageMaterialReport(c *gin.Context) {
	_ = entity.StudentUsageMaterialReportRequest{}
}
