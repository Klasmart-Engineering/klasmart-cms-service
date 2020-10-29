package da

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
	"sync"
	"time"
)

type IAssessmentDA interface {
	dbo.DataAccesser
	GetExcludeSoftDeleted(ctx context.Context, tx *dbo.DBContext, id string) (*entity.Assessment, error)
	UpdateStatus(ctx context.Context, tx *dbo.DBContext, id string, status entity.AssessmentStatus) error
	BatchGetAssessmentsByScheduleIDs(ctx context.Context, tx *dbo.DBContext, scheduleIDs []string) ([]*entity.Assessment, error)
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

type QueryAssessmentsCondition struct {
	Status     *entity.AssessmentStatus       `json:"status"`
	TeacherIDs *[]string                      `json:"teacher_ids"`
	OrderBy    *entity.ListAssessmentsOrderBy `json:"order_by"`
	Page       int                            `json:"page"`
	PageSize   int                            `json:"page_size"`
}

func (c *QueryAssessmentsCondition) GetConditions() ([]string, []interface{}) {
	var (
		formats []string
		values  []interface{}
	)

	formats = append(formats, "delete_at = 0")

	if c.Status != nil {
		formats = append(formats, "status = ?")
		values = append(values, *c.Status)
	}

	if c.TeacherIDs != nil && len(*c.TeacherIDs) > 0 {
		var (
			teacherFormats = make([]string, len(*c.TeacherIDs))
			teacherValues  = make([]interface{}, len(*c.TeacherIDs))
		)
		for i, teacherID := range *c.TeacherIDs {
			teacherFormats[i] = fmt.Sprintf("json_contains(teacher_ids, json_array(?))")
			teacherValues[i] = teacherID
		}
		formats = append(formats, "("+strings.Join(teacherFormats, " or ")+")")
		values = append(values, teacherValues...)
	}

	return formats, values
}

func (c *QueryAssessmentsCondition) GetPager() *dbo.Pager {
	return &dbo.Pager{
		Page:     c.Page,
		PageSize: c.PageSize,
	}
}

func (c *QueryAssessmentsCondition) GetOrderBy() string {
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
