package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
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
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
