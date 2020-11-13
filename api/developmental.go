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

// @Summary getDevelopmental
// @ID getDevelopmental
// @Description get developmental
// @Accept json
// @Produce json
// @Param program_id query string false "program id"
// @Tags developmental
// @Success 200 {array} entity.Developmental
// @Failure 500 {object} InternalServerErrorResponse
// @Router /developmentals [get]
func (s *Server) getDevelopmental(c *gin.Context) {
	ctx := c.Request.Context()
	programID := c.Query("program_id")
	result, err := model.GetDevelopmentalModel().Query(ctx, &da.DevelopmentalCondition{
		ProgramID: sql.NullString{
			String: programID,
			Valid:  len(programID) != 0,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary getDevelopmentalByID
// @ID getDevelopmentalByID
// @Description get developmental by id
// @Accept json
// @Produce json
// @Param id path string true "developmental id"
// @Tags developmental
// @Success 200 {object} entity.Developmental
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /developmentals/{id} [get]
func (s *Server) getDevelopmentalByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	result, err := model.GetDevelopmentalModel().GetByID(ctx, id)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary addDevelopmental
// @ID addDevelopmental
// @Description addDevelopmental
// @Accept json
// @Produce json
// @Param developmental body entity.Developmental true "developmental"
// @Tags developmental
// @Success 200 {object} entity.IDResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /developmentals [post]
func (s *Server) addDevelopmental(c *gin.Context) {
	ctx := c.Request.Context()
	data := new(entity.Developmental)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "invalid data", log.Err(err), log.Any("data", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	id, err := model.GetDevelopmentalModel().Add(ctx, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusOK, entity.IDResponse{ID: id})
}

// @Summary updateDevelopmental
// @ID updateDevelopmental
// @Description updateDevelopmental
// @Accept json
// @Produce json
// @Param id path string true "developmental id"
// @Param developmental body entity.Developmental true "developmental"
// @Tags developmental
// @Success 200 {object} entity.IDResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /developmentals/{id} [put]
func (s *Server) updateDevelopmental(c *gin.Context) {
	ctx := c.Request.Context()
	data := new(entity.Developmental)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "invalid data", log.Err(err), log.Any("data", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	data.ID = c.Param("id")
	id, err := model.GetDevelopmentalModel().Update(ctx, data)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
