package api

import (
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

type ShortcodeRequest struct {
	Kind entity.ShortcodeKind `json:"kind" form:"kind"`
}
type ShortcodeResponse struct {
	Shortcode string `json:"shortcode" form:"shortcode"`
}

// @ID generateShortcode
// @Summary generate Shortcode
// @Tags shortcode
// @Description generate shortcode
// @Accept json
// @Produce json
// @Param kind body ShortcodeRequest false "learning outcome"
// @Success 200 {object} ShortcodeResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 409 {object} ConflictResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /shortcode [post]
func (s *Server) generateShortcode(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	var data ShortcodeRequest
	err := c.ShouldBindJSON(&data)
	if err != nil && err.Error() != "EOF" {
		log.Warn(ctx, "generateShortcode: ShouldBindJSON failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.CreateLearningOutcome)
	if err != nil {
		log.Warn(ctx, "generateShortcode: HasOrganizationPermission failed", log.Any("op", op), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "generateShortcode: no permission", log.Any("op", op), log.String("perm", string(external.CreateLearningOutcome)))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}

	var shortcode string
	switch data.Kind {
	case entity.KindMileStone:
		shortcode, err = model.GetMilestoneModel().GenerateShortcode(ctx, op)
	case entity.KindOutcome:
		shortcode, err = model.GetOutcomeModel().GenerateShortcode(ctx, op)
	default:
		log.Warn(ctx, "generateShortcode: kind not allowed", log.Any("shortcode_kind", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	switch err {
	case nil:
		c.JSON(http.StatusOK, &ShortcodeResponse{Shortcode: shortcode})
	case constant.ErrConflict:
		c.JSON(http.StatusConflict, L(AssessMsgExistShortcode))
	case constant.ErrExceededLimit:
		c.JSON(http.StatusConflict, L(AssessMsgExistShortcode))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
