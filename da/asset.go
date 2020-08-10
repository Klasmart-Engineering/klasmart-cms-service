package da

import (
	"context"
	"errors"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var(
	ErrRecordNotFound = errors.New("record not found")
	ErrPageOutOfRange = errors.New("page out of range")
)


type IAssetDA interface {
	CreateAsset(ctx context.Context, data entity.AssetObject) (int64, error)
	UpdateAsset(ctx context.Context, data entity.UpdateAssetRequest) error
	DeleteAsset(ctx context.Context, id int64) error

	GetAssetByID(ctx context.Context, id int64) (*entity.AssetObject, error)
	SearchAssets(ctx context.Context, condition *SearchAssetCondition) (int64, []*entity.AssetObject, error)
}

type MysqlDBAssetDA struct {
	dbo.BaseDA
}

func (m *MysqlDBAssetDA) CreateAsset(ctx context.Context, data entity.AssetObject) (int64, error) {
	now := time.Now()
	data.CreatedAt = &now
	data.UpdatedAt = &now
	_, err := m.InsertTx(ctx, dbo.MustGetDB(ctx), data)
	if err != nil {
		log.Error(ctx, "create asset failed", log.Err(err))
		return -1, err
	}
	return data.ID, nil
}

func (m *MysqlDBAssetDA) UpdateAsset(ctx context.Context, data entity.UpdateAssetRequest) error {
	now := time.Now()
	_, err := m.UpdateTx(ctx, dbo.MustGetDB(ctx), &entity.AssetObject{
		ID:            data.ID,
		Name:          data.Name,
		UpdatedAt:     &now,
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *MysqlDBAssetDA) DeleteAsset(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := m.UpdateTx(ctx, dbo.MustGetDB(ctx), entity.AssetObject{ID: id, DeletedAt: &now})
	if err != nil {
		return err
	}
	return nil
}

func (m *MysqlDBAssetDA) GetAssetByID(ctx context.Context, id int64) (*entity.AssetObject, error) {
	obj := new(entity.AssetObject)
	err := m.GetTx(ctx, dbo.MustGetDB(ctx), id, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (m *MysqlDBAssetDA) SearchAssets(ctx context.Context, condition *SearchAssetCondition) (int64, []*entity.AssetObject, error) {
	res := make([]*entity.AssetObject, 0)
	m.PageTx(ctx, dbo.MustGetDB(ctx), condition, res)
	return 0, nil, nil
}

type SearchAssetCondition struct {
	ID        []int64 `json:"id"`
	Name      string `json:"name"`
	OrgID	  string `json:"org_id"`

	SearchWords []string `json:"search_words"`
	Author      []string    `json:"author"`

	OrderBy  AssetsOrderBy `json:"order_by"`
	PageSize int `json:"page_size"`
	Page     int `json:"page"`
}

func (s SearchAssetCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var parameters []interface{}

	if len(s.ID) > 0 {
		wheres = append(wheres, "id in (?)")
		parameters = append(parameters, s.ID)
	}

	if s.Name != "" {
		wheres = append(wheres, "name like %?%")
		parameters = append(parameters, s.Name)
	}

	if len(s.SearchWords) > 0 {
		for i := range s.SearchWords {
			wheres = append(wheres, "name like %?% or keywords like %?% or description like %?% or author_name like %?%")
			parameters = append(parameters, s.SearchWords[i])
			parameters = append(parameters, s.SearchWords[i])
			parameters = append(parameters, s.SearchWords[i])
			parameters = append(parameters, s.SearchWords[i])
		}
	}

	if len(s.Author) > 0 {
		wheres = append(wheres, "author in (?)")
		parameters = append(parameters, s.Author)
	}
	return wheres, parameters
}

func (s SearchAssetCondition) GetPager() *dbo.Pager {
	return &dbo.Pager{
		Page:     s.Page,
		PageSize: s.PageSize,
	}
}

func (s SearchAssetCondition) GetOrderBy() string {
	return s.OrderBy.ToSQL()
}


type AssetsOrderBy int

const (
	AssetsOrderByName = iota
	AssetsOrderByNameDesc
	AssetsOrderByAuthor
	AssetsOrderByAuthorDesc
	AssetsOrderByCreatedAt
	AssetsOrderByCreatedAtDesc
)

// NewUserOrderBy parse order by
func NewAssetsOrderBy(orderby string) AssetsOrderBy {
	switch orderby {
	case "name":
		return AssetsOrderByName
	case "-name":
		return AssetsOrderByNameDesc
	case "author":
		return AssetsOrderByAuthor
	case "-author":
		return AssetsOrderByAuthorDesc
	case "created_at":
		return AssetsOrderByCreatedAt
	case "-created_at":
		return AssetsOrderByCreatedAtDesc
	default:
		return AssetsOrderByName
	}
}

// ToSQL enum to order by sql
func (s AssetsOrderBy) ToSQL() string {
	switch s {
	case AssetsOrderByName:
		return "name"
	case AssetsOrderByNameDesc:
		return "name desc"
	case AssetsOrderByAuthor:
		return "author"
	case AssetsOrderByAuthorDesc:
		return "author desc"
	case AssetsOrderByCreatedAt:
		return "created_at"
	case AssetsOrderByCreatedAtDesc:
		return "created_at desc"
	default:
		return "name"
	}
}


var _assetDA IAssetDA
var _assetDAOnce sync.Once

func GetAssetDA() IAssetDA {
	_assetDAOnce.Do(func() {
		_assetDA = new(MysqlDBAssetDA)
	})
	return _assetDA
}
