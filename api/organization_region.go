package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
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
