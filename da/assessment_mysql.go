package da

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type IAssessmentDA interface {
	dbo.DataAccesser
	GetExcludeSoftDeleted(ctx context.Context, tx *dbo.DBContext, id string) (*entity.Assessment, error)
	UpdateStatus(ctx context.Context, tx *dbo.DBContext, id string, status entity.AssessmentStatus) error
	SoftDelete(ctx context.Context, tx *dbo.DBContext, ids string) error
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
	cacheResult, err := GetAssessmentRedisDA().Item(ctx, id)
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
		return nil, err
	}

	if err := GetAssessmentRedisDA().CacheItem(ctx, id, &item); err != nil {
		log.Info(ctx, "get assessment exclude soft deleted: cache item failed",
			log.Err(err),
			log.String("id", id),
		)
	}

	return &item, nil
}

func (a *assessmentDA) UpdateStatus(ctx context.Context, tx *dbo.DBContext, id string, status entity.AssessmentStatus) error {
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

	if err := GetAssessmentRedisDA().CleanItem(ctx, id); err != nil {
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

func (a *assessmentDA) SoftDelete(ctx context.Context, tx *dbo.DBContext, id string) error {
	if err := tx.Where(a.filterSoftDeletedTemplate()).Update("delete_at", time.Now().Unix()).Error; err != nil {
		log.Error(ctx, "SoftDelete: update failed",
			log.Err(err),
			log.String("id", id),
		)
		return err
	}
	return nil
}

func (a *assessmentDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.Assessment) error {
	if len(items) == 0 {
		return nil
	}
	columns := []string{
		"id",
		"schedule_id",
		"`type`",
		"title",
		"complete_time",
		"status",
		"create_at",
		"update_at",
		"delete_at",
		"class_length",
		"class_end_time",
	}
	var matrix [][]interface{}
	for _, item := range items {
		if item.ID == "" {
			item.ID = utils.NewID()
		}
		matrix = append(matrix, []interface{}{
			item.ID,
			item.ScheduleID,
			item.Type,
			item.CompleteTime,
			item.Status,
			item.CreateAt,
			item.UpdateAt,
			item.DeleteAt,
			item.ClassLength,
			item.ClassEndTime,
		})
	}
	format, values := SQLBatchInsert(entity.Assessment{}.TableName(), columns, matrix)
	if err := tx.Exec(format, values...).Error; err != nil {
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
	Type                         entity.NullAssessmentType                         `json:"type"`
	OrgID                        entity.NullString                                 `json:"org_id"`
	Status                       entity.NullAssessmentStatus                       `json:"status"`
	ScheduleIDs                  entity.NullStrings                                `json:"schedule_ids"`
	TeacherIDs                   entity.NullStrings                                `json:"teacher_ids"`
	AllowTeacherIDs              entity.NullStrings                                `json:"allow_teacher_ids"`
	AllowTeacherIDAndStatusPairs entity.NullAssessmentAllowTeacherIDAndStatusPairs `json:"teacher_id_and_status_pairs"`
	ClassType                    entity.NullScheduleClassType                      `json:"class_type"`
	ClassIDs                     entity.NullStrings                                `json:"class_ids"`
	ClassIDsOrTeacherIDs         NullClassIDsOrTeacherIDs                          `json:"class_ids_or_teacher_ids"`

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

	if c.Type.Valid && c.Type.Value.Valid() {
		t.Appendf("`type` = ?", c.Type.Value)
	}

	if c.OrgID.Valid {
		t.Appendf("exists (select 1 from schedules"+
			" where org_id = ? and delete_at = 0 and assessments.schedule_id = schedules.id)", c.OrgID.String)
	}

	if c.ClassType.Valid {
		t.Appendf("exists (select 1 from schedules"+
			" where class_type = ? and delete_at = 0 and assessments.schedule_id = schedules.id)", c.OrgID.String)
	}

	if c.Status.Valid {
		t.Appendf("status = ?", c.Status.Value)
	}

	if c.TeacherIDs.Valid {
		t.Appendf("exists (select 1 from assessments_attendances"+
			" where assessments.id = assessments_attendances.assessment_id and role = 'teacher' and attendance_id in (?))",
			utils.SliceDeduplication(c.TeacherIDs.Strings))
	}

	if c.AllowTeacherIDs.Valid {
		t.Appendf("exists (select 1 from assessments_attendances"+
			" where assessments.id = assessments_attendances.assessment_id and role = 'teacher' and attendance_id in (?))",
			utils.SliceDeduplication(c.AllowTeacherIDs.Strings))
	}

	if c.AllowTeacherIDAndStatusPairs.Valid {
		t2 := NewSQLTemplate("")
		for _, p := range c.AllowTeacherIDAndStatusPairs.Values {
			t2.Appendf("(attendance_id = ? and assessments.status = ?)", p.TeacherID, p.Status)
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
			" where assessments.id = assessments_attendances.assessment_id and role = 'teacher' and attendance_id in (?))",
			utils.SliceDeduplication(c.TeacherIDs.Strings))
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
	case entity.ListAssessmentsOrderByClassEndTime:
		return "class_end_time"
	case entity.ListAssessmentsOrderByClassEndTimeDesc:
		return "class_end_time desc"
	case entity.ListAssessmentsOrderByCompleteTime:
		return "complete_time"
	case entity.ListAssessmentsOrderByCompleteTimeDesc:
		return "complete_time desc"
	}
	return ""
}
