package da

import (
	"sync"
)

type IReportDA interface {
	DataAccessor
	ITeacherLoadAssessment
	ITeacherLoadLesson
	IStudentProgressAssignment
	IStudentProgressLearnOutcomeAchievement
	IClassAttendance
	ILearnOutcome
	ILearnerWeekly
}
type ReportDA struct {
	BaseDA
}

var _reportDA *ReportDA
var _reportDAOnce sync.Once

func GetReportDA() IReportDA {
	_reportDAOnce.Do(func() {
		_reportDA = new(ReportDA)
	})
	return _reportDA
}
