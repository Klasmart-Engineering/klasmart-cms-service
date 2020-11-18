package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
	"time"
)

type IFolderDA interface {
	CreateFolder(ctx context.Context, tx *dbo.DBContext, f entity.FolderItem) (string ,error)
	UpdateFolder(ctx context.Context, tx *dbo.DBContext, fid string, f entity.FolderItem) error

	BatchUpdateFolderPath(ctx context.Context, tx *dbo.DBContext, fids []string, path string) error
	BatchUpdateFolderVisibilitySettings(ctx context.Context, tx *dbo.DBContext, fids []string, path string) error

	DeleteFolder(ctx context.Context, tx *dbo.DBContext, fid string) error
	GetFolderByID(ctx context.Context, tx *dbo.DBContext, fid string) (*entity.FolderItem, error)

	GetFolderByIDList(ctx context.Context, tx *dbo.DBContext, fids []string) ([]*entity.FolderItem, error)

	SearchFolder(ctx context.Context, tx *dbo.DBContext, condition FolderCondition) (int, []*entity.FolderItem, error)
}

type FolderDA struct{
	s dbo.BaseDA
}

func (fda *FolderDA) CreateFolder(ctx context.Context, tx *dbo.DBContext, f entity.FolderItem) (string, error) {
	now := time.Now()
	if f.ID == "" {
		f.ID = utils.NewID()
	}
	f.UpdateAt = now.Unix()
	f.CreateAt = now.Unix()
	_, err := fda.s.InsertTx(ctx, tx, &f)
	if err != nil {
		log.Warn(ctx, "create folder da failed", log.Err(err), log.Any("req", f))
		return "", err
	}
	return f.ID, nil
}

func (fda *FolderDA) UpdateFolder(ctx context.Context, tx *dbo.DBContext, fid string, f entity.FolderItem) error {
	f.ID = fid
	f.UpdateAt = time.Now().Unix()
	log.Info(ctx, "Update folder da", log.String("id", f.ID))
	_, err := fda.s.UpdateTx(ctx, tx, &f)
	if err != nil {
		log.Warn(ctx, "update folder da failed", log.Err(err), log.String("id", fid), log.Any("req", f))
		return err
	}

	return nil
}
func (fda *FolderDA) BatchUpdateFolderPath(ctx context.Context, tx *dbo.DBContext, fids []string, path string) error{
	err := tx.Model(entity.FolderItem{}).Where("id IN (?)", fids).Updates(map[string]interface{}{"path": path}).Error
	if err != nil {
		log.Warn(ctx, "update folder da failed", log.Err(err), log.Strings("fids", fids), log.String("path", path))
		return err
	}

	return nil
}
func (fda *FolderDA) BatchUpdateFolderVisibilitySettings(ctx context.Context, tx *dbo.DBContext, link []string, path string) error{
	err := tx.Model(entity.FolderItem{}).Where("link IN (?)", link).Updates(map[string]interface{}{"path": path}).Error
	if err != nil {
		log.Warn(ctx, "update folder da failed", log.Err(err), log.Strings("link", link), log.String("path", path))
		return err
	}

	return nil
}

func (fda *FolderDA) DeleteFolder(ctx context.Context, tx *dbo.DBContext, fid string) error {
	folderItem, err := fda.GetFolderByID(ctx, tx, fid)
	if err != nil{
		return err
	}
	folderItem.ID = fid
	folderItem.DeleteAt = time.Now().Unix()
	_, err = fda.s.UpdateTx(ctx, tx, folderItem)
	if err != nil {
		log.Warn(ctx, "delete folder da failed", log.Err(err), log.String("id", fid))
		return err
	}
	return nil
}

func (fda *FolderDA) GetFolderByID(ctx context.Context, tx *dbo.DBContext, fid string) (*entity.FolderItem, error) {
	obj := new(entity.FolderItem)
	err := fda.s.GetTx(ctx, tx, fid, obj)
	if err != nil {
		log.Warn(ctx, "get folder da failed", log.Err(err), log.String("id", fid))
		return nil, err
	}
	if obj.DeleteAt > 0 {
		log.Info(ctx, "folder was deleted", log.String("id", fid))
		return nil, dbo.ErrRecordNotFound
	}

	return obj, nil
}

func (fda *FolderDA) GetFolderByIDList(ctx context.Context, tx *dbo.DBContext, fids []string) ([]*entity.FolderItem, error) {
	objs := make([]*entity.FolderItem, 0)
	err := fda.s.QueryTx(ctx, tx, &FolderCondition{
		IDs:           fids,
	}, &objs)
	if err != nil {
		return nil, err
	}

	return objs, nil
}

func (fda *FolderDA) SearchFolder(ctx context.Context, tx *dbo.DBContext, condition FolderCondition) (int, []*entity.FolderItem, error) {
	objs := make([]*entity.FolderItem, 0)
	count, err := fda.s.PageTx(ctx, tx, &condition, &objs)
	if err != nil {
		return 0, nil, err
	}

	return count, objs, nil
}


type FolderOrderBy int
const (
	FolderOrderByCreatedAtDesc = iota
	FolderOrderByCreatedAt
	FolderOrderByIdDesc
	FolderOrderById
	FolderOrderByUpdatedAt
	FolderOrderByUpdatedAtDesc
)

// NewTeacherOrderBy parse order by
func NewFolderOrderBy(orderby string) FolderOrderBy {
	switch orderby {
	case "id":
		return FolderOrderById
	case "-id":
		return FolderOrderByIdDesc
	case "create_at":
		return FolderOrderByCreatedAt
	case "-create_at":
		return FolderOrderByCreatedAtDesc
	case "update_at":
		return FolderOrderByUpdatedAt
	case "-update_at":
		return FolderOrderByUpdatedAtDesc
	default:
		return FolderOrderByCreatedAtDesc
	}
}

// ToSQL enum to order by sql
func (s FolderOrderBy) ToSQL() string {
	switch s {
	case FolderOrderById:
		return "id"
	case FolderOrderByIdDesc:
		return "id desc"
	case FolderOrderByCreatedAt:
		return "create_at"
	case FolderOrderByCreatedAtDesc:
		return "create_at desc"
	case FolderOrderByUpdatedAt:
		return "update_at"
	case FolderOrderByUpdatedAtDesc:
		return "update_at desc"
	default:
		return "create_at desc"
	}
}

type FolderCondition struct {
	IDs []string
	OwnerType int
	ItemType int
	Owner string
	ParentId string
	VisibilitySetting []string
	Link string

	Name string
	Path string

	ExactPath string

	OrderBy FolderOrderBy `json:"order_by"`
	Pager   utils.Pager
}

func (s *FolderCondition) GetConditions() ([]string, []interface{}) {
	conditions := make([]string, 0)
	params := make([]interface{}, 0)

	if len(s.IDs) > 0 {
		condition := " id in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.IDs)
	}
	if s.OwnerType > 0 {
		conditions = append(conditions, "owner_type = ?")
		params = append(params, s.OwnerType)
	}
	if s.Owner != "" {
		conditions = append(conditions, "owner = ?")
		params = append(params, s.Owner)
	}
	if s.ItemType != 0 {
		conditions = append(conditions, "item_type = ?")
		params = append(params, s.ItemType)
	}

	if s.ParentId != "" {
		conditions = append(conditions, "parent_id = ?")
		params = append(params, s.ParentId)
	}
	if s.Link != "" {
		conditions = append(conditions, "link = ?")
		params = append(params, s.Link)
	}
	if s.Name != "" {
		conditions = append(conditions, "name = ?")
		params = append(params, s.Name)
	}
	if s.Path != "" {
		conditions = append(conditions, "path like ?")
		params = append(params, s.Path + "%")
	}
	if s.ExactPath != "" {
		conditions = append(conditions, "path like ?")
		params = append(params, s.ExactPath)
	}

	if len(s.VisibilitySetting) != 0 {
		conditions = append(conditions, "visibility_setting in (?)")
		params = append(params, s.VisibilitySetting)
	}

	conditions = append(conditions, " delete_at = 0")
	return conditions, params
}

func (f *FolderCondition) GetPager() *dbo.Pager {
	return &dbo.Pager{
		Page:     int(f.Pager.PageIndex),
		PageSize: int(f.Pager.PageSize),
	}
}

func (f *FolderCondition) GetOrderBy() string {
	return f.OrderBy.ToSQL()
}

var (
	folderDA    IFolderDA
	_folderDAOnce sync.Once
)

func GetFolderDA() IFolderDA {
	_folderDAOnce.Do(func() {
		folderDA = new(FolderDA)
	})

	return folderDA
}
