package external

type PermissionName string

func (p PermissionName) String() string {
	return string(p)
}

var AllPermissionNames = []PermissionName{
	CreateContentPage201,

	ViewMyPending212,
	ViewOrgPending213,
	ViewMyPublished214,
	ViewOrgPublished215,
	ViewMyArchived216,
	ViewOrgArchived217,
	ViewMySchoolPublished218,

	CreateLessonMaterial220,
	CreateLessonPlan221,
	CreateMySchoolsContent223,
	CreateAllSchoolsContent224,
	ViewMySchoolPending225,
	ViewMySchoolArchived226,
	ViewAllSchoolsPublished227,
	ViewAllSchoolsPending228,
	ViewAllSchoolsArchived229,

	EditMyUnpublishedContent230,
	EditMyPublishedContent234,
	EditOrgPublishedContent235,
	EditLessonMaterialMetadataAndContent236,
	EditLessonPlanMetadata237,
	EditLessonPlanContent238,

	DeleteMyUnpublishedContent240,
	DeleteMySchoolsPending241,
	RemoveMySchoolsPublished242,
	DeleteMySchoolsArchived243,
	DeleteAllSchoolsPending244,
	RemoveAllSchoolsPublished245,
	DeleteAllSchoolsArchived246,
	EditMySchoolsPublished247,
	EditAllSchoolsPublished249,

	DeleteMyPending251,
	DeleteOrgPendingContent252,
	DeleteOrgArchivedContent253,
	RemoveOrgPublishedContent254,

	ApprovePendingContent271,
	RejectPendingContent272,
	RepublishArchivedContent274,
	CreateFolder289,

	FullContentMmanagement294,

	PublishFeaturedContentForAllHub,
	PublishFeaturedContentForAllOrgsHub,
	PublishFeaturedContentForSpecificOrgsHub,

	CreateAsset320,
	DeleteAsset340,

	ScheduleViewOrgCalendar,
	ScheduleViewSchoolCalendar,
	ScheduleViewMyCalendar,
	ScheduleViewPendingCalendar,
	ScheduleCreateEvent,
	ScheduleCreateMyEvent,
	ScheduleCreateMySchoolEvent,
	ScheduleCreateReviewEvent,
	ScheduleCreateLiveCalendarEvents,
	ScheduleCreateClassCalendarEvents,
	ScheduleCreateStudyCalendarEvents,
	ScheduleCreateHomefunCalendarEvents,

	ScheduleDeleteEvent,
	LiveClassTeacher,
	LiveClassStudent,

	ViewMyUnpublishedLearningOutcome,
	ViewOrgUnpublishedLearningOutcome,
	ViewMyPendingLearningOutcome,
	ViewOrgPendingLearningOutcome,
	ViewPublishedLearningOutcome,
	CreateLearningOutcome,
	EditMyUnpublishedLearningOutcome,
	EditOrgUnpublishedLearningOutcome,
	EditPublishedLearningOutcome,
	DeleteMyUnpublishedLearningOutcome,
	DeleteOrgUnpublishedLearningOutcome,
	DeleteMyPendingLearningOutcome,
	DeleteOrgPendingLearningOutcome,
	DeletePublishedLearningOutcome,
	ApprovePendingLearningOutcome,
	RejectPendingLearningOutcome,

	ReportTeacherReports603,
	ReportViewReports610,

	ReportViewMySchoolReports611,
	ReportSchoolsSkillsTaught641,
	ReportSchoolsClassAchievements647,

	ReportViewMyOrganizationsReports612,
	ReportOrganizationsSkillsTaught640,
	ReportOrganizationsClassAchievements646,

	ReportViewMyReports614,
	ReportMySkillsTaught642,
	ReportMyClassAchievments648,

	AssessmentViewCompletedAssessments414,
	AssessmentViewInProgressAssessments415,
	AssessmentEditInProgressAssessment439,
	AssessmentViewOrgCompletedAssessments424,
	AssessmentViewOrgInProgressAssessments425,
	AssessmentViewSchoolCompletedAssessments426,
	AssessmentViewSchoolInProgressAssessments427,
	AssessmentViewTeacherFeedback670,

	ViewUnPublishedMilestone,
	ViewPublishedMilestone,
	CreateMilestone,
	ViewMyUnpublishedMilestone,
	ViewMyPendingMilestone,
	EditUnpublishedMilestone,
	EditPublishedMilestone,
	DeleteUnpublishedMilestone,
	DeletePublishedMilestone,
	ViewPendingMilestone,
	EditMyUnpublishedMilestone,
	DeleteMyUnpublishedMilestone,
	DeleteOrgPendingMilestone,
	DeleteMyPendingMilestone,
	ApprovePendingMilestone,
	RejectPendingMilestone,

	LearningSummaryReport,
	ReportLearningSummaryStudent,
	ReportLearningSummaryTeacher,
	ReportLearningSummarySchool,
	ReportLearningSummmaryOrg,
	ReportStudentProgressReportView,
	ReportStudentProgressReportOrganization,
	ReportStudentProgressReportSchool,
	ReportStudentProgressReportTeacher,
	ReportStudentProgressReportStudent,
	ReportOrganizationStudentUsage,
	ReportSchoolStudentUsage,
	ReportTeacherStudentUsage,

	ViewProgram20111,
	ViewSubjects20115,

	ShowAllFolder295,
}

// Important: If you add a new permission, you must also add it to the AllPermissionNames
const (
	CreateContentPage201 PermissionName = "create_content_page_201"

	ViewMyPending212         PermissionName = "view_my_pending_212"
	ViewOrgPending213        PermissionName = "view_org_pending_213"
	ViewMyPublished214       PermissionName = "view_my_published_214"
	ViewOrgPublished215      PermissionName = "view_org_published_215"
	ViewMyArchived216        PermissionName = "view_my_archived_216"
	ViewOrgArchived217       PermissionName = "view_org_archived_217"
	ViewMySchoolPublished218 PermissionName = "view_my_school_published_218"

	CreateLessonMaterial220    PermissionName = "create_lesson_material_220"
	CreateLessonPlan221        PermissionName = "create_lesson_plan_221"
	CreateMySchoolsContent223  PermissionName = "create_my_schools_content_223"
	CreateAllSchoolsContent224 PermissionName = "create_all_schools_content_224"
	ViewMySchoolPending225     PermissionName = "view_my_school_pending_225"
	ViewMySchoolArchived226    PermissionName = "view_my_school_archived_226"
	ViewAllSchoolsPublished227 PermissionName = "view_all_schools_published_227"
	ViewAllSchoolsPending228   PermissionName = "view_all_schools_pending_228"
	ViewAllSchoolsArchived229  PermissionName = "view_all_schools_archived_229"

	EditMyUnpublishedContent230             PermissionName = "edit_my_unpublished_content_230"
	EditMyPublishedContent234               PermissionName = "edit_my_published_content_234"
	EditOrgPublishedContent235              PermissionName = "edit_org_published_content_235"
	EditLessonMaterialMetadataAndContent236 PermissionName = "edit_lesson_material_metadata_and_content_236"
	EditLessonPlanMetadata237               PermissionName = "edit_lesson_plan_metadata_237"
	EditLessonPlanContent238                PermissionName = "edit_lesson_plan_content_238"

	DeleteMyUnpublishedContent240 PermissionName = "delete_my_unpublished_content_240"
	DeleteMySchoolsPending241     PermissionName = "delete_my_schools_pending_241"
	RemoveMySchoolsPublished242   PermissionName = "remove_my_schools_published_242"
	DeleteMySchoolsArchived243    PermissionName = "delete_my_schools_archived_243"
	DeleteAllSchoolsPending244    PermissionName = "delete_all_schools_pending_244"
	RemoveAllSchoolsPublished245  PermissionName = "remove_all_schools_published_245"
	DeleteAllSchoolsArchived246   PermissionName = "delete_all_schools_archived_246"
	EditMySchoolsPublished247     PermissionName = "edit_my_schools_published_247"
	EditAllSchoolsPublished249    PermissionName = "edit_all_schools_published_249"

	DeleteMyPending251           PermissionName = "delete_my_pending_251"
	DeleteOrgPendingContent252   PermissionName = "delete_org_pending_content_252"
	DeleteOrgArchivedContent253  PermissionName = "delete_org_archived_content_253"
	RemoveOrgPublishedContent254 PermissionName = "remove_org_published_content_254"

	ApprovePendingContent271    PermissionName = "approve_pending_content_271"
	RejectPendingContent272     PermissionName = "reject_pending_content_272"
	RepublishArchivedContent274 PermissionName = "republish_archived_content_274"

	CreateFolder289 PermissionName = "create_folder_289"

	FullContentMmanagement294 PermissionName = "full_content_management_294"

	PublishFeaturedContentForAllHub          PermissionName = "publish_featured_content_for_all_hub_79000"
	PublishFeaturedContentForAllOrgsHub      PermissionName = "publish_featured_content_for_all_orgs_79002"
	PublishFeaturedContentForSpecificOrgsHub PermissionName = "publish_featured_content_for_specific_orgs_79001"

	CreateAsset320 PermissionName = "create_asset_320"
	DeleteAsset340 PermissionName = "delete_asset_340"

	ScheduleViewOrgCalendar             PermissionName = "view_org_calendar_511"
	ScheduleViewSchoolCalendar          PermissionName = "view_school_calendar_512"
	ScheduleViewMyCalendar              PermissionName = "view_my_calendar_510"
	ScheduleViewPendingCalendar         PermissionName = "view_pending_calendar_events_513"
	ScheduleCreateEvent                 PermissionName = "create_event_520"
	ScheduleCreateMyEvent               PermissionName = "create_my_schedule_events_521"
	ScheduleCreateMySchoolEvent         PermissionName = "create_my_schools_schedule_events_522"
	ScheduleCreateReviewEvent           PermissionName = "create_review_calendar_events_523"
	ScheduleCreateLiveCalendarEvents    PermissionName = "create_live_calendar_events_524"
	ScheduleCreateClassCalendarEvents   PermissionName = "create_class_calendar_events_525"
	ScheduleCreateStudyCalendarEvents   PermissionName = "create_study_calendar_events_526"
	ScheduleCreateHomefunCalendarEvents PermissionName = "create_home_fun_calendar_events_527"

	ScheduleDeleteEvent PermissionName = "delete_event_540"
	LiveClassTeacher    PermissionName = "attend_live_class_as_a_teacher_186"
	LiveClassStudent    PermissionName = "attend_live_class_as_a_student_187"

	ViewMyUnpublishedLearningOutcome    PermissionName = "view_my_unpublished_learning_outcome_410"
	ViewOrgUnpublishedLearningOutcome   PermissionName = "view_org_unpublished_learning_outcome_411"
	ViewMyPendingLearningOutcome        PermissionName = "view_my_pending_learning_outcome_412"
	ViewOrgPendingLearningOutcome       PermissionName = "view_org_pending_learning_outcome_413"
	ViewPublishedLearningOutcome        PermissionName = "view_published_learning_outcome_416"
	CreateLearningOutcome               PermissionName = "create_learning_outcome_421"
	EditMyUnpublishedLearningOutcome    PermissionName = "edit_my_unpublished_learning_outcome_430"
	EditOrgUnpublishedLearningOutcome   PermissionName = "edit_org_unpublished_learning_outcome_431"
	EditPublishedLearningOutcome        PermissionName = "edit_published_learning_outcome_436"
	DeleteMyUnpublishedLearningOutcome  PermissionName = "delete_my_unpublished_learning_outcome_444"
	DeleteOrgUnpublishedLearningOutcome PermissionName = "delete_org_unpublished_learning_outcome_445"
	DeleteMyPendingLearningOutcome      PermissionName = "delete_my_pending_learning_outcome_446"
	DeleteOrgPendingLearningOutcome     PermissionName = "delete_org_pending_learning_outcome_447"
	DeletePublishedLearningOutcome      PermissionName = "delete_published_learning_outcome_448"
	ApprovePendingLearningOutcome       PermissionName = "approve_pending_learning_outcome_481"
	RejectPendingLearningOutcome        PermissionName = "reject_pending_learning_outcome_482"

	ReportTeacherReports603 PermissionName = "teacher_reports_603"
	ReportViewReports610    PermissionName = "view_reports_610"

	ReportViewMySchoolReports611      PermissionName = "view_my_school_reports_611"
	ReportSchoolsSkillsTaught641      PermissionName = "report_schools_skills_taught_641"
	ReportSchoolsClassAchievements647 PermissionName = "report_schools_class_achievements_647"

	ReportViewMyOrganizationsReports612     PermissionName = "view_my_organizations_reports_612"
	ReportOrganizationsSkillsTaught640      PermissionName = "report_organizations_skills_taught_640"
	ReportOrganizationsClassAchievements646 PermissionName = "report_organizations_class_achievements_646"

	ReportViewMyReports614      PermissionName = "view_my_reports_614"
	ReportMySkillsTaught642     PermissionName = "report_my_skills_taught_642"
	ReportMyClassAchievments648 PermissionName = "report_my_class_achievments_648"

	AssessmentViewCompletedAssessments414        PermissionName = "view_completed_assessments_414"
	AssessmentViewInProgressAssessments415       PermissionName = "view_in_progress_assessments_415"
	AssessmentEditInProgressAssessment439        PermissionName = "edit_in_progress_assessment_439"
	AssessmentViewOrgCompletedAssessments424     PermissionName = "view_org_completed_assessments_424"
	AssessmentViewOrgInProgressAssessments425    PermissionName = "view_org_in_progress_assessments_425"
	AssessmentViewSchoolCompletedAssessments426  PermissionName = "view_school_completed_assessments_426"
	AssessmentViewSchoolInProgressAssessments427 PermissionName = "view_school_in_progress_assessments_427"
	AssessmentViewTeacherFeedback670             PermissionName = "view_teacher_feedback_670"

	ViewUnPublishedMilestone     PermissionName = "view_unpublished_milestone_417"
	ViewPublishedMilestone       PermissionName = "view_published_milestone_418"
	CreateMilestone              PermissionName = "create_milestone_422"
	ViewMyUnpublishedMilestone   PermissionName = "view_my_unpublished_milestone_428"
	ViewMyPendingMilestone       PermissionName = "view_my_pending_milestone_429"
	EditUnpublishedMilestone     PermissionName = "edit_unpublished_milestone_440"
	EditPublishedMilestone       PermissionName = "edit_published_milestone_441"
	DeleteUnpublishedMilestone   PermissionName = "delete_unpublished_milestone_449"
	DeletePublishedMilestone     PermissionName = "delete_published_milestone_450"
	ViewPendingMilestone         PermissionName = "view_pending_milestone_486"
	EditMyUnpublishedMilestone   PermissionName = "edit_my_unpublished_milestone_487"
	DeleteMyUnpublishedMilestone PermissionName = "delete_my_unpublished_milestone_488"
	DeleteOrgPendingMilestone    PermissionName = "delete_org_pending_milestone_489"
	DeleteMyPendingMilestone     PermissionName = "delete_my_pending_milestone_490"
	ApprovePendingMilestone      PermissionName = "approve_pending_milestone_491"
	RejectPendingMilestone       PermissionName = "reject_pending_milestone_492"

	LearningSummaryReport                   PermissionName = "learning_summary_report_653"
	ReportLearningSummaryStudent            PermissionName = "report_learning_summary_student_649"
	ReportLearningSummaryTeacher            PermissionName = "report_learning_summary_teacher_650"
	ReportLearningSummarySchool             PermissionName = "report_learning_summary_school_651"
	ReportLearningSummmaryOrg               PermissionName = "report_learning_summary_org_652"
	ReportStudentProgressReportView         PermissionName = "student_progress_report_662"
	ReportStudentProgressReportOrganization PermissionName = "report_student_progress_organization_658"
	ReportStudentProgressReportSchool       PermissionName = "report_student_progress_school_659"
	ReportStudentProgressReportTeacher      PermissionName = "report_student_progress_teacher_660"
	ReportStudentProgressReportStudent      PermissionName = "report_student_progress_student_661"
	ReportOrganizationStudentUsage          PermissionName = "report_organization_student_usage_654"
	ReportSchoolStudentUsage                PermissionName = "report_school_student_usage_655"
	ReportTeacherStudentUsage               PermissionName = "report_teacher_student_usage_656"

	ViewProgram20111  PermissionName = "view_program_20111"
	ViewSubjects20115 PermissionName = "view_subjects_20115"

	ShowAllFolder295 = "show_all_folders_295"
)

type TeacherViewPermissionParams struct {
	ViewSchoolReports PermissionName
	ViewOrgReports    PermissionName
	ViewMyReports     PermissionName
}
