package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary getDevelopmental
// @ID getDevelopmental
// @Description get developmental
// @Accept json
// @Produce json
// @Tags developmental
// @Success 200 {array} entity.Developmental
// @Failure 500 {object} InternalServerErrorResponse
// @Router /developmentals [get]
func (s *Server) getDevelopmental(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}

// @Summary getDevelopmentalByID
// @ID getDevelopmentalByID
// @Description get developmental by id
// @Accept json
// @Produce json
// @Param id path string true "developmental id"
// @Tags developmental
// @Success 200 {object} entity.Developmental
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /developmentals/{id} [get]
func (s *Server) getDevelopmentalByID(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}
