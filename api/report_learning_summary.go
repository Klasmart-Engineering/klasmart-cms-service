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

// @Summary query specified learning summary filter items
// @Description list specified learning summary filter items
// @Tags reports/learningSummary
// @ID queryLearningSummaryFilterItems
// @Accept json
// @Produce json
// @Param type query string true "filter type" enums(year,week,school,class,teacher,student,subject)
// @Param year query integer false "year"
// @Param week_start query integer false "week start timestamp(unit: second)"
// @Param week_end query integer false "week end timestamp(unit: second)"
// @Param school_id query string false "school id"
// @Param class_id query string false "class id"
// @Param teacher_id query string false "teacher_id"
// @Param student_id query string false "student_id"
// @Success 200 {array} entity.QueryLearningSummaryFilterResultItem
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/learning_summary/filters [get]
func (s *Server) queryLearningSummaryFilterItems(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)

	strType := c.Query("type")
	typo := entity.LearningSummaryFilterType(strType)
	if !typo.Valid() {
		log.Error(ctx, "parse learning summary filter: invalid type field", log.String("type", strType))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	filter, err := s.parseLearningSummaryFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	args := entity.QueryLearningSummaryFilterItemsArgs{
		Type:                  typo,
		LearningSummaryFilter: filter,
	}
	result, err := model.GetLearningSummaryReportModel().QueryFilterItems(ctx, dbo.MustGetDB(ctx), operator, &args)
	if err != nil {
		log.Error(ctx, "query learning summary filter items: query filter items failed",
			log.Err(err),
			log.Any("args", args),
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
// @Success 200 {array} entity.QueryLiveClassesSummaryResult
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
// @Success 200 {array} entity.QueryAssignmentsSummaryResult
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

	strYear := c.Query("year")
	year, err := strconv.Atoi(strYear)
	if err != nil {
		log.Error(ctx, "parse learning summary filter: parse year field failed",
			log.Err(err),
			log.String("year", strYear),
		)
		return nil, err
	}
	strWeekStart := c.Query("year")
	weekStart, err := strconv.ParseInt(strWeekStart, 10, 64)
	if err != nil {
		log.Error(ctx, "parse learning summary filter: parse week_start field failed",
			log.Err(err),
			log.String("week_start", strWeekStart),
		)
		return nil, err
	}
	strWeekEnd := c.Query("week_end")
	weekEnd, err := strconv.ParseInt(strWeekEnd, 10, 64)
	if err != nil {
		log.Error(ctx, "parse learning summary filter: parse week_end field failed",
			log.Err(err),
			log.String("week_end", strWeekEnd),
		)
		return nil, err
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
