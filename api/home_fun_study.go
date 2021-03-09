package api

import "github.com/gin-gonic/gin"

// @Summary list home fun studies
// @Description list home fun studies
// @Tags homeFunStudies
// @ID listHomeFunStudies
// @Accept json
// @Produce json
// @Param query query string false "fuzzy query teacher name and student name"
// @Param status query string false "query status" enums(all,in_progress,complete)
// @Param order_by query string false "list order by" enums(latest_feedback_at,-latest_feedback_at,complete_at,-complete_at) default(-latest_feedback_at)
// @Param page query int false "page number" default(1)
// @Param page_size query integer false "page size" format(int) default(10)
// @Success 200 {object} entity.ListHomeFunStudiesResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /home_fun_studies [get]
func (s *Server) listHomeFunStudies(c *gin.Context) {
	panic("not implemented")
}

// @Summary get home fun study
// @Description get home fun study detail
// @Tags homeFunStudies
// @ID getHomeFunStudy
// @Accept json
// @Produce json
// @Param id path string true "home fun study id"
// @Success 200 {object} entity.ListHomeFunStudiesResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /home_fun_studies/{id} [get]
func (s *Server) getHomeFunStudy(c *gin.Context) {
	panic("not implemented")
}

// @Summary assess home fun study
// @Description assess home fun study
// @Tags homeFunStudies
// @ID assessHomeFunStudy
// @Accept json
// @Produce json
// @Param id path string true "home fun study id"
// @Param assess_home_fun_study_args body entity.AssessHomeFunStudyArgs true "assess home fun study args"
// @Success 200 {string} string "OK"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /home_fun_studies/{id}/assess [put]
func (s *Server) assessHomeFunStudy(c *gin.Context) {
	panic("not implemented")
}
