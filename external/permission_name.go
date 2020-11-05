package external

type PermissionName string

func (p PermissionName) String() string {
	return string(p)
}

const (
	CreateContentPage201 PermissionName = "create_content_page_201"

	ViewOrgCalendar PermissionName = "view_org_calendar__511"
	ViewMyCalendar  PermissionName = "view_my_calendar_510"
	CreateEvent     PermissionName = "create_event__520"
	EditEvent       PermissionName = "edit_event__530"
	DeleteEvent     PermissionName = "delete_event_540"
)
