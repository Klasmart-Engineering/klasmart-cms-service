package api

import (
	"database/sql"
	"github.com/dgrijalva/jwt-go"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

// @Summary list assessments
// @Description list assessments
// @Tags assessments
// @ID listAssessment
// @Accept json
// @Produce json
// @Param status query string false "status search"
// @Param teacher_name query string false "teacher name fuzzy search"
// @Param page query int false "page number" default(1)
// @Param page_size query integer false "page size" format(int) default(10)
// @Param order_by query string false "list order by" enums(class_end_time,-class_end_time,complete_time,-complete_time) default(-class_end_time)
// @Success 200 {object} entity.ListAssessmentsResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments [get]
func (s *Server) listAssessments(c *gin.Context) {
	ctx := c.Request.Context()

	cmd := entity.ListAssessmentsQuery{}
	{
		status := c.Query("status")
		if status != "" {
			status := entity.ListAssessmentsStatus(status)
			if !status.Valid() {
				log.Info(ctx, "list assessments: invalid list assessments status",
					log.String("status", string(status)),
				)
				c.JSON(http.StatusBadRequest, L(GeneralUnknown))
				return
			}
			if status != entity.ListAssessmentsStatusAll {
				temp := status.AssessmentStatus()
				cmd.Status = &temp
			}
		}

		teacherName := c.Query("teacher_name")
		if teacherName != "" {
			cmd.TeacherName = &teacherName
		}

		orderBy := c.Query("order_by")
		if orderBy != "" {
			orderBy := entity.ListAssessmentsOrderBy(orderBy)
			if !orderBy.Valid() {
				log.Info(ctx, "list assessments: invalid order by",
					log.String("status", string(status)),
				)
				c.JSON(http.StatusBadRequest, L(GeneralUnknown))
				return
			}
			cmd.OrderBy = &orderBy
		} else {
			orderBy := entity.ListAssessmentsOrderByClassEndTimeDesc
			cmd.OrderBy = &orderBy
		}

		pager := utils.GetDboPager(c.Query("page"), c.Query("page_size"))
		cmd.Page, cmd.PageSize = pager.Page, pager.PageSize
	}

	result, err := model.GetAssessmentModel().ListAssessments(ctx, dbo.MustGetDB(ctx), s.getOperator(c), cmd)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	default:
		log.Error(ctx, "list assessments: list failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
}

// @Summary add assessments
// @Description add assessments
// @Tags assessments
// @ID addAssessment
// @Accept json
// @Produce json
// @Param assessment body entity.AddAssessmentCommand true "add assessment command"
// @Success 200 {object} entity.AddAssessmentResult
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments [post]
func (s *Server) addAssessment(c *gin.Context) {
	ctx := c.Request.Context()

	log.Debug(ctx, "add assessment jwt: call")
	body := struct {
		Token string `json:"token"`
	}{}
	if err := c.ShouldBind(&body); err != nil {
		log.Info(ctx, "add assessment jwt: bind failed",
			log.Err(err),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	cmd := entity.AddAssessmentCommand{}
	if _, err := jwt.ParseWithClaims(body.Token, &cmd, func(token *jwt.Token) (interface{}, error) {
		return config.Get().Assessment.AddAssessmentSecret, nil
	}); err != nil {
		log.Error(ctx, "add assessment jwt: parse with claims failed",
			log.Err(err),
			log.Any("token", body.Token),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	log.Debug(ctx, "add assessment jwt: fill cmd", log.Any("cmd", cmd), log.String("token", body.Token))
	newID, err := model.GetAssessmentModel().AddAssessment(ctx, s.getOperator(c), cmd)
	switch err {
	case nil:
		log.Debug(ctx, "add assessment jwt success",
			log.Any("cmd", cmd),
			log.String("new_id", newID),
		)
		c.JSON(http.StatusOK, entity.AddAssessmentResult{ID: newID})
	default:
		log.Error(ctx, "add assessment jwt: add failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary add assessments for test
// @Description add assessments for test
// @Tags assessments
// @ID addAssessmentForTest
// @Accept json
// @Produce json
// @Param assessment body entity.AddAssessmentCommand true "add assessment command"
// @Success 200 {object} entity.AddAssessmentResult
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments_for_test [post]
func (s *Server) addAssessmentForTest(c *gin.Context) {
	ctx := c.Request.Context()

	cmd := entity.AddAssessmentCommand{}
	if err := c.ShouldBind(&cmd); err != nil {
		log.Info(ctx, "add assessment: bind failed",
			log.Err(err),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	newID, err := model.GetAssessmentModel().AddAssessment(ctx, s.getOperator(c), cmd)
	switch err {
	case nil:
		c.JSON(http.StatusOK, entity.AddAssessmentResult{ID: newID})
	default:
		log.Error(ctx, "add assessment: add failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary get assessment detail
// @Description get assessment detail
// @Tags assessments
// @ID getAssessment
// @Accept json
// @Produce json
// @Param id path string true "assessment id"
// @Success 200 {object} entity.AssessmentDetailView
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments/{id} [get]
func (s *Server) getAssessmentDetail(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		log.Info(ctx, "get assessment detail: require id")
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	item, err := model.GetAssessmentModel().GetAssessment(ctx, dbo.MustGetDB(ctx), s.getOperator(c), id)
	switch err {
	case nil:
		c.JSON(http.StatusOK, item)
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case constant.ErrRecordNotFound, sql.ErrNoRows:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	default:
		log.Info(ctx, "get assessment detail: get detail failed",
			log.Err(err),
			log.String("id", id),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary update assessment
// @Description update assessment
// @Tags assessments
// @ID updateAssessment
// @Accept json
// @Produce json
// @Param id path string true "assessment id"
// @Param update_assessment_command body entity.UpdateAssessmentCommand true "update assessment assessment command"
// @Success 200 {string} string "OK"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments/{id} [put]
func (s *Server) updateAssessment(c *gin.Context) {
	ctx := c.Request.Context()

	cmd := entity.UpdateAssessmentCommand{}
	{
		if err := c.ShouldBind(&cmd); err != nil {
			log.Info(ctx, "update assessment: bind failed",
				log.Err(err),
			)
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
			return
		}

		id := c.Param("id")
		if id == "" {
			log.Info(ctx, "update assessment: require id")
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
			return
		}
		cmd.ID = id
	}

	{
		if cmd.AttendanceIDs != nil && len(*cmd.AttendanceIDs) == 0 {
			c.JSON(http.StatusBadRequest, L(AssessMsgOneStudent))
			return
		}
	}

	err := model.GetAssessmentModel().UpdateAssessment(ctx, s.getOperator(c), cmd)
	switch err {
	case nil:
		c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	default:
		log.Info(ctx, "update assessment: update failed")
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary list home fun studies
// @Description list home fun studies
// @Tags assessments
// @ID listHomeFunStudies
// @Accept json
// @Produce json
// @Param query query string false "fuzzy query teacher name and student name"
// @Param status query string false "query status" enums(all,in_progress,complete)
// @Param order_by query string false "list order by" enums(latest_submit_time,-latest_submit_time,complete_time,-complete_time) default(-class_end_time)
// @Param page query int false "page number" default(1)
// @Param page_size query integer false "page size" format(int) default(10)
// @Success 200 {object} entity.ListHomeFunStudiesResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments/home_fun_studies [get]
func (s *Server) listHomeFunStudies(c *gin.Context) {
	panic("not implemented")
}

// @Summary get home fun study
// @Description get home fun study detail
// @Tags assessments
// @ID getHomeFunStudy
// @Accept json
// @Produce json
// @Param id path string true "home fun study id"
// @Success 200 {object} entity.ListHomeFunStudiesResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments/home_fun_studies/{id} [get]
func (s *Server) getHomeFunStudy(c *gin.Context) {
	panic("not implemented")
}

// @Summary assess home fun study
// @Description assess home fun study
// @Tags assessments
// @ID assessHomeFunStudy
// @Accept json
// @Produce json
// @Param id path string true "home fun study id"
// @Param assess_home_fun_study_args body entity.AssessHomeFunStudyArgs true "assess home fun study args"
// @Success 200 {string} string "OK"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments/home_fun_studies/{id}/assess [put]
func (s *Server) assessHomeFunStudy(c *gin.Context) {
	panic("not implemented")
}
