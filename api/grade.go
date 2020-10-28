package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary getGrade
// @ID getGrade
// @Description get grade
// @Accept json
// @Produce json
// @Tags grade
// @Success 200 {array} entity.Grade
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /grades [get]
func (s *Server) getGrade(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}

// @Summary getGradeByID
// @ID getGradeByID
// @Description get grade by id
// @Accept json
// @Produce json
// @Param id path string true "grade id"
// @Tags grade
// @Success 200 {object} entity.Grade
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /grades/{id} [get]
func (s *Server) getGradeByID(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}
