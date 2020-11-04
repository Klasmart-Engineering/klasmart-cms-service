package api

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
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
		c.JSON(http.StatusInternalServerError, L(Unknown))
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
		c.JSON(http.StatusNotFound, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(Unknown))
	}
}
