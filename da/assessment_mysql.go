package da

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
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
	BatchGetAssessmentsByScheduleIDs(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) ([]*entity.Assessment, error)
	FilterCompletedAssessmentIDs(ctx context.Context, tx *dbo.DBContext, ids []string) ([]string, error)
	FilterCompletedAssessments(ctx context.Context, tx *dbo.DBContext, ids []string) ([]*entity.Assessment, error)
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

type QueryAssessmentConditions struct {
	OrgID                   entity.NullString                            `json:"org_id"`
	Status                  entity.NullAssessmentStatus                  `json:"status"`
	ScheduleIDs             entity.NullStrings                           `json:"schedule_ids"`
	TeacherIDs              entity.NullStrings                           `json:"teacher_ids"`
	AllowTeacherIDs         entity.NullStrings                           `json:"allow_teacher_ids"`
	TeacherIDAndStatusPairs entity.NullAssessmentTeacherIDAndStatusPairs `json:"teacher_id_and_status_pairs"`
	ClassType               entity.NullScheduleClassType                 `json:"class_type"`
	OrderBy                 entity.NullListAssessmentsOrderBy            `json:"order_by"`
	Pager                   dbo.Pager                                    `json:"pager"`
}

func (c *QueryAssessmentConditions) GetConditions() ([]string, []interface{}) {
	t := NewSQLTemplate("delete_at = 0")

	if c.OrgID.Valid {
		t.Appendf("exists (select 1 from schedules"+
			" where org_id = ? and delete_at = 0 and assessments.schedule_id = schedules.id)", c.OrgID.Valid)
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

	if c.TeacherIDAndStatusPairs.Valid {
		t2 := NewSQLTemplate("")
		for _, p := range c.TeacherIDAndStatusPairs.Values {
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

func (a *assessmentDA) BatchGetAssessmentsByScheduleIDs(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) ([]*entity.Assessment, error) {
	var result []*entity.Assessment
	if err := tx.Where("schedule_id in (?)", scheduleIDs).Find(&result).Error; err != nil {
		log.Error(ctx, "batch get assessments by schedule ids: find failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
		)
		return nil, err
	}
	return result, nil
}

func (a *assessmentDA) FilterCompletedAssessmentIDs(ctx context.Context, tx *dbo.DBContext, ids []string) ([]string, error) {
	var items []struct {
		ID string `gorm:"column:id"`
	}
	if err := tx.Table(entity.Assessment{}.TableName()).
		Select("id").
		Where("id in (?) and status = ?", ids, entity.AssessmentStatusComplete).
		Find(&items).Error; err != nil {
		log.Error(ctx, "filter completed assessment ids failed",
			log.Err(err),
			log.Strings("ids", ids),
		)
		if gorm.IsRecordNotFoundError(err) {
			return nil, constant.ErrRecordNotFound
		}
		return nil, err
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.ID)
	}
	return result, nil
}

func (a *assessmentDA) FilterCompletedAssessments(ctx context.Context, tx *dbo.DBContext, ids []string) ([]*entity.Assessment, error) {
	var result []*entity.Assessment
	if err := tx.Model(entity.Assessment{}).
		Where("id in (?) and status = ?", ids, entity.AssessmentStatusComplete).
		Find(&result).Error; err != nil {
		log.Error(ctx, "call FilterCompletedAssessments failed",
			log.Err(err),
			log.Strings("ids", ids),
		)
		if gorm.IsRecordNotFoundError(err) {
			return nil, constant.ErrRecordNotFound
		}
		return nil, err
	}
	return result, nil
}
