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

// @Summary getSubject
// @ID getSubject
// @Description get subjects
// @Accept json
// @Produce json
// @Param program_id query string false "program id"
// @Tags subject
// @Success 200 {array} entity.Subject
// @Failure 500 {object} InternalServerErrorResponse
// @Router /subjects [get]
func (s *Server) getSubject(c *gin.Context) {
	ctx := c.Request.Context()
	programID := c.Query("program_id")
	result, err := model.GetSubjectModel().Query(ctx, &da.SubjectCondition{
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

// @Summary getSubjectByID
// @ID getSubjectByID
// @Description get subjects by id
// @Accept json
// @Produce json
// @Param id path string true "subject id"
// @Tags subject
// @Success 200 {object} entity.Subject
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /subjects/{id} [get]
func (s *Server) getSubjectByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	result, err := model.GetSubjectModel().GetByID(ctx, id)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary addSubject
// @ID addSubject
// @Description addSubject
// @Accept json
// @Produce json
// @Param subject body entity.Subject true "subject"
// @Tags subject
// @Success 200 {object} entity.IDResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /subjects [post]
func (s *Server) addSubject(c *gin.Context) {
	op := GetOperator(c)
	ctx := c.Request.Context()
	data := new(entity.Subject)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "invalid data", log.Err(err), log.Any("data", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	id, err := model.GetSubjectModel().Add(ctx, op, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusOK, entity.IDResponse{ID: id})
}

// @Summary updateSubject
// @ID updateSubject
// @Description updateSubject
// @Accept json
// @Produce json
// @Param id path string true "subject id"
// @Param subject body entity.Subject true "subject"
// @Tags subject
// @Success 200 {object} entity.IDResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /subjects/{id} [put]
func (s *Server) updateSubject(c *gin.Context) {
	op := GetOperator(c)
	ctx := c.Request.Context()
	data := new(entity.Subject)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "invalid data", log.Err(err), log.Any("data", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	data.ID = c.Param("id")
	id, err := model.GetSubjectModel().Update(ctx, op, data)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary deleteSubject
// @ID deleteSubject
// @Description deleteSubject
// @Accept json
// @Produce json
// @Param id path string true "subject id"
// @Tags subject
// @Success 200 {object} entity.IDResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /subjects/{id} [delete]
func (s *Server) deleteSubject(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	id := c.Param("id")
	err := model.GetSubjectModel().Delete(ctx, op, id)
	switch err {
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
