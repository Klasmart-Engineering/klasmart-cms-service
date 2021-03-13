package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

// @Summary getSkill
// @ID getSkill
// @Description get skill
// @Accept json
// @Produce json
// @Param program_id query string false "program id"
// @Param developmental_id query string false "developmental id"
// @Tags skill
// @Success 200 {array} entity.Skill
// @Failure 500 {object} InternalServerErrorResponse
// @Router /skills [get]
func (s *Server) getSkill(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	var subCategories []*external.SubCategory
	var err error

	categoryID := c.Query("developmental_id")
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
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
