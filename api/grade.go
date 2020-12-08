package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary getGrade
// @ID getGrade
// @Description get grade
// @Accept json
// @Produce json
// @Param program_id query string false "program id"
// @Tags grade
// @Success 200 {array} entity.Grade
// @Failure 500 {object} InternalServerErrorResponse
// @Router /grades [get]
func (s *Server) getGrade(c *gin.Context) {
	ctx := c.Request.Context()
	programID := c.Query("program_id")
	result, err := model.GetGradeModel().Query(ctx, &da.GradeCondition{
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

// @Summary getGradeByID
// @ID getGradeByID
// @Description get grade by id
// @Accept json
// @Produce json
// @Param id path string true "grade id"
// @Tags grade
// @Success 200 {object} entity.Grade
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /grades/{id} [get]
func (s *Server) getGradeByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	result, err := model.GetGradeModel().GetByID(ctx, id)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary addGrade
// @ID addGrade
// @Description addGrade
// @Accept json
// @Produce json
// @Param grade body entity.Grade true "Grade"
// @Tags grade
// @Success 200 {object} entity.IDResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /grades [post]
func (s *Server) addGrade(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	data := new(entity.Grade)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "invalid data", log.Err(err), log.Any("data", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	id, err := model.GetGradeModel().Add(ctx, op, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusOK, entity.IDResponse{ID: id})
}

// @Summary updateGrade
// @ID updateGrade
// @Description updateGrade
// @Accept json
// @Produce json
// @Param id path string true "grade id"
// @Param grade body entity.Grade true "grade"
// @Tags grade
// @Success 200 {object} entity.IDResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /grades/{id} [put]
func (s *Server) updateGrade(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	data := new(entity.Grade)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "invalid data", log.Err(err), log.Any("data", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	data.ID = c.Param("id")
	id, err := model.GetGradeModel().Update(ctx, op, data)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary deleteGrade
// @ID deleteGrade
// @Description deleteGrade
// @Accept json
// @Produce json
// @Param id path string true "grade id"
// @Tags grade
// @Success 200 {object} entity.IDResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /grades/{id} [delete]
func (s *Server) deleteGrade(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	id := c.Param("id")
	err := model.GetGradeModel().Delete(ctx, op, id)
	switch err {
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
