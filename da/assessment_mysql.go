package da

import (
	"context"
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
	OrgID                   *string                                    `json:"org_id"`
	Status                  *entity.AssessmentStatus                   `json:"status"`
	ScheduleIDs             []string                                   `json:"schedule_ids"`
	TeacherIDs              []string                                   `json:"teacher_ids"`
	AllowTeacherIDs         []string                                   `json:"allow_teacher_ids"`
	TeacherIDAndStatusPairs []*entity.AssessmentTeacherIDAndStatusPair `json:"teacher_id_and_status_pairs"`
	OrderBy                 *entity.ListAssessmentsOrderBy             `json:"order_by"`
	Page                    int                                        `json:"page"`
	PageSize                int                                        `json:"page_size"`
}

func (c *QueryAssessmentConditions) GetConditions() ([]string, []interface{}) {
	b := NewSQLBuilder().Append("delete_at = 0")

	if c.OrgID != nil {
		b.Append("exists (select 1 from schedules"+
			" where org_id = ? and delete_at = 0 and assessments.schedule_id = schedules.id)", c.OrgID)
	}

	if c.Status != nil {
		b.Append("status = ?", *c.Status)
	}

	if c.TeacherIDs != nil {
		if len(c.TeacherIDs) == 0 {
			return FalseSQLTemplate().DBOConditions()
		}
		teacherIDs := utils.SliceDeduplication(c.TeacherIDs)
		t := NewSQLTemplate("")
		for _, tid := range teacherIDs {
			t.Or("json_contains(teacher_ids, json_array(?))", tid)
		}
		b.AppendTemplate(t.WrapBracket())
	}

	if c.AllowTeacherIDs != nil {
		if len(c.AllowTeacherIDs) == 0 {
			return FalseSQLTemplate().DBOConditions()
		}
		allowTeacherIDs := utils.SliceDeduplication(c.AllowTeacherIDs)
		t := NewSQLTemplate("")
		for _, tid := range allowTeacherIDs {
			t.Or("json_contains(teacher_ids, json_array(?))", tid)
		}
		b.AppendTemplate(t.WrapBracket())
	}

	if len(c.TeacherIDAndStatusPairs) > 0 {
		for _, p := range c.TeacherIDAndStatusPairs {
			b.Append("((not json_contains(teacher_ids, json_array(?))) or (json_contains(teacher_ids, json_array(?)) and status = ?))",
				p.TeacherID, p.TeacherID, string(p.Status))
		}
	}

	if c.ScheduleIDs != nil {
		if len(c.ScheduleIDs) == 0 {
			return FalseSQLTemplate().DBOConditions()
		}
		b.Append("schedule_id in (?)", c.ScheduleIDs)
	}

	return b.MergeWithAnd().DBOConditions()
}

func (c *QueryAssessmentConditions) GetPager() *dbo.Pager {
	return &dbo.Pager{
		Page:     c.Page,
		PageSize: c.PageSize,
	}
}

func (c *QueryAssessmentConditions) GetOrderBy() string {
	if c.OrderBy == nil {
		return ""
	}
	switch *c.OrderBy {
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
