package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary getSkill
// @ID getSkill
// @Description get skill
// @Accept json
// @Produce json
// @Param developmental_id query string false "developmental id"
// @Tags skill
// @Success 200 {array} entity.Skill
// @Failure 500 {object} InternalServerErrorResponse
// @Router /skills [get]
func (s *Server) getSkill(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}

// @Summary getSkillByID
// @ID getSkillByID
// @Description get skill by id
// @Accept json
// @Produce json
// @Param id path string true "skill id"
// @Tags skill
// @Success 200 {object} entity.Skill
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /skills/{id} [get]
func (s *Server) getSkillByID(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}
