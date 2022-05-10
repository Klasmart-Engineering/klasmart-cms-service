package api

import (
	"net/http"

	"github.com/KL-Engineering/kidsloop-cms-service/model"
	"github.com/gin-gonic/gin"
)

// @Summary getProgramGroup
// @ID getProgramGroup
// @Description get program groups
// @Accept json
// @Produce json
// @Tags program
// @Success 200 {array}  string
// @Failure 500 {object} InternalServerErrorResponse
// @Router /programs_groups [get]
func (s *Server) getProgramGroup(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := model.GetProgramGroupModel().AllGroupNames(ctx)
	if err != nil {
		s.defaultErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
