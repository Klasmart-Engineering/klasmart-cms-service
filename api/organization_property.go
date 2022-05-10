package api

import (
	"net/http"

	"github.com/KL-Engineering/kidsloop-cms-service/model"
	"github.com/gin-gonic/gin"
)

// @Summary getOrganizationPropertyByID
// @ID getOrganizationPropertyByID
// @Description get organization property by id
// @Accept json
// @Produce json
// @Param id path string true "organization id"
// @Tags organizationProperty
// @Success 200 {object} entity.OrganizationProperty
// @Failure 500 {object} InternalServerErrorResponse
// @Router /organizations_propertys/{id} [get]
func (s *Server) getOrganizationPropertyByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	result, err := model.GetOrganizationPropertyModel().GetOrDefault(ctx, id)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}
