package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary getSubject
// @ID getSubject
// @Description get subjects
// @Accept json
// @Produce json
// @Tags subject
// @Success 200 {array} entity.Subject
// @Failure 500 {object} InternalServerErrorResponse
// @Router /subjects [get]
func (s *Server) getSubject(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}

// @Summary getSubjectByID
// @ID getSubjectByID
// @Description get subjects by id
// @Accept json
// @Produce json
// @Param id path string true "subject id"
// @Tags subject
// @Success 200 {object} entity.Subject
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /subjects/{id} [get]
func (s *Server) getSubjectByID(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}
