package entity

import (
	"encoding/json"
	"testing"
)

func TestScheduleAddViewJson(t *testing.T) {
	scheduleJson := `
{"attachment_path":"","class_id":"3d10726b-9add-4c3c-9dc8-187d668fca6f","class_type":"Homework","description":"test 666","due_at":1616169600,"is_all_day":false,"is_force":true,"is_repeat":false,"lesson_plan_id":"","program_id":"04c630cc-fabe-4176-80f2-30a029907a33","repeat":{},"subject_id":"49c8d5ee-472b-47a6-8c57-58daf863c2e1","teacher_ids":["f2133261-ff0f-5101-9461-2e20f8206bd3","7eaec14a-39a7-5b1c-8a95-694c7b3ec73a","05257f55-dd23-5ec6-ae11-9cad1d9a1749","f42b2cd2-ff02-4871-a9e4-63c2d4b6cdd4","59fe6ce7-0bee-5713-b3cb-ea6edb4f2501"],"title":"test Name 666","start_at":1616169600,"end_at":1616169600,"attachment":{"id":"","name":""},"time_zone_offset":28800,"is_home_fun":true,"participants_student_ids":[],"participants_teacher_ids":[],"class_roster_student_ids":["f42b2cd2-ff02-4871-a9e4-63c2d4b6cdd4","59fe6ce7-0bee-5713-b3cb-ea6edb4f2501"],"class_roster_teacher_ids":["f2133261-ff0f-5101-9461-2e20f8206bd3","7eaec14a-39a7-5b1c-8a95-694c7b3ec73a","05257f55-dd23-5ec6-ae11-9cad1d9a1749"]}
`
	var scheduleObj = new(ScheduleAddView)
	err := json.Unmarshal([]byte(scheduleJson), scheduleObj)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(scheduleObj)
}

func TestJsonTemp(t *testing.T) {
	data := []*ScheduleFilterClass{
		{
			ID:               ScheduleFilterUndefinedClass,
			Name:             ScheduleFilterUndefinedClass,
			OperatorRoleType: ScheduleRoleTypeUnknown,
		},
		{
			ID:               "1",
			Name:             "班级一",
			OperatorRoleType: ScheduleRoleTypeUnknown,
		},
		{
			ID:               "2",
			Name:             "班级二",
			OperatorRoleType: ScheduleRoleTypeStudent,
		},
		{
			ID:               "3",
			Name:             "班级三",
			OperatorRoleType: ScheduleRoleTypeTeacher,
		},
		{
			ID:               "4",
			Name:             "班级四",
			OperatorRoleType: ScheduleRoleTypeStudent,
		},
	}
	b, _ := json.Marshal(data)
	t.Log(string(b))
}
