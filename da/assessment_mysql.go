package da

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type IAssessmentDA interface {
	dbo.DataAccesser
	GetExcludeSoftDeleted(ctx context.Context, tx *dbo.DBContext, id string) (*entity.Assessment, error)
	UpdateStatus(ctx context.Context, tx *dbo.DBContext, id string, status entity.AssessmentStatus) error
	BatchSoftDelete(ctx context.Context, tx *dbo.DBContext, ids []string) error
	BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.Assessment) error
}

var (
	assessmentDAInstance     IAssessmentDA
	assessmentDAInstanceOnce = sync.Once{}
)

func GetAssessmentDA() IAssessmentDA {
	assessmentDAInstanceOnce.Do(func() {
		assessmentDAInstance = &assessmentDA{}
	})
	return assessmentDAInstance
}

type assessmentDA struct {
	dbo.BaseDA
}

func (a *assessmentDA) GetExcludeSoftDeleted(ctx context.Context, tx *dbo.DBContext, id string) (*entity.Assessment, error) {
	tx.ResetCondition()

	cacheResult, err := GetAssessmentRedisDA().GetAssessment(ctx, id)
	if err != nil {
		log.Info(ctx, "get assessment exclude soft deleted: get failed from cache",
			log.Err(err),
			log.String("id", id),
		)
	} else if cacheResult != nil {
		log.Info(ctx, "get assessment exclude soft deleted: hit cache")
		return cacheResult, nil
	}

	item := entity.Assessment{}
	if err := tx.Model(entity.Assessment{}).
		Where(a.filterSoftDeletedTemplate()).
		Where("id = ?", id).
		First(&item).
		Error; err != nil {
		log.Error(ctx, "get assessment: get from db failed",
			log.Err(err),
			log.String("id", id),
		)
		if gorm.IsRecordNotFoundError(err) {
			return nil, constant.ErrRecordNotFound
		}
		return nil, err
	}

	if err := GetAssessmentRedisDA().SetAssessment(ctx, id, &item); err != nil {
		log.Warn(ctx, "get assessment exclude soft deleted: cache item failed",
			log.Err(err),
			log.String("id", id),
		)
	}

	return &item, nil
}

func (a *assessmentDA) UpdateStatus(ctx context.Context, tx *dbo.DBContext, id string, status entity.AssessmentStatus) error {
	tx.ResetCondition()

	nowUnix := time.Now().Unix()
	updateFields := map[string]interface{}{
		"status":    status,
		"update_at": nowUnix,
	}
	if status == entity.AssessmentStatusComplete {
		updateFields["complete_time"] = nowUnix
	}
	if err := tx.Model(entity.Assessment{}).
		Where(a.filterSoftDeletedTemplate()).
		Where("id = ?", id).
		Updates(updateFields).
		Error; err != nil {
		log.Error(ctx, "update assessment status: update failed in db",
			log.Err(err),
			log.String("id", id),
			log.String("status", string(status)),
		)
		return err
	}

	if err := GetAssessmentRedisDA().CleanAssessment(ctx, id); err != nil {
		log.Info(ctx, "update assessment status: clean item cache failed",
			log.Err(err),
			log.String("id", id),
			log.String("status", string(status)),
		)
	}

	return nil
}

func (a *assessmentDA) filterSoftDeletedTemplate() string {
	return "delete_at = 0"
}

func (a *assessmentDA) BatchSoftDelete(ctx context.Context, tx *dbo.DBContext, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	tx.ResetCondition()
	if err := tx.Model(entity.Assessment{}).Where(a.filterSoftDeletedTemplate()).
		Where("id in (?)", ids).
		Update("delete_at", time.Now().Unix()).Error; err != nil {
		log.Error(ctx, "BatchSoftDelete: update failed",
			log.Err(err),
			log.Strings("ids", ids),
		)
		return err
	}
	return nil
}

func (a *assessmentDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.Assessment) error {
	if len(items) == 0 {
		return nil
	}
	_, err := a.InsertTx(ctx, tx, &items)
	if err != nil {
		log.Error(ctx, "BatchInsert: batch insert failed",
			log.Err(err),
			log.Any("items", items),
		)
		return err
	}
	return nil
}

type QueryAssessmentConditions struct {
	IDs                          entity.NullStrings                                `json:"ids"`
	OrgID                        entity.NullString                                 `json:"org_id"`
	Status                       entity.NullAssessmentStatus                       `json:"status"`
	ScheduleIDs                  entity.NullStrings                                `json:"schedule_ids"`
	StudentIDs                   entity.NullStrings                                `json:"student_ids"`
	TeacherIDs                   entity.NullStrings                                `json:"teacher_ids"`
	AllowTeacherIDs              entity.NullStrings                                `json:"allow_teacher_ids"`
	AllowTeacherIDAndStatusPairs entity.NullAssessmentAllowTeacherIDAndStatusPairs `json:"teacher_id_and_status_pairs"`
	ClassTypes                   entity.NullScheduleClassTypes                     `json:"class_types"`
	ClassIDs                     entity.NullStrings                                `json:"class_ids"`
	ClassIDsOrTeacherIDs         NullClassIDsOrTeacherIDs                          `json:"class_ids_or_teacher_ids"`

	CreatedBetween  entity.NullTimeRange `json:"created_between"`
	UpdateBetween   entity.NullTimeRange `json:"update_between"`
	CompleteBetween entity.NullTimeRange `json:"complete_between"`
	ClassType       entity.NullString    `json:"class_type"`

	OrderBy entity.NullAssessmentsOrderBy `json:"order_by"`
	Pager   dbo.Pager                     `json:"pager"`
}

type ClassIDsOrTeacherIDs struct {
	ClassIDs   []string `json:"class_ids"`
	TeacherIDs []string `json:"teacher_ids"`
}

type NullClassIDsOrTeacherIDs struct {
	Value ClassIDsOrTeacherIDs
	Valid bool
}

func (c *QueryAssessmentConditions) GetConditions() ([]string, []interface{}) {
	t := NewSQLTemplate("delete_at = 0")

	if c.IDs.Valid {
		t.Appendf("id in (?)", c.IDs.Strings)
	}

	if c.OrgID.Valid {
		t.Appendf("exists (select 1 from schedules"+
			" where org_id = ? and delete_at = 0 and assessments.schedule_id = schedules.id)", c.OrgID.String)
	}

	if c.ClassTypes.Valid {
		t.Appendf("exists (select 1 from schedules"+
			" where class_type in (?) and delete_at = 0 and assessments.schedule_id = schedules.id)", c.ClassTypes.Value)
	}

	if c.Status.Valid {
		t.Appendf("status = ?", c.Status.Value)
	}

	if c.TeacherIDs.Valid {
		t.Appendf("exists (select 1 from assessments_attendances"+
			" where assessments.id = assessments_attendances.assessment_id and role = 'teacher' and attendance_id in (?))",
			utils.SliceDeduplication(c.TeacherIDs.Strings))
	}

	if c.CreatedBetween.Valid {
		t.Appendf("(create_at BETWEEN ? AND ?)", c.CreatedBetween.StartAt, c.CreatedBetween.EndAt)
	}
	if c.UpdateBetween.Valid {
		t.Appendf("(update_at BETWEEN ? AND ?)", c.UpdateBetween.StartAt, c.UpdateBetween.EndAt)
	}
	if c.CompleteBetween.Valid {
		t.Appendf("(complete_time BETWEEN ? AND ?)", c.CompleteBetween.StartAt, c.CompleteBetween.EndAt)
	}
	if c.ClassType.Valid {
		t.Appendf("exists (select 1 from schedules"+
			" where assessments.schedule_id = schedules.id and class_type = ? and is_home_fun=false)",
			c.ClassType.String)
	}

	if c.StudentIDs.Valid {
		t.Appendf("exists (select 1 from assessments_attendances"+
			" where assessments.id = assessments_attendances.assessment_id and role = 'student' and attendance_id in (?))",
			utils.SliceDeduplication(c.StudentIDs.Strings))
	}

	if c.AllowTeacherIDs.Valid {
		t.Appendf("exists (select 1 from assessments_attendances"+
			" where assessments.id = assessments_attendances.assessment_id and role = 'teacher' and attendance_id in (?))",
			utils.SliceDeduplication(c.AllowTeacherIDs.Strings))
	}

	if c.AllowTeacherIDAndStatusPairs.Valid {
		t2 := NewSQLTemplate("")
		for _, p := range c.AllowTeacherIDAndStatusPairs.Values {
			t2.Appendf("(attendance_id = ? and assessments.status in (?))", p.TeacherID, p.Status)
		}
		format, values := t2.Or()
		format = fmt.Sprintf("exists (select 1 from assessments_attendances "+
			"where assessments.id = assessments_attendances.assessment_id and role = 'teacher' and %s)", format)
		t.Appendf(format, values...)
	}

	if c.ScheduleIDs.Valid {
		t.Appendf("schedule_id in (?)", c.ScheduleIDs.Strings)
	}

	if c.ClassIDs.Valid {
		t.Appendf("exists (select 1 from schedules"+
			" where class_id in (?) and delete_at = 0 and assessments.schedule_id = schedules.id)", c.ClassIDs.Strings)
	}

	if c.ClassIDsOrTeacherIDs.Valid {
		t2 := NewSQLTemplate("")
		t2.Appendf("exists (select 1 from schedules"+
			" where class_id in (?) and delete_at = 0 and assessments.schedule_id = schedules.id)", c.ClassIDs.Strings)
		t2.Appendf("exists (select 1 from assessments_attendances"+
			" where assessments.id = assessments_attendances.assessment_id and role = ? and attendance_id in (?))",
			entity.AssessmentAttendanceRoleTeacher, utils.SliceDeduplication(c.TeacherIDs.Strings))
		t.AppendResult(t2.Or())
	}

	return t.DBOConditions()
}

func (c *QueryAssessmentConditions) GetPager() *dbo.Pager {
	return &c.Pager
}

func (c *QueryAssessmentConditions) GetOrderBy() string {
	if !c.OrderBy.Valid {
		return ""
	}
	switch c.OrderBy.Value {
	case entity.AssessmentOrderByClassEndTime:
		return "class_end_time"
	case entity.AssessmentOrderByClassEndTimeDesc:
		return "class_end_time desc"
	case entity.AssessmentOrderByCompleteTime:
		return "status desc, complete_time, create_at"
	case entity.AssessmentOrderByCompleteTimeDesc:
		return "status desc, complete_time desc, create_at desc"
	case entity.AssessmentOrderByCreateAt:
		return "create_at"
	case entity.AssessmentOrderByCreateAtDesc:
		return "create_at desc"
	case entity.AssessmentOrderByUpdateAt:
		return "update_at"
	case entity.AssessmentOrderByUpdateAtDesc:
		return "update_at desc"
	}
	return ""
}
