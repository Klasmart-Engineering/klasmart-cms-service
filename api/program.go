package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary getProgram
// @ID getProgram
// @Description get program
// @Accept json
// @Produce json
// @Tags program
// @Success 200 {array} entity.Program
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /programs [get]
func (s *Server) getProgram(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}

// @Summary getProgramByID
// @ID getProgramByID
// @Description get program by id
// @Accept json
// @Produce json
// @Param id path string true "program id"
// @Tags program
// @Success 200 {object} entity.Program
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /programs/{id} [get]
func (s *Server) getProgramByID(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}
