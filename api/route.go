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
		assets.GET("/", s.mustLogin, s.searchAssets)
		assets.POST("/", s.mustLogin, s.createAsset)
		assets.GET("/:id", s.mustLogin, s.getAssetByID)
		assets.PUT("/:id", s.mustLogin, s.updateAsset)
		assets.DELETE("/:id", s.mustLogin, s.deleteAsset)
	}
	content := s.engine.Group("/v1")
	{
		content.POST("/contents", s.mustLogin, s.createContent)
		content.GET("/contents/:content_id", s.mustLogin, s.getContent)
		content.PUT("/contents/:content_id", s.mustLogin, s.updateContent)
		content.PUT("/contents/:content_id/lock", s.mustLogin, s.lockContent)
		content.PUT("/contents/:content_id/publish", s.mustLogin, s.publishContent)
		content.PUT("/contents/:content_id/publish/assets", s.mustLogin, s.publishContentWithAssets)
		content.PUT("/contents/:content_id/review/approve", s.mustLogin, s.approve)
		content.PUT("/contents/:content_id/review/reject", s.mustLogin, s.reject)
		content.PUT("/contents_review/approve", s.mustLogin, s.approveBulk)
		content.PUT("/contents_review/reject", s.mustLogin, s.rejectBulk)

		content.DELETE("/contents/:content_id", s.mustLogin, s.deleteContent)
		content.GET("/contents", s.mustLogin, s.queryContent)
		content.GET("/contents/:content_id/statistics", s.mustLogin, s.contentDataCount)
		content.GET("/contents_private", s.mustLogin, s.queryPrivateContent)
		content.GET("/contents_pending", s.mustLogin, s.queryPendingContent)
		content.GET("/contents_folders", s.mustLogin, s.queryFolderContent)

		content.PUT("/contents_bulk/publish", s.mustLogin, s.publishContentBulk)
		content.DELETE("/contents_bulk", s.mustLogin, s.deleteContentBulk)

		content.GET("/contents_resources", s.mustLogin, s.getUploadPath)
		content.GET("/contents_resources/:resource_id", s.mustLoginWithoutOrgID, s.getPath)
		content.GET("/contents/:content_id/live/token", s.mustLogin, s.getContentLiveToken)
	}
	schedules := s.engine.Group("/v1")
	{
		schedules.PUT("/schedules/:id", s.mustLogin, s.updateSchedule)
		schedules.DELETE("/schedules/:id", s.mustLogin, s.deleteSchedule)
		schedules.POST("/schedules", s.mustLogin, s.addSchedule)
		schedules.GET("/schedules/:id", s.mustLogin, s.getScheduleByID)
		schedules.GET("/schedules", s.mustLogin, s.querySchedule)
		schedules.GET("/schedules_time_view", s.mustLogin, s.getScheduleTimeView)
		schedules.GET("/schedules/:id/live/token", s.mustLogin, s.getScheduleLiveToken)
		schedules.PUT("/schedules/:id/status", s.mustLogin, s.updateScheduleStatus)
		schedules.GET("/schedules_participate/class", s.mustLogin, s.getParticipateClass)
		schedules.GET("/schedules_lesson_plans", s.mustLogin, s.getLessonPlans)
	}

	assessments := s.engine.Group("/v1")
	{
		assessments.GET("/assessments", s.mustLogin, s.listAssessments)
		assessments.POST("/assessments", s.addAssessment)
		assessments.POST("/assessments_for_test", s.mustLogin, s.addAssessmentForTest)
		assessments.GET("/assessments/:id", s.mustLogin, s.getAssessmentDetail)
		assessments.PUT("/assessments/:id", s.mustLogin, s.updateAssessment)
	}

	reports := s.engine.Group("/v1")
	{
		reports.GET("/reports/students", s.mustLogin, s.listStudentsReport)
		reports.GET("/reports/students/:id", s.mustLogin, s.getStudentReport)
		reports.GET("/reports/teachers/:id", s.mustLogin, s.getTeacherReport)
	}

	outcomes := s.engine.Group("/v1")
	{
		outcomes.POST("/learning_outcomes", s.mustLogin, s.createOutcome)
		outcomes.GET("/learning_outcomes/:id", s.mustLogin, s.getOutcome)
		outcomes.PUT("/learning_outcomes/:id", s.mustLogin, s.updateOutcome)
		outcomes.DELETE("/learning_outcomes/:id", s.mustLogin, s.deleteOutcome)
		outcomes.GET("/learning_outcomes", s.mustLogin, s.queryOutcomes)

		outcomes.PUT("/learning_outcomes/:id/lock", s.mustLogin, s.lockOutcome)
		outcomes.PUT("/learning_outcomes/:id/publish", s.mustLogin, s.publishOutcome)
		outcomes.PUT("/learning_outcomes/:id/approve", s.mustLogin, s.approveOutcome)
		outcomes.PUT("/learning_outcomes/:id/reject", s.mustLogin, s.rejectOutcome)

		outcomes.PUT("/bulk_approve/learning_outcomes", s.mustLogin, s.bulkApproveOutcome)
		outcomes.PUT("/bulk_reject/learning_outcomes", s.mustLogin, s.bulkRejectOutcome)

		outcomes.PUT("/bulk_publish/learning_outcomes", s.mustLogin, s.bulkPublishOutcomes)
		outcomes.DELETE("/bulk/learning_outcomes", s.mustLogin, s.bulkDeleteOutcomes)

		outcomes.GET("/private_learning_outcomes", s.mustLogin, s.queryPrivateOutcomes)
		outcomes.GET("/pending_learning_outcomes", s.mustLogin, s.queryPendingOutcomes)
	}

	folders := s.engine.Group("/v1/folders")
	{
		folders.POST("", s.mustLogin, s.createFolder)
		folders.POST("/items", s.mustLogin, s.addFolderItem)
		folders.DELETE("/items/:item_id", s.mustLogin, s.removeFolderItem)
		folders.DELETE("/items", s.mustLogin, s.removeFolderItemBulk)
		folders.PUT("/items/details/:item_id", s.mustLogin, s.updateFolderItem)
		folders.PUT("/items/move", s.mustLogin, s.moveFolderItem)
		folders.PUT("/items/bulk/move", s.mustLogin, s.moveFolderItemBulk)

		folders.GET("/items/list/:item_id", s.mustLogin, s.listFolderItems)
		folders.GET("/items/search/private", s.mustLogin, s.searchPrivateFolderItems)
		folders.GET("/items/search/org", s.mustLogin, s.searchOrgFolderItems)
		folders.GET("/items/details/:item_id", s.mustLogin, s.getFolderItemByID)

	}

	crypto := s.engine.Group("/v1/crypto")
	{
		crypto.GET("/h5p/signature", s.mustLogin, s.h5pSignature)
		crypto.GET("/h5p/jwt", s.mustLogin, s.generateH5pJWT)
	}

	ages := s.engine.Group("/v1/ages")
	{
		ages.GET("", s.mustLoginWithoutOrgID, s.getAge)
		ages.GET("/:id", s.mustLoginWithoutOrgID, s.getAgeByID)
		ages.POST("", s.mustLoginWithoutOrgID, s.addAge)
		ages.PUT("/:id", s.mustLoginWithoutOrgID, s.updateAge)
		ages.DELETE("/:id", s.mustLoginWithoutOrgID, s.deleteAge)
	}
	classTypes := s.engine.Group("/v1/class_types")
	{
		classTypes.GET("", s.mustLoginWithoutOrgID, s.getClassType)
		classTypes.GET("/:id", s.mustLoginWithoutOrgID, s.getClassTypeByID)
	}
	developmental := s.engine.Group("/v1/developmentals")
	{
		developmental.GET("", s.mustLoginWithoutOrgID, s.getDevelopmental)
		developmental.GET("/:id", s.mustLoginWithoutOrgID, s.getDevelopmentalByID)
		developmental.POST("", s.mustLoginWithoutOrgID, s.addDevelopmental)
		developmental.PUT("/:id", s.mustLoginWithoutOrgID, s.updateDevelopmental)
		developmental.DELETE("/:id", s.mustLoginWithoutOrgID, s.deleteDevelopmental)
	}
	grade := s.engine.Group("/v1/grades")
	{
		grade.GET("", s.mustLoginWithoutOrgID, s.getGrade)
		grade.GET("/:id", s.mustLoginWithoutOrgID, s.getGradeByID)
		grade.POST("", s.mustLoginWithoutOrgID, s.addGrade)
		grade.PUT("/:id", s.mustLoginWithoutOrgID, s.updateGrade)
		grade.DELETE("/:id", s.mustLoginWithoutOrgID, s.deleteGrade)
	}
	lessonTypes := s.engine.Group("/v1/lesson_types")
	{
		lessonTypes.GET("", s.mustLoginWithoutOrgID, s.getLessonType)
		lessonTypes.GET("/:id", s.mustLoginWithoutOrgID, s.getLessonTypeByID)
	}
	programs := s.engine.Group("/v1/programs")
	{
		programs.GET("", s.mustLoginWithoutOrgID, s.getProgram)
		programs.GET("/:id", s.mustLoginWithoutOrgID, s.getProgramByID)
		programs.POST("", s.mustLoginWithoutOrgID, s.addProgram)
		programs.PUT("/:id", s.mustLoginWithoutOrgID, s.updateProgram)
		programs.DELETE("/:id", s.mustLoginWithoutOrgID, s.deleteProgram)

		programs.PUT("/:id/ages", s.mustLoginWithoutOrgID, s.SetAge)
		programs.PUT("/:id/grades", s.mustLoginWithoutOrgID, s.SetGrade)
		programs.PUT("/:id/subjects", s.mustLoginWithoutOrgID, s.SetSubject)
		programs.PUT("/:id/developments", s.mustLoginWithoutOrgID, s.SetDevelopmental)
		programs.PUT("/:id/skills", s.mustLoginWithoutOrgID, s.SetSkill)
	}
	skills := s.engine.Group("/v1/skills")
	{
		skills.GET("", s.mustLoginWithoutOrgID, s.getSkill)
		skills.GET("/:id", s.mustLoginWithoutOrgID, s.getSkillByID)
		skills.POST("", s.mustLoginWithoutOrgID, s.addSkill)
		skills.PUT("/:id", s.mustLoginWithoutOrgID, s.updateSkill)
		skills.DELETE("/:id", s.mustLoginWithoutOrgID, s.deleteSkill)
	}
	subjects := s.engine.Group("/v1/subjects")
	{
		subjects.GET("", s.mustLoginWithoutOrgID, s.getSubject)
		subjects.GET("/:id", s.mustLoginWithoutOrgID, s.getSubjectByID)
		subjects.POST("", s.mustLoginWithoutOrgID, s.addSubject)
		subjects.PUT("/:id", s.mustLoginWithoutOrgID, s.updateSubject)
		subjects.DELETE("/:id", s.mustLoginWithoutOrgID, s.deleteSubject)
	}
	visibilitySettings := s.engine.Group("/v1/visibility_settings")
	{
		visibilitySettings.GET("", s.mustLogin, s.getVisibilitySetting)
		visibilitySettings.GET("/:id", s.mustLogin, s.getVisibilitySettingByID)
	}
	userSettings := s.engine.Group("/v1/user_settings")
	{
		userSettings.POST("", s.mustLoginWithoutOrgID, s.setUserSetting)
		userSettings.GET("", s.mustLoginWithoutOrgID, s.getUserSettingByOperator)
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
