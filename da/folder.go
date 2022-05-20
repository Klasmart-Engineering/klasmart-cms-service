package da

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/KL-Engineering/kidsloop-cms-service/constant"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
	"github.com/jinzhu/gorm"
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

	GetSharedContentParentPath(ctx context.Context, tx *dbo.DBContext, orgIDs []string) ([]string, error)

	UpdateEmptyField(ctx context.Context, tx *dbo.DBContext, fIDs []string) error
	GetPrivateTree(ctx context.Context, contentCondition *ContentCondition, folderCondition *FolderCondition) (data []*entity.TreeData, err error)
	GetAllTree(ctx context.Context, combineCondition *CombineConditions, folderCondition *FolderCondition) (data []*entity.TreeData, err error)
}

type FolderDA struct {
	s dbo.BaseDA
}

func (fda *FolderDA) UpdateEmptyField(ctx context.Context, tx *dbo.DBContext, fIDs []string) error {
	if len(fIDs) <= 0 {
		log.Info(ctx, "UpdateEmptyField folder id is empty")
		return nil
	}
	sql := `
update cms_folder_items set has_descendant = (
        case when exists (select id from cms_contents where cms_contents.dir_path like concat(if(cms_folder_items.dir_path='/', '', cms_folder_items.dir_path), '/', cms_folder_items.id, '%')  and publish_status='published' and delete_at=0) 
        then 1 else 0 end
) where id in (?);`
	err := tx.Exec(sql, fIDs).Error
	if err != nil {
		log.Error(ctx, "UpdateEmptyField exec sql failed",
			log.Err(err),
			log.String("sql", sql),
			log.Strings("folder_ids", fIDs))
		return err
	}
	return nil
}

func (fda *FolderDA) GetSharedContentParentPath(ctx context.Context, tx *dbo.DBContext, orgIDs []string) ([]string, error) {
	if len(orgIDs) <= 0 {
		return []string{}, nil
	}

	sql := fmt.Sprintf(`select distinct if(dir_path ='/', concat(dir_path, id), concat(dir_path, '/', id)) as parent_path from %s
	where id in (select folder_id from %s where org_id in (?) and delete_at = 0 )`, entity.FolderItem{}.TableName(), entity.SharedFolderRecord{}.TableName())

	//	sql := fmt.Sprintf(`select distinct if(dir_path ='/', concat(dir_path, id), concat(dir_path, '/', id)) as parent_path from %s
	//where id in (select folder_id from %s where org_id in (?))`, entity.FolderItem{}.TableName(), entity.SharedFolderRecord{}.TableName())

	var result []struct {
		ParentPath string `json:"parent_path" gorm:"column:parent_path"`
	}
	err := tx.Raw(sql, orgIDs).Scan(&result).Error
	if err != nil {
		log.Error(ctx, "GetSharedContentParentPath failed",
			log.Strings("orgs", orgIDs),
			log.String("sql", sql),
			log.Err(err))
		return nil, err
	}
	parentsPath := make([]string, len(result))
	for i, v := range result {
		parentsPath[i] = v.ParentPath
	}
	return parentsPath, nil
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
	tx.ResetCondition()
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

	tx.ResetCondition()
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
	tx.ResetCondition()
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
	tx.ResetCondition()
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
	tx.ResetCondition()
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
	tx.ResetCondition()
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
	tx.ResetCondition()
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
func (fda *FolderDA) GetPrivateTree(ctx context.Context, contentCondition *ContentCondition, folderCondition *FolderCondition) (data []*entity.TreeData, err error) {
	sql := getPrivateTreeSql(contentCondition)
	params := map[string]interface{}{
		"PublishStatus":   entity.ContentStatusPublished,
		"Author":          contentCondition.Author,
		"VisibilityID":    contentCondition.VisibilitySettings,
		"JoinUser":        contentCondition.JoinUserIDList,
		"Name":            contentCondition.Name,
		"ContentName":     contentCondition.ContentName,
		"Path":            constant.RootPath,
		"OwnerType":       entity.OwnerTypeOrganization,
		"FolderPartition": entity.FolderPartitionMaterialAndPlans,
		"OrgID":           folderCondition.Owner,
		"FolderItemType":  entity.FolderItemTypeFolder,
	}
	err = fda.s.QueryRawSQL(ctx, &data, sql, params)
	if err != nil {
		log.Error(ctx, "exec GetPrivateTree sql failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("params1", params))
		return
	}
	return
}

func getPrivateTreeSql(contentCondition *ContentCondition) string {
	var sql []string
	var whereRootForMeContentSql []string
	var whereNoRootForMeContentSql []string
	var whereRootForMeContentSqlString string
	var whereNoRootForMeContentSqlString string
	whereRootForMeContentSql = append(whereRootForMeContentSql, `( content_type in (1,2,10)
                      and publish_status in (@PublishStatus) and author =@Author and delete_at=0 `)
	if len(contentCondition.VisibilitySettings) > 0 {
		whereRootForMeContentSql = append(whereRootForMeContentSql, ` and id 
           IN (SELECT content_id FROM cms_content_visibility_settings WHERE visibility_setting IN (@VisibilityID)) `)
	}
	if contentCondition.Name != "" {
		whereRootForMeContentSql = append(whereRootForMeContentSql, ` and (
                      match(content_name, description, keywords) against(@Name in boolean mode) `)
		if len(contentCondition.JoinUserIDList) > 0 {
			whereRootForMeContentSql = append(whereRootForMeContentSql, ` OR author in (@JoinUser) `)
		}
		whereRootForMeContentSql = append(whereRootForMeContentSql, ` ) `)

	}
	if contentCondition.ContentName != "" {
		whereRootForMeContentSql = append(whereRootForMeContentSql, ` and content_name like @ContentName  `)
	}
	whereNoRootForMeContentSql = append(whereNoRootForMeContentSql, whereRootForMeContentSql...)
	whereRootForMeContentSql = append(whereRootForMeContentSql, ` and dir_path = @Path ) `)
	whereNoRootForMeContentSql = append(whereNoRootForMeContentSql, ` and dir_path <> @Path ) `)
	whereRootForMeContentSqlString = strings.Join(whereRootForMeContentSql, " ")
	whereNoRootForMeContentSqlString = strings.Join(whereNoRootForMeContentSql, " ")

	sql = append(sql, `select * from 
           (
              select cms_folder.parent_id, cms_folder.id,cms_folder.name,cms_folder.dir_path,
              ifnull(cms_folder_content.content_count,0) as content_count,1 as item_type,
             `)
	if contentCondition.Name != "" {
		sql = append(sql, `case when match(cms_folder.name, cms_folder.description, cms_folder.keywords) 
                against(@Name in boolean mode)  then 1 else 0 end as has_search_self ,`)
	}
	if contentCondition.ContentName != "" {
		sql = append(sql, `case when cms_folder.name like @ContentName then 1 else 0 end as has_search_self ,`)
	}
	if contentCondition.Name == "" && contentCondition.ContentName == "" {
		sql = append(sql, `1 as has_search_self ,`)
	}
	sql = append(sql, `cms_folder.has_descendant from cms_folder_items cms_folder left join 
          (
            SELECT folder.id,sum(case when isnull(content.content_name) then 0 else 1 end)  content_count FROM cms_folder_items folder
            left join 
            (
             select * from cms_contents where `+whereNoRootForMeContentSqlString+`
             ) content 
             on folder.id=content.parent_folder
             where  folder.owner_type = @OwnerType and folder.partition = @FolderPartition 
             and folder.owner = @OrgID and folder.item_type = @FolderItemType  and folder.delete_at = 0
             group by  folder.id
          ) cms_folder_content
          on cms_folder.id=cms_folder_content.id
          where cms_folder.owner_type = @OwnerType and cms_folder.partition = @FolderPartition 
          and cms_folder.owner = @OrgID and cms_folder.item_type =@FolderItemType  and cms_folder.delete_at = 0
          union all
          select '' as parent_id,'' as  id,'' as name,'' as dir_path,(select count(*) from cms_contents where `+whereRootForMeContentSqlString+
		`) as content_count,0 as item_type,0 as has_search_self,0 has_descendant
          ) tree_data
          order by name `)
	querySql := strings.Join(sql, "")
	return querySql
}

func getAllTreeSql(forMeCondition *ContentCondition, forOtherCondition *ContentCondition) string {
	var sql []string
	var whereRootForMeContentSql []string
	var whereNoRootForMeContentSql []string
	var whereRootForOtherContentSql []string
	var whereNoRootForOtherContentSql []string
	var whereRootForAllContentSqlString string
	var whereNoRootForAllContentSqlString string

	// for me condition
	whereRootForMeContentSql = append(whereRootForMeContentSql, ` ( content_type in (1,2,10)
                     and publish_status in (@PublishStatus) and author =@Author and delete_at=0 `)
	if len(forMeCondition.VisibilitySettings) > 0 {
		whereRootForMeContentSql = append(whereRootForMeContentSql, ` and id
          IN (SELECT content_id FROM cms_content_visibility_settings WHERE visibility_setting IN (@MeVisibilityID)) `)
	}
	if forMeCondition.Name != "" {
		whereRootForMeContentSql = append(whereRootForMeContentSql, ` and (
                     match(content_name, description, keywords) against(@Name in boolean mode) `)
		if len(forMeCondition.JoinUserIDList) > 0 {
			whereRootForMeContentSql = append(whereRootForMeContentSql, ` OR author in (@JoinMeUser) `)
		}
		whereRootForMeContentSql = append(whereRootForMeContentSql, ` ) `)

	}
	if forMeCondition.ContentName != "" {
		whereRootForMeContentSql = append(whereRootForMeContentSql, ` and content_name like @ContentName  `)
	}
	whereNoRootForMeContentSql = append(whereNoRootForMeContentSql, whereRootForMeContentSql...)
	whereRootForMeContentSql = append(whereRootForMeContentSql, ` and dir_path = @Path ) `)
	whereNoRootForMeContentSql = append(whereNoRootForMeContentSql, ` and dir_path <> @Path ) `)

	//for other condition
	whereRootForOtherContentSql = append(whereRootForMeContentSql, ` ( content_type in (1,2,10)
                     and publish_status in (@PublishStatus) and delete_at=0 `)
	if len(forOtherCondition.VisibilitySettings) > 0 {
		whereRootForOtherContentSql = append(whereRootForOtherContentSql, ` and id
          IN (SELECT content_id FROM cms_content_visibility_settings WHERE visibility_setting IN (@OtherVisibilityID)) `)
	}
	if forOtherCondition.Name != "" {
		whereRootForOtherContentSql = append(whereRootForOtherContentSql, ` and (
                     match(content_name, description, keywords) against(@Name in boolean mode) `)
		if len(forOtherCondition.JoinUserIDList) > 0 {
			whereRootForOtherContentSql = append(whereRootForOtherContentSql, ` OR author in (@JoinOtherUser) `)
		}
		whereRootForOtherContentSql = append(whereRootForOtherContentSql, ` ) `)

	}
	if forOtherCondition.ContentName != "" {
		whereRootForOtherContentSql = append(whereRootForOtherContentSql, ` and content_name like @ContentName  `)
	}
	whereNoRootForOtherContentSql = append(whereNoRootForOtherContentSql, whereRootForOtherContentSql...)
	whereRootForOtherContentSql = append(whereRootForOtherContentSql, ` and dir_path = @Path ) `)
	whereNoRootForOtherContentSql = append(whereNoRootForOtherContentSql, ` and dir_path <> @Path ) `)

	//query sql
	rootForMeContentSqlString := strings.Join(whereRootForMeContentSql, " ")
	rootForOtherContentSqlString := strings.Join(whereRootForOtherContentSql, " ")
	noRootForMeContentSqlString := strings.Join(whereNoRootForMeContentSql, " ")
	noRootForOtherContentSqlString := strings.Join(whereNoRootForOtherContentSql, " ")
	whereRootForAllContentSqlString = " ( " + rootForMeContentSqlString + " or " + rootForOtherContentSqlString + ")"
	whereNoRootForAllContentSqlString = " ( " + noRootForMeContentSqlString + " or " + noRootForOtherContentSqlString + ")"

	sql = append(sql, `select * from 
           (
              select cms_folder.parent_id, cms_folder.id,cms_folder.name,cms_folder.dir_path,
              ifnull(cms_folder_content.content_count,0) as content_count,1 as item_type,
             `)
	if forMeCondition.Name != "" {
		sql = append(sql, `case when match(cms_folder.name, cms_folder.description, cms_folder.keywords) 
                against(@Name in boolean mode)  then 1 else 0 end as has_search_self ,`)
	}
	if forMeCondition.ContentName != "" {
		sql = append(sql, `case when cms_folder.name like @ContentName then 1 else 0 end as has_search_self ,`)
	}
	if forMeCondition.Name == "" && forMeCondition.ContentName == "" {
		sql = append(sql, `1 as has_search_self ,`)
	}
	sql = append(sql, `cms_folder.has_descendant from cms_folder_items cms_folder left join 
          (
            SELECT folder.id,sum(case when isnull(content.content_name) then 0 else 1 end)  content_count FROM cms_folder_items folder
            left join 
            (
             select * from cms_contents where `+whereNoRootForAllContentSqlString+`
             ) content 
             on folder.id=content.parent_folder
             where  folder.owner_type = @OwnerType and folder.partition = @FolderPartition 
             and folder.owner = @OrgID and folder.item_type = @FolderItemType  and folder.delete_at = 0
             group by  folder.id
          ) cms_folder_content
          on cms_folder.id=cms_folder_content.id
          where cms_folder.owner_type = @OwnerType and cms_folder.partition = @FolderPartition 
          and cms_folder.owner = @OrgID and cms_folder.item_type =@FolderItemType  and cms_folder.delete_at = 0
          union all
          select '' as parent_id,'' as  id,'' as name,'' as dir_path,(select count(*) from cms_contents where `+whereRootForAllContentSqlString+
		`) as content_count,0 as item_type,0 as has_search_self,0 has_descendant
          ) tree_data
          order by name `)
	querySql := strings.Join(sql, "")
	return querySql
}
func (fda *FolderDA) GetAllTree(ctx context.Context, combineCondition *CombineConditions, folderCondition *FolderCondition) (data []*entity.TreeData, err error) {
	forMeCondition := combineCondition.SourceCondition.(*ContentCondition)
	forOtherCondition := combineCondition.TargetCondition.(*ContentCondition)
	sql := getAllTreeSql(forMeCondition, forOtherCondition)
	params := map[string]interface{}{
		"PublishStatus":     entity.ContentStatusPublished,
		"Author":            forMeCondition.Author,
		"MeVisibilityID":    forMeCondition.VisibilitySettings,
		"JoinMeUser":        forMeCondition.JoinUserIDList,
		"Name":              forMeCondition.Name,
		"ContentName":       forMeCondition.ContentName,
		"Path":              constant.RootPath,
		"OwnerType":         entity.OwnerTypeOrganization,
		"FolderPartition":   entity.FolderPartitionMaterialAndPlans,
		"OrgID":             folderCondition.Owner,
		"FolderItemType":    entity.FolderItemTypeFolder,
		"OtherVisibilityID": forOtherCondition.VisibilitySettings,
		"JoinOtherUser":     forOtherCondition.JoinUserIDList,
	}
	err = fda.s.QueryRawSQL(ctx, &data, sql, params)
	if err != nil {
		log.Error(ctx, "exec GetAllTree sql failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("params1", params))
		return
	}
	return
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

	ShowEmptyFolder entity.NullBool `json:"show_empty_folder"`

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

	if s.ShowEmptyFolder.Valid && !s.ShowEmptyFolder.Bool {
		conditions = append(conditions, "has_descendant = ?")
		params = append(params, 1)
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
