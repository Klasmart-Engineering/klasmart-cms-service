package entity

type ScheduleFilterClass struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	HasStudentFlag bool   `json:"has_student_flag"`
}

type ScheduleFilterSchool struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
