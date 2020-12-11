package da

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

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
	ContentOrderByName
	ContentOrderByNameDesc
)

type IContentDA interface {
	CreateContent(ctx context.Context, tx *dbo.DBContext, co entity.Content) (string, error)
	UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, co entity.Content) error
	DeleteContent(ctx context.Context, tx *dbo.DBContext, cid string) error
	GetContentByID(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.Content, error)

	GetContentByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.Content, error)
	SearchContent(ctx context.Context, tx *dbo.DBContext, condition ContentCondition) (int, []*entity.Content, error)
	SearchContentUnSafe(ctx context.Context, tx *dbo.DBContext, condition dbo.Conditions) (int, []*entity.Content, error)
	Count(context.Context, dbo.Conditions) (int, error)

	SearchFolderContent(ctx context.Context, tx *dbo.DBContext, condition1 ContentCondition, condition2 FolderCondition) (int, []*entity.FolderContent, error)
	SearchFolderContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition1 CombineConditions, condition2 FolderCondition) (int, []*entity.FolderContent, error)
	CountFolderContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition1 CombineConditions, condition2 FolderCondition) (int, error)

	BatchReplaceContentPath(ctx context.Context, tx *dbo.DBContext, cids []string, oldPath, path string) error
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
	IDS           []string `json:"ids"`
	Name          string   `json:"name"`
	ContentType   []int    `json:"content_type"`
	Scope         []string `json:"scope"`
	PublishStatus []string `json:"publish_status"`
	Author        string   `json:"author"`
	Org           string   `json:"org"`
	Program       string   `json:"program"`
	SourceID      string   `json:"source_id"`
	LatestID      string   `json:"latest_id"`
	SourceType    string   `json:"source_type"`
	DirPath       string   `json:"dir_path"`

	AuthedContentFlag bool           `json:"authed_content"`
	AuthedOrgID       string         `json:"authed_org_id"`
	OrderBy           ContentOrderBy `json:"order_by"`
	Pager             utils.Pager

	JoinUserIdList []string `json:"join_user_id_list"`
}

func (s *ContentCondition) GetConditions() ([]string, []interface{}) {
	conditions := make([]string, 0)
	params := make([]interface{}, 0)

	if len(s.IDS) > 0 {
		condition := " id in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.IDS)
	}
	if s.Name != "" {
		condition := "match(content_name, description, keywords) against(? in boolean mode)"
		if len(s.JoinUserIdList) > 0 {
			condition1 := "author in (?)"
			where := "(" + condition + " OR " + condition1 + ")"

			conditions = append(conditions, where)
			params = append(params, s.Name)
			params = append(params, s.JoinUserIdList)
		} else {
			conditions = append(conditions, condition)
			params = append(params, s.Name)
		}
	}

	if s.Program != "" {
		conditions = append(conditions, "program = ?")
		params = append(params, s.Program)
	}

	if len(s.ContentType) > 0 {
		var subConditions []string

		for _, item := range s.ContentType {
			subConditions = append(subConditions, "content_type in (?) ")
			params = append(params, fmt.Sprintf("%d", item))
		}
		conditions = append(conditions, "("+strings.Join(subConditions, " or ")+")")
	}

	//Search authed content
	if s.AuthedContentFlag && s.AuthedOrgID != "" {
		sql := fmt.Sprintf(`select content_id from %v where org_id = ?`, entity.AuthedContentRecord{}.TableName())
		condition := fmt.Sprintf("id in (%v)", sql)
		conditions = append(conditions, condition)
		params = append(params, s.AuthedOrgID)
	}

	if len(s.Scope) > 0 {
		condition := " publish_scope in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.Scope)
	}
	if s.SourceType != "" {
		condition := "source_type = ?"
		conditions = append(conditions, condition)
		params = append(params, s.SourceType)
	}
	if s.DirPath != "" {
		condition := "dir_path = ?"
		conditions = append(conditions, condition)
		params = append(params, s.DirPath)
	}

	if len(s.PublishStatus) > 0 {
		condition := " publish_status in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.PublishStatus)
	}
	if s.SourceID != "" {
		condition := " source_id = ?"
		conditions = append(conditions, condition)
		params = append(params, s.SourceID)
	}
	if s.LatestID != "" {
		condition := " latest_id = ?"
		conditions = append(conditions, condition)
		params = append(params, s.LatestID)
	}

	if s.Author != "" {
		condition := " author = ? "
		conditions = append(conditions, condition)
		params = append(params, s.Author)
	}
	if s.Org != "" {
		condition := " org = ? "
		conditions = append(conditions, condition)
		params = append(params, s.Org)
	}

	conditions = append(conditions, " delete_at = 0")
	return conditions, params
}
func (s *ContentCondition) GetPager() *dbo.Pager {
	return &dbo.Pager{
		Page:     int(s.Pager.PageIndex),
		PageSize: int(s.Pager.PageSize),
	}
}
func (s *ContentCondition) GetOrderBy() string {
	return s.OrderBy.ToSQL()
}

func IsChineseChar(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) || (regexp.MustCompile("[\u3002\uff1b\uff0c\uff1a\u201c\u201d\uff08\uff09\u3001\uff1f\u300a\u300b]").MatchString(string(r))) {
			return true
		}
	}
	return false
}

// NewTeacherOrderBy parse order by
func NewContentOrderBy(orderby string) ContentOrderBy {
	switch orderby {
	case "id":
		return ContentOrderById
	case "-id":
		return ContentOrderByIdDesc
	case "create_at":
		return ContentOrderByCreatedAt
	case "-create_at":
		return ContentOrderByCreatedAtDesc
	case "update_at":
		return ContentOrderByUpdatedAt
	case "-update_at":
		return ContentOrderByUpdatedAtDesc
	case "content_name":
		return ContentOrderByName
	case "-content_name":
		return ContentOrderByNameDesc
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
		return "create_at"
	case ContentOrderByCreatedAtDesc:
		return "create_at desc"
	case ContentOrderByUpdatedAt:
		return "update_at"
	case ContentOrderByUpdatedAtDesc:
		return "update_at desc"
	case ContentOrderByName:
		return "content_name"
	case ContentOrderByNameDesc:
		return "content_name desc"
	default:
		return "content_name"
	}
}

type DBContentDA struct {
	s dbo.BaseDA
}

func (cd *DBContentDA) CreateContent(ctx context.Context, tx *dbo.DBContext, co entity.Content) (string, error) {
	co.ID = utils.NewID()
	_, err := cd.s.InsertTx(ctx, tx, &co)
	if err != nil {
		return "", err
	}
	return co.ID, nil
}
func (cd *DBContentDA) UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, co entity.Content) error {
	co.ID = cid
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
	content.DeleteAt = now.Unix()
	_, err := cd.s.UpdateTx(ctx, tx, content)
	if err != nil {
		return err
	}
	return nil
}
func (cd *DBContentDA) GetContentByID(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.Content, error) {
	obj := new(entity.Content)
	err := cd.s.GetTx(ctx, tx, cid, obj)
	if err != nil {
		return nil, err
	}
	if obj.DeleteAt > 0 {
		return nil, dbo.ErrRecordNotFound
	}

	return obj, nil
}
func (cd *DBContentDA) GetContentByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.Content, error) {
	objs := make([]*entity.Content, 0)
	err := cd.s.QueryTx(ctx, tx, &ContentCondition{
		IDS: cids,
	}, &objs)
	if err != nil {
		return nil, err
	}

	return objs, nil
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

func (cm *DBContentDA) BatchReplaceContentPath(ctx context.Context, tx *dbo.DBContext, cids []string, oldPath, path string) error {
	// err := tx.Model(entity.FolderItem{}).Where("id IN (?)", fids).Updates(map[string]interface{}{"path": path}).Error
	if len(cids) < 1 {
		//若fids为空，则不更新
		return nil
	}
	fidsSQLParts := make([]string, len(cids))
	params := []interface{}{oldPath, path}
	for i := range cids {
		fidsSQLParts[i] = "?"
		params = append(params, cids[i])
	}
	fidsSQL := strings.Join(fidsSQLParts, ",")

	sql := fmt.Sprintf(`UPDATE cms_contents SET dir_path = replace(dir_path,?,?) WHERE id IN (%s)`, fidsSQL)
	err := tx.Exec(sql, params...).Error

	log.Info(ctx, "update folder",
		log.String("sql", sql),
		log.Any("params", params))
	if err != nil {
		log.Error(ctx, "update folder da failed", log.Err(err),
			log.Strings("fids", cids),
			log.String("path", string(path)),
			log.String("oldPath", string(oldPath)),
			log.String("sql", sql),
			log.Any("params", params))
		return err
	}

	return nil
}

type TotalContentResponse struct {
	Total int
}

func (cd *DBContentDA) SearchFolderContent(ctx context.Context, tx *dbo.DBContext, condition1 ContentCondition, condition2 FolderCondition) (int, []*entity.FolderContent, error) {
	return cd.doSearchFolderContent(ctx, tx, &condition1, &condition2)
}

func (cd *DBContentDA) SearchFolderContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition1 CombineConditions, condition2 FolderCondition) (int, []*entity.FolderContent, error) {
	return cd.doSearchFolderContent(ctx, tx, &condition1, &condition2)
}

func (cd *DBContentDA) CountFolderContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition1 CombineConditions, condition2 FolderCondition) (int, error) {
	query1, params1 := condition1.GetConditions()
	query2, params2 := condition2.GetConditions()

	params1 = append(params2, params1...)
	var total TotalContentResponse
	var err error
	//获取数量
	query := cd.countFolderContentSQL(query1, query2)
	err = tx.Raw(query, params1...).Scan(&total).Error
	if err != nil {
		log.Error(ctx, "count raw sql failed", log.Err(err),
			log.String("query", query), log.Any("params", params1),
			log.Any("condition1", condition1), log.Any("condition2", condition2))
		return 0, err
	}
	return total.Total, nil
}

func (cd *DBContentDA) doSearchFolderContent(ctx context.Context, tx *dbo.DBContext, condition1 dbo.Conditions, condition2 dbo.Conditions) (int, []*entity.FolderContent, error) {
	query1, params1 := condition1.GetConditions()
	query2, params2 := condition2.GetConditions()

	params1 = append(params2, params1...)
	var total TotalContentResponse
	var err error
	//获取数量
	query := cd.countFolderContentSQL(query1, query2)
	err = tx.Raw(query, params1...).Scan(&total).Error
	if err != nil {
		log.Error(ctx, "count raw sql failed", log.Err(err),
			log.String("query", query), log.Any("params", params1),
			log.Any("condition1", condition1), log.Any("condition2", condition2))
		return 0, nil, err
	}

	//查询
	folderContents := make([]*entity.FolderContent, 0)
	db := tx.Raw(cd.searchFolderContentSQL(ctx, query1, query2), params1...)
	orderBy := condition1.GetOrderBy()
	if orderBy != "" {
		db = db.Order(orderBy)
	}

	pager := condition1.GetPager()
	if pager != nil && pager.Enable() {
		// pagination
		offset, limit := pager.Offset()
		db = db.Offset(offset).Limit(limit)
	}
	err = db.Find(&folderContents).Error
	if err != nil {
		log.Error(ctx, "query raw sql failed", log.Err(err),
			log.String("query", query), log.Any("params", params1),
			log.Any("condition1", condition1), log.Any("condition2", condition2))
		return 0, nil, err
	}
	return total.Total, folderContents, nil
}

func (cd *DBContentDA) Count(ctx context.Context, condition dbo.Conditions) (int, error) {
	return cd.s.Count(ctx, condition, &entity.Content{})
}

func (cd *DBContentDA) searchFolderContentSQL(ctx context.Context, query1, query2 []string) string {
	rawQuery1 := strings.Join(query1, " and ")
	rawQuery2 := strings.Join(query2, " and ")
	sql := fmt.Sprintf(`SELECT 
	id, %v as content_type, name AS content_name, items_count, '' AS description, '' as keywords, creator as author, dir_path, 'published' as publish_status, thumbnail, '' as data, create_at, update_at 
	FROM cms_folder_items 
	WHERE  %v
	UNION ALL SELECT 
	id, content_type, content_name, 0 AS items_count, description, keywords, author, dir_path, publish_status, thumbnail, data, create_at, update_at
	FROM cms_contents 
	WHERE %v`, entity.AliasContentTypeFolder, rawQuery2, rawQuery1)

	log.Info(ctx, "search folder content", log.String("sql", sql))
	return sql
}

func (cd *DBContentDA) countFolderContentSQL(query1, query2 []string) string {
	rawQuery1 := strings.Join(query1, " and ")
	rawQuery2 := strings.Join(query2, " and ")
	return `SELECT COUNT(*) AS total FROM 
(SELECT 
    name
FROM
    cms_folder_items 
WHERE ` + rawQuery2 + `
UNION ALL (SELECT 
    content_name
FROM
    cms_contents
WHERE ` + rawQuery1 + `
)) AS records;`
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
