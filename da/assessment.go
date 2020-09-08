package da

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
		Where("id = ?").
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
		"status":      status,
		"update_time": nowUnix,
	}
	if status == entity.AssessmentStatusComplete {
		updateFields["complete_time"] = nowUnix
	}
	if err := tx.Model(entity.Assessment{}).
		Where(a.filterSoftDeletedTemplate()).
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
	return "delete_time == 0"
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

	formats = append(formats, "delete_time = 0")

	if c.Status != nil {
		formats = append(formats, "status = ?")
		values = append(values, *c.Status)
	}

	if c.TeacherIDs != nil {
		temp := entity.NullStrings{
			Strings: *c.TeacherIDs,
			Valid:   true,
		}
		formats = append(formats, fmt.Sprintf("teacher_id in (%s)", temp.SQLPlaceHolder()))
		values = append(values, temp.ToInterfaceSlice()...)
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
