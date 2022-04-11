package api

import (
	"net/http"
	"strconv"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary query learning summary time filter
// @Description  query learning summary time filter
// @Tags reports/learningSummary
// @ID queryLearningSummaryTimeFilter
// @Accept json
// @Produce json
// @Param time_offset query integer true "time offset (unit: second)"
// @Param summary_type query string true "learning summary type" enums(live_class,assignment)
// @Param school_ids query string false "school ids, use commas to separate"
// @Param teacher_id query string false "teacher_id"
// @Param student_id query string false "student_id"
// @Success 200 {array} entity.LearningSummaryFilterYear
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/learning_summary/time_filter [get]
func (s *Server) queryLearningSummaryTimeFilter(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	// parse args
	strTimeOffset := c.Query("time_offset")
	timeOffset, err := strconv.Atoi(strTimeOffset)
	if err != nil {
		log.Error(ctx, "query learning summary remaining filter: require time offset",
			log.Err(err),
			log.String("time_offset", strTimeOffset),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	args := entity.QueryLearningSummaryTimeFilterArgs{
		TimeOffset: timeOffset,
		OrgID:      operator.OrgID,
		SchoolIDs:  utils.SliceDeduplicationExcludeEmpty(strings.Split(c.Query("school_ids"), ",")),
		TeacherID:  c.Query("teacher_id"),
		StudentID:  c.Query("student_id"),
	}

	// call business model
	result, err := model.GetLearningSummaryReportModel().QueryTimeFilter(ctx, dbo.MustGetDB(ctx), operator, &args)
	if err != nil {
		log.Error(ctx, "query learning summary time filter failed",
			log.Err(err),
			log.Any("args", args),
		)
	}
	switch err {
	case nil:
		if result == nil {
			result = []*entity.LearningSummaryFilterYear{}
		}
		c.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(ReportMsgNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary query live classes summary
// @Description query live classes summary
// @Tags reports/learningSummary
// @ID queryLiveClassesSummary
// @Accept json
// @Produce json
// @Param year query integer false "year"
// @Param week_start query integer false "week start timestamp(unit: second)"
// @Param week_end query integer false "week end timestamp(unit: second)"
// @Param school_id query string false "school id"
// @Param class_id query string false "class id"
// @Param teacher_id query string false "teacher_id"
// @Param student_id query string false "student_id"
// @Param subject_id query string false "subject_id"
// @Success 200 {object} entity.QueryLiveClassesSummaryResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/learning_summary/live_classes [get]
func (s *Server) queryLiveClassesSummary(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	filter, err := s.parseLearningSummaryFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	if filter.StudentID == "" {
		log.Error(ctx, "query live classes summary: require student id")
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	result, err := model.GetLearningSummaryReportModel().QueryLiveClassesSummary(ctx, dbo.MustGetDB(ctx), operator, filter)
	if err != nil {
		log.Error(ctx, "query live classes summary failed",
			log.Err(err),
			log.Any("filter", filter),
		)
	}
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(ReportMsgNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary query live classes summary
// @Description query live classes summary
// @Tags reports/learningSummary
// @ID queryLiveClassesSummaryV2
// @Accept json
// @Produce json
// @Param year query integer false "year"
// @Param week_start query integer false "week start timestamp(unit: second)"
// @Param week_end query integer false "week end timestamp(unit: second)"
// @Param school_id query string false "school id"
// @Param class_id query string false "class id"
// @Param teacher_id query string false "teacher_id"
// @Param student_id query string false "student_id"
// @Param subject_id query string false "subject_id"
// @Success 200 {object} entity.QueryLiveClassesSummaryResultV2
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/learning_summary/live_classes_v2 [get]
func (s *Server) queryLiveClassesSummaryV2(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	filter, err := s.parseLearningSummaryFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	if filter.StudentID == "" {
		log.Error(ctx, "query live classes summary: require student id")
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	result, err := model.GetLearningSummaryReportModel().QueryLiveClassesSummaryV2(ctx, dbo.MustGetDB(ctx), operator, filter)
	if err != nil {
		log.Error(ctx, "query live classes summary failed",
			log.Err(err),
			log.Any("filter", filter),
		)
	}
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(ReportMsgNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary query outcomes for  live classes summary
// @Description query outcomes for  live classes summary
// @Tags reports/learningSummary
// @ID queryOutcomesByAssessmentID
// @Accept json
// @Produce json
// @Param assessment_id query string false "assessment_id"
// @Success 200 {object} []entity.LearningSummaryOutcome
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/learning_summary/outcomes [get]
func (s *Server) queryOutcomesByAssessmentID(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)
	assessmentID := c.Query("assessment_id")
	studentID := c.Query("student_id")
	result, err := model.GetLearningSummaryReportModel().QueryOutcomesByAssessmentID(ctx, operator, assessmentID, studentID)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(ReportMsgNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary query live classes summary
// @Description query live classes summary
// @Tags reports/learningSummary
// @ID queryAssignmentsSummaryV2
// @Accept json
// @Produce json
// @Param year query integer false "year"
// @Param week_start query integer false "week start timestamp(unit: second)"
// @Param week_end query integer false "week end timestamp(unit: second)"
// @Param school_id query string false "school id"
// @Param class_id query string false "class id"
// @Param teacher_id query string false "teacher_id"
// @Param student_id query string false "student_id"
// @Param subject_id query string false "subject_id"
// @Success 200 {object} entity.QueryAssignmentsSummaryResultV2
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/learning_summary/assignments_v2 [get]
func (s *Server) queryAssignmentsSummaryV2(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	filter, err := s.parseLearningSummaryFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	if filter.StudentID == "" {
		log.Error(ctx, "query assignments summary: require student id")
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	result, err := model.GetLearningSummaryReportModel().QueryAssignmentsSummaryV2(ctx, dbo.MustGetDB(ctx), operator, filter)
	if err != nil {
		log.Error(ctx, "query assignments summary failed",
			log.Err(err),
			log.Any("filter", filter),
		)
	}
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(ReportMsgNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary query live classes summary
// @Description query live classes summary
// @Tags reports/learningSummary
// @ID queryAssignmentsSummary
// @Accept json
// @Produce json
// @Param year query integer false "year"
// @Param week_start query integer false "week start timestamp(unit: second)"
// @Param week_end query integer false "week end timestamp(unit: second)"
// @Param school_id query string false "school id"
// @Param class_id query string false "class id"
// @Param teacher_id query string false "teacher_id"
// @Param student_id query string false "student_id"
// @Param subject_id query string false "subject_id"
// @Success 200 {object} entity.QueryAssignmentsSummaryResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/learning_summary/assignments [get]
func (s *Server) queryAssignmentsSummary(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	filter, err := s.parseLearningSummaryFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	if filter.StudentID == "" {
		log.Error(ctx, "query assignments summary: require student id")
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	result, err := model.GetLearningSummaryReportModel().QueryAssignmentsSummary(ctx, dbo.MustGetDB(ctx), operator, filter)
	if err != nil {
		log.Error(ctx, "query assignments summary failed",
			log.Err(err),
			log.Any("filter", filter),
		)
	}
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(ReportMsgNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}

func (s *Server) parseLearningSummaryFilter(c *gin.Context) (*entity.LearningSummaryFilter, error) {
	ctx := c.Request.Context()
	var err error

	year := 0
	strYear := c.Query("year")
	if strYear != "" {
		year, err = strconv.Atoi(strYear)
		if err != nil {
			log.Error(ctx, "parse learning summary filter: parse year field failed",
				log.Err(err),
				log.String("year", strYear),
			)
			return nil, err
		}
	}
	weekStart := int64(0)
	strWeekStart := c.Query("week_start")
	if strWeekStart != "" {
		weekStart, err = strconv.ParseInt(strWeekStart, 10, 64)
		if err != nil {
			log.Error(ctx, "parse learning summary filter: parse week_start field failed",
				log.Err(err),
				log.String("week_start", strWeekStart),
			)
			return nil, err
		}
	}
	weekEnd := int64(0)
	strWeekEnd := c.Query("week_end")
	if strWeekEnd != "" {
		weekEnd, err = strconv.ParseInt(strWeekEnd, 10, 64)
		if err != nil {
			log.Error(ctx, "parse learning summary filter: parse week_end field failed",
				log.Err(err),
				log.String("week_end", strWeekEnd),
			)
			return nil, err
		}
	}
	filter := entity.LearningSummaryFilter{
		Year:      year,
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		SchoolIDs: utils.SliceDeduplicationExcludeEmpty(strings.Split(c.Query("school_id"), ",")),
		ClassID:   c.Query("class_id"),
		TeacherID: c.Query("teacher_id"),
		StudentID: c.Query("student_id"),
		SubjectID: c.Query("subject_id"),
	}
	return &filter, nil
}
