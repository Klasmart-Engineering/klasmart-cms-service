package api

import (
	"net/http"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
	"github.com/gin-gonic/gin"
)

// @Summary getSkill
// @ID getSkill
// @Description get skill
// @Accept json
// @Produce json
// @Param program_id query string false "program id"
// @Param developmental_id query string false "developmental id"
// @Tags skill
// @Success 200 {array} external.SubCategory
// @Failure 500 {object} InternalServerErrorResponse
// @Router /skills [get]
func (s *Server) getSkill(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	var subCategories []*external.SubCategory
	var err error

	categoryID := c.Query("developmental_id")

	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.ViewSubjects20115)
	if err != nil {
		log.Error(ctx, "getSkill: HasOrganizationPermission failed",
			log.Any("op", operator),
			log.String("perm", string(external.ViewSubjects20115)),
			log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "getSkill: HasOrganizationPermission failed",
			log.Any("op", operator),
			log.String("perm", string(external.ViewSubjects20115)))
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
		return
	}

	if categoryID != "" {
		subCategories, err = external.GetSubCategoryServiceProvider().GetByCategory(ctx, operator, categoryID)
	} else {
		// get all subcategory
		subCategories, err = external.GetSubCategoryServiceProvider().GetByOrganization(ctx, operator)
	}

	switch err {
	case nil:
		c.JSON(http.StatusOK, subCategories)
	default:
		s.defaultErrorHandler(c, err)
	}
}
