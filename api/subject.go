package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary getSubject
// @ID getSubject
// @Description get subjects
// @Accept json
// @Produce json
// @Param program_id query string false "program id"
// @Tags subject
// @Success 200 {array} external.Subject
// @Failure 500 {object} InternalServerErrorResponse
// @Router /subjects [get]
func (s *Server) getSubject(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	var result []*external.Subject
	var err error

	programID := c.Query("program_id")

	if programID != "" {
		result, err = external.GetSubjectServiceProvider().GetByOrganization(ctx, operator)
	} else {
		result, err = external.GetSubjectServiceProvider().GetByProgram(ctx, operator, programID)
	}

	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
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
// @Success 200 {object} IDResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /subjects [post]
func (s *Server) addSubject(c *gin.Context) {
	op := s.getOperator(c)
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
	c.JSON(http.StatusOK, IDResponse{ID: id})
}

// @Summary updateSubject
// @ID updateSubject
// @Description updateSubject
// @Accept json
// @Produce json
// @Param id path string true "subject id"
// @Param subject body entity.Subject true "subject"
// @Tags subject
// @Success 200 {object} IDResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /subjects/{id} [put]
func (s *Server) updateSubject(c *gin.Context) {
	op := s.getOperator(c)
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
		c.JSON(http.StatusOK, IDResponse{ID: id})
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
// @Success 200 {object} IDResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /subjects/{id} [delete]
func (s *Server) deleteSubject(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	id := c.Param("id")
	err := model.GetSubjectModel().Delete(ctx, op, id)
	switch err {
	case nil:
		c.JSON(http.StatusOK, IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
