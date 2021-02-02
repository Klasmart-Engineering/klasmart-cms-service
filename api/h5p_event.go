package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
)

// @Summary createH5PEvent
// @ID createH5PEvent
// @Description create h5p event
// @Accept json
// @Produce json
// @Param content body entity.CreateH5pEventRequest true "create request"
// @Tags h5p
// @Success 200 {object} CreateContentResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /h5p/events [post]
func (s *Server) createH5PEvent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	data := new(entity.CreateH5pEventRequest)
	err := c.ShouldBind(&data)
	if err != nil {
		log.Error(ctx, "create h5p event failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	err = model.GetH5PEventModel().CreateH5PEvent(ctx, dbo.MustGetDB(ctx), data, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, gin.H{})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
