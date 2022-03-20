package da

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IContentDA interface {
	CreateContent(ctx context.Context, tx *dbo.DBContext, co entity.Content) (string, error)
	UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, co entity.Content) error
	GetContentByID(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.Content, error)

	SearchContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, condition *ContentVisibilitySettingsCondition) ([]*entity.ContentVisibilitySetting, error)
	GetContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error)
	BatchCreateContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, cid string, scope []string) error
	BatchDeleteContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, cid string, visibilitySettings []string) error
	BatchUpdateContentPath(ctx context.Context, tx *dbo.DBContext, cids []string, dirPath entity.Path) error

	GetContentByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.Content, error)
	SearchContent(ctx context.Context, tx *dbo.DBContext, condition *ContentCondition) (int, []*entity.Content, error)
	SearchContentUnSafe(ctx context.Context, tx *dbo.DBContext, condition dbo.Conditions) (int, []*entity.Content, error)
	QueryContent(cgtx context.Context, tx *dbo.DBContext, condition *ContentCondition) ([]*entity.Content, error)
	SearchFolderContent(ctx context.Context, tx *dbo.DBContext, condition1 ContentCondition, condition2 *FolderCondition) (int, []*entity.FolderContent, error)
	SearchFolderContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition1 dbo.Conditions, condition2 *FolderCondition) (int, []*entity.FolderContent, error)
	CountFolderContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition1 dbo.Conditions, condition2 *FolderCondition) (int, error)

	BatchReplaceContentPath(ctx context.Context, tx *dbo.DBContext, cids []string, oldPath, path string) error

	GetLessonPlansCanSchedule(ctx context.Context, op *entity.Operator, cond *entity.ContentConditionRequest, condOrgContent dbo.Conditions, programGroups []*entity.ProgramGroup) (total int, lps []*entity.LessonPlanForSchedule, err error)
}

var (
	contentDA    IContentDA
	_contentOnce sync.Once
)

func GetContentDA() IContentDA {
	_contentOnce.Do(func() {
		contentDA = &ContentDA{
			redisDA: GetContentRedis(),
			mysqlDA: new(ContentMySQLDA),
		}
	})

	return contentDA
}

type ContentDA struct {
	redisDA IContentRedis
	mysqlDA *ContentMySQLDA
}

func (c ContentDA) CreateContent(ctx context.Context, tx *dbo.DBContext, co entity.Content) (string, error) {
	return c.mysqlDA.CreateContent(ctx, tx, co)
}

func (c ContentDA) UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, co entity.Content) error {
	return c.mysqlDA.UpdateContent(ctx, tx, cid, co)
}

func (c ContentDA) GetContentByID(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.Content, error) {
	return c.mysqlDA.GetContentByID(ctx, tx, cid)
}

func (c ContentDA) SearchContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, condition *ContentVisibilitySettingsCondition) ([]*entity.ContentVisibilitySetting, error) {
	return c.mysqlDA.SearchContentVisibilitySettings(ctx, tx, condition)
}

func (c ContentDA) GetContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error) {
	return c.mysqlDA.GetContentVisibilitySettings(ctx, tx, cid)
}

func (c ContentDA) BatchCreateContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, cid string, scope []string) error {
	return c.mysqlDA.BatchCreateContentVisibilitySettings(ctx, tx, cid, scope)
}

func (c ContentDA) BatchDeleteContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, cid string, visibilitySettings []string) error {
	return c.mysqlDA.BatchDeleteContentVisibilitySettings(ctx, tx, cid, visibilitySettings)
}

func (c ContentDA) BatchUpdateContentPath(ctx context.Context, tx *dbo.DBContext, cids []string, dirPath entity.Path) error {
	return c.mysqlDA.BatchUpdateContentPath(ctx, tx, cids, dirPath)
}

func (c ContentDA) GetContentByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.Content, error) {
	return c.mysqlDA.GetContentByIDList(ctx, tx, cids)
}

func (c ContentDA) SearchContent(ctx context.Context, tx *dbo.DBContext, condition *ContentCondition) (int, []*entity.Content, error) {
	return c.mysqlDA.SearchContent(ctx, tx, condition)
}

func (c ContentDA) SearchContentUnSafe(ctx context.Context, tx *dbo.DBContext, condition dbo.Conditions) (int, []*entity.Content, error) {
	return c.mysqlDA.SearchContentUnSafe(ctx, tx, condition)
}

func (c ContentDA) QueryContent(ctx context.Context, tx *dbo.DBContext, condition *ContentCondition) ([]*entity.Content, error) {
	return c.mysqlDA.QueryContent(ctx, tx, condition)
}

func (c ContentDA) SearchFolderContent(ctx context.Context, tx *dbo.DBContext, condition1 ContentCondition, condition2 *FolderCondition) (int, []*entity.FolderContent, error) {
	return c.mysqlDA.SearchFolderContent(ctx, tx, condition1, condition2)
}

func (c ContentDA) SearchFolderContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition1 dbo.Conditions, condition2 *FolderCondition) (int, []*entity.FolderContent, error) {
	return c.mysqlDA.SearchFolderContentUnsafe(ctx, tx, condition1, condition2)
}

func (c ContentDA) CountFolderContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition1 dbo.Conditions, condition2 *FolderCondition) (int, error) {
	return c.mysqlDA.CountFolderContentUnsafe(ctx, tx, condition1, condition2)
}

func (c ContentDA) BatchReplaceContentPath(ctx context.Context, tx *dbo.DBContext, cids []string, oldPath, path string) error {
	return c.mysqlDA.BatchReplaceContentPath(ctx, tx, cids, oldPath, path)
}

func (c ContentDA) GetLessonPlansCanSchedule(ctx context.Context, op *entity.Operator, cond *entity.ContentConditionRequest, condOrgContent dbo.Conditions, programGroups []*entity.ProgramGroup) (total int, lps []*entity.LessonPlanForSchedule, err error) {
	return c.mysqlDA.GetLessonPlansCanSchedule(ctx, op, cond, condOrgContent, programGroups)
}
