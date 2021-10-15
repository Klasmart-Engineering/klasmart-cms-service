package api

import "github.com/gin-gonic/gin"

// @Summary  getLearnOutcomeAchievement
// @Tags reports/studentProgress
// @ID getLearnOutcomeAchievement
// @Accept json
// @Produce json
// @Param request body entity.LearnOutcomeAchievementRequest true "request "
// @Success 200 {object} entity.LearnOutcomeAchievementResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_progress/learn_outcome_achievements [post]
func (s *Server) getLearnOutcomeAchievement(c *gin.Context) {

}
