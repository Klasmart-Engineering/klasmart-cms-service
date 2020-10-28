package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary getVisibilitySetting
// @ID getVisibilitySetting
// @Description get visibilitySetting
// @Accept json
// @Produce json
// @Tags visibilitySetting
// @Success 200 {array} entity.VisibilitySetting
// @Failure 500 {object} InternalServerErrorResponse
// @Router /visibility_settings [get]
func (s *Server) getVisibilitySetting(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}

// @Summary getVisibilitySettingByID
// @ID getVisibilitySettingByID
// @Description get visibilitySetting by id
// @Accept json
// @Produce json
// @Param id path string true "visibilitySetting id"
// @Tags visibilitySetting
// @Success 200 {object} entity.VisibilitySetting
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /visibility_settings/{id} [get]
func (s *Server) getVisibilitySettingByID(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}
