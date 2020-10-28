package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary getLessonType
// @ID getLessonType
// @Description get lessonType
// @Accept json
// @Produce json
// @Tags lessonType
// @Success 200 {array} entity.LessonType
// @Failure 500 {object} InternalServerErrorResponse
// @Router /lesson_types [get]
func (s *Server) getLessonType(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}

// @Summary getLessonTypeByID
// @ID getLessonTypeByID
// @Description get lessonType by id
// @Accept json
// @Produce json
// @Param id path string true "lessonType id"
// @Tags lessonType
// @Success 200 {object} entity.LessonType
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /lesson_types/{id} [get]
func (s *Server) getLessonTypeByID(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}
