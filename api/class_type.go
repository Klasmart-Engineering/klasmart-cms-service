package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary getClassType
// @ID getClassType
// @Description get class type
// @Accept json
// @Produce json
// @Tags classType
// @Success 200 {array} entity.ClassType
// @Failure 500 {object} InternalServerErrorResponse
// @Router /class_types [get]
func (s *Server) getClassType(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}

// @Summary getClassTypeByID
// @ID getClassTypeByID
// @Description get classType by id
// @Accept json
// @Produce json
// @Param id path string true "classType id"
// @Tags classType
// @Success 200 {object} entity.ClassType
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /class_types/{id} [get]
func (s *Server) getClassTypeByID(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}
