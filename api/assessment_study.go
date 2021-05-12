package api

import (
	"github.com/gin-gonic/gin"
)

// @Summary list studies
// @Description list studies
// @Tags studies
// @ID listStudies
// @Accept json
// @Produce json
// @Param query query string false "query teacher name and class name"
// @Param status query string false "query status" enums(all,in_progress,complete) default(all)
// @Param order_by query string false "list order by" enums(create_at,-create_at,complete_at,-complete_at) default(-create_at)
// @Param page query int false "page number" default(1)
// @Param page_size query integer false "page size" format(int) default(10)
// @Success 200 {object} entity.ListStudiesResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /studies [get]
func (s *Server) listStudies(c *gin.Context) {
	panic("not implemented")
}

// @Summary get study detail
// @Description get study detail
// @Tags studies
// @ID getStudyDetail
// @Accept json
// @Produce json
// @Param id path string true "study id"
// @Success 200 {object} entity.GetStudyDetailResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /studies/{id} [get]
func (s *Server) getStudyDetail(c *gin.Context) {
	panic("not implemented")
}
