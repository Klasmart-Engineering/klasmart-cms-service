package api

import (
	"github.com/gin-gonic/gin"
)

// @Summary list h5p assessments
// @Description list h5p assessments
// @Tags h5pAssessments
// @ID listH5PAssessments
// @Accept json
// @Produce json
// @Param type query string false "h5p assessment type" enums(study_h5p,class_and_live_h5p)
// @Param query query string false "query teacher name or class name"
// @Param query_type query string false "query type" enums(all,class_name,teacher_name) default(all)
// @Param status query string false "query status" enums(all,in_progress,complete) default(all)
// @Param order_by query string false "list order by" enums(create_at,-create_at,complete_at,-complete_at) default(-complete_at)
// @Param page query int false "page number" default(1)
// @Param page_size query integer false "page size" format(int) default(10)
// @Success 200 {object} entity.ListH5PAssessmentsResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /h5p_assessments [get]
func (s *Server) listH5PAssessments(c *gin.Context) {
	panic("not implemented")
}

// @Summary get h5p assessment detail
// @Description get h5p assessment detail
// @Tags h5pAssessments
// @ID getH5PAssessmentDetail
// @Accept json
// @Produce json
// @Param id path string true "h5p assessment id"
// @Success 200 {object} entity.GetH5PAssessmentDetailResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /h5p_assessments/{id} [get]
func (s *Server) getH5PAssessmentDetail(c *gin.Context) {
	panic("not implemented")
}

// @Summary
// @Description update h5p assessment
// @Tags h5pAssessments
// @ID updateH5PAssessment
// @Accept json
// @Produce json
// @Param id path string true "h5p assessment id"
// @Param update_h5p_assessment_args body entity.UpdateH5PAssessmentArgs true "update h5p assessment args"
// @Success 200 {string} string "OK"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /h5p_assessments/{id}/update [put]
func (s *Server) updateH5PAssessment(c *gin.Context) {
	panic("not implemented")
}
