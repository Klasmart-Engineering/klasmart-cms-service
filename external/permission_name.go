package external

type PermissionName string

func (p PermissionName) String() string {
	return string(p)
}

const (
	CreateContentPage201 PermissionName = "create_content_page_201"

	UnpublishedContentPage202               PermissionName = "unpublished_content_page_202"
	PendingContentPage203                   PermissionName = "pending_content_page_203"
	PublishedContentPage204                 PermissionName = "published_content_page_204"
	ArchivedContentPage205                  PermissionName = "archived_content_page_205"
	CreateLessonMaterial220                 PermissionName = "create_lesson_material_220"
	CreateLessonPlan221                     PermissionName = "create_lesson_plan_221"
	EditOrgPublishedContent235              PermissionName = "edit_org_published_content_235"
	EditLessonMaterialMetadataAndContent236 PermissionName = "edit_lesson_material_metadata_and_content_236"
	EditLessonPlanMetadata237               PermissionName = "edit_lesson_plan_metadata_237"
	EditLessonPlanContent238                PermissionName = "edit_lesson_plan_content_238"
	ApprovePendingContent271                PermissionName = "approve_pending_content_271"
	RejectPendingContent272                 PermissionName = "reject_pending_content_272"
	ArchivePublishedContent273              PermissionName = "archive_published_content_273"
	RepublishArchivedContent274             PermissionName = "republish_archived_content_274"
	DeleteArchivedContent275                PermissionName = "delete_archived_content_275"
	AssociateLearningOutcomes284            PermissionName = "associate_learning_outcomes_284"

	CreateAssetPage301 PermissionName = "create_asset_page_301"
	CreateAsset320     PermissionName = "create_asset_320"
	DeleteAsset340     PermissionName = "delete_asset_340"

	ScheduleViewOrgCalendar    PermissionName = "view_org_calendar__511"
	ScheduleViewSchoolCalendar PermissionName = "view_school_calendar_512"
	ScheduleViewMyCalendar     PermissionName = "view_my_calendar_510"
	ScheduleCreateEvent        PermissionName = "create_event__520"
	ScheduleEditEvent          PermissionName = "edit_event__530"
	ScheduleDeleteEvent        PermissionName = "delete_event_540"
	LiveClassTeacher           PermissionName = "attend_live_class_as_a_teacher_186"
	LiveClassStudent           PermissionName = "attend_live_class_as_a_student_187"

	ReportTeacherReports603 PermissionName = "teacher_reports_603"
	ReportViewReports610    PermissionName = "view_reports_610"
	ReportViewMyReports614  PermissionName = "view_my_reports_614"

	Assessments400                                     PermissionName = "assessments_400"
	AssessmentPage406                                  PermissionName = "assessments_page_406"
	AssessmentViewCompletedAssessments414              PermissionName = "view_completed_assessments_414"
	AssessmentViewInProgressAssessments415             PermissionName = "view_in_progress_assessments_415"
	AssessmentEditAttendanceForInProgressAssessment438 PermissionName = "edit_attendance_for_in_progress_assessment_438"
	AssessmentEditInProgressAssessment439              PermissionName = "edit_in_progress_assessment_439"
	AssessmentViewOrgCompletedAssessments424           PermissionName = "view_org_completed_assessments_424"
	AssessmentViewOrgInProgressAssessments425          PermissionName = "view_org_in_progress_assessments_425"
	AssessmentViewSchoolCompletedAssessments426        PermissionName = "view_school_completed_assessments_426"
	AssessmentViewSchoolInProgressAssessments427       PermissionName = "view_school_in_progress_assessments_427"
)
