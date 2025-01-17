package da

import (
	"context"
	"sync"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
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
	tx.ResetCondition()

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
	tx.ResetCondition()

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

func (as *assessmentAttendanceDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentAttendance) error {
	if len(items) == 0 {
		return nil
	}
	var models []entity.AssessmentAttendance
	for _, item := range items {
		if item.ID == "" {
			item.ID = utils.NewID()
		}
		models = append(models, entity.AssessmentAttendance{ID: item.ID, AssessmentID: item.AssessmentID,
			AttendanceID: item.AttendanceID, Checked: item.Checked, Origin: item.Origin, Role: item.Role})
	}
	_, err := as.InsertTx(ctx, tx, &models)
	if err != nil {
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
	tx.ResetCondition()

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
	tx.ResetCondition()

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
	AssessmentIDs entity.NullStrings
	Role          entity.NullAssessmentAttendanceRole
	Checked       entity.NullBool
	AttendanceID  entity.NullString
}

func (c *QueryAssessmentAttendanceConditions) GetConditions() ([]string, []interface{}) {
	t := NewSQLTemplate("")
	if c.AssessmentIDs.Valid {
		t.Appendf("assessment_id in (?)", c.AssessmentIDs.Strings)
	}
	if c.Role.Valid {
		t.Appendf("role = ?", c.Role.Value)
	}
	if c.Checked.Valid {
		t.Appendf("checked = ?", c.Checked.Bool)
	}
	if c.AttendanceID.Valid {
		t.Appendf("attendance_id = ?", c.AttendanceID.String)
	}
	return t.DBOConditions()
}

func (c *QueryAssessmentAttendanceConditions) GetPager() *dbo.Pager { return nil }

func (c *QueryAssessmentAttendanceConditions) GetOrderBy() string { return "" }
