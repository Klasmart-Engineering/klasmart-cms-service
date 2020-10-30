package api

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/basicdata"
	"net/http"
)

// @Summary getSkill
// @ID getSkill
// @Description get skill
// @Accept json
// @Produce json
// @Param developmental_id query string false "developmental id"
// @Tags skill
// @Success 200 {array} entity.Skill
// @Failure 500 {object} InternalServerErrorResponse
// @Router /skills [get]
func (s *Server) getSkill(c *gin.Context) {
	ctx := c.Request.Context()
	developmentalID := c.Query("developmental_id")
	result, err := basicdata.GetSkillModel().Query(ctx, &da.SkillCondition{
		DevelopmentalID: sql.NullString{
			String: developmentalID,
			Valid:  developmentalID != "",
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(Unknown))
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary getSkillByID
// @ID getSkillByID
// @Description get skill by id
// @Accept json
// @Produce json
// @Param id path string true "skill id"
// @Tags skill
// @Success 200 {object} entity.Skill
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /skills/{id} [get]
func (s *Server) getSkillByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	result, err := basicdata.GetSkillModel().GetByID(ctx, id)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(Unknown))
	}
}
