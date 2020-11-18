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

	ViewLearningOutcomePage PermissionName = "learning_outcome_page_404"
	ViewMyUnpublishedLearningOutcome PermissionName = "view_my_unpublished_learning_outcome_410"
	ViewOrgUnpublishedLearningOutcome PermissionName = "view_org_unpublished_learning_outcome_411"
	ViewMyPendingLearningOutcome PermissionName = "view_my_pending_learning_outcome_412"
	ViewOrgPendingLearningOutcome PermissionName = "view_org_pending_learning_outcome_413"
	ViewPublishedLearningOutcome PermissionName = "view_published_learning_outcome_416"
	CreateLearningOutcome PermissionName = "create_learning_outcome_421"
	EditMyUnpublishedLearningOutcome PermissionName = "edit_my_unpublished_learning_outcome_430"
	EditOrgUnpublishedLearningOutcome PermissionName = "edit_org_unpublished_learning_outcome_431"
	EditPublishedLearningOutcome PermissionName = "edit_published_learning_outcome_436"
	DeleteMyUnpublishedLearningOutcome PermissionName = "delete_my_unpublished_learning_outcome_444"
	DeleteOrgUnpublishedLearningOutcome PermissionName = "delete_org_unpublished_learning_outcome_445"
	DeleteMyPendingLearningOutcome PermissionName = "delete_my_pending_learning_outcome_446"
	DeleteOrgPendingLearningOutcome PermissionName = "delete_org_pending_learning_outcome_447"
	DeletePublishedLearningOutcome PermissionName = "delete_published_learning_outcome_448"
	ApprovePendingLearningOutcome PermissionName = "approve_pending_learning_outcome_481"
	RejectPendingLearningOutcome PermissionName = "reject_pending_learning_outcome_482"
)
