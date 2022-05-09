package da

import (
	"context"
	"fmt"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
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
	IDS                entity.NullStrings `json:"ids"`
	Name               string             `json:"name"`
	ContentType        []int              `json:"content_type"`
	VisibilitySettings []string           `json:"visibility_settings"`
	PublishStatus      []string           `json:"publish_status"`
	Author             string             `json:"author"`
	Org                string             `json:"org"`
	Program            []string           `json:"program"`
	SourceID           string             `json:"source_id"`
	LatestID           string             `json:"latest_id"`
	SourceType         string             `json:"source_type"`
	DirPath            entity.NullStrings `json:"dir_path"`

	DirPathRecursion     string   `json:"dir_path_recursion"`
	DirPathRecursionList []string `json:"dir_path_recursion_list"`
	ContentName          string   `json:"content_name"`

	ParentPath entity.NullStrings `json:"parent_path"`
	OrderBy    ContentOrderBy     `json:"order_by"`
	Pager      utils.Pager

	DataSourceID string `json:"data_source_id"`

	CreateAtLe int `json:"create_at_le"`
	CreateAtGe int `json:"create_at_ge"`

	JoinUserIDList []string `json:"join_user_id_list"`
	IncludeDeleted bool
}

func (s *ContentCondition) GetConditions() ([]string, []interface{}) {
	conditions := make([]string, 0)
	params := make([]interface{}, 0)

	if s.IDS.Valid {
		condition := " id in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.IDS.Strings)
	}
	if s.Name != "" {
		condition := "match(content_name, description, keywords) against(? in boolean mode)"
		if len(s.JoinUserIDList) > 0 {
			condition1 := "author in (?)"
			where := "(" + condition + " OR " + condition1 + ")"

			conditions = append(conditions, where)
			params = append(params, s.Name)
			params = append(params, s.JoinUserIDList)
		} else {
			conditions = append(conditions, condition)
			params = append(params, s.Name)
		}
	}
	if s.ContentName != "" {
		conditions = append(conditions, "content_name LIKE ?")
		params = append(params, s.ContentName)
	}
	if s.DataSourceID != "" {
		conditions = append(conditions, "data->'$.source' = ?")
		params = append(params, s.DataSourceID)
	}

	if len(s.Program) > 0 {
		conditions = append(conditions, "(EXISTS (SELECT 1 FROM cms_content_properties WHERE property_type=? AND property_id IN (?) AND cms_content_properties.content_id=cms_contents.id))")
		params = append(params, entity.ContentPropertyTypeProgram)
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

	if s.ParentPath.Valid {
		if len(s.ParentPath.Strings) > 0 {
			conds := make([]string, len(s.ParentPath.Strings))
			for i, v := range s.ParentPath.Strings {
				conds[i] = "dir_path like " + "'" + v + "%" + "'"
			}
			condition := fmt.Sprintf("(%s)", strings.Join(conds, " or "))
			conditions = append(conditions, condition)
		} else {
			condition := "1=0"
			conditions = append(conditions, condition)
		}
	}

	if len(s.VisibilitySettings) > 0 {
		condition := "id IN (SELECT content_id FROM cms_content_visibility_settings WHERE visibility_setting IN (?))"
		conditions = append(conditions, condition)
		params = append(params, s.VisibilitySettings)
	}

	if s.SourceType != "" {
		condition := "source_type = ?"
		conditions = append(conditions, condition)
		params = append(params, s.SourceType)
	}
	if s.DirPath.Valid {
		condition := "dir_path IN (?)"
		conditions = append(conditions, condition)
		params = append(params, s.DirPath.Strings)
	}
	if s.DirPathRecursion != "" {
		condition := "dir_path LIKE ?"
		conditions = append(conditions, condition)
		params = append(params, s.DirPathRecursion+"%")
	}
	if len(s.DirPathRecursionList) > 0 {
		subCondition := make([]string, len(s.DirPathRecursionList))
		for i := range s.DirPathRecursionList {
			subCondition[i] = "dir_path like ?"
			params = append(params, s.DirPathRecursionList[i]+"%")
		}
		condition := "(" + strings.Join(subCondition, " or ") + ")"
		conditions = append(conditions, condition)
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

	if s.CreateAtLe > 0 {
		condition := " create_at < ? "
		conditions = append(conditions, condition)
		params = append(params, s.CreateAtLe)
	}

	if s.CreateAtGe > 0 {
		condition := " create_at > ? "
		conditions = append(conditions, condition)
		params = append(params, s.CreateAtGe)
	}

	if !s.IncludeDeleted {
		conditions = append(conditions, "delete_at=0")
	}

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

type ContentMySQLDA struct {
	s dbo.BaseDA
}

func (cd *ContentMySQLDA) CreateContent(ctx context.Context, tx *dbo.DBContext, co entity.Content) (string, error) {
	co.ID = utils.NewID()
	_, err := cd.s.InsertTx(ctx, tx, &co)
	if err != nil {
		return "", err
	}
	return co.ID, nil
}

func (cd *ContentMySQLDA) BatchCreateContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, cid string, scope []string) error {
	cv := make([]entity.ContentVisibilitySetting, len(scope))
	for i := range scope {
		cv[i] = entity.ContentVisibilitySetting{
			ContentID:         cid,
			VisibilitySetting: scope[i],
		}
	}
	for i := range cv {
		_, err := cd.s.InsertTx(ctx, tx, &cv[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (cd *ContentMySQLDA) BatchDeleteContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, cid string, visibilitySettings []string) error {
	err := tx.Where("content_id = ? AND visibility_setting IN (?)",
		cid, visibilitySettings).Delete(entity.ContentVisibilitySetting{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (cd *ContentMySQLDA) UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, co entity.Content) error {
	co.ID = cid
	log.Info(ctx, "Update contentdata path", log.String("id", co.ID))
	_, err := cd.s.UpdateTx(ctx, tx, &co)
	if err != nil {
		return err
	}

	return nil
}
func (cd *ContentMySQLDA) BatchUpdateContentPath(ctx context.Context, tx *dbo.DBContext, cids []string, dirPath entity.Path) error {
	if len(cids) < 1 {
		return nil
	}
	tx.ResetCondition()
	err := tx.Model(entity.Content{}).Where("id IN (?)", cids).Updates(entity.Content{DirPath: dirPath, ParentFolder: dirPath.Parent()}).Error
	if err != nil {
		return err
	}
	return nil
}

func (cd *ContentMySQLDA) GetContentByID(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.Content, error) {
	obj := new(entity.Content)
	err := cd.s.GetTx(ctx, tx, cid, obj)
	if err != nil {
		return nil, err
	}
	if obj.DeleteAt > 0 {
		log.Error(ctx, "record deleted", log.String("id", cid), log.Any("content", obj))
		return nil, dbo.ErrRecordNotFound
	}

	return obj, nil
}
func (cd *ContentMySQLDA) GetContentByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.Content, error) {
	if len(cids) < 1 {
		return nil, nil
	}
	objs := make([]*entity.Content, 0)

	err := cd.s.QueryTx(ctx, tx, &ContentCondition{
		IDS: entity.NullStrings{Strings: cids, Valid: true},
	}, &objs)
	if err != nil {
		return nil, err
	}

	return objs, nil
}

func (cd *ContentMySQLDA) GetContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error) {
	objs := make([]*entity.ContentVisibilitySetting, 0)
	if cid == "" {
		return nil, nil
	}
	err := cd.s.QueryTx(ctx, tx, &ContentVisibilitySettingsCondition{
		ContentIDs: []string{cid},
	}, &objs)
	if err != nil {
		return nil, err
	}
	ret := make([]string, len(objs))
	for i := range objs {
		ret[i] = objs[i].VisibilitySetting
	}

	return ret, nil
}

func (cd *ContentMySQLDA) SearchContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, condition *ContentVisibilitySettingsCondition) ([]*entity.ContentVisibilitySetting, error) {
	if condition.Empty() {
		return nil, nil
	}
	objs := make([]*entity.ContentVisibilitySetting, 0)
	err := cd.s.QueryTx(ctx, tx, condition, &objs)
	if err != nil {
		return nil, err
	}
	return objs, nil
}
func (cd *ContentMySQLDA) SearchContent(ctx context.Context, tx *dbo.DBContext, condition *ContentCondition) (int, []*entity.Content, error) {
	objs := make([]*entity.Content, 0)
	count, err := cd.s.PageTx(ctx, tx, condition, &objs)
	if err != nil {
		return 0, nil, err
	}

	return count, objs, nil
}
func (cd *ContentMySQLDA) SearchContentUnSafe(ctx context.Context, tx *dbo.DBContext, condition dbo.Conditions) (int, []*entity.Content, error) {
	objs := make([]*entity.Content, 0)
	count, err := cd.s.PageTx(ctx, tx, condition, &objs)
	if err != nil {
		return 0, nil, err
	}

	return count, objs, nil
}

func (cd *ContentMySQLDA) QueryContent(ctx context.Context, tx *dbo.DBContext, condition *ContentCondition) ([]*entity.Content, error) {
	objs := make([]*entity.Content, 0)
	err := cd.s.QueryTx(ctx, tx, condition, &objs)
	return objs, err
}

func (cm *ContentMySQLDA) BatchReplaceContentPath(ctx context.Context, tx *dbo.DBContext, cids []string, oldPath, path string) error {
	// err := tx.Model(entity.FolderItem{}).Where("id IN (?)", fids).Updates(map[string]interface{}{"path": path}).Error
	if len(cids) < 1 {
		//若fids为空，则不更新
		//if fids is nil, no need to update
		return nil
	}
	fidsSQLParts := make([]string, len(cids))
	params := []interface{}{oldPath, path}
	for i := range cids {
		fidsSQLParts[i] = "?"
		params = append(params, cids[i])
	}
	fidsSQL := strings.Join(fidsSQLParts, constant.StringArraySeparator)

	tx.ResetCondition()
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

type ContentVisibilitySettingsCondition struct {
	IDS                []string `json:"ids"`
	VisibilitySettings []string `json:"visibility_settings"`
	ContentIDs         []string `json:"content_i_ds"`
	Pager              utils.Pager
}

func (s *ContentVisibilitySettingsCondition) Empty() bool {
	if len(s.IDS) == 0 && len(s.VisibilitySettings) == 0 && len(s.ContentIDs) == 0 {
		return true
	}
	return false
}

func (s *ContentVisibilitySettingsCondition) GetConditions() ([]string, []interface{}) {
	conditions := make([]string, 0)
	params := make([]interface{}, 0)

	if len(s.IDS) > 0 {
		condition := " id in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.IDS)
	}

	if len(s.ContentIDs) > 0 {
		condition := " content_id in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.ContentIDs)
	}

	if len(s.VisibilitySettings) > 0 {
		condition := " visibility_setting in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.IDS)
	}
	return conditions, params
}

func (s *ContentVisibilitySettingsCondition) GetPager() *dbo.Pager {
	return &dbo.Pager{
		Page:     int(s.Pager.PageIndex),
		PageSize: int(s.Pager.PageSize),
	}
}
func (s *ContentVisibilitySettingsCondition) GetOrderBy() string {
	return "content_id"
}

type TotalContentResponse struct {
	Total int
}

func (cd *ContentMySQLDA) SearchFolderContent(ctx context.Context, tx *dbo.DBContext, condition1 ContentCondition, condition2 *FolderCondition) (int, []*entity.FolderContent, error) {
	return cd.doSearchFolderContent(ctx, tx, &condition1, condition2)
}

func (cd *ContentMySQLDA) SearchFolderContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition1 dbo.Conditions, condition2 *FolderCondition) (int, []*entity.FolderContent, error) {
	return cd.doSearchFolderContent(ctx, tx, condition1, condition2)
}

func (cd *ContentMySQLDA) CountFolderContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition1 dbo.Conditions, condition2 *FolderCondition) (int, error) {
	query1, params1 := condition1.GetConditions()
	query2, params2 := condition2.GetConditions()

	params1 = append(params2, params1...)
	var total TotalContentResponse
	var err error
	//获取数量
	//get folder total
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

func (cd *ContentMySQLDA) appendConditionParamWithOrderAndPage(sql string, orderBy string, pager *dbo.Pager) string {

	if orderBy != "" {
		sql = fmt.Sprintf("%s order by %s", sql, orderBy)
	}
	if pager != nil && pager.Enable() {
		offset, limit := pager.Offset()
		sql = fmt.Sprintf("%s limit %d offset %d", sql, limit, offset)
	}
	return sql
}

func (cd *ContentMySQLDA) doSearchFolderContent(ctx context.Context, tx *dbo.DBContext, condition1 dbo.Conditions, condition2 dbo.Conditions) (int, []*entity.FolderContent, error) {
	query1, params1 := condition1.GetConditions()
	query2, params2 := condition2.GetConditions()

	params1 = append(params2, params1...)
	var total TotalContentResponse
	var err error
	//获取数量
	//get folder total
	query := cd.countFolderContentSQL(query1, query2)
	err = cd.s.QueryRawSQLTx(ctx, tx, &total, query, params1...)
	if err != nil {
		log.Error(ctx, "count raw sql failed", log.Err(err),
			log.String("query", query), log.Any("params", params1),
			log.Any("condition1", condition1), log.Any("condition2", condition2))
		return 0, nil, err
	}

	//查询
	//Query
	folderContents := make([]*entity.FolderContent, 0)

	orderBy := condition1.GetOrderBy()
	pager := condition1.GetPager()
	exSql := cd.appendConditionParamWithOrderAndPage(cd.searchFolderContentSQL(ctx, query1, query2), orderBy, pager)
	err = cd.s.QueryRawSQLTx(ctx, tx, &folderContents, exSql, params1...)
	if err != nil {
		log.Error(ctx, "query raw sql failed", log.Err(err),
			log.String("query", query), log.Any("params", params1),
			log.Any("condition1", condition1), log.Any("condition2", condition2))
		return 0, nil, err
	}
	return total.Total, folderContents, nil
}

func (cd *ContentMySQLDA) searchFolderContentSQL(ctx context.Context, query1, query2 []string) string {
	rawQuery1 := strings.Join(query1, " and ")
	rawQuery2 := strings.Join(query2, " and ")
	sql := fmt.Sprintf(`SELECT 
	id, %v as content_type, name AS content_name, items_count, description AS description, keywords as keywords, creator as author, dir_path, 'published' as publish_status, thumbnail, '' as data, create_at, update_at 
	FROM cms_folder_items 
	WHERE  %v
	UNION ALL SELECT 
	id, content_type, content_name, 0 AS items_count, description, keywords, author, dir_path, publish_status, thumbnail, data, create_at, update_at
	FROM cms_contents 
	WHERE %v`, entity.AliasContentTypeFolder, rawQuery2, rawQuery1)

	log.Info(ctx, "search folder content", log.String("sql", sql))
	return sql
}

func (cd *ContentMySQLDA) countFolderContentSQL(query1, query2 []string) string {
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

func (cd *ContentMySQLDA) SearchSharedContentV2(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, op *entity.Operator) (response entity.QuerySharedContentV2Response, err error) {
	response.Items = []*entity.QuerySharedContentV2Item{}
	sbFolderID := NewSqlBuilder(ctx, `
select id from cms_folder_items cfi 
where cfi.id in (
	SELECT
		csf.folder_id
	from
		cms_shared_folders csf
	where
		csf.org_id in (?)
		and csf.delete_at = 0
		 
	) 
and cfi.parent_id = ?
and cfi.delete_at = 0
`,
		[]interface{}{op.OrgID, constant.ShareToAll},
		condition.ParentID,
	)

	sbFolderSelect := NewSqlBuilder(ctx, `
cfi.id,
? as content_type,
cfi.name,
cfi.thumbnail,
cfi.creator as author,
? as publish_status, 
cfi.update_at,
cfi.create_at,
cfi.name as content_name
`, entity.AliasContentTypeFolder, entity.ContentStatusPublished)
	sbParentID := NewSqlBuilder(ctx, `
and cfi.parent_id = ?
`, condition.ParentID)
	sbName := NewSqlBuilder(ctx, ``)
	if condition.Name != "" {
		sbName = NewSqlBuilder(ctx, `and match(cfi.name, cfi.description, cfi.keywords) against(? in boolean mode)`, condition.Name)
	}
	sbContentName := NewSqlBuilder(ctx, ``)
	if condition.ContentName != "" {
		sbContentName = NewSqlBuilder(ctx, `and cfi.name= ? `, condition.ContentName)
	}
	sbFolder := NewSqlBuilder(ctx, `
select 
	{{.sbFolderSelect}}
from cms_folder_items cfi 
where  cfi.id  in ({{.sbFolderID}})
{{.sbParentID}}
{{.sbName}}
{{.sbContentName}}
`).Replace(ctx, "sbFolderSelect", sbFolderSelect).
		Replace(ctx, "sbFolderID", sbFolderID).
		Replace(ctx, "sbParentID", sbParentID).
		Replace(ctx, "sbName", sbName).
		Replace(ctx, "sbContentName", sbContentName)
	sb := NewSqlBuilder(ctx, `
{{.sbFolder}}
{{.sbContent}}
`).Replace(ctx, "sbFolder", sbFolder)
	sbContent := NewSqlBuilder(ctx, ``)
	if condition.ParentID != "/" {
		sqlWhere := strings.Builder{}
		sqlWhere.WriteString(`
where  cc.parent_folder = ?
and cc.publish_status = ?
and cc.delete_at = 0
`)
		argsWhere := []interface{}{
			condition.ParentID,
			entity.ContentStatusPublished,
		}
		if utils.ContainsString(condition.GroupNames, entity.LessonPlanGroupNameMoreFeaturedContent.String()) {
			sqlWhere.WriteString(`
and EXISTS ( 
	SELECT
		1
	FROM
		cms_content_properties ccp
	WHERE
		ccp.property_type = ?
		and ccp.content_id = cc.id
		AND ccp.property_id not IN (select program_id from programs_groups  )
)
`)
			argsWhere = append(argsWhere, entity.ContentPropertyTypeProgram)

		} else if len(condition.Program) > 0 {
			sqlWhere.WriteString(`
and EXISTS ( 
	SELECT
		1
	FROM
		cms_content_properties ccp
	WHERE
		ccp.property_type = ?
		and ccp.content_id = cc.id
		AND ccp.property_id  IN (?)
)
`)
			argsWhere = append(argsWhere, entity.ContentPropertyTypeProgram, condition.Program)
		}

		if condition.Name != "" {
			sqlWhere.WriteString(`
and match(cc.content_name, cc.description, cc.keywords) against(? in boolean mode)
`)
			argsWhere = append(argsWhere, condition.Name)
		}
		if condition.ContentName != "" {
			sqlWhere.WriteString(`
and cc.content_name = ?
`)
			argsWhere = append(argsWhere, condition.ContentName)
		}
		sbWhere := NewSqlBuilder(ctx, sqlWhere.String(), argsWhere...)

		sqlContent := `
union all
select 
	cc.id,
	cc.	content_type,
	cc.content_name as name,
	cc.thumbnail ,
	cc.author ,
	cc.publish_status , 
	cc.update_at,
	cc.create_at,
	cc.content_name
from cms_contents cc 
{{.sbWhere}}
`
		sbContent = NewSqlBuilder(ctx, sqlContent).Replace(ctx, "sbWhere", sbWhere)
	}
	sb = sb.Replace(ctx, "sbContent", sbContent)
	sql, args, err := sb.Build(ctx)
	if err != nil {
		return
	}

	response.Total, err = cd.PageRawSQL(ctx, &response.Items, NewContentOrderBy(condition.OrderBy).ToSQL(), sql, dbo.Pager{
		Page:     int(condition.Pager.PageIndex),
		PageSize: int(condition.Pager.PageSize),
	}, args...)
	if err != nil {
		return
	}
	return
}
