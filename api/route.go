package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	errRouteNotFound = errors.New("route not found")
)

func (s Server) registeRoute() {
	s.engine.NoRoute(func(c *gin.Context) {
		c.AbortWithError(http.StatusNotFound, errRouteNotFound)
	})

	s.engine.GET("/v1/ping", s.ping)

	assets := s.engine.Group("/v1/assets")
	{
		assets.GET("/", MustLogin, s.searchAssets)
		assets.POST("/", MustLogin, s.createAsset)
		assets.GET("/:id", MustLogin, s.getAssetByID)
		assets.PUT("/:id", MustLogin, s.updateAsset)
		assets.DELETE("/:id", MustLogin, s.deleteAsset)
	}
	content := s.engine.Group("/v1")
	{
		content.POST("/contents", MustLogin, s.createContent)
		content.GET("/contents/:content_id", MustLogin, s.getContent)
		content.PUT("/contents/:content_id", MustLogin, s.updateContent)
		content.PUT("/contents/:content_id/lock", MustLogin, s.lockContent)
		content.PUT("/contents/:content_id/publish", MustLogin, s.publishContent)
		content.PUT("/contents/:content_id/publish/assets", MustLogin, s.publishContentWithAssets)
		content.PUT("/contents/:content_id/review/approve", MustLogin, s.approve)
		content.PUT("/contents/:content_id/review/reject", MustLogin, s.reject)
		content.DELETE("/contents/:content_id", MustLogin, s.deleteContent)
		content.GET("/contents", MustLogin, s.queryContent)
		content.GET("/contents/:content_id/statistics", MustLogin, s.contentDataCount)
		content.GET("/contents_private", MustLogin, s.queryPrivateContent)
		content.GET("/contents_pending", MustLogin, s.queryPendingContent)

		content.PUT("/contents_bulk/publish", MustLogin, s.publishContentBulk)
		content.DELETE("/contents_bulk", MustLogin, s.deleteContentBulk)

		content.GET("/contents_resources", MustLogin, s.getUploadPath)
		content.GET("/contents_resources/:resource_id", MustLogin, s.getPath)
		content.GET("/contents/:content_id/live/token", MustLogin, s.getContentLiveToken)
	}
	schedules := s.engine.Group("/v1")
	{
		schedules.PUT("/schedules/:id", MustLogin, s.updateSchedule)
		schedules.DELETE("/schedules/:id", MustLogin, s.deleteSchedule)
		schedules.POST("/schedules", MustLogin, s.addSchedule)
		schedules.GET("/schedules/:id", MustLogin, s.getScheduleByID)
		schedules.GET("/schedules", MustLogin, s.querySchedule)
		schedules.GET("/schedules_time_view", MustLogin, s.getScheduleTimeView)
		schedules.GET("/schedules/:id/live/token", MustLogin, s.getScheduleLiveToken)
		schedules.PUT("/schedules/:id/status", MustLogin, s.updateScheduleStatus)
		schedules.GET("/schedules_participate/class", MustLogin, s.getParticipateClass)
		schedules.GET("/schedules_lesson_plans", MustLogin, s.getLessonPlans)
	}

	assessments := s.engine.Group("/v1")
	{
		assessments.GET("/assessments", MustLogin, s.listAssessments)
		assessments.POST("/assessments", MustLogin, s.addAssessment)
		assessments.GET("/assessments/:id", MustLogin, s.getAssessmentDetail)
		assessments.PUT("/assessments/:id", MustLogin, s.updateAssessment)
	}

	reports := s.engine.Group("/v1")
	{
		reports.GET("/reports", MustLogin, s.listStudentsReport)
		reports.GET("/reports/students/:id", MustLogin, s.getStudentDetailReport)
	}

	outcomes := s.engine.Group("/v1")
	{
		outcomes.POST("/learning_outcomes", MustLogin, s.createOutcome)
		outcomes.GET("/learning_outcomes/:id", MustLogin, s.getOutcome)
		outcomes.PUT("/learning_outcomes/:id", MustLogin, s.updateOutcome)
		outcomes.DELETE("/learning_outcomes/:id", MustLogin, s.deleteOutcome)
		outcomes.GET("/learning_outcomes", MustLogin, s.queryOutcomes)

		outcomes.PUT("/learning_outcomes/:id/lock", MustLogin, s.lockOutcome)
		outcomes.PUT("/learning_outcomes/:id/publish", MustLogin, s.publishOutcome)
		outcomes.PUT("/learning_outcomes/:id/approve", MustLogin, s.approveOutcome)
		outcomes.PUT("/learning_outcomes/:id/reject", MustLogin, s.rejectOutcome)

		outcomes.PUT("/bulk_publish/learning_outcomes", MustLogin, s.bulkPublishOutcomes)
		outcomes.DELETE("/bulk/learning_outcomes", MustLogin, s.bulkDeleteOutcomes)

		outcomes.GET("/private_learning_outcomes", MustLogin, s.queryPrivateOutcomes)
		outcomes.GET("/pending_learning_outcomes", MustLogin, s.queryPendingOutcomes)
	}

	crypto := s.engine.Group("/v1/crypto")
	{
		crypto.GET("/h5p/signature", MustLogin, s.h5pSignature)
	}
}

// Ping godoc
// @ID ping
// @Summary Ping
// @Description Ping and test service
// @Tags common
// @Accept  json
// @Produce  plain
// @Success 200 {object} string
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /ping [get]
func (s Server) ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}
