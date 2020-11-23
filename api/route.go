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
		content.GET("/contents_folders", MustLogin, s.queryFolderContent)
		content.GET("/contents/:content_id/statistics", MustLogin, s.contentDataCount)
		content.GET("/contents_private", MustLogin, s.queryPrivateContent)
		content.GET("/contents_pending", MustLogin, s.queryPendingContent)

		content.PUT("/contents_bulk/publish", MustLogin, s.publishContentBulk)
		content.DELETE("/contents_bulk", MustLogin, s.deleteContentBulk)

		content.GET("/contents_resources", MustLogin, s.getUploadPath)
		content.GET("/contents_resources/:resource_id", MustLoginWithoutOrgID, s.getPath)
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
		reports.GET("/reports/students", MustLogin, s.listStudentsReport)
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

	folders := s.engine.Group("/folders")
	{
		folders.POST("", MustLogin, s.createFolder)
		folders.POST("/items", MustLogin, s.addFolderItem)
		folders.DELETE("/items/:item_id", MustLogin, s.removeFolderItem)
		folders.PUT("/items/details/:item_id", MustLogin, s.updateFolderItem)
		folders.PUT("/items/move/:item_id", MustLogin, s.moveFolderItem)
		folders.PUT("/items/bulk/move", MustLogin, s.moveFolderItemBulk)

		folders.GET("/items/list/:item_id", MustLogin, s.listFolderItems)
		folders.GET("/items/search/private", MustLogin, s.searchPrivateFolderItems)
		folders.GET("/items/search/org", MustLogin, s.searchOrgFolderItems)
		folders.GET("/items/details/:item_id", MustLogin, s.getFolderItemByID)
	}

	crypto := s.engine.Group("/crypto")
	{
		crypto.GET("/h5p/signature", MustLogin, s.h5pSignature)
		crypto.GET("/h5p/jwt", MustLogin, s.generateH5pJWT)
	}

	ages := s.engine.Group("/v1/ages")
	{
		ages.GET("", MustLoginWithoutOrgID, s.getAge)
		ages.GET("/:id", MustLoginWithoutOrgID, s.getAgeByID)
		ages.POST("", MustLoginWithoutOrgID, s.addAge)
		ages.PUT("/:id", MustLoginWithoutOrgID, s.updateAge)
		ages.DELETE("/:id", MustLoginWithoutOrgID, s.deleteAge)
	}
	classTypes := s.engine.Group("/v1/class_types")
	{
		classTypes.GET("", MustLoginWithoutOrgID, s.getClassType)
		classTypes.GET("/:id", MustLoginWithoutOrgID, s.getClassTypeByID)
	}
	developmental := s.engine.Group("/v1/developmentals")
	{
		developmental.GET("", MustLoginWithoutOrgID, s.getDevelopmental)
		developmental.GET("/:id", MustLoginWithoutOrgID, s.getDevelopmentalByID)
		developmental.POST("", MustLoginWithoutOrgID, s.addDevelopmental)
		developmental.PUT("/:id", MustLoginWithoutOrgID, s.updateDevelopmental)
		developmental.DELETE("/:id", MustLoginWithoutOrgID, s.deleteDevelopmental)
	}
	grade := s.engine.Group("/v1/grades")
	{
		grade.GET("", MustLoginWithoutOrgID, s.getGrade)
		grade.GET("/:id", MustLoginWithoutOrgID, s.getGradeByID)
		grade.POST("", MustLoginWithoutOrgID, s.addGrade)
		grade.PUT("/:id", MustLoginWithoutOrgID, s.updateGrade)
		grade.DELETE("/:id", MustLoginWithoutOrgID, s.deleteGrade)
	}
	lessonTypes := s.engine.Group("/v1/lesson_types")
	{
		lessonTypes.GET("", MustLoginWithoutOrgID, s.getLessonType)
		lessonTypes.GET("/:id", MustLoginWithoutOrgID, s.getLessonTypeByID)
	}
	programs := s.engine.Group("/v1/programs")
	{
		programs.GET("", MustLoginWithoutOrgID, s.getProgram)
		programs.GET("/:id", MustLoginWithoutOrgID, s.getProgramByID)
		programs.POST("", MustLoginWithoutOrgID, s.addProgram)
		programs.PUT("/:id", MustLoginWithoutOrgID, s.updateProgram)
		programs.DELETE("/:id", MustLoginWithoutOrgID, s.deleteProgram)

		programs.PUT("/:id/ages", MustLoginWithoutOrgID, s.SetAge)
		programs.PUT("/:id/grades", MustLoginWithoutOrgID, s.SetGrade)
		programs.PUT("/:id/subjects", MustLoginWithoutOrgID, s.SetSubject)
		programs.PUT("/:id/developments", MustLoginWithoutOrgID, s.SetDevelopmental)
		programs.PUT("/:id/skills", MustLoginWithoutOrgID, s.SetSkill)
	}
	skills := s.engine.Group("/v1/skills")
	{
		skills.GET("", MustLoginWithoutOrgID, s.getSkill)
		skills.GET("/:id", MustLoginWithoutOrgID, s.getSkillByID)
		skills.POST("", MustLoginWithoutOrgID, s.addSkill)
		skills.PUT("/:id", MustLoginWithoutOrgID, s.updateSkill)
		skills.DELETE("/:id", MustLoginWithoutOrgID, s.deleteSkill)
	}
	subjects := s.engine.Group("/v1/subjects")
	{
		subjects.GET("", MustLoginWithoutOrgID, s.getSubject)
		subjects.GET("/:id", MustLoginWithoutOrgID, s.getSubjectByID)
		subjects.POST("", MustLoginWithoutOrgID, s.addSubject)
		subjects.PUT("/:id", MustLoginWithoutOrgID, s.updateSubject)
		subjects.DELETE("/:id", MustLoginWithoutOrgID, s.deleteSubject)
	}
	visibilitySettings := s.engine.Group("/v1/visibility_settings")
	{
		visibilitySettings.GET("", MustLogin, s.getVisibilitySetting)
		visibilitySettings.GET("/:id", MustLogin, s.getVisibilitySettingByID)
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
