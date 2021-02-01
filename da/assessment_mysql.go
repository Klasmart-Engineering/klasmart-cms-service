package da

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"strings"
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

type QueryAssessmentsCondition struct {
	Status                         *entity.AssessmentStatus                `json:"status"`
	ScheduleIDs                    []string                                `json:"schedule_ids"`
	TeacherIDs                     []string                                `json:"teacher_ids"`
	TeacherAssessmentStatusFilters []*entity.TeacherAssessmentStatusFilter `json:"teacher_assessment_status_filters"`
	OrderBy                        *entity.ListAssessmentsOrderBy          `json:"order_by"`
	Page                           int                                     `json:"page"`
	PageSize                       int                                     `json:"page_size"`
}

func (c *QueryAssessmentsCondition) Simple() {
	c.TeacherIDs = utils.SliceDeduplication(c.TeacherIDs)
	var newTeacherAssessmentStatusFilters []*entity.TeacherAssessmentStatusFilter
	{
		statusMap := map[string]*struct {
			HasStatusInProgress bool
			HasStatusComplete   bool
		}{}
		for _, item := range c.TeacherAssessmentStatusFilters {
			if statusMap[item.TeacherID] == nil {
				statusMap[item.TeacherID] = &struct {
					HasStatusInProgress bool
					HasStatusComplete   bool
				}{}
			}
			switch item.Status {
			case entity.AssessmentStatusComplete:
				statusMap[item.TeacherID].HasStatusComplete = true
			case entity.AssessmentStatusInProgress:
				statusMap[item.TeacherID].HasStatusInProgress = true
			}
		}
		for teacherID, status := range statusMap {
			if status.HasStatusInProgress && status.HasStatusComplete {
				continue
			}
			if !status.HasStatusInProgress && !status.HasStatusComplete {
				continue
			}
			if status.HasStatusComplete {
				newTeacherAssessmentStatusFilters = append(newTeacherAssessmentStatusFilters, &entity.TeacherAssessmentStatusFilter{
					TeacherID: teacherID,
					Status:    entity.AssessmentStatusComplete,
				})
			}
			if status.HasStatusInProgress {
				newTeacherAssessmentStatusFilters = append(newTeacherAssessmentStatusFilters, &entity.TeacherAssessmentStatusFilter{
					TeacherID: teacherID,
					Status:    entity.AssessmentStatusInProgress,
				})
			}
		}
	}
	c.TeacherAssessmentStatusFilters = newTeacherAssessmentStatusFilters
}

func (c *QueryAssessmentsCondition) GetConditions() ([]string, []interface{}) {
	c.Simple()

	var (
		formats []string
		values  []interface{}
	)

	formats = append(formats, "delete_at = 0")

	if c.Status != nil {
		formats = append(formats, "status = ?")
		values = append(values, *c.Status)
	}

	if c.TeacherIDs != nil {
		if len(c.TeacherIDs) == 0 {
			return []string{"1 = 2"}, nil
		}
		var (
			partFormats = make([]string, 0, len(c.TeacherIDs))
			partValues  = make([]interface{}, 0, len(c.TeacherIDs))
		)
		for _, teacherID := range c.TeacherIDs {
			partFormats = append(partFormats, fmt.Sprintf("json_contains(teacher_ids, json_array(?))"))
			partValues = append(partValues, teacherID)
		}
		formats = append(formats, "("+strings.Join(partFormats, " or ")+")")
		values = append(values, partValues...)
	}

	if len(c.TeacherAssessmentStatusFilters) > 0 {
		var (
			partFormats = make([]string, 0, len(c.TeacherAssessmentStatusFilters))
			partValues  = make([]interface{}, 0, len(c.TeacherAssessmentStatusFilters)*2)
		)
		for _, item := range c.TeacherAssessmentStatusFilters {
			partFormats = append(partFormats, fmt.Sprintf("(not json_contains(teacher_ids, json_array(?))) or (json_contains(teacher_ids, json_array(?)) and status = ?)"))
			partValues = append(partValues, item.TeacherID, item.TeacherID, string(item.Status))
		}
		formats = append(formats, "("+strings.Join(partFormats, " and ")+")")
		values = append(values, partValues...)
	}

	if c.ScheduleIDs != nil {
		if len(c.ScheduleIDs) == 0 {
			return []string{"1 = 2"}, nil
		}
		formats = append(formats, "schedule_id in (?)")
		values = append(values, c.ScheduleIDs)
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
