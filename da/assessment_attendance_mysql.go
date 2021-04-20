package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type IAssessmentAttendanceDA interface {
	dbo.Querier
	GetTeacherIDsByAssessmentID(ctx context.Context, tx *dbo.DBContext, assessmentID string) ([]string, error)
	GetStudentIDsByAssessmentID(ctx context.Context, tx *dbo.DBContext, assessmentID string) ([]string, error)
	BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentAttendance) error
	UncheckStudents(ctx context.Context, db *dbo.DBContext, assessmentID string) error
	BatchCheck(ctx context.Context, db *dbo.DBContext, assessmentID string, attendanceIDs []string) error
}

var (
	assessmentAttendanceDAInstance     IAssessmentAttendanceDA
	assessmentAttendanceDAInstanceOnce = sync.Once{}
)

func GetAssessmentAttendanceDA() IAssessmentAttendanceDA {
	assessmentAttendanceDAInstanceOnce.Do(func() {
		assessmentAttendanceDAInstance = &assessmentAttendanceDA{}
	})
	return assessmentAttendanceDAInstance
}

type assessmentAttendanceDA struct {
	dbo.BaseDA
}

func (*assessmentAttendanceDA) GetTeacherIDsByAssessmentID(ctx context.Context, tx *dbo.DBContext, assessmentID string) ([]string, error) {
	var (
		items []*entity.AssessmentAttendance
		ids   []string
	)
	if err := tx.Where("assessment_id = ? and role = ?", assessmentID, entity.AssessmentAttendanceRoleTeacher).
		Find(&items).
		Error; err != nil {
		log.Error(ctx, "GetTeacherIDsByAssessmentID: Find: find failed",
			log.String("assessment_id", assessmentID),
		)
		return nil, err
	}
	for _, item := range items {
		ids = append(ids, item.AttendanceID)
	}
	return ids, nil
}

func (*assessmentAttendanceDA) GetStudentIDsByAssessmentID(ctx context.Context, tx *dbo.DBContext, assessmentID string) ([]string, error) {
	var (
		items []*entity.AssessmentAttendance
		ids   []string
	)
	if err := tx.Where("assessment_id = ? and role = ?", assessmentID, entity.AssessmentAttendanceRoleStudent).
		Find(&items).
		Error; err != nil {
		log.Error(ctx, "GetStudentIDsByAssessmentID: Find: find failed",
			log.String("assessment_id", assessmentID),
		)
		return nil, err
	}
	for _, item := range items {
		ids = append(ids, item.AttendanceID)
	}
	return ids, nil
}

func (*assessmentAttendanceDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentAttendance) error {
	if len(items) == 0 {
		return nil
	}
	columns := []string{"id", "assessment_id", "attendance_id", "checked", "origin", "role"}
	var matrix [][]interface{}
	for _, item := range items {
		if item.ID == "" {
			item.ID = utils.NewID()
		}
		matrix = append(matrix, []interface{}{item.ID, item.AssessmentID, item.AttendanceID, item.Checked, item.Origin, item.Role})
	}
	format, values := SQLBatchInsert(entity.AssessmentAttendance{}.TableName(), columns, matrix)
	if err := tx.Exec(format, values...).Error; err != nil {
		log.Error(ctx, "batch insert assessments_attendances: batch insert failed",
			log.Err(err),
			log.Any("items", items),
		)
		return err
	}
	return nil
}

// Uncheck assessment all students
func (a *assessmentAttendanceDA) UncheckStudents(ctx context.Context, tx *dbo.DBContext, assessmentID string) error {
	if err := tx.Model(&entity.AssessmentAttendance{}).Where("assessment_id = ? and role = ?", assessmentID, entity.AssessmentAttendanceRoleStudent).
		Update("checked", false).
		Error; err != nil {
		log.Error(ctx, "uncheck assessment attendance: update failed",
			log.Err(err),
			log.String("assessment_id", assessmentID),
		)
		return err
	}
	return nil
}

func (a *assessmentAttendanceDA) BatchCheck(ctx context.Context, tx *dbo.DBContext, assessmentID string, attendanceIDs []string) error {
	if len(attendanceIDs) == 0 {
		return nil
	}
	if err := tx.Model(&entity.AssessmentAttendance{}).Where("assessment_id = ? and attendance_id in (?)", assessmentID, attendanceIDs).
		Update("checked", true).
		Error; err != nil {
		log.Error(ctx, "uncheck assessment attendance: update failed",
			log.Err(err),
			log.String("assessment_id", assessmentID),
			log.Strings("attendance_ids", attendanceIDs),
		)
		return err
	}
	return nil
}

type QueryAssessmentAttendanceConditions struct {
	AssessmentIDs []string
	Role          *entity.AssessmentAttendanceRole
	Checked       *bool
}

func (c *QueryAssessmentAttendanceConditions) GetConditions() ([]string, []interface{}) {
	t := NewSQLTemplate("")
	if len(c.AssessmentIDs) > 0 {
		t.Appendf("assessment_id in (?)", c.AssessmentIDs)
	}
	if c.Role != nil {
		t.Appendf("role = ?", *c.Role)
	}
	if c.Checked != nil {
		t.Appendf("checked = ?", *c.Checked)
	}
	return t.DBOConditions()
}

func (c *QueryAssessmentAttendanceConditions) GetPager() *dbo.Pager { return nil }

func (c *QueryAssessmentAttendanceConditions) GetOrderBy() string { return "" }
