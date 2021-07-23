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
	IDs         entity.NullStrings           `json:"ids"`
	OrgID       entity.NullString            `json:"org_id"`
	ScheduleID  entity.NullString            `json:"schedule_id"`
	ScheduleIDs entity.NullStrings           `json:"schedule_ids"`
	TeacherIDs  utils.NullSQLJSONStringArray `json:"teacher_ids"`
	StudentIDs  entity.NullStrings           `json:"student_ids"`
	Status      entity.NullAssessmentStatus  `json:"status"`

	CreatedBetween  NullTimeRange     `json:"created_between"`
	UpdateBetween   NullTimeRange     `json:"update_between"`
	CompleteBetween NullTimeRange     `json:"complete_between"`
	ClassType       entity.NullString `json:"class_type"`

	OrderBy                      entity.NullListHomeFunStudiesOrderBy              `json:"order_by"`
	AllowTeacherIDs              entity.NullStrings                                `json:"allow_teacher_ids"`
	AllowTeacherIDAndStatusPairs entity.NullAssessmentAllowTeacherIDAndStatusPairs `json:"allow_pairs"`
	Pager                        dbo.Pager                                         `json:"pager"`
}

func (c *QueryHomeFunStudyCondition) GetConditions() ([]string, []interface{}) {
	b := NewSQLTemplate("delete_at = 0")

	if c.IDs.Valid {
		b.Appendf("id in (?)", c.IDs.Strings)
	}

	if c.OrgID.Valid {
		b.Appendf("exists (select 1 from schedules"+
			" where org_id = ? and delete_at = 0 and home_fun_studies.schedule_id = schedules.id)", c.OrgID.String)
	}

	if c.ScheduleID.Valid {
		b.Appendf("schedule_id = ?", c.ScheduleID.String)
	}

	if c.ScheduleIDs.Valid {
		b.Appendf("schedule_id in (?)", c.ScheduleIDs.Strings)
	}

	if c.TeacherIDs.Valid || c.StudentIDs.Valid {
		flag := false
		t2 := NewSQLTemplate("")
		if len(c.TeacherIDs.Values) > 0 {
			teacherIDs := utils.SliceDeduplication(c.TeacherIDs.Values)
			for _, teacherID := range teacherIDs {
				t2.Appendf("json_contains(teacher_ids, json_array(?))", teacherID)
				flag = true
			}
		}
		if len(c.StudentIDs.Strings) > 0 {
			t2.Appendf("student_id in (?)", c.StudentIDs.Strings)
			flag = true
		}
		if !flag {
			return FalseSQLTemplate().DBOConditions()
		}
		b.AppendResult(t2.Or())
	}

	if c.Status.Valid {
		b.Appendf("status = ?", c.Status.Value)
	}

	if c.CreatedBetween.Valid {
		b.Appendf("(create_at BETWEEN ? AND ?)", c.CreatedBetween.StartAt, c.CreatedBetween.EndAt)
	}
	if c.UpdateBetween.Valid {
		b.Appendf("(update_at BETWEEN ? AND ?)", c.UpdateBetween.StartAt, c.UpdateBetween.EndAt)
	}
	if c.CompleteBetween.Valid {
		b.Appendf("(complete_at BETWEEN ? AND ?)", c.CompleteBetween.StartAt, c.CompleteBetween.EndAt)
	}
	if c.ClassType.Valid {
		b.Appendf("exists (select 1 from schedules"+
			" where home_fun_studies.schedule_id = schedules.id and class_type = ? and is_home_fun=true)",
			c.ClassType.String)
	}

	if c.AllowTeacherIDs.Valid {
		allowTeacherIDs := utils.SliceDeduplication(c.AllowTeacherIDs.Strings)
		t2 := NewSQLTemplate("")
		for _, tid := range allowTeacherIDs {
			t2.Appendf("json_contains(teacher_ids, json_array(?))", tid)
		}
		b.AppendResult(t2.Or())
	}

	if c.AllowTeacherIDAndStatusPairs.Valid {
		for _, pair := range c.AllowTeacherIDAndStatusPairs.Values {
			b.Appendf("((not json_contains(teacher_ids, json_array(?))) or (json_contains(teacher_ids, json_array(?)) and status = ?))",
				pair.TeacherID, pair.TeacherID, string(pair.Status))
		}
	}

	return b.DBOConditions()
}

func (c *QueryHomeFunStudyCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

func (c *QueryHomeFunStudyCondition) GetOrderBy() string {
	if !c.OrderBy.Valid {
		return ""
	}
	switch c.OrderBy.Value {
	case entity.ListHomeFunStudiesOrderByCompleteAt:
		return "complete_at"
	case entity.ListHomeFunStudiesOrderByCompleteAtDesc:
		return "complete_at desc"
	case entity.ListHomeFunStudiesOrderByLatestFeedbackAt:
		return "latest_feedback_at"
	case entity.ListHomeFunStudiesOrderByLatestFeedbackAtDesc:
		return "latest_feedback_at desc"
	case entity.ListHomeFunStudiesOrderByCreateAt:
		return "create_at"
	case entity.ListHomeFunStudiesOrderByCreateAtDesc:
		return "create_at desc"
	case entity.ListHomeFunStudiesOrderByUpdateAt:
		return "update_at"
	case entity.ListHomeFunStudiesOrderByUpdateAtDesc:
		return "update_at desc"
	}
	return ""
}
