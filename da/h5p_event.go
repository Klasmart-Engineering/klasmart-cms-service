package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
	"time"
)

type IH5PEventDA interface {
	Create(ctx context.Context, tx *dbo.DBContext, f *entity.H5PEvent) error
	QueryTx(ctx context.Context, db *dbo.DBContext, condition dbo.Conditions, values interface{}) error
}

type H5PEventDA struct {
	s dbo.BaseDA
}

func (h *H5PEventDA) Create(ctx context.Context, tx *dbo.DBContext, x *entity.H5PEvent) error {
	now := time.Now()
	//if x.ID == "" {
	//	x.ID = utils.NewID()
	//}
	x.CreateAt = now.Unix()
	_, err := h.s.InsertTx(ctx, tx, x)
	if err != nil {
		log.Error(ctx, "create folder da failed", log.Err(err), log.Any("req", x))
		return err
	}
	return nil
}

type H5PEventCondition struct {
	LessonPlanIDs     []string
	MaterialIDs       []string
	UserIDs           []string
	VerbIDs           []string
	LocalLibraryNames []string
}

func (c H5PEventCondition) GetConditions() ([]string, []interface{}) {
	var (
		conditions []string
		values     []interface{}
	)
	if c.LessonPlanIDs != nil {
		if len(c.LessonPlanIDs) == 0 {
			return c.falseConditions()
		}
		conditions = append(conditions, "lesson_plan_id in (?)")
		values = append(values, c.LessonPlanIDs)
	}
	if c.MaterialIDs != nil {
		if len(c.MaterialIDs) == 0 {
			return c.falseConditions()
		}
		conditions = append(conditions, "material_id in (?)")
		values = append(values, c.MaterialIDs)
	}
	if c.UserIDs != nil {
		if len(c.UserIDs) == 0 {
			return c.falseConditions()
		}
		conditions = append(conditions, "user_id in (?)")
		values = append(values, c.UserIDs)
	}
	if c.VerbIDs != nil {
		if len(c.VerbIDs) == 0 {
			return c.falseConditions()
		}
		conditions = append(conditions, "verb_id in (?)")
		values = append(values, c.VerbIDs)
	}
	if c.LocalLibraryNames != nil {
		if len(c.LocalLibraryNames) == 0 {
			return c.falseConditions()
		}
		conditions = append(conditions, "local_library_name in (?)")
		values = append(values, c.LocalLibraryNames)
	}
	return conditions, values
}

func (c H5PEventCondition) GetPager() *dbo.Pager { return nil }

func (c H5PEventCondition) GetOrderBy() string { return "" }

func (c H5PEventCondition) falseConditions() ([]string, []interface{}) {
	return []string{"false"}, nil
}

func (h *H5PEventDA) QueryTx(ctx context.Context, db *dbo.DBContext, condition dbo.Conditions, values interface{}) error {
	return h.s.QueryTx(ctx, db, condition, values)
}

var (
	_h5pEvent     IH5PEventDA
	_h5pEventOnce sync.Once
)

func GetH5PEventDA() IH5PEventDA {
	_h5pEventOnce.Do(func() {
		_h5pEvent = new(H5PEventDA)
	})
	return _h5pEvent
}
