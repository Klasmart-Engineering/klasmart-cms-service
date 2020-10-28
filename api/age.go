package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary getAge
// @ID getAge
// @Description get age
// @Accept json
// @Produce json
// @Tags age
// @Success 200 {array} entity.Age
// @Failure 500 {object} InternalServerErrorResponse
// @Router /ages [get]
func (s *Server) getAge(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}

// @Summary getAgeByID
// @ID getAgeByID
// @Description get age by id
// @Accept json
// @Produce json
// @Param id path string true "age id"
// @Tags age
// @Success 200 {object} entity.Age
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /ages/{id} [get]
func (s *Server) getAgeByID(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}
