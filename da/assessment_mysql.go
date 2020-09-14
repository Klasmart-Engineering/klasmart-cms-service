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
	GetExcludeSoftDeleted(ctx context.Context, tx *dbo.DBContext, id string) (*entity.Assessment, error)
	UpdateStatus(ctx context.Context, tx *dbo.DBContext, id string, status entity.AssessmentStatus) error
	dbo.DataAccesser
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
			teacherFormats []string
			teacherValues  []interface{}
		)
		for _, teacherID := range *c.TeacherIDs {
			teacherFormats = append(teacherFormats, fmt.Sprintf("json_contains(teacher_ids, json_array(?))"))
			teacherValues = append(teacherValues, teacherID)
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
