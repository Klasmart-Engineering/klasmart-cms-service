package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/KL-Engineering/common-log/log"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"

	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
	"github.com/KL-Engineering/kidsloop-cms-service/model"
)

// @Summary assessments query
// @Description assessments query
// @Tags assessments
// @ID queryAssessmentV2
// @Accept json
// @Produce json
// @Param status query string false "status search,multiple states are separated by commas,optional value is Pending,NotStarted,Started,Draft,Complete"
// @Param query_key query string false "query key search"
// @Param query_type query string false "query type" enums(TeacherID)
// @Param assessment_type query string false "assessment type,multiple states are separated by commas,optional value is :OnlineClass,OfflineClass,OnlineStudy,ReviewStudy,OfflineStudy"
// @Param due_at_le query integer false "get less than due_at data"
// @Param class_id query string false "query by class id"
// @Param page query int false "page number" default(1)
// @Param page_size query integer false "page size" format(int) default(10)
// @Param order_by query string false "query order by" enums(class_end_at,-class_end_at,complete_at,-complete_at,create_at,-create_at) default(-create_at)
// @Success 200 {object} v2.AssessmentPageReply
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments_v2 [get]
func (s *Server) queryAssessmentV2(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	req := new(v2.AssessmentQueryReq)
	if err := c.ShouldBind(req); err != nil {
		return
	}

	if req.PageSize <= 0 || req.PageIndex <= 0 {
		req.PageIndex = constant.DefaultPageIndex
		req.PageSize = constant.DefaultPageSize
	}

	log.Debug(ctx, "queryAssessment request", log.Any("req", req), log.Any("op", op))

	result, err := model.GetAssessmentModelV2().Page(ctx, op, req)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary get assessment detail
// @Description get assessment detail
// @Tags assessments
// @ID getAssessmentDetailV2
// @Accept json
// @Produce json
// @Param id path string true "assessment id"
// @Success 200 {object} v2.AssessmentDetailReply
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments_v2/{id} [get]
func (s *Server) getAssessmentDetailV2(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	id := c.Param("id")

	log.Debug(ctx, "param", log.String("id", id))

	result, err := model.GetAssessmentModelV2().GetByID(ctx, op, id)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case constant.ErrRecordNotFound, sql.ErrNoRows:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary
// @Description update assessment
// @Tags assessments
// @ID updateAssessmentV2
// @Accept json
// @Produce json
// @Param id path string true "assessment id"
// @Param req body v2.AssessmentUpdateReq true "update assessment args"
// @Success 200 {string} string "OK"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments_v2/{id} [put]
func (s *Server) updateAssessmentV2(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	req := new(v2.AssessmentUpdateReq)

	if err := c.ShouldBind(req); err != nil {
		log.Error(ctx, "update assessment: bind body json failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	req.ID = c.Param("id")

	log.Debug(ctx, "request", log.Any("req", req))

	err := model.GetAssessmentModelV2().Update(ctx, op, req)

	switch err {
	case nil:
		c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case model.ErrAssessmentHasCompleted, constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary get assessments summary
// @Description get assessments summary
// @Tags assessments
// @ID getAssessmentsSummary
// @Accept json
// @Produce json
// @Param status query string false "search status, multiple states are separated by commas,optional value is Started,Draft,Complete"
// @Success 200 {object} v2.AssessmentsSummary
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments_summary [get]
func (s *Server) getAssessmentsSummary(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)
	status := c.Query("status")

	result, err := model.GetAssessmentModelV2().StatisticsCount(ctx, operator, &v2.StatisticsCountReq{Status: status})
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary get student assessments
// @Description get student assessments
// @Tags assessments
// @ID getStudentAssessments
// @Accept json
// @Produce json
// @Param type query string true "type search, the value of 'home_fun_study' is deprecated,please use 'OfflineStudy' instead" enums(OfflineClass,OnlineClass,OnlineStudy,OfflineStudy,home_fun_study,ReviewStudy)
// @Param status query string false "status search" enums(NotStarted,InProgress,Done,Resubmitted,Completed)
// @Param teacher_id query string false "teacher id search"
// @Param assessment_id query string false "assessment id search"
// @Param schedule_ids query string false "schedule ids search"
// @Param order_by query string false "order by" enums(create_at,-create_at,in_progress_at,-in_progress_at,done_at,-done_at,resubmitted_at,-resubmitted_at,completed_at,-completed_at) default(-create_at)
// @Param created_ge query string false "created_at greater search"
// @Param created_le query string false "created_at less search"
// @Param in_progress_ge query string false "in_progress_at greater search"
// @Param in_progress_le query string false "in_progress_at less search"
// @Param done_ge query string false "done_at greater search"
// @Param done_le query string false "done_at less search"
// @Param resubmitted_ge query string false "resubmitted_at greater search"
// @Param resubmitted_le query string false "resubmitted_at less search"
// @Param completed_ge query string false "complete_at greater search"
// @Param completed_le query string false "complete_at less search"
// @Param complete_at_ge query string false "complete_at greater search (deprecated)"
// @Param complete_at_le query string false "complete_at less search (deprecated)"
// @Param page query string false "page search"
// @Param page_size query string false "page size search"
// @Success 200 {object} v2.SearchStudentAssessmentsResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments_for_student [get]
func (s *Server) getStudentAssessments(c *gin.Context) {
	ctx := c.Request.Context()

	op := s.getOperator(c)
	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.AssessmentViewTeacherFeedback670)
	if err != nil {
		s.defaultErrorHandler(c, err)
		return
	}
	if !hasPermission {
		log.Warn(ctx, "No permission", log.Any("operator", op))
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
		return
	}

	conditions := new(v2.StudentQueryAssessmentConditions)
	err = c.ShouldBindQuery(conditions)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	conditions.OrgID = op.OrgID
	conditions.StudentID = op.UserID

	log.Info(ctx, "request params", log.Any("conditions", conditions))

	total, result, err := model.GetAssessmentModelV2().QueryStudentAssessment(ctx, op, conditions)

	switch err {
	case nil:
		c.JSON(http.StatusOK, &v2.SearchStudentAssessmentsResponse{
			List:  result,
			Total: total,
		})
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary add assessments
// @Description add assessments
// @Tags assessments
// @ID addAssessment
// @Accept json
// @Produce json
// @Param assessment body v2.ScheduleEndClassCallBackReq true "add assessment command"
// @Success 200 {object} v2.ScheduleEndClassCallBackResp
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

	log.Info(ctx, "add assessment call back info",
		log.String("log type", "report"),
		log.String("token", body.Token),
		log.String("step", "REPORT step1"))

	args := new(v2.ScheduleEndClassCallBackReq)
	if _, err := jwt.ParseWithClaims(body.Token, args, func(token *jwt.Token) (interface{}, error) {
		return config.Get().Assessment.AddAssessmentSecret, nil
	}); err != nil {
		log.Error(ctx, "add assessment jwt: parse with claims failed",
			log.Err(err),
			log.Any("token", body.Token),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	log.Info(ctx, "add assessment jwt: fill args", log.Any("args", args), log.String("token", body.Token))

	operator := s.getOperator(c)
	err := model.GetLiveRoomEventBusModel().PubEndClass(ctx, operator, args)
	switch err {
	case nil:
		log.Debug(ctx, "add assessment jwt: add success",
			log.Any("args", args),
		)
		c.JSON(http.StatusOK, &v2.ScheduleEndClassCallBackResp{
			ScheduleID: args.ScheduleID,
		})
	case constant.ErrInvalidArgs:
		log.Error(ctx, "add assessment jwt: add failed",
			log.Err(err),
			log.Any("args", args),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	default:
		log.Error(ctx, "add assessment jwt: add failed",
			log.Err(err),
			log.Any("args", args),
		)
		s.defaultErrorHandler(c, err)
	}
}

// @Summary list assessments
// @Description list assessments
// @Tags assessments
// @ID listAssessment
// @Accept json
// @Produce json
// @Param status query string false "status search"
// @Param page query int false "page number" default(1)
// @Param page_size query integer false "page size" format(int) default(10)
// @Param order_by query string false "list order by" enums(class_end_time,-class_end_time,complete_time,-complete_time) default(-class_end_time)
// @Success 200 {object} v2.ListAssessmentsResultForHomePage
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments [get]
func (s *Server) queryAssessments(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	req := new(v2.AssessmentQueryReq)
	if err := c.ShouldBind(req); err != nil {
		return
	}

	if req.Status == v2.AssessmentStatusCompliantCompleted.String() {
		req.Status = v2.AssessmentStatusComplete.String()
	} else if req.Status == v2.AssessmentStatusCompliantNotCompleted.String() {
		req.Status = fmt.Sprintf("%s,%s", v2.AssessmentStatusStarted.String(), v2.AssessmentStatusInDraft.String())
	}

	if req.PageSize <= 0 || req.PageIndex <= 0 {
		req.PageIndex = constant.DefaultPageIndex
		req.PageSize = constant.DefaultPageSize
	}

	log.Debug(ctx, "queryAssessment request", log.Any("req", req), log.Any("op", op))

	result, err := model.GetAssessmentModelV2().PageForHomePage(ctx, op, req)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}
