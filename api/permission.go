package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

type HasPermissionRequest struct {
	PermissionNames []external.PermissionName `json:"permission_name" swaggertype:"array,string"`
}

type HasPermissionResponse map[external.PermissionName]bool

// @Summary hasOrganizationPermissions
// @ID hasOrganizationPermissions
// @Description has organization permission
// @Accept json
// @Produce json
// @Param PermissionNames body HasPermissionRequest true "permission names"
// @Tags permission
// @Success 200 {object} HasPermissionResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /organization_permissions [post]
func (s *Server) hasOrganizationPermissions(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	request := &HasPermissionRequest{}
	if err := c.ShouldBind(&request); err != nil {
		log.Info(ctx, "invalid check organization permission request", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	if len(request.PermissionNames) == 0 {
		log.Info(ctx, "invalid check organization permission request", log.String("permission_name", c.Query("permission_name")))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	permissions := make([]external.PermissionName, len(request.PermissionNames))
	for index, permissionName := range request.PermissionNames {
		if permissionName.String() == "" {
			log.Info(ctx, "permission name is invalid", log.String("permission_name", c.Query("permission_name")))
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
			return
		}

		permissions[index] = permissionName
	}

	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, operator, permissions)
	switch err {
	case nil:
		log.Info(ctx, "check permission success",
			log.String("permission_name", c.Query("permission_name")),
			log.Any("hasPermission", hasPermission))
		c.JSON(http.StatusOK, HasPermissionResponse(hasPermission))
	default:
		log.Error(ctx, "check permission failed",
			log.Err(err),
			log.String("permission_name", c.Query("permission_name")))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
