package api

import (
	"github.com/gin-gonic/gin"
)

// @Summary list content assessments
// @Description list content assessments
// @Tags contentAssessments
// @ID listContentAssessments
// @Accept json
// @Produce json
// @Param type query string false "content assessment type" enums(study,class_and_live)
// @Param query query string false "query teacher name or class name"
// @Param query_type query string false "query type" enums(class_name,teacher_name)
// @Param status query string false "query status" enums(in_progress,complete)
// @Param order_by query string false "list order by" enums(create_at,-create_at,complete_at,-complete_at) default(-complete_at)
// @Param page query int false "page number" default(1)
// @Param page_size query integer false "page size" format(int) default(10)
// @Success 200 {object} entity.ListContentAssessmentsResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /content_assessments [get]
func (s *Server) listContentAssessments(c *gin.Context) {
	panic("not implemented")
}

// @Summary get content assessment detail
// @Description get content assessment detail
// @Tags contentAssessments
// @ID getContentAssessmentDetail
// @Accept json
// @Produce json
// @Param id path string true "content assessment id"
// @Success 200 {object} entity.GetContentAssessmentDetailResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /content_assessments/{id} [get]
func (s *Server) getContentAssessmentDetail(c *gin.Context) {
	panic("not implemented")
}

// @Summary
// @Description update content assessment
// @Tags contentAssessments
// @ID updateContentAssessment
// @Accept json
// @Produce json
// @Param id path string true "content assessment id"
// @Param update_content_assessment_args body entity.UpdateContentAssessmentArgs true "update content assessment args"
// @Success 200 {string} string "OK"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /content_assessments/{id}/update [put]
func (s *Server) updateContentAssessment(c *gin.Context) {
	panic("not implemented")
}
