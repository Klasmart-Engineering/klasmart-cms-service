package api

import (
	"errors"
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

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

	if config.Get().KidsLoopRegion == constant.KidsloopCN {
		users := s.engine.Group("/v1/users")
		{
			users.GET("/check_account", s.checkAccount)
			users.POST("/send_code", s.sendCode)
			users.POST("/login", s.login)
			users.POST("/register", s.register)
			users.POST("/forgotten_pwd", s.forgottenPassword)
			users.PUT("/reset_password", s.mustLogin, s.resetPassword)
			users.POST("/invite_notify", s.mustAms, s.inviteNotify)
		}
	}

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
		content.POST("/contents/copy", s.mustLogin, s.copyContent)
		//Inherent & unchangeable
		content.GET("/contents/:content_id", s.mustLogin, s.getContent)
		//Inherent & unchangeable
		content.GET("/contents", s.mustLogin, s.queryContent)

		content.PUT("/contents/:content_id", s.mustLogin, s.updateContent)
		content.PUT("/contents/:content_id/lock", s.mustLogin, s.lockContent)
		content.PUT("/contents/:content_id/publish", s.mustLogin, s.publishContent)
		content.PUT("/contents/:content_id/publish/assets", s.mustLogin, s.publishContentWithAssets)
		content.PUT("/contents/:content_id/review/approve", s.mustLogin, s.approve)
		content.PUT("/contents/:content_id/review/reject", s.mustLogin, s.reject)
		content.PUT("/contents_review/approve", s.mustLogin, s.approveBulk)
		content.PUT("/contents_review/reject", s.mustLogin, s.rejectBulk)

		content.DELETE("/contents/:content_id", s.mustLogin, s.deleteContent)
		content.GET("/contents/:content_id/statistics", s.mustLogin, s.contentDataCount)
		content.GET("/contents_private", s.mustLogin, s.queryPrivateContent)
		content.GET("/contents_pending", s.mustLogin, s.queryPendingContent)
		content.GET("/contents_folders", s.mustLogin, s.queryFolderContent)
		content.GET("/contents_authed", s.mustLogin, s.queryAuthContent)

		content.PUT("/contents_bulk/publish", s.mustLogin, s.publishContentBulk)
		content.DELETE("/contents_bulk", s.mustLogin, s.deleteContentBulk)

		content.GET("/contents_resources", s.mustLogin, s.getUploadPath)
		content.GET("/contents_resources/:resource_id", s.mustLoginWithoutOrgID, s.getContentResourcePath)
		content.GET("/contents_resources/:resource_id/download", s.mustLoginWithoutOrgID, s.getDownloadPath)
		content.GET("/contents/:content_id/live/token", s.mustLogin, s.getContentLiveToken)
	}
	h5pEvents := s.engine.Group("/v1")
	{
		h5pEvents.POST("/h5p/events", s.mustLogin, s.createH5PEvent)
	}

	authedContents := s.engine.Group("/v1")
	{
		authedContents.POST("/contents_auth", s.mustLogin, s.addAuthedContent)
		authedContents.POST("/contents_auth/batch", s.mustLogin, s.batchAddAuthedContent)
		authedContents.DELETE("/contents_auth", s.mustLogin, s.deleteAuthedContent)
		authedContents.DELETE("/contents_auth/batch", s.mustLogin, s.batchDeleteAuthedContent)
		authedContents.GET("/contents_auth/org", s.mustLogin, s.getOrgAuthedContent)
		authedContents.GET("/contents_auth/content", s.mustLogin, s.getContentAuthedOrg)
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
		schedules.GET("/schedules_lesson_plans", s.mustLogin, s.getLessonPlans)
		schedules.GET("/schedules_time_view/dates", s.mustLogin, s.getScheduledDates)
		schedules.GET("/schedules_filter/schools", s.mustLogin, s.getSchoolInScheduleFilter)
		schedules.GET("/schedules_filter/classes", s.mustLogin, s.getClassesInScheduleFilter)
		schedules.PUT("/schedules/:id/show_option", s.mustLogin, s.updateScheduleShowOption)
		schedules.GET("/schedules/:id/operator/newest_feedback", s.mustLogin, s.getScheduleNewestFeedbackByOperator)
		schedules.GET("/schedules_filter/programs", s.mustLogin, s.getProgramsInScheduleFilter)
		schedules.GET("/schedules_filter/subjects", s.mustLogin, s.getSubjectsInScheduleFilter)
		schedules.GET("/schedules_filter/class_types", s.mustLogin, s.getClassTypesInScheduleFilter)
		schedules.GET("/schedules_view/:id", s.mustLogin, s.getScheduleViewByID)

		schedules.POST("/schedules_time_view/dates", s.mustLogin, s.postScheduledDates)
		schedules.POST("/schedules_time_view", s.mustLogin, s.postScheduleTimeView)
	}
	scheduleFeedback := s.engine.Group("/v1/schedules_feedbacks")
	{
		scheduleFeedback.POST("", s.mustLogin, s.addScheduleFeedback)
		scheduleFeedback.GET("", s.mustLogin, s.queryFeedback)
	}

	assessments := s.engine.Group("/v1")
	{
		assessments.GET("/assessments", s.mustLogin, s.listAssessments)
		assessments.POST("/assessments", s.addAssessment)
		assessments.POST("/assessments_for_test", s.mustLogin, s.addAssessmentForTest)
		assessments.GET("/assessments/:id", s.mustLogin, s.getAssessmentDetail)
		assessments.PUT("/assessments/:id", s.mustLogin, s.updateAssessment)

		assessments.GET("/assessments_summary", s.mustLogin, s.getAssessmentsSummary)
	}

	homeFunStudies := s.engine.Group("/v1")
	{
		homeFunStudies.GET("/home_fun_studies", s.mustLogin, s.listHomeFunStudies)
		homeFunStudies.GET("/home_fun_studies/:id", s.mustLogin, s.getHomeFunStudy)
		homeFunStudies.PUT("/home_fun_studies/:id/assess", s.mustLogin, s.assessHomeFunStudy)
	}

	reports := s.engine.Group("/v1")
	{
		reports.GET("/reports/students", s.mustLogin, s.listStudentsAchievementReport)
		reports.GET("/reports/students/:id", s.mustLogin, s.getStudentAchievementReport)
		reports.GET("/reports/teachers/:id", s.mustLogin, s.getTeacherReport)

		reports.GET("/reports/performance/students", s.mustLogin, s.listStudentsPerformanceReport)
		reports.GET("/reports/performance/students/:id", s.mustLogin, s.getStudentPerformanceReport)

		reports.GET("/reports/performance/h5p/students", s.mustLogin, s.listStudentsPerformanceH5PReport)
		reports.GET("/reports/performance/h5p/students/:id", s.mustLogin, s.getStudentPerformanceH5PReport)

		reports.GET("/reports/teaching_loading", s.mustLogin, s.listTeachingLoadReport)
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

	shortcode := s.engine.Group("/v1")
	{
		shortcode.POST("/shortcode", s.mustLogin, s.generateShortcode)
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

		folders.GET("/share", s.mustLogin, s.getFoldersSharedRecords)
		folders.PUT("/share", s.mustLogin, s.shareFolders)

	}

	crypto := s.engine.Group("/v1/crypto")
	{
		crypto.GET("/h5p/signature", s.mustLogin, s.h5pSignature)
		crypto.GET("/h5p/jwt", s.mustLogin, s.generateH5pJWT)
	}

	classTypes := s.engine.Group("/v1/class_types")
	{
		classTypes.GET("", s.mustLoginWithoutOrgID, s.getClassType)
		classTypes.GET("/:id", s.mustLoginWithoutOrgID, s.getClassTypeByID)
	}

	lessonTypes := s.engine.Group("/v1/lesson_types")
	{
		lessonTypes.GET("", s.mustLoginWithoutOrgID, s.getLessonType)
		lessonTypes.GET("/:id", s.mustLoginWithoutOrgID, s.getLessonTypeByID)
	}

	programGroups := s.engine.Group("/v1/programs_groups")
	{
		programGroups.GET("", s.mustLoginWithoutOrgID, s.getProgramGroup)
	}

	programs := s.engine.Group("/v1/programs")
	{
		programs.GET("", s.mustLoginWithoutOrgID, s.getProgram)
	}

	subjects := s.engine.Group("/v1/subjects")
	{
		subjects.GET("", s.mustLoginWithoutOrgID, s.getSubject)
	}

	developmental := s.engine.Group("/v1/developmentals")
	{
		developmental.GET("", s.mustLoginWithoutOrgID, s.getDevelopmental)
	}

	skills := s.engine.Group("/v1/skills")
	{
		skills.GET("", s.mustLoginWithoutOrgID, s.getSkill)
	}

	ages := s.engine.Group("/v1/ages")
	{
		ages.GET("", s.mustLoginWithoutOrgID, s.getAge)
	}

	grade := s.engine.Group("/v1/grades")
	{
		grade.GET("", s.mustLoginWithoutOrgID, s.getGrade)
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

	organizationProperties := s.engine.Group("/v1/organizations_propertys")
	{
		organizationProperties.GET("/:id", s.mustLoginWithoutOrgID, s.getOrganizationPropertyByID)
	}
	organizationRegions := s.engine.Group("/v1/organizations_region")
	{
		organizationRegions.GET("", s.mustLoginWithoutOrgID, s.getOrganizationByHeadquarterForDetails)
	}

	learningOutcomeSet := s.engine.Group("/v1/sets")
	{
		learningOutcomeSet.POST("", s.mustLogin, s.createOutcomeSet)
		learningOutcomeSet.POST("/bulk_bind", s.mustLogin, s.bulkBindOutcomeSet)
		learningOutcomeSet.GET("", s.mustLogin, s.pullOutcomeSet)
	}

	classes := s.engine.Group("/v1")
	{
		// ams-class add members event
		classes.POST("/classes_members", s.classAddMembersEvent)
		classes.DELETE("/classes_members", s.classDeleteMembersEvent)
	}
	milestone := s.engine.Group("/v1")
	{
		milestone.POST("/milestones", s.mustLogin, s.createMilestone)
		milestone.GET("/milestones/:id", s.mustLogin, s.obtainMilestone)

		milestone.PUT("/milestones/:id/occupy", s.mustLogin, s.occupyMilestone)
		milestone.PUT("/milestones/:id", s.mustLogin, s.updateMilestone)

		milestone.DELETE("/milestones", s.mustLogin, s.deleteMilestone)

		milestone.GET("/milestones", s.mustLogin, s.searchMilestone)

		milestone.POST("/milestones/publish", s.mustLogin, s.publishMilestone)
	}

	organizationPermissions := s.engine.Group("/v1/organization_permissions")
	{
		organizationPermissions.POST("", s.mustLogin, s.hasOrganizationPermissions)
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
