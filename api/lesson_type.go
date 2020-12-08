package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
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
	ctx := c.Request.Context()
	result, err := model.GetLessonTypeModel().Query(ctx, &da.LessonTypeCondition{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusOK, result)
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
	ctx := c.Request.Context()
	id := c.Param("id")
	result, err := model.GetLessonTypeModel().GetByID(ctx, id)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
