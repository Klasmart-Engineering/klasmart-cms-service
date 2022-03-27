package da

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
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
		da := &ContentDA{
			redisDA: GetContentRedis(),
			mysqlDA: new(ContentMySQLDA),
		}

		cache, err := NewLazyRefreshCache(&LazyRefreshCacheOption{
			RedisKeyPrefix:  RedisKeyPrefixContentFolderQuery,
			RefreshDuration: constant.ContentFolderQueryRefreshDuration,
			RawQuery:        da.queryContentsAndFolders})
		if err != nil {
			log.Panic(context.Background(), "create content and folder cache failed", log.Err(err))
		}

		da.contentFolderCache = cache

		contentDA = da
	})

	return contentDA
}

type ContentDA struct {
	redisDA            IContentRedis
	mysqlDA            *ContentMySQLDA
	contentFolderCache *LazyRefreshCache
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
	if !config.Get().RedisConfig.OpenCache {
		return c.mysqlDA.SearchFolderContentUnsafe(ctx, tx, condition1, condition2)
	}

	request := &contentFolderRequest{
		Tx:         tx,
		Condition1: condition1,
		Condition2: condition2,
	}

	response := &contentFolderResponse{Records: []*entity.FolderContent{}}

	err := c.contentFolderCache.Get(ctx, request, response)
	if err != nil {
		return 0, nil, err
	}

	return response.Total, response.Records, nil
}

func (c ContentDA) queryContentsAndFolders(ctx context.Context, condition interface{}) (interface{}, error) {
	request, ok := condition.(*contentFolderRequest)
	if !ok {
		log.Error(ctx, "invalid request", log.Any("condition", condition))
		return nil, constant.ErrInvalidArgs
	}

	total, records, err := c.mysqlDA.SearchFolderContentUnsafe(ctx, request.Tx, request.Condition1, request.Condition2)
	if err != nil {
		return nil, err
	}

	return &contentFolderResponse{Total: total, Records: records}, nil
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

type contentFolderRequest struct {
	Tx         *dbo.DBContext
	Condition1 dbo.Conditions
	Condition2 *FolderCondition
}

type contentFolderResponse struct {
	Total   int `json:"total"`
	Records []*entity.FolderContent
}
