package api

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
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
	programID := c.Query("program_id")
	developmentalID := c.Query("developmental_id")
	if len(programID) == 0 || len(developmentalID) == 0 {
		log.Info(ctx, "param is invalid", log.String("programID", programID), log.String("developmentalID", developmentalID))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	result, err := model.GetSkillModel().Query(ctx, &da.SkillCondition{
		ProgramIDAndDevelopmentalID: []sql.NullString{
			sql.NullString{
				String: programID,
				Valid:  true,
			},
			sql.NullString{
				String: developmentalID,
				Valid:  true,
			},
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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
	result, err := model.GetSkillModel().GetByID(ctx, id)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary addSkill
// @ID addSkill
// @Description addSkill
// @Accept json
// @Produce json
// @Param skill body entity.Skill true "skill"
// @Tags skill
// @Success 200 {object} entity.IDResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /skills [post]
func (s *Server) addSkill(c *gin.Context) {
	op := GetOperator(c)
	ctx := c.Request.Context()
	data := new(entity.Skill)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "invalid data", log.Err(err), log.Any("data", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	id, err := model.GetSkillModel().Add(ctx, op, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusOK, entity.IDResponse{ID: id})
}

// @Summary updateSkill
// @ID updateSkill
// @Description updateSkill
// @Accept json
// @Produce json
// @Param id path string true "skill id"
// @Param skill body entity.Skill true "skill"
// @Tags skill
// @Success 200 {object} entity.IDResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /skills/{id} [put]
func (s *Server) updateSkill(c *gin.Context) {
	op := GetOperator(c)
	ctx := c.Request.Context()
	data := new(entity.Skill)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "invalid data", log.Err(err), log.Any("data", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	data.ID = c.Param("id")
	id, err := model.GetSkillModel().Update(ctx, op, data)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
