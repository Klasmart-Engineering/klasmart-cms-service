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

// @Summary getAge
// @ID getAge
// @Description get age
// @Accept json
// @Produce json
// @Param program_id query string false "program id"
// @Tags age
// @Success 200 {array} entity.Age
// @Failure 500 {object} InternalServerErrorResponse
// @Router /ages [get]
func (s *Server) getAge(c *gin.Context) {
	ctx := c.Request.Context()
	programID := c.Query("program_id")
	result, err := model.GetAgeModel().Query(ctx, &da.AgeCondition{
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

// @Summary getAgeByID
// @ID getAgeByID
// @Description get age by id
// @Accept json
// @Produce json
// @Param id path string true "age id"
// @Tags age
// @Success 200 {object} entity.Age
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /ages/{id} [get]
func (s *Server) getAgeByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	result, err := model.GetAgeModel().GetByID(ctx, id)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary addAge
// @ID addAge
// @Description add age
// @Accept json
// @Produce json
// @Param age body entity.Age true "age"
// @Tags age
// @Success 200 {object} entity.IDResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /ages [post]
func (s *Server) addAge(c *gin.Context) {
	ctx := c.Request.Context()
	data := new(entity.Age)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "invalid data", log.Err(err), log.Any("data", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	id, err := model.GetAgeModel().Add(ctx, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusOK, entity.IDResponse{ID: id})
}

// @Summary updateAge
// @ID updateAge
// @Description updateAge
// @Accept json
// @Produce json
// @Param id path string true "age id"
// @Param age body entity.Age true "age"
// @Tags age
// @Success 200 {object} entity.IDResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /ages/{id} [put]
func (s *Server) updateAge(c *gin.Context) {
	ctx := c.Request.Context()
	data := new(entity.Age)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "invalid data", log.Err(err), log.Any("data", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	data.ID = c.Param("id")
	id, err := model.GetAgeModel().Update(ctx, data)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
