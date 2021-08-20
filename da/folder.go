package da

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"github.com/jinzhu/gorm"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	ErrNoFids = errors.New("fids length is 0")
)

type IFolderDA interface {
	CreateFolder(ctx context.Context, tx *dbo.DBContext, f *entity.FolderItem) (string, error)
	UpdateFolder(ctx context.Context, tx *dbo.DBContext, fid string, f *entity.FolderItem) error
	// AddFolderItemsCount(ctx context.Context, tx *dbo.DBContext, fid string, addon int) error

	BatchReplaceFolderPath(ctx context.Context, tx *dbo.DBContext, fids []string, oldPath, path entity.Path) error
	BatchUpdateFolderPathPrefix(ctx context.Context, tx *dbo.DBContext, fids []string, prefix entity.Path) error
	BatchUpdateFoldersPath(ctx context.Context, tx *dbo.DBContext, fids []string, dirPath entity.Path) error

	DeleteFolder(ctx context.Context, tx *dbo.DBContext, fid string) error
	BatchDeleteFolders(ctx context.Context, tx *dbo.DBContext, fids []string) error

	GetFolderByID(ctx context.Context, tx *dbo.DBContext, fid string) (*entity.FolderItem, error)

	GetFolderByIDList(ctx context.Context, tx *dbo.DBContext, fids []string) ([]*entity.FolderItem, error)

	SearchFolderPage(ctx context.Context, tx *dbo.DBContext, condition FolderCondition) (int, []*entity.FolderItem, error)
	SearchFolderCount(ctx context.Context, tx *dbo.DBContext, condition FolderCondition) (int, error)
	SearchFolder(ctx context.Context, tx *dbo.DBContext, condition FolderCondition) ([]*entity.FolderItem, error)

	BatchGetFolderItemsCount(ctx context.Context, tx *dbo.DBContext, fids []string) ([]*entity.FolderItemsCount, error)
	BatchUpdateFolderItemsCount(ctx context.Context, tx *dbo.DBContext, req []*entity.UpdateFolderItemsCountRequest) error
}

type FolderDA struct {
	s dbo.BaseDA
}

func (fda *FolderDA) CreateFolder(ctx context.Context, tx *dbo.DBContext, f *entity.FolderItem) (string, error) {
	now := time.Now()
	if f.ID == "" {
		f.ID = utils.NewID()
	}
	f.UpdateAt = now.Unix()
	f.CreateAt = now.Unix()
	_, err := fda.s.InsertTx(ctx, tx, f)
	if err != nil {
		log.Error(ctx, "create folder da failed", log.Err(err), log.Any("req", f))
		return "", err
	}
	return f.ID, nil
}

func (fda *FolderDA) UpdateFolder(ctx context.Context, tx *dbo.DBContext, fid string, f *entity.FolderItem) error {
	f.ID = fid
	f.UpdateAt = time.Now().Unix()
	log.Info(ctx, "Update folder da", log.String("id", f.ID))
	_, err := fda.s.UpdateTx(ctx, tx, f)
	if err != nil {
		log.Error(ctx, "update folder da failed", log.Err(err), log.String("id", fid), log.Any("req", f))
		return err
	}

	return nil
}

func (fda *FolderDA) AddFolderItemsCount(ctx context.Context, tx *dbo.DBContext, fid string, addon int) error {
	err := tx.Model(&entity.FolderItem{ID: fid}).UpdateColumn("items_count", gorm.Expr("items_count + ?", addon)).Error
	if err != nil {
		log.Error(ctx, "update folder items count failed", log.Err(err), log.Int("addon", addon), log.String("fid", fid))
		return err
	}

	return nil
}

func (fda *FolderDA) BatchGetFolderItemsCount(ctx context.Context, tx *dbo.DBContext, fids []string) ([]*entity.FolderItemsCount, error) {
	if len(fids) < 1 {
		//若fids为空，则不更新
		//if fids is nil, no need to update
		return nil, nil
	}
	folders, err := fda.GetFolderByIDList(ctx, tx, fids)
	if err != nil {
		log.Error(ctx, "GetFolderByIDList failed", log.Err(err),
			log.Strings("fids", fids))
		return nil, err
	}
	pathList := make([]string, len(folders))

	for i := range folders {
		pathList[i] = string(folders[i].ChildrenPath())
	}
	if len(pathList) < 1 {
		return nil, nil
	}
	sql := `
	SELECT "content" as classify,dir_path, count(*) as count FROM cms_contents WHERE publish_status = "published" AND dir_path IN (?) AND delete_at=0 GROUP BY dir_path
		UNION ALL
	SELECT "folder" as classify, dir_path, count(*) as count FROM cms_folder_items WHERE item_type=1 AND dir_path IN (?) AND delete_at=0 GROUP BY dir_path;
	`
	res := make([]*entity.FolderItemsCount, 0)
	err = tx.Raw(sql, pathList, pathList).Scan(&res).Error
	if err != nil {
		log.Error(ctx, "Query group dir_path sql failed", log.Err(err),
			log.Strings("fids", fids),
			log.Strings("pathList", pathList),
			log.String("sql", sql))
		return nil, err
	}
	for i := range res {
		//Handle root path
		if res[i].DirPath == constant.FolderRootPath {
			res[i].ID = constant.FolderRootPath
		} else {
			pairs := strings.Split(res[i].DirPath, "/")
			res[i].ID = pairs[len(pairs)-1]
		}
	}
	log.Info(ctx, "query sql", log.String("sql", sql), log.Strings("pathList", pathList))
	return res, nil
}

func (fda *FolderDA) BatchUpdateFolderItemsCount(ctx context.Context, tx *dbo.DBContext, req []*entity.UpdateFolderItemsCountRequest) error {
	if len(req) < 1 {
		//if req is nil, no need to update
		return nil
	}
	sql := `UPDATE cms_folder_items SET items_count = (case id `
	doubleSize := len(req) * 2
	params := make([]interface{}, doubleSize+1)
	ids := make([]string, len(req))
	for i := range req {
		sql = sql + "WHEN ? THEN ? \n"
		params[i*2] = req[i].ID
		params[i*2+1] = req[i].Count
		ids[i] = req[i].ID
	}

	sql = sql + " end) WHERE id IN (?)"
	params[doubleSize] = ids

	err := tx.Exec(sql, params...).Error
	if err != nil {
		log.Error(ctx,
			"BatchGetFolderItemsCount failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("params", params))
		return err
	}
	log.Info(ctx,
		"BatchGetFolderItemsCount success",
		log.String("sql", sql),
		log.Any("params", params))

	return nil
}

//Unused because when old path is root path("/"), replace function in mysql will replace all "/" in path
func (fda *FolderDA) BatchReplaceFolderPath(ctx context.Context, tx *dbo.DBContext, fids []string, oldPath, path entity.Path) error {
	// err := tx.Model(entity.FolderItem{}).Where("id IN (?)", fids).Updates(map[string]interface{}{"path": path}).Error
	if len(fids) < 1 {
		//若fids为空，则不更新
		//if fids is nil, no need to update
		return nil
	}
	fidsSQLParts := make([]string, len(fids))
	params := []interface{}{oldPath, path}
	for i := range fids {
		fidsSQLParts[i] = "?"
		params = append(params, fids[i])
	}
	fidsSQL := strings.Join(fidsSQLParts, constant.StringArraySeparator)

	sql := fmt.Sprintf(`UPDATE cms_folder_items SET dir_path = replace(dir_path,?,?) WHERE id IN (%s)`, fidsSQL)
	err := tx.Exec(sql, params...).Error

	log.Info(ctx, "update folder",
		log.String("sql", sql),
		log.Any("params", params))
	if err != nil {
		log.Error(ctx, "update folder da failed", log.Err(err),
			log.Strings("fids", fids),
			log.String("path", string(path)),
			log.String("oldPath", string(oldPath)),
			log.String("sql", sql),
			log.Any("params", params))
		return err
	}

	return nil
}

func (fda *FolderDA) BatchUpdateFolderPathPrefix(ctx context.Context, tx *dbo.DBContext, fids []string, prefix entity.Path) error {
	// err := tx.Model(entity.FolderItem{}).Where("id IN (?)", fids).Updates(map[string]interface{}{"path": path}).Error
	if len(fids) < 1 {
		//若fids为空，则不更新
		//if fids is nil, no need to update
		return nil
	}

	fidsSQLParts := make([]string, len(fids))
	params := []interface{}{prefix}
	for i := range fids {
		fidsSQLParts[i] = "?"
		params = append(params, fids[i])
	}
	fidsSQL := strings.Join(fidsSQLParts, ",")

	sql := fmt.Sprintf(`UPDATE cms_folder_items SET dir_path = CONCAT(?, dir_path) WHERE id IN (%s)`, fidsSQL)
	err := tx.Exec(sql, params...).Error

	log.Info(ctx, "update folder",
		log.String("sql", sql),
		log.Any("params", params))
	if err != nil {
		log.Error(ctx, "update folder da failed", log.Err(err),
			log.Strings("fids", fids),
			log.String("path", string(prefix)),
			log.String("sql", sql),
			log.Any("params", params))
		return err
	}

	return nil
}
func (fda *FolderDA) BatchUpdateFolderPathByLink(ctx context.Context, tx *dbo.DBContext, link []string, path entity.Path) error {
	err := tx.Model(entity.FolderItem{}).Where("link IN (?)", link).Updates(map[string]interface{}{"path": path}).Error
	if err != nil {
		log.Error(ctx, "update folder da failed", log.Err(err), log.Strings("link", link), log.String("path", string(path)))
		return err
	}

	return nil
}

func (fda *FolderDA) BatchUpdateFoldersPath(ctx context.Context, tx *dbo.DBContext, fids []string, dirPath entity.Path) error {
	if len(fids) < 1 {
		return nil
	}
	err := tx.Model(entity.FolderItem{}).Where("id IN (?)", fids).Updates(entity.FolderItem{DirPath: dirPath, ParentID: dirPath.Parent()}).Error
	if err != nil {
		return err
	}
	return nil
}

func (fda *FolderDA) BatchDeleteFolders(ctx context.Context, tx *dbo.DBContext, fids []string) error {
	if len(fids) < 1 {
		return nil
	}
	err := tx.Model(entity.FolderItem{}).Where("id IN (?)", fids).Updates(entity.FolderItem{DeleteAt: time.Now().Unix()}).Error
	if err != nil {
		return err
	}
	return nil
}

func (fda *FolderDA) DeleteFolder(ctx context.Context, tx *dbo.DBContext, fid string) error {
	folderItem, err := fda.GetFolderByID(ctx, tx, fid)
	if err != nil {
		return err
	}
	folderItem.ID = fid
	folderItem.DeleteAt = time.Now().Unix()
	_, err = fda.s.UpdateTx(ctx, tx, folderItem)
	if err != nil {
		log.Error(ctx, "delete folder da failed", log.Err(err), log.String("id", fid))
		return err
	}
	return nil
}

func (fda *FolderDA) GetFolderByID(ctx context.Context, tx *dbo.DBContext, fid string) (*entity.FolderItem, error) {
	obj := new(entity.FolderItem)
	err := fda.s.GetTx(ctx, tx, fid, obj)
	if err != nil {
		log.Error(ctx, "get folder da failed", log.Err(err), log.String("id", fid))
		return nil, err
	}
	if obj.DeleteAt > 0 {
		log.Info(ctx, "folder was deleted", log.String("id", fid))
		return nil, dbo.ErrRecordNotFound
	}

	return obj, nil
}

func (fda *FolderDA) GetFolderByIDList(ctx context.Context, tx *dbo.DBContext, fids []string) ([]*entity.FolderItem, error) {
	if len(fids) < 1 {
		return nil, nil
	}
	objs := make([]*entity.FolderItem, 0)
	err := fda.s.QueryTx(ctx, tx, &FolderCondition{
		IDs: entity.NullStrings{
			Strings: fids,
			Valid:   true,
		},
	}, &objs)
	if err != nil {
		return nil, err
	}

	return objs, nil
}

func (fda *FolderDA) SearchFolderPage(ctx context.Context, tx *dbo.DBContext, condition FolderCondition) (int, []*entity.FolderItem, error) {
	objs := make([]*entity.FolderItem, 0)
	count, err := fda.s.PageTx(ctx, tx, &condition, &objs)
	if err != nil {
		return 0, nil, err
	}

	return count, objs, nil
}
func (fda *FolderDA) SearchFolder(ctx context.Context, tx *dbo.DBContext, condition FolderCondition) ([]*entity.FolderItem, error) {
	objs := make([]*entity.FolderItem, 0)
	err := fda.s.QueryTx(ctx, tx, &condition, &objs)
	if err != nil {
		return nil, err
	}

	return objs, nil
}
func (fda *FolderDA) SearchFolderCount(ctx context.Context, tx *dbo.DBContext, condition FolderCondition) (int, error) {
	total, err := fda.s.CountTx(ctx, tx, &condition, entity.FolderItem{})
	if err != nil {
		return 0, err
	}

	return total, nil
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
	IDs       entity.NullStrings
	OwnerType int
	ItemType  int
	Owner     string
	ParentID  string
	Partition entity.FolderPartition
	Link      string

	NameLike          string
	Name              string
	DirDescendant     string
	DirDescendantList []string

	Editors []string
	Disable bool

	ExactDirPath string

	OrderBy FolderOrderBy `json:"order_by"`
	Pager   utils.Pager
}

func (s *FolderCondition) GetConditions() ([]string, []interface{}) {
	conditions := make([]string, 0)
	params := make([]interface{}, 0)

	if s.IDs.Valid {
		condition := " id in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.IDs.Strings)
	}
	if s.OwnerType > 0 {
		conditions = append(conditions, "owner_type = ?")
		params = append(params, s.OwnerType)
	}
	if s.Partition != "" {
		conditions = append(conditions, "`partition` = ?")
		params = append(params, string(s.Partition))
	}
	if s.Owner != "" {
		conditions = append(conditions, "owner = ?")
		params = append(params, s.Owner)
	}
	if s.ItemType != 0 {
		conditions = append(conditions, "item_type = ?")
		params = append(params, s.ItemType)
	}
	if len(s.Editors) > 0 {
		conditions = append(conditions, "editor in (?)")
		params = append(params, s.Editors)
	}

	if s.Disable {
		conditions = append(conditions, "1!=1")
	}

	if s.ParentID != "" {
		conditions = append(conditions, "parent_id = ?")
		params = append(params, s.ParentID)
	}
	if s.Link != "" {
		conditions = append(conditions, "link = ?")
		params = append(params, s.Link)
	}
	if s.NameLike != "" {
		condition := "match(name, description, keywords) against(? in boolean mode)"
		conditions = append(conditions, condition)
		params = append(params, s.NameLike)
	}
	if s.Name != "" {
		conditions = append(conditions, "name = ?")
		params = append(params, s.Name)
	}

	if s.DirDescendant != "" {
		conditions = append(conditions, "dir_path like ?")
		params = append(params, s.DirDescendant+"%")
	}
	if len(s.DirDescendantList) > 0 {
		subCondition := make([]string, len(s.DirDescendantList))
		for i := range s.DirDescendantList {
			subCondition[i] = "dir_path like ?"
			params = append(params, s.DirDescendantList[i]+"%")
		}
		condition := "(" + strings.Join(subCondition, " or ") + ")"
		conditions = append(conditions, condition)
	}

	if s.ExactDirPath != "" {
		conditions = append(conditions, "dir_path = ?")
		params = append(params, s.ExactDirPath)
	}

	//if len(s.VisibilitySetting) != 0 {
	//	conditions = append(conditions, "visibility_setting in (?)")
	//	params = append(params, s.VisibilitySetting)
	//}

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
	folderDA      IFolderDA
	_folderDAOnce sync.Once
)

func GetFolderDA() IFolderDA {
	_folderDAOnce.Do(func() {
		folderDA = new(FolderDA)
	})

	return folderDA
}
