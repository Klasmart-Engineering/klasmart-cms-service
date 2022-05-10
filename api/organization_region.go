package api

import (
	"net/http"

	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/model"
	"github.com/gin-gonic/gin"
)

type OrganizationRegionInfoResponse struct {
	Orgs []*entity.RegionOrganizationInfo `json:"orgs"`
}

// @Summary getOrganizationByHeadquarterForDetails
// @ID getOrganizationByHeadquarterForDetails
// @Description get organization region by user org
// @Tags organizationProperty
// @Success 200 {object} OrganizationRegionInfoResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /organizations_region [get]
func (s *Server) getOrganizationByHeadquarterForDetails(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	result, err := model.GetOrganizationRegionModel().GetOrganizationByHeadquarterForDetails(ctx, dbo.MustGetDB(ctx), op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, OrganizationRegionInfoResponse{Orgs: result})
	default:
		s.defaultErrorHandler(c, err)
	}
}
