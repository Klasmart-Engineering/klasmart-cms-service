package entity

type TeacherLoadLessonSummaryRequest TeacherLoadLessonSummaryArgs
type TeacherLoadLessonSummaryResponse TeacherLoadLessonSummaryRes

func (t TeacherLoadLessonSummaryRequest) Validate() (TeacherLoadLessonSummaryArgs, error) {
	return TeacherLoadLessonSummaryArgs(t), nil
}

type TeacherLoadLessonSummaryArgs struct {
	TeacherIDs []string  `json:"teacher_ids" form:"teacher_ids"`
	ClassIDs   []string  `json:"class_ids" form:"class_ids"`
	Duration   TimeRange `json:"duration" form:"duration"`
}

type TeacherLoadLesson struct {
	TeacherID               string `json:"teacher_id"`
	NumberOfClasses         int    `json:"number_of_classes"`
	NumberOfStudents        int    `json:"number_of_students"`
	CompletedLiveLessons    int    `json:"completed_live_Lessons"`
	CompletedInClassLessons int    `json:"completed_in_class_lessons"`
	MissedLiveLessons       int    `json:"missed_live_lessons"`
	MissedInClassLessons    int    `json:"missed_in_class_lessons"`
	TotalScheduled          int    `json:"total_scheduled"`
}

type TeacherLoadLessonSummary struct {
	CompletedLiveLessons    int `json:"completed_live_lessons"`
	CompletedInClassLessons int `json:"completed_in_class_lessons"`
	MissedLiveLessons       int `json:"missed_live_lessons"`
	MissedInClassLessons    int `json:"missed_in_class_lessons"`
}

type TeacherLoadLessonSummaryRes struct {
	List    []*TeacherLoadLesson     `json:"list"`
	Summary TeacherLoadLessonSummary `json:"summary"`
}
