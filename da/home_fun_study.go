package da

import (
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type IHomeFunStudyDA interface {
	dbo.DataAccesser
}

var (
	homeFunStudyDAInstance     IHomeFunStudyDA
	homeFunStudyDAInstanceOnce = sync.Once{}
)

func GetHomeFunStudyDA() IHomeFunStudyDA {
	homeFunStudyDAInstanceOnce.Do(func() {
		homeFunStudyDAInstance = &homeFunStudyDA{}
	})
	return homeFunStudyDAInstance
}

type homeFunStudyDA struct {
	dbo.BaseDA
}

type QueryHomeFunStudyCondition struct {
	OrgID           *string                                    `json:"org_id"`
	ScheduleID      *string                                    `json:"schedule_id"`
	TeacherIDs      utils.SQLJSONStringArray                   `json:"teacher_ids"`
	StudentIDs      []string                                   `json:"student_ids"`
	Status          *entity.AssessmentStatus                   `json:"status"`
	OrderBy         *entity.ListHomeFunStudiesOrderBy          `json:"order_by"`
	AllowTeacherIDs []string                                   `json:"allow_teacher_ids"`
	AllowPairs      []*entity.AssessmentTeacherIDAndStatusPair `json:"allow_pairs"`
	Page            int                                        `json:"page"`
	PageSize        int                                        `json:"page_size"`
}

func (c *QueryHomeFunStudyCondition) GetConditions() ([]string, []interface{}) {
	b := NewSQLTemplate("delete_at = 0")

	if c.OrgID != nil {
		b.Appendf("exists (select 1 from schedules"+
			" where org_id = ? and delete_at = 0 and home_fun_studies.schedule_id = schedules.id)", c.OrgID)
	}

	if c.ScheduleID != nil {
		b.Appendf("schedule_id = ?", c.ScheduleID)
	}

	if c.TeacherIDs != nil || c.StudentIDs != nil {
		flag := false
		t2 := NewSQLTemplate("")
		if len(c.TeacherIDs) > 0 {
			teacherIDs := utils.SliceDeduplication(c.TeacherIDs)
			for _, teacherID := range teacherIDs {
				t2.Appendf("json_contains(teacher_ids, json_array(?))", teacherID)
				flag = true
			}
		}
		if len(c.StudentIDs) > 0 {
			t2.Appendf("student_id in (?)", c.StudentIDs)
			flag = true
		}
		if !flag {
			return FalseSQLTemplate().DBOConditions()
		}
		b.AppendResult(t2.Or())
	}

	if c.Status != nil {
		b.Appendf("status = ?", *c.Status)
	}

	if c.AllowTeacherIDs != nil {
		if len(c.AllowTeacherIDs) == 0 {
			return FalseSQLTemplate().DBOConditions()
		}
		allowTeacherIDs := utils.SliceDeduplication(c.AllowTeacherIDs)
		t2 := NewSQLTemplate("")
		for _, tid := range allowTeacherIDs {
			t2.Appendf("json_contains(teacher_ids, json_array(?))", tid)
		}
		b.AppendResult(t2.Or())
	}

	for _, pair := range c.AllowPairs {
		b.Appendf("((not json_contains(teacher_ids, json_array(?))) or (json_contains(teacher_ids, json_array(?)) and status = ?))",
			pair.TeacherID, pair.TeacherID, string(pair.Status))
	}

	return b.DBOConditions()
}

func (c *QueryHomeFunStudyCondition) GetPager() *dbo.Pager {
	return &dbo.Pager{
		Page:     c.Page,
		PageSize: c.PageSize,
	}
}

func (c *QueryHomeFunStudyCondition) GetOrderBy() string {
	if c.OrderBy == nil {
		return ""
	}
	switch *c.OrderBy {
	case entity.ListHomeFunStudiesOrderByCompleteAt:
		return "complete_at"
	case entity.ListHomeFunStudiesOrderByCompleteAtDesc:
		return "complete_at desc"
	case entity.ListHomeFunStudiesOrderByLatestFeedbackAt:
		return "latest_feedback_at"
	case entity.ListHomeFunStudiesOrderByLatestFeedbackAtDesc:
		return "latest_feedback_at desc"
	}
	return ""
}
