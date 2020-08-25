package da

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"strings"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ContentOrderBy int

const (
	ContentOrderByCreatedAtDesc = iota
	ContentOrderByCreatedAt
	ContentOrderByIdDesc
	ContentOrderById
	ContentOrderByUpdatedAt
	ContentOrderByUpdatedAtDesc
)

type IContentDA interface {
	CreateContent(ctx context.Context, tx *dbo.DBContext, co entity.Content) (string, error)
	UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, co entity.Content) error
	DeleteContent(ctx context.Context, tx *dbo.DBContext, cid string) error
	GetContentById(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.Content, error)

	SearchContent(ctx context.Context, tx *dbo.DBContext, condition ContentCondition) (int, []*entity.Content, error)
	SearchContentUnSafe(ctx context.Context, tx *dbo.DBContext, condition dbo.Conditions) (int, []*entity.Content, error)
	Count(context.Context, dbo.Conditions) (int, error)
}

type CombineConditions struct {
	SourceCondition dbo.Conditions
	TargetCondition dbo.Conditions
}

func (s *CombineConditions) GetConditions() ([]string, []interface{}) {
	sourceWhereRaw, sourceParams := s.SourceCondition.GetConditions()
	targetWhereRaw, targetParams := s.TargetCondition.GetConditions()
	where := ""
	if len(sourceWhereRaw) < 1 || len(targetWhereRaw) < 1 {
		return s.SourceCondition.GetConditions()
	}

	sourceWhere := "(" + strings.Join(sourceWhereRaw, " and ") + ")"
	targetWhere := "(" + strings.Join(targetWhereRaw, " and ") + ")"
	where = sourceWhere + " OR " + targetWhere
	params := append(sourceParams, targetParams...)
	return []string{where}, params
}

func (s *CombineConditions) GetPager() *dbo.Pager {
	return s.SourceCondition.GetPager()
}
func (s *CombineConditions) GetOrderBy() string {
	return s.SourceCondition.GetOrderBy()
}

type ContentCondition struct {
	IDS          []string `json:"ids"`
	Name         []string `json:"name"`
	ContentType  []int `json:"content_type"`
	Scope        []string `json:"scope"`
	PublishStatus []string `json:"publish_status"`
	Author      string `json:"author"`
	Org    		string `json:"org"`

	OrderBy  ContentOrderBy `json:"order_by"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func (s *ContentCondition) GetConditions() ([]string, []interface{}) {
	conditions := make([]string, 0)
	params := make([]interface{}, 0)
	if len(s.IDS) > 0 {
		condition := " id in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.IDS)
	}
	if len(s.Name) != 0 {
		var fullparam string
		for i := range s.Name {
			if s.Name[i] != " " && s.Name[i] != "" {
				fullparam = fullparam + "+" + s.Name[i] + " "
			}
		}
		fullparam = strings.TrimSuffix(fullparam, " ")
		conditions = append(conditions, "match(name, description, author_name, keywords) against(? in boolean mode)")
		params = append(params, fullparam)
	}
	if len(s.ContentType) > 0 {
		var subConditions []string
		for _, item := range s.ContentType {
			subConditions = append(subConditions, "content_type in (?) ")
			params = append(params, fmt.Sprintf("%d", item))
		}
		conditions = append(conditions, "("+strings.Join(subConditions, " or ")+")")
	}
	if len(s.Scope) > 0 {
		condition := " publish_scope in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.Scope)
	}

	if len(s.PublishStatus) > 0 {
		condition := " publish_status in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.PublishStatus)
	}
	if s.Author != "" {
		condition := " author = ? "
		conditions = append(conditions, condition)
		params = append(params, s.Author)
	}
	if s.Org != ""{
		condition := " org = ? "
		conditions = append(conditions, condition)
		params = append(params, s.Org)
	}

	conditions = append(conditions, " deleted_at IS NULL ")
	return conditions, params
}
func (s *ContentCondition) GetPager() *dbo.Pager {
	if s.PageSize < 1 {
		s.PageSize = 20
	}
	return &dbo.Pager{
		Page:     s.Page,
		PageSize: s.PageSize,
	}
}
func (s *ContentCondition) GetOrderBy() string {
	return s.OrderBy.ToSQL()
}

// NewTeacherOrderBy parse order by
func NewContentOrderBy(orderby string) ContentOrderBy {
	switch orderby {
	case "id":
		return ContentOrderById
	case "-id":
		return ContentOrderByIdDesc
	case "created_at":
		return ContentOrderByCreatedAt
	case "-created_at":
		return ContentOrderByCreatedAtDesc
	case "updated_at":
		return ContentOrderByUpdatedAt
	case "-updated_at":
		return ContentOrderByUpdatedAtDesc
	default:
		return ContentOrderByCreatedAtDesc
	}
}

// ToSQL enum to order by sql
func (s ContentOrderBy) ToSQL() string {
	switch s {
	case ContentOrderById:
		return "id"
	case ContentOrderByIdDesc:
		return "id desc"
	case ContentOrderByCreatedAt:
		return "created_at"
	case ContentOrderByCreatedAtDesc:
		return "created_at desc"
	case ContentOrderByUpdatedAt:
		return "updated_at"
	case ContentOrderByUpdatedAtDesc:
		return "updated_at desc"
	default:
		return "upload_at desc"
	}
}

type DBContentDA struct {
	s dbo.BaseDA
}

func (cd *DBContentDA) CreateContent(ctx context.Context, tx *dbo.DBContext, co entity.Content) (string, error) {
	now := time.Now()
	co.ID = utils.NewID()
	co.UpdatedAt = now.Unix()
	co.CreatedAt = now.Unix()
	err := cd.s.SaveTx(ctx, tx, &co)
	if err != nil {
		return "", err
	}
	return co.ID, nil
}
func (cd *DBContentDA) UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, co entity.Content) error {
	now := time.Now()
	co.ID = cid
	co.UpdatedAt = now.Unix()
	log.Info(ctx, "Update contentdata da", log.String("id", co.ID))
	_, err := cd.s.UpdateTx(ctx, tx, &co)
	if err != nil {
		return err
	}

	return nil
}
func (cd *DBContentDA) DeleteContent(ctx context.Context, tx *dbo.DBContext, cid string) error {
	now := time.Now()
	content := new(entity.Content)
	content.ID = cid
	content.DeletedAt = now.Unix()
	_, err := cd.s.UpdateTx(ctx, tx, content)
	if err != nil {
		return err
	}
	return nil
}
func (cd *DBContentDA) GetContentById(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.Content, error) {
	obj := new(entity.Content)
	err := cd.s.GetTx(ctx, tx, cid, obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (cd *DBContentDA) SearchContent(ctx context.Context, tx *dbo.DBContext, condition ContentCondition) (int, []*entity.Content, error) {
	objs := make([]*entity.Content, 0)
	count, err := cd.s.PageTx(ctx, tx, &condition, &objs)
	if err != nil {
		return 0, nil, err
	}

	return count, objs, nil
}
func (cd *DBContentDA) SearchContentUnSafe(ctx context.Context, tx *dbo.DBContext, condition dbo.Conditions) (int, []*entity.Content, error) {
	objs := make([]*entity.Content, 0)
	count, err := cd.s.PageTx(ctx, tx, condition, &objs)
	if err != nil {
		return 0, nil, err
	}

	return count, objs, nil
}

func (cd *DBContentDA) Count(ctx context.Context, condition dbo.Conditions) (int, error) {
	return cd.s.Count(ctx, condition, &entity.Content{})
}

var (
	contentDA    IContentDA
	_contentOnce sync.Once
)

func GetContentDA() IContentDA {
	_contentOnce.Do(func() {
		contentDA = new(DBContentDA)
	})

	return contentDA
}
