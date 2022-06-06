package da

import (
	"context"
	"sync"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
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

	CleanCache(ctx context.Context)
	SearchSharedContentV2(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, op *entity.Operator) (*entity.QuerySharedContentV2Response, error)
}

var (
	contentDA    IContentDA
	_contentOnce sync.Once
)

func GetContentDA() IContentDA {
	_contentOnce.Do(func() {
		da := &ContentDA{
			ContentMySQLDA: new(ContentMySQLDA),
		}

		folderCache, err := utils.NewLazyRefreshCache(&utils.LazyRefreshCacheOption{
			RedisKeyPrefix:  RedisKeyPrefixContentFolderQuery,
			Expiration:      constant.ContentFolderQueryCacheExpiration,
			RefreshDuration: constant.ContentFolderQueryCacheRefreshDuration,
			RawQuery:        da.queryContentsAndFolders})
		if err != nil {
			log.Panic(context.Background(), "create content and folder cache failed", log.Err(err))
		}
		da.contentFolderCache = folderCache

		sharedCache, err := utils.NewLazyRefreshCache(&utils.LazyRefreshCacheOption{
			RedisKeyPrefix:  RedisKeyPrefixContentSharedV2,
			Expiration:      constant.SharedContentQueryCacheExpiration,
			RefreshDuration: constant.SharedContentQueryCacheRefreshDuration,
			RawQuery:        da.querySharedContents})
		if err != nil {
			log.Panic(context.Background(), "create shared content cache failed", log.Err(err))
		}
		da.sharedContentCache = sharedCache

		contentDA = da
	})

	return contentDA
}

type ContentDA struct {
	*ContentMySQLDA
	contentFolderCache *utils.LazyRefreshCache
	sharedContentCache *utils.LazyRefreshCache
}

func (c ContentDA) SearchFolderContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition1 dbo.Conditions, condition2 *FolderCondition) (int, []*entity.FolderContent, error) {
	if !config.Get().RedisConfig.OpenCache {
		return c.ContentMySQLDA.SearchFolderContentUnsafe(ctx, tx, condition1, condition2)
	}

	request := &contentFolderRequest{
		Condition1: condition1,
		Condition2: condition2,
	}

	resp, err := c.queryContentsAndFolders(ctx, request)
	response := resp.(*contentFolderResponse)
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

	total, records, err := c.ContentMySQLDA.SearchFolderContentUnsafe(ctx, dbo.MustGetDB(ctx), request.Condition1, request.Condition2)
	if err != nil {
		return nil, err
	}

	return &contentFolderResponse{Total: total, Records: records}, nil
}

type sharedContentQueryCondition struct {
	Condition *entity.ContentConditionRequest
	Operator  *entity.Operator
}

func (c ContentDA) SearchSharedContentV2(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, op *entity.Operator) (*entity.QuerySharedContentV2Response, error) {
	if !config.Get().RedisConfig.OpenCache {
		return c.ContentMySQLDA.SearchSharedContentV2(ctx, tx, condition, op)
	}

	request := &sharedContentQueryCondition{
		Condition: condition,
		Operator:  op,
	}

	response := &entity.QuerySharedContentV2Response{Items: []*entity.QuerySharedContentV2Item{}}

	err := c.sharedContentCache.Get(ctx, request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c ContentDA) querySharedContents(ctx context.Context, condition interface{}) (interface{}, error) {
	request, ok := condition.(*sharedContentQueryCondition)
	if !ok {
		log.Error(ctx, "invalid request", log.Any("condition", condition))
		return nil, constant.ErrInvalidArgs
	}

	response, err := c.ContentMySQLDA.SearchSharedContentV2(ctx, dbo.MustGetDB(ctx), request.Condition, request.Operator)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c ContentDA) CleanCache(ctx context.Context) {
	err := c.contentFolderCache.Clean(ctx)
	if err != nil {
		log.Warn(ctx, "clean content folder cache failed", log.Err(err))
	}
}

type contentFolderRequest struct {
	Condition1 dbo.Conditions
	Condition2 *FolderCondition
}

type contentFolderResponse struct {
	Total   int `json:"total"`
	Records []*entity.FolderContent
}
