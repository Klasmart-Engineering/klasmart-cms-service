package api

import (
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"net/http"
	"strconv"

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
// @Param summary_type query string true "learning summary type" enums(live_class,assignment)
// @Success 200 {array} entity.LearningSummaryFilterYear
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/learning_summary/time_filter [get]
func (s *Server) queryLearningSummaryTimeFilter(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	// parse args
	strSummaryType := c.Query("summary_type")
	summaryType := entity.LearningSummaryType(strSummaryType)
	if !summaryType.Valid() {
		log.Error(ctx, "parse learning summary time filter: invalid summary type", log.String("summary_type", strSummaryType))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	args := entity.QueryLearningSummaryTimeFilterArgs{
		SummaryType: summaryType,
		OrgID:       operator.OrgID,
	}

	// call business model
	result, err := model.GetLearningSummaryReportModel().QueryTimeFilter(ctx, dbo.MustGetDB(ctx), operator, &args)
	if err != nil {
		log.Error(ctx, "query learning summary time filter failed",
			log.Err(err),
			log.Any("args", args),
		)
	}

	// handle error
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(ReportMsgNoPermission))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary query remaining learning summary filter
// @Description query remaining learning summary filter
// @Tags reports/learningSummary
// @ID queryLearningSummaryRemainingFilter
// @Accept json
// @Produce json
// @Param summary_type query string true "learning summary type" enums(live_class,assignment)
// @Param week_start query integer false "week start timestamp(unit: second)"
// @Param week_end query integer false "week end timestamp(unit: second)"
// @Success 200 {array} entity.LearningSummaryFilterSchool
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/learning_summary/remaining_filter [get]
func (s *Server) queryLearningSummaryRemainingFilter(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	// parse args
	strSummaryType := c.Query("summary_type")
	summaryType := entity.LearningSummaryType(strSummaryType)
	if !summaryType.Valid() {
		log.Error(ctx, "parse learning summary remaining filter: invalid summary type", log.String("summary_type", strSummaryType))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	var err error
	weekStart := int64(0)
	strWeekStart := c.Query("week_start")
	if strWeekStart != "" {
		weekStart, err = strconv.ParseInt(strWeekStart, 10, 64)
		if err != nil {
			log.Error(ctx, "query learning summary remaining filter: parse week_start field failed",
				log.Err(err),
				log.String("week_start", strWeekStart),
			)
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		}
	}
	weekEnd := int64(0)
	strWeekEnd := c.Query("week_end")
	if strWeekEnd != "" {
		weekEnd, err = strconv.ParseInt(strWeekEnd, 10, 64)
		if err != nil {
			log.Error(ctx, "query learning summary remaining filter: parse week_end field failed",
				log.Err(err),
				log.String("week_end", strWeekEnd),
			)
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		}
	}
	args := entity.QueryLearningSummaryRemainingFilterArgs{
		SummaryType: summaryType,
		OrgID:       operator.OrgID,
		WeekStart:   weekStart,
		WeekEnd:     weekEnd,
	}

	// call business model
	result, err := model.GetLearningSummaryReportModel().QueryRemainingFilter(ctx, dbo.MustGetDB(ctx), operator, &args)
	if err != nil {
		log.Error(ctx, "query learning summary remaining filter failed",
			log.Err(err),
			log.Any("args", args),
		)
	}

	// handle error
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(ReportMsgNoPermission))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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
		SchoolID:  c.Query("school_id"),
		ClassID:   c.Query("class_id"),
		TeacherID: c.Query("teacher_id"),
		StudentID: c.Query("student_id"),
		SubjectID: c.Query("subject_id"),
	}
	return &filter, nil
}
