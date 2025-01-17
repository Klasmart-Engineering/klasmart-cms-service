package api

import (
	"net/http"
	"strconv"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/model"
	"github.com/gin-gonic/gin"
)

// @Summary getVisibilitySetting
// @ID getVisibilitySetting
// @Description get visibilitySetting
// @Accept json
// @Produce json
// @Tags visibilitySetting
// @Param content_type query string true "content type"
// @Success 200 {array} entity.VisibilitySetting
// @Success 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /visibility_settings [get]
func (s *Server) getVisibilitySetting(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	contentType := c.Query("content_type")

	contentTypeInt, err := strconv.Atoi(contentType)
	if err != nil {
		log.Error(ctx, "request error", log.Err(err), log.String("contentType", contentType))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	result, err := model.GetVisibilitySettingModel().Query(ctx, contentTypeInt, op)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.String("contentType", contentType), log.Int("contentTypeInt", contentTypeInt))
		s.defaultErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
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
	ctx := c.Request.Context()
	id := c.Param("id")
	op := s.getOperator(c)
	result, err := model.GetVisibilitySettingModel().GetByID(ctx, id, op)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}
