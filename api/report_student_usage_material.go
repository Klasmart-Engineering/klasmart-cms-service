package api

import (
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

// @Summary get student usage of material report
// @Description get student usage of material report
// @Tags reports/studentUsage
// @ID getStudentUsageMaterialReport
// @Accept json
// @Produce json
// @Param request query entity.StudentUsageMaterialReportRequest true "request"
// @Success 200 {object} entity.StudentUsageMaterialReportResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/material [post]
func (s *Server) getStudentUsageMaterialReport(c *gin.Context) {
	ctx := c.Request.Context()
	var err error
	defer func() {
		if err == nil {
			return
		}
		switch err {
		case constant.ErrInvalidArgs:
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		default:
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		}
	}()
	op := s.getOperator(c)
	req := entity.StudentUsageMaterialReportRequest{}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		log.Error(ctx, "invalid request", log.Err(err))
		err = constant.ErrInvalidArgs
		return
	}
	res, err := model.GetReportModel().GetStudentUsageMaterial(ctx, op, &req)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary get student usage of material report
// @Description get student usage of material report
// @Tags reports/studentUsage
// @ID getStudentUsageMaterialViewCountReport
// @Accept json
// @Produce json
// @Param request query entity.StudentUsageMaterialViewCountReportRequest true "request"
// @Success 200 {object} entity.StudentUsageMaterialViewCountReportResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/material_view_count [post]
func (s *Server) getStudentUsageMaterialViewCountReport(c *gin.Context) {
	ctx := c.Request.Context()
	var err error
	defer func() {
		if err == nil {
			return
		}
		switch err {
		case constant.ErrInvalidArgs:
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		default:
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		}
	}()
	op := s.getOperator(c)
	req := entity.StudentUsageMaterialViewCountReportRequest{}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		log.Error(ctx, "invalid request", log.Err(err))
		err = constant.ErrInvalidArgs
		return
	}
	res, err := model.GetReportModel().GetStudentUsageMaterialViewCount(ctx, op, &req)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, res)
}
