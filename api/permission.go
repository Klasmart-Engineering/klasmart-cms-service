package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

type HasPermissionResponse map[external.PermissionName]bool

// @Summary hasOrganizationPermissions
// @ID hasOrganizationPermissions
// @Description has organization permission
// @Accept json
// @Produce json
// @Param permission_name query string true "permission_name separated by commas"
// @Tags permission
// @Success 200 {object} HasPermissionResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /organization_permissions [get]
func (s *Server) hasOrganizationPermissions(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	permissionNames := strings.Split(c.Query("permission_name"), constant.StringArraySeparator)
	if len(permissionNames) == 0 {
		log.Info(ctx, "permission name is invalid", log.String("permission_name", c.Query("permission_name")))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	permissions := make([]external.PermissionName, len(permissionNames))
	for index, permissionName := range permissionNames {
		permissions[index] = external.PermissionName(permissionName)
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
