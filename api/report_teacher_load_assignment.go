package api

import (
	"net/http"

	"github.com/KL-Engineering/kidsloop-cms-service/model"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/gin-gonic/gin"
)

// @Summary list teaching load report
// @Description list teaching load report
// @Tags reports/teacherLoad
// @ID getTeacherLoadReportOfAssignment
// @Accept json
// @Produce json
// @Param request body entity.TeacherLoadAssignmentRequest true "request "
// @Success 200 {object} []entity.TeacherLoadAssignmentResponseItem
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/teacher_load/assignments [post]
func (s *Server) getTeacherLoadReportOfAssignment(c *gin.Context) {
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
			s.defaultErrorHandler(c, err)
		}
	}()
	op := s.getOperator(c)
	req := entity.TeacherLoadAssignmentRequest{}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		log.Error(ctx, "invalid request", log.Err(err))
		err = constant.ErrInvalidArgs
		return
	}
	err = req.ClassTypeList.Validate(ctx)
	if err != nil {
		return
	}
	res, err := model.GetReportModel().GetTeacherLoadReportOfAssignment(ctx, op, &req)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, res)
}
