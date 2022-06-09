package api

import (
	"errors"
	"net/http"

	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"

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
	s.engine.GET("/v1/version", s.version)

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

	content := s.engine.Group("/v1")
	{
		content.POST("/contents", s.mustLogin, s.createContent)

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
		content.GET("/contents_authed", s.mustLogin, s.querySharedContent)
		content.GET("/contents_shared", s.mustLogin, s.querySharedContentV2)

		content.PUT("/contents_bulk/publish", s.mustLogin, s.publishContentBulk)
		content.DELETE("/contents_bulk", s.mustLogin, s.deleteContentBulk)

		content.GET("/contents_resources", s.mustLogin, s.getUploadPath)
		content.GET("/contents_resources/:resource_id", s.mustLoginWithoutOrgID, s.getContentResourcePath)
		content.GET("/contents_resources/:resource_id/download", s.mustLoginWithoutOrgID, s.getDownloadPath)
		content.GET("/contents_resources/:resource_id/check", s.mustLoginWithoutOrgID, s.checkExist)
		content.GET("/contents/:content_id/live/token", s.mustLogin, s.getContentLiveToken)
		content.POST("/contents_lesson_plans", s.mustLogin, s.getLessonPlansCanSchedule)

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
		schedules.GET("/schedules/:id/contents", s.mustLogin, s.getScheduleLiveLessonPlan)
		schedules.GET("/schedules_lesson_plans", s.mustLogin, s.getLessonPlans)
		schedules.GET("/schedules_time_view/dates", s.mustLogin, s.getScheduledDates)
		schedules.PUT("/schedules/:id/show_option", s.mustLogin, s.updateScheduleShowOption)
		schedules.GET("/schedules/:id/operator/newest_feedback", s.mustLogin, s.getScheduleNewestFeedbackByOperator)
		schedules.GET("/schedules_filter/programs", s.mustLogin, s.getProgramsInScheduleFilter)
		schedules.GET("/schedules_filter/subjects", s.mustLogin, s.getSubjectsInScheduleFilter)
		schedules.GET("/schedules_view/:id", s.mustLogin, s.getScheduleViewByID)

		schedules.POST("/schedules_time_view/dates", s.mustLogin, s.postScheduledDates)
		schedules.POST("/schedules_time_view", s.mustLogin, s.postScheduleTimeView)
		schedules.POST("/schedules_time_view/list", s.mustLogin, s.getScheduleTimeViewList)
		schedules.POST("/schedules/review/check_data", s.mustLogin, s.checkScheduleReviewData)
	}
	scheduleFeedback := s.engine.Group("/v1/schedules_feedbacks")
	{
		scheduleFeedback.POST("", s.mustLogin, s.addScheduleFeedback)
		scheduleFeedback.GET("", s.mustLogin, s.queryFeedback)
	}

	assessments := s.engine.Group("/v1")
	{
		// onlineStudy, onlineClass, offlineClass
		assessments.GET("/assessments_v2", s.mustLogin, s.queryAssessmentV2)
		assessments.GET("/assessments_v2/:id", s.mustLogin, s.getAssessmentDetailV2)
		assessments.PUT("/assessments_v2/:id", s.mustLogin, s.updateAssessmentV2)

		// live room callback
		assessments.POST("/assessments", s.addAssessment)

		// offlineStudy
		//assessments.GET("/user_offline_study", s.mustLogin, s.queryUserOfflineStudy)
		//assessments.GET("/user_offline_study/:id", s.mustLogin, s.getUserOfflineStudyByID)
		//assessments.PUT("/user_offline_study/:id", s.mustLogin, s.updateUserOfflineStudy)

		// home page
		assessments.GET("/assessments_summary", s.mustLogin, s.getAssessmentsSummary)
		assessments.GET("/assessments_for_student", s.mustLogin, s.getStudentAssessments)
		assessments.GET("/assessments", s.mustLogin, s.queryAssessments)
	}

	reports := s.engine.Group("/v1")
	{
		reports.GET("/reports/students_achievement_overview", s.mustLogin, s.getLearningOutcomeOverView)
		reports.GET("/reports/students", s.mustLogin, s.listStudentsAchievementReport)
		reports.GET("/reports/students/:id", s.mustLogin, s.getStudentAchievementReport)
		reports.GET("/reports/teachers/:id", s.mustLogin, s.getTeacherReport)
		reports.GET("/reports/teachers", s.mustLogin, s.getTeachersReport)

		reports.GET("/reports/performance/students", s.mustLogin, s.listStudentsPerformanceReport)
		reports.GET("/reports/performance/students/:id", s.mustLogin, s.getStudentPerformanceReport)

		reports.POST("/reports/teaching_loading", s.mustLogin, s.listTeachingLoadReport)
		reports.POST("/reports/teacher_load/lessons_list", s.mustLogin, s.listTeacherLoadLessons)
		reports.POST("/reports/teacher_load/lessons_summary", s.mustLogin, s.summaryTeacherLoadLessons)
		reports.POST("/reports/teacher_load/assignments", s.mustLogin, s.getTeacherLoadReportOfAssignment)
		reports.POST("/reports/teacher_load/missed_lessons", s.mustLogin, s.listTeacherMissedLessons)
		reports.GET("/reports/teacher_load_overview", s.mustLogin, s.getTeacherLoadOverview)

		reports.GET("/reports/learning_summary/time_filter", s.mustLogin, s.queryLearningSummaryTimeFilter)
		reports.GET("/reports/learning_summary/live_classes", s.mustLogin, s.queryLiveClassesSummary)
		reports.GET("/reports/learning_summary/live_classes_v2", s.mustLogin, s.queryLiveClassesSummaryV2)
		reports.GET("/reports/learning_summary/outcomes", s.mustLogin, s.queryOutcomesByAssessmentID)
		reports.GET("/reports/learning_summary/assignments", s.mustLogin, s.queryAssignmentsSummary)
		reports.GET("/reports/learning_summary/assignments_v2", s.mustLogin, s.queryAssignmentsSummaryV2)
		reports.GET("/reports/learner_weekly_overview", s.mustLogin, s.getLearnerWeeklyReportOverview)
		reports.GET("/reports/learner_monthly_overview", s.mustLogin, s.getLearnerMonthlyReportOverview)

		reports.GET("/reports/student_usage/organization_registration", s.mustLogin, s.getStudentUsageOrganizationRegistration)
		reports.POST("/reports/student_usage/class_registration", s.mustLogin, s.getStudentUsageClassRegistration)
		reports.POST("/reports/student_usage/material_view_count", s.mustLogin, s.getStudentUsageMaterialViewCountReport)
		reports.POST("/reports/student_usage/material", s.mustLogin, s.getStudentUsageMaterialReport)
		reports.POST("/reports/student_usage/classes_assignments_overview", s.mustLogin, s.getClassesAssignmentsOverview)
		reports.POST("/reports/student_usage/classes_assignments", s.mustLogin, s.getClassesAssignments)
		reports.POST("/reports/student_usage/classes_assignments/:class_id/unattended", s.mustLogin, s.getClassesAssignmentsUnattended)

		reports.POST("/reports/student_progress/learn_outcome_achievement", s.mustLogin, s.getLearnOutcomeAchievement)
		reports.POST("/reports/student_progress/class_attendance", s.mustLogin, s.getClassAttendance)
		reports.POST("/reports/student_progress/assignment_completion", s.mustLogin, s.getAssignmentsCompletion)
		reports.GET("/reports/student_progress/app/insight_message", s.mustLogin, s.getAppInsightMessage)
		reports.POST("/reports/learner_usage/overview", s.mustLogin, s.getLearnerUsageOverview)
	}

	outcomes := s.engine.Group("/v1")
	{
		outcomes.POST("/learning_outcomes", s.mustLogin, s.createOutcome)
		outcomes.GET("/learning_outcomes/:id", s.mustLogin, s.getOutcome)
		outcomes.PUT("/learning_outcomes/:id", s.mustLogin, s.updateOutcome)
		outcomes.DELETE("/learning_outcomes/:id", s.mustLogin, s.deleteOutcome)
		outcomes.GET("/learning_outcomes", s.mustLogin, s.queryOutcomes)
		outcomes.POST("/learning_outcomes/export", s.mustLogin, s.exportOutcomes)
		outcomes.POST("/learning_outcomes/verify_import", s.mustLogin, s.verifyImportOutcomes)
		outcomes.POST("/learning_outcomes/import", s.mustLogin, s.importOutcomes)

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
		outcomes.POST("/published_learning_outcomes", s.mustLogin, s.queryPublishedOutcomes)
	}

	shortcode := s.engine.Group("/v1")
	{
		shortcode.POST("/shortcode", s.mustLogin, s.generateShortcode)
	}

	folders := s.engine.Group("/v1/folders")
	{
		folders.POST("", s.mustLogin, s.createFolder)
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
		folders.GET("/tree", s.mustLogin, s.getTree)
	}

	crypto := s.engine.Group("/v1/crypto")
	{
		crypto.GET("/h5p/signature", s.mustLogin, s.h5pSignature)
		crypto.GET("/h5p/jwt", s.mustLogin, s.generateH5pJWT)
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
		milestone.GET("/private_milestones", s.mustLogin, s.searchPrivateMilestone)
		milestone.GET("/pending_milestones", s.mustLogin, s.searchPendingMilestone)

		milestone.PUT("/bulk_publish/milestones", s.mustLogin, s.bulkPublishMilestone)
		milestone.PUT("/bulk_approve/milestones", s.mustLogin, s.bulkApproveMilestone)
		milestone.PUT("/bulk_reject/milestones", s.mustLogin, s.bulkRejectMilestone)
	}

	organizationPermissions := s.engine.Group("/v1/organization_permissions")
	{
		organizationPermissions.POST("", s.mustLogin, s.hasOrganizationPermissions)
	}

	studentUsageReport := s.engine.Group("/v1/student_usage_record")
	{
		studentUsageReport.POST("/event", s.addStudentUsageRecordEvent)
	}

	internal := s.engine.Group("/v1/internal")
	{
		internal.GET("/contents", s.mustLoginWithoutOrgID, s.queryContentInternal)
		internal.GET("/schedules", s.mustLoginWithoutOrgID, s.queryScheduleInternal)
		internal.GET("/schedules/:id/relation_ids", s.mustLoginWithoutOrgID, s.queryScheduleRelationIDsInternal)
		internal.POST("/schedules/update_review_status", s.mustDataService, s.updateScheduleReviewStatus)
		internal.GET("/schedule_counts", s.getScheduleAttendance)
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

func (s Server) version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"git_hash":        constant.GitHash,
		"build_timestamp": constant.BuildTimestamp,
		"latest_migrate":  constant.LatestMigrate,
	})
}
