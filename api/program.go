package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
	"strings"
)

// @Summary getProgram
// @ID getProgram
// @Description get program
// @Accept json
// @Produce json
// @Tags program
// @Success 200 {array} entity.Program
// @Failure 500 {object} InternalServerErrorResponse
// @Router /programs [get]
func (s *Server) getProgram(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := model.GetProgramModel().Query(ctx, &da.ProgramCondition{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary getProgramByID
// @ID getProgramByID
// @Description get program by id
// @Accept json
// @Produce json
// @Param id path string true "program id"
// @Tags program
// @Success 200 {object} entity.Program
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /programs/{id} [get]
func (s *Server) getProgramByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	result, err := model.GetProgramModel().GetByID(ctx, id)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary addProgram
// @ID addProgram
// @Description addProgram
// @Accept json
// @Produce json
// @Param program body entity.Program true "program"
// @Tags program
// @Success 200 {object} entity.IDResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /programs [post]
func (s *Server) addProgram(c *gin.Context) {
	ctx := c.Request.Context()
	data := new(entity.Program)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "invalid data", log.Err(err), log.Any("data", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	id, err := model.GetProgramModel().Add(ctx, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusOK, entity.IDResponse{ID: id})
}

// @Summary updateProgram
// @ID updateProgram
// @Description updateProgram
// @Accept json
// @Produce json
// @Param id path string true "program id"
// @Param program body entity.Program true "program"
// @Tags program
// @Success 200 {object} entity.IDResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /programs/{id} [put]
func (s *Server) updateProgram(c *gin.Context) {
	ctx := c.Request.Context()
	data := new(entity.Program)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "invalid data", log.Err(err), log.Any("data", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	data.ID = c.Param("id")
	id, err := model.GetProgramModel().Update(ctx, data)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary SetAge
// @ID SetAge
// @Description SetAge
// @Accept json
// @Produce json
// @Param id path string true "program id"
// @Param age_ids query string true "age_ids"
// @Tags program
// @Success 200 {string} string
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /programs/{id}/ages [put]
func (s *Server) SetAge(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	ageIDs := c.Query("age_ids")
	ageIDArr := strings.Split(ageIDs, ",")
	err := model.GetProgramModel().SetAge(ctx, id, ageIDArr)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary SetGrade
// @ID SetGrade
// @Description SetGrade
// @Accept json
// @Produce json
// @Param id path string true "program id"
// @Param grade_ids query string true "grade_ids"
// @Tags program
// @Success 200 {string} string
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /programs/{id}/grades [put]
func (s *Server) SetGrade(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	gradeIDs := c.Query("grade_ids")
	gradeIDArr := strings.Split(gradeIDs, ",")
	err := model.GetProgramModel().SetGrade(ctx, id, gradeIDArr)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary SetSubject
// @ID SetSubject
// @Description SetSubject
// @Accept json
// @Produce json
// @Param id path string true "program id"
// @Param subject_ids query string true "subject_ids"
// @Tags program
// @Success 200 {string} string
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /programs/{id}/subjects [put]
func (s *Server) SetSubject(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	subjectIDs := c.Query("subject_ids")
	subjectIDArr := strings.Split(subjectIDs, ",")
	err := model.GetProgramModel().SetSubject(ctx, id, subjectIDArr)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary SetDevelopmental
// @ID SetDevelopmental
// @Description SetDevelopmental
// @Accept json
// @Produce json
// @Param id path string true "program id"
// @Param development_ids query string true "development_ids"
// @Tags program
// @Success 200 {string} string
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /programs/{id}/developments [put]
func (s *Server) SetDevelopmental(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	deveIDs := c.Query("development_ids")
	deveIDArr := strings.Split(deveIDs, ",")
	err := model.GetProgramModel().SetDevelopment(ctx, id, deveIDArr)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary SetSkill
// @ID SetSkill
// @Description SetSkill
// @Accept json
// @Produce json
// @Param id path string true "program id"
// @Param development_id query string true "development_id"
// @Param skill_ids query string true "skill_ids"
// @Tags program
// @Success 200 {string} string
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /programs/{id}/skills [put]
func (s *Server) SetSkill(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	deveID := c.Query("development_id")
	skillIDs := c.Query("skill_ids")
	skillIDArr := strings.Split(skillIDs, ",")
	err := model.GetProgramModel().SetDeveSkill(ctx, id, deveID, skillIDArr)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
