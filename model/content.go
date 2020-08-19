package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/cache"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/contentdata"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	mutex "gitlab.badanamu.com.cn/calmisland/kidsloop2/mutext"
	"sync"
	"time"
)

type IContentModel interface {
	CreateContent(ctx context.Context, tx *dbo.DBContext, c entity.CreateContentRequest, operator *entity.Operator) (string, error)
	UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, data entity.CreateContentRequest, user *entity.Operator) error
	PublishContent(ctx context.Context, tx *dbo.DBContext, cid, scope string, user *entity.Operator) error

	LockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error)
	//UnlockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error

	GetContentById(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error)
	GetContentByIdList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) ([]*entity.ContentInfoWithDetails, error)

	DeleteContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error
	CloneContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error)

	UpdateContentPublishStatus(ctx context.Context, tx *dbo.DBContext, cid, reason, status string) error
	CheckContentAuthorization(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error

	SearchUserContent(ctx context.Context, tx *dbo.DBContext, condition da.DyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error)
	SearchUserPrivateContent(ctx context.Context, tx *dbo.DBContext, condition da.DyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error)
	ListPendingContent(ctx context.Context, tx *dbo.DBContext, condition da.DyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error)
	SearchContent(ctx context.Context, tx *dbo.DBContext, condition da.DyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error)

	GetVisibleContentById(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error)
}

type ContentModel struct {
}

func (cm *ContentModel) handleSourceContent(ctx context.Context, tx *dbo.DBContext, contentId, sourceId string) error{
	sourceContent, err := da.GetDyContentDA().GetContentById(ctx, sourceId)
	if err != nil{
		return err
	}
	now := time.Now()
	sourceContent.PublishStatus = entity.ContentStatusHidden
	sourceContent.LatestId = contentId
	//解锁source content
	sourceContent.LockedBy = "-"
	sourceContent.UpdatedAt = &now
	err = da.GetDyContentDA().UpdateContent(ctx, sourceId, *sourceContent)
	if err != nil {
		log.Error(ctx, "update source content failed", log.Err(err))
		return ErrUpdateContentFailed
	}
	return nil
}

func (cm *ContentModel) preparePublishContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	err := cm.checkPublishContent(ctx, tx, content, user)
	if err != nil {
		log.Error(ctx, "check content scope & sub content scope failed", log.Err(err))
		return err
	}
	if user.OrgID == content.Org && user.Role != "teacher" {
		content.PublishStatus = entity.ContentStatusPublished
		//直接发布，则顶替旧
		if content.SourceId != "" {
			//存在前序content，则隐藏前序
			err = cm.handleSourceContent(ctx, tx, content.ID, content.SourceId)
			if err != nil{
				return err
			}
		}

	} else {
		//TODO:更新发布状态
		content.PublishStatus = entity.ContentStatusPending
	}
	return nil
}

func (cm *ContentModel) doPublishContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	//TODO:Maybe wrong
	err := cm.preparePublishContent(ctx, tx, content, user)
	if err != nil {
		log.Error(ctx, "prepare publish failed", log.Err(err))
		return err
	}

	err = da.GetDyContentDA().UpdateContent(ctx, content.ID, *content)
	if err != nil {
		log.Error(ctx, "update lesson plan failed", log.Err(err))
		return ErrUpdateContentFailed
	}

	return nil
}

func (cm ContentModel) checkContentInfo(ctx context.Context, c entity.CreateContentRequest, created bool) error {
	//TODO:Check age, category...
	return nil
}

func (cm ContentModel) checkUpdateContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) (*entity.Content, error) {
	//TODO:maybe wrong
	if content.Author != user.UserID {
		return nil, ErrNoAuth
	}
	if content.PublishStatus == entity.ContentStatusPending ||
		content.PublishStatus == entity.ContentStatusArchive ||
		content.PublishStatus == entity.ContentStatusHidden ||
		content.PublishStatus == entity.ContentStatusAttachment{
		return nil, ErrInvalidPublishStatus
	}
	if content.PublishStatus == entity.ContentStatusPublished {
		return nil, ErrInvalidPublishStatus
	}
	return content, nil
}

func (cm ContentModel) checkPublishContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	//若content为已发布状态或发布中状态，则创建新content
	if content.PublishStatus != entity.ContentStatusDraft && content.PublishStatus != entity.ContentStatusRejected{
		//报错
		return ErrInvalidContentStatusToPublish
	}

	//TODO:检查子内容是否合法
	contentData, err := contentdata.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil{
		return err
	}
	subContentIds, err := contentData.SubContentIds(ctx)
	if err != nil{
		return err
	}
	_, contentList, err := da.GetDyContentDA().SearchContent(ctx, &da.DyContentCondition{
		IDS: subContentIds,
	})
	if err != nil {
		return err
	}
	err = checkPublishContentChildren(ctx, content, contentList)
	if err != nil {
		return err
	}

	return nil
}

func (cm *ContentModel) searchContent(ctx context.Context, tx *dbo.DBContext, condition *da.DyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error) {
	cachedContents := cache.GetRedisContentCache().GetContentCacheBySearchCondition(ctx, condition)
	if cachedContents != nil{
		return cachedContents.Key, cachedContents.ContentList, nil
	}

	key, objs, err := da.GetDyContentDA().SearchContent(ctx, condition)
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Err(err))
		return "", nil, ErrReadContentFailed
	}
	response := make([]*entity.ContentInfo, len(objs))
	for i := range objs {
		temp, err := contentdata.ConvertContentObj(ctx, objs[i])
		if err != nil {
			log.Error(ctx, "Can't parse contentdata, contentId: %v, error: %v", log.String("id", objs[i].ID), log.Err(err))
			return "", nil, err
		}
		response[i] = temp
	}
	contentWithDetails, err := buildContentWithDetails(ctx, response, user)
	if err != nil {
		return "", nil, err
	}
	return key, contentWithDetails, nil
}

func (cm *ContentModel) searchContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition *da.DyCombineContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error) {
	cachedContents := cache.GetRedisContentCache().GetContentCacheBySearchCondition(ctx, condition)
	if cachedContents != nil{
		return cachedContents.Key, cachedContents.ContentList, nil
	}

	key, objs, err := da.GetDyContentDA().SearchContent(ctx, condition)
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Err(err))
		return "", nil, ErrReadContentFailed
	}
	response := make([]*entity.ContentInfo, len(objs))
	for i := range objs {
		temp, err := contentdata.ConvertContentObj(ctx, objs[i])
		if err != nil {
			log.Error(ctx, "can't parse contentdata", log.Err(err))
			return "", nil, err
		}
		response[i] = temp
	}
	contentWithDetails, err := buildContentWithDetails(ctx, response, user)
	if err != nil {
		return "", nil, err
	}
	return key, contentWithDetails, nil
}

func (cm *ContentModel) CreateContent(ctx context.Context, tx *dbo.DBContext, c entity.CreateContentRequest, operator *entity.Operator) (string, error) {
	//检查数据信息是否正确
	err := cm.checkContentInfo(ctx, c, true)
	if err != nil {
		return "", err
	}

	//组装要创建的内容
	obj, err := cm.prepareCreateContentParams(ctx, c, operator)
	if err != nil {
		return "", err
	}

	//若要发布，则过滤状态
	if c.DoPublish{
		err = cm.preparePublishContent(ctx, tx, obj, operator)
		if err != nil {
			log.Error(ctx, "publish content failed", log.Err(err))
			return "", err
		}
	}

	//添加内容
	pid, err := da.GetDyContentDA().CreateContent(ctx, *obj)
	if err != nil {
		log.Error(ctx, "can't create contentdata", log.Err(err))
		return "", ErrCreateContentFailed
	}

	return pid, nil
}

func (cm *ContentModel) UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, data entity.CreateContentRequest, user *entity.Operator) error {
	err := cm.checkContentInfo(ctx, data, false)
	if err != nil {
		return err
	}
	content, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil {
		log.Error(ctx, "can't read contentdata on update contentdata", log.Err(err))
		return ErrNoContentData
	}

	content, err = cm.checkUpdateContent(ctx, tx, content, user)
	if err != nil {
		return err
	}
	//Check status
	//Check Authorization
	obj, err := cm.prepareUpdateContentParams(ctx, content, &data)
	if err != nil {
		log.Error(ctx, "can't prepare params contentdata on update contentdata", log.Err(err))
		return err
	}

	//若要发布，则过滤状态
	if data.DoPublish{
		err = cm.preparePublishContent(ctx, tx, obj, user)
		if err != nil {
			log.Error(ctx, "publish content failed", log.Err(err))
			return err
		}
	}

	//更新数据库
	err = da.GetDyContentDA().UpdateContent(ctx, cid, *obj)
	if err != nil {
		log.Error(ctx, "update contentdata failed", log.Err(err))
		return ErrUpdateContentFailed
	}

	cache.GetRedisContentCache().CleanContentCache(ctx, []string{cid})

	return nil
}

func (cm *ContentModel) UpdateContentPublishStatus(ctx context.Context, tx *dbo.DBContext, cid, reason, status string) error {
	content, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil {
		log.Error(ctx, "can't read contentdata on update contentdata", log.Err(err))
		return ErrReadContentFailed
	}

	content.PublishStatus = entity.NewContentPublishStatus(status)
	content.RejectReason = reason
	err = da.GetDyContentDA().UpdateContent(ctx, cid, *content)
	if err != nil {
		log.Error(ctx, "update contentdata scope failed", log.Err(err))
		return ErrUpdateContentFailed
	}
	if status == entity.ContentStatusPublished {
		//处理source content
		err = cm.handleSourceContent(ctx, tx, content.ID, content.SourceId)
		if err != nil{
			return err
		}
	}

	cache.GetRedisContentCache().CleanContentCache(ctx, []string{cid, content.SourceId})

	return nil
}

func (cm *ContentModel) UnlockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error{
	content, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil {
		log.Error(ctx, "can't read contentdata for publishing", log.Err(err))
		return ErrNoContent
	}
	//TODO:检查权限
	//if content.LockedBy != user.UserID {
	//	return ErrNoAuth
	//}
	now := time.Now()
	content.LockedBy = "-"
	content.UpdatedAt = &now
	return da.GetDyContentDA().UpdateContent(ctx, cid, *content)
}
func (cm *ContentModel) LockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error){
	locker, err := mutex.NewLock(ctx, "content", "lock", cid)
	if err != nil{
		return "", err
	}
	locker.Lock()
	defer locker.Unlock()

	content, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil {
		log.Error(ctx, "can't read contentdata for publishing", log.Err(err))
		return  "", ErrNoContent
	}
	if content.PublishStatus != entity.ContentStatusPublished {
		return  "", ErrInvalidPublishStatus
	}

	//更新锁定状态
	if content.LockedBy != "" && content.LockedBy != "-" {
		return "", ErrContentAlreadyLocked
	}
	now := time.Now()
	content.LockedBy = cid
	content.UpdatedAt = &now
	err = da.GetDyContentDA().UpdateContent(ctx, cid, *content)
	if err != nil {
		return "", err
	}
	//TODO:检查权限
	//克隆Content
	ccid, err := cm.CloneContent(ctx, tx, content.ID, user)
	if err != nil {
		return "", err
	}

	//Update with new content
	return ccid, nil

}

func (cm *ContentModel) PublishContent(ctx context.Context, tx *dbo.DBContext, cid, scope string, user *entity.Operator) error {
	content, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil {
		log.Error(ctx, "can't read contentdata for publishing", log.Err(err))
		return ErrNoContent
	}

	//发布
	content.PublishScope = scope
	err = cm.doPublishContent(ctx, tx, content, user)
	if err != nil {
		return err
	}

	cache.GetRedisContentCache().CleanContentCache(ctx, []string{cid, content.SourceId})
	return nil
}

func (cm *ContentModel) DeleteContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error {
	content, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil {
		log.Error(ctx, "can't read contentdata on delete contentdata", log.Err(err))
		return ErrReadContentFailed
	}

	if content.Author != user.UserID {
		return ErrNoAuth
	}

	obj, err := cm.prepareDeleteContentParams(ctx, content, content.PublishStatus)
	if err != nil {
		log.Error(ctx, "prepare contentdata params failed", log.Err(err))
		return ErrDeleteContentFailed
	}

	err = da.GetDyContentDA().UpdateContent(ctx, cid, *obj)
	if err != nil {
		log.Error(ctx, "delete contentdata failed", log.Err(err))
		return ErrDeleteContentFailed
	}

	//解锁source content
	if content.SourceId != "" {
		err = cm.UnlockContent(ctx, tx, content.SourceId, user)
		if err != nil{
			return err
		}
	}

	cache.GetRedisContentCache().CleanContentCache(ctx, []string{cid, content.SourceId})
	return nil
}

func (cm *ContentModel) CloneContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error){
	content, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil {
		log.Error(ctx, "can't read contentdata on update contentdata", log.Err(err))
		return "", ErrNoContent
	}

	//检查是否有克隆权限
	err = cm.CheckContentAuthorization(ctx, tx, &entity.Content{
		ID:            content.ID,
		PublishScope:  content.PublishScope,
		PublishStatus: content.PublishStatus,
		Author:      	content.Author,
		Org:    		content.Org,
	}, user)
	if err != nil {
		log.Error(ctx, "no auth to read content for cloning", log.Err(err))
		return "", ErrCloneContentFailed
	}

	obj := cm.prepareCloneContentParams(ctx, content, user)

	id, err := da.GetDyContentDA().CreateContent(ctx, *obj)
	if err != nil {
		log.Error(ctx, "clone contentdata failed", log.Err(err))
		return "", ErrUpdateContentFailed
	}

	return id, nil
}

func (cm *ContentModel) CheckContentAuthorization(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	if user.UserID == content.Author {
		return nil
	}
	//TODO:maybe wrong
	if user.Role != "teacher" {
		return nil
	}

	if content.PublishStatus == entity.ContentStatusAttachment ||
		content.PublishStatus == entity.ContentStatusHidden{
		return nil
	}

	if content.PublishStatus != entity.ContentStatusPublished {
		log.Error(ctx, "read unpublished content, userId: %v, contentId: %v", log.String("userID", user.UserID), log.String("contentID", content.ID))
		return ErrGetUnpublishedContent
	}
	//TODO: Check org scope

	return ErrGetUnauthorizedContent
}

func (cm *ContentModel) GetContentById(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error) {
	cachedContent := cache.GetRedisContentCache().GetContentCacheById(ctx, cid)
	if cachedContent != nil{
		return cachedContent, nil
	}

	obj, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Err(err))
		return nil, ErrNoContent
	}
	content, err := contentdata.ConvertContentObj(ctx, obj)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Err(err))
		return nil, ErrParseContentDataFailed
	}

	//补全相关内容
	err = content.Data.PrepareResult(ctx)
	if err != nil {
		log.Error(ctx, "can't get contentdata for details", log.Err(err))
		return nil, ErrParseContentDataDetailsFailed
	}

	contentWithDetails, err := buildContentWithDetails(ctx, []*entity.ContentInfo{content}, user)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Err(err))
		return nil, ErrReadContentFailed
	}
	if len(contentWithDetails) < 1 {
		return &entity.ContentInfoWithDetails{
			ContentInfo: *content,
		}, nil
	}

	cache.GetRedisContentCache().SaveContentCacheList(ctx, contentWithDetails)
	return contentWithDetails[0], nil
}

func (cm *ContentModel) GetContentByIdList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) ([]*entity.ContentInfoWithDetails, error) {
	nid, cachedContent := cache.GetRedisContentCache().GetContentCacheByIdList(ctx, cids)
	if len(nid) < 1{
		return cachedContent, nil
	}

	if len(cids) < 1 {
		return nil, nil
	}
	_, data, err := da.GetDyContentDA().SearchContent(ctx, &da.DyContentCondition{
		IDS: cids,
	})
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Err(err))
		return nil, ErrReadContentFailed
	}
	res := make([]*entity.ContentInfo, len(data))
	for i := range data {
		temp, err := contentdata.ConvertContentObj(ctx, data[i])
		if err != nil {
			log.Error(ctx, "can't parse contentdata", log.String("id", data[i].ID), log.Err(err))
			return nil, ErrReadContentFailed
		}
		res[i] = temp
	}
	contentWithDetails, err := buildContentWithDetails(ctx, res, user)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Err(err))
		return nil, ErrReadContentFailed
	}

	cache.GetRedisContentCache().SaveContentCacheList(ctx, contentWithDetails)

	return contentWithDetails, nil
}

func (cm *ContentModel) SearchUserContent(ctx context.Context, tx *dbo.DBContext, condition da.DyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error) {
	condition1 := condition
	condition2 := condition

	//condition1为自己的内容
	condition1.Author = user.UserID
	condition1.PublishStatus = filterInvisiblePublishStatus(ctx, condition1.PublishStatus)

	//condition2为别人的内容
	//过滤未发布的内容
	condition2.PublishStatus = []string{entity.ContentStatusPublished}

	//过滤可见范围以外的内容
	scopes := listVisibleScopes(ctx, user)

	condition2.Scope = scopes
	//过滤（已购买 + 免费） => （cid in () or price == -1）

	combineCondition := &da.DyCombineContentCondition{
		Condition1: &condition1,
		Condition2: &condition2,
	}
	//where, params := combineCondition.GetConditions()
	//logger.WithContext(ctx).WithField("subject", "content").Infof("Combine condition: %#v, params: %#v", where, params)
	return cm.searchContentUnsafe(ctx, tx, combineCondition, user)
}

func (cm *ContentModel) SearchUserPrivateContent(ctx context.Context, tx *dbo.DBContext, condition da.DyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error) {
	condition.Author = user.UserID
	condition.PublishStatus = filterInvisiblePublishStatus(ctx, condition.PublishStatus)

	return cm.searchContent(ctx, tx, &condition, user)
}

func (cm *ContentModel) ListPendingContent(ctx context.Context, tx *dbo.DBContext, condition da.DyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error) {
	condition.PublishStatus = []string{entity.ContentStatusPending}
	return cm.searchContent(ctx, tx, &condition, user)
}

func (cm *ContentModel) SearchContent(ctx context.Context, tx *dbo.DBContext, condition da.DyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error) {
	return cm.searchContent(ctx, tx, &condition, user)
}

func (cm *ContentModel) GetVisibleContentById(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error){
	var err error
	var contentData *entity.Content

	cachedContent := cache.GetRedisContentCache().GetContentCacheById(ctx, cid)
	if cachedContent != nil {
		if cachedContent.LatestID == "" {
			return cachedContent, nil
		}else{
			latestCachedContent := cache.GetRedisContentCache().GetContentCacheById(ctx, cachedContent.LatestID)
			if latestCachedContent != nil {
				return latestCachedContent, nil
			}else{
				contentData = &entity.Content{LatestId: cachedContent.LatestID}
			}
		}
	}

	if contentData == nil {
		contentData, err = da.GetDyContentDA().GetContentById(ctx, cid)
		if err != nil{
			return nil, err
		}
	}

	if contentData.LatestId != "" {
		newContentData, err := da.GetDyContentDA().GetContentById(ctx, contentData.LatestId)
		if err != nil{
			return nil, err
		}
		contentData = newContentData
	}

	content, err := contentdata.ConvertContentObj(ctx, contentData)
	if err != nil{
		return nil, err
	}
	contentWithDetails, err := buildContentWithDetails(ctx, []*entity.ContentInfo{content}, user)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Err(err))
		return nil, ErrReadContentFailed
	}
	if len(contentWithDetails) < 1 {
		return &entity.ContentInfoWithDetails{
			ContentInfo: *content,
		}, nil
	}
	cache.GetRedisContentCache().SaveContentCacheList(ctx, contentWithDetails)
	return contentWithDetails[0], nil
}

func buildContentWithDetails(ctx context.Context, contentList []*entity.ContentInfo, user *entity.Operator) ([]*entity.ContentInfoWithDetails, error) {
	return nil, nil
}

func listVisibleScopes(ctx context.Context, operator *entity.Operator) []string {
	return nil
}

func checkPublishContentChildren(ctx context.Context, c *entity.Content, children []*entity.Content) error{
	return nil
}

func filterInvisiblePublishStatus(ctx context.Context, status []string) []string {
	newStatus := make([]string, 0)
	for i := range status {
		if status[i] != entity.ContentStatusAttachment &&
			status[i] != entity.ContentStatusArchive &&
			status[i] != entity.ContentStatusHidden{
			newStatus = append(newStatus, status[i])
		}
	}
	return newStatus
}

var(
	_contentModel IContentModel
	_contentModelOnce sync.Once
)

func GetContentModel() IContentModel {
	_contentModelOnce.Do(func() {
		_contentModel = new(ContentModel)
	})
	return _contentModel
}