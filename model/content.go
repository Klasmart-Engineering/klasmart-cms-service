package model

import (
	"context"
	"strings"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/contentdata"
	mutex "gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
)

type IContentModel interface {
	CreateContent(ctx context.Context, tx *dbo.DBContext, c entity.CreateContentRequest, operator *entity.Operator) (string, error)
	UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, data entity.CreateContentRequest, user *entity.Operator) error
	PublishContent(ctx context.Context, tx *dbo.DBContext, cid, scope string, user *entity.Operator) error
	LockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error)
	DeleteContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error
	CloneContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error)

	PublishContentBulk(ctx context.Context, tx *dbo.DBContext, ids []string, user *entity.Operator) error
	DeleteContentBulk(ctx context.Context, tx *dbo.DBContext, ids []string, user *entity.Operator) error

	GetContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error)
	GetContentByIdList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) ([]*entity.ContentInfoWithDetails, error)
	GetContentNameByID(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.ContentName, error)
	GetContentNameByIdList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.ContentName, error)

	UpdateContentPublishStatus(ctx context.Context, tx *dbo.DBContext, cid, reason, status string) error
	CheckContentAuthorization(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error

	SearchUserContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)
	SearchUserPrivateContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)
	ListPendingContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)
	SearchContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)
	SearchContentByDynamoKey(ctx context.Context, tx *dbo.DBContext, condition da.DyKeyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error)

	GetVisibleContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error)
}

type ContentModel struct {
}

func (cm *ContentModel) handleSourceContent(ctx context.Context, tx *dbo.DBContext, contentId, sourceId string) error {
	sourceContent, err := da.GetContentDA().GetContentById(ctx, tx, sourceId)
	if err != nil {
		return err
	}
	sourceContent.PublishStatus = entity.ContentStatusHidden
	sourceContent.LatestID = contentId
	//解锁source content
	sourceContent.LockedBy = "-"
	err = da.GetContentDA().UpdateContent(ctx, tx, sourceId, *sourceContent)
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
	//if user.OrgID == content.Org && user.Role != "teacher" {
	//	content.PublishStatus = entity.ContentStatusPublished
	//	//直接发布，则顶替旧
	//	if content.SourceID != "" {
	//		//存在前序content，则隐藏前序
	//		err = cm.handleSourceContent(ctx, tx, content.ID, content.SourceID)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//
	//} else {
	//	//TODO:更新发布状态
	//	content.PublishStatus = entity.ContentStatusPending
	//}
	content.PublishStatus = entity.ContentStatusPending
	return nil
}

func (cm *ContentModel) doPublishContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	//TODO:Maybe wrong
	err := cm.preparePublishContent(ctx, tx, content, user)
	if err != nil {
		log.Error(ctx, "prepare publish failed", log.Err(err), log.String("cid", content.ID), log.String("uid", user.UserID))
		return err
	}

	err = da.GetContentDA().UpdateContent(ctx, tx, content.ID, *content)
	if err != nil {
		log.Error(ctx, "update lesson plan failed", log.Err(err), log.String("cid", content.ID), log.String("uid", user.UserID))
		return ErrUpdateContentFailed
	}

	return nil
}

func (cm ContentModel) checkContentInfo(ctx context.Context, c entity.CreateContentRequest, created bool) error {
	//TODO:Check age, category...
	if c.Thumbnail != "" {
		parts := strings.Split(c.Thumbnail, "-")
		if len(parts) != 2 {
			return ErrInvalidResourceId
		}
		// _, exist := storage.DefaultStorage().ExistFile(ctx, parts[0], parts[1])
		// if !exist {
		// 	return ErrResourceNotFound
		// }
	}
	return nil
}

func (cm ContentModel) checkUpdateContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) (*entity.Content, error) {

	//若为asset，直接发布
	if content.ContentType.IsAsset() {
		return content, nil
	}

	//TODO:maybe wrong
	if content.Author != user.UserID {
		return nil, ErrNoAuth
	}
	if content.PublishStatus == entity.ContentStatusPending ||
		content.PublishStatus == entity.ContentStatusArchive ||
		content.PublishStatus == entity.ContentStatusHidden ||
		content.PublishStatus == entity.ContentStatusAttachment {
		return nil, ErrInvalidPublishStatus
	}
	if content.PublishStatus == entity.ContentStatusPublished {
		return nil, ErrInvalidPublishStatus
	}
	return content, nil
}

func (cm ContentModel) checkPublishContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	//若content为已发布状态或发布中状态，则创建新content
	if content.PublishStatus != entity.ContentStatusDraft && content.PublishStatus != entity.ContentStatusRejected &&
		content.PublishStatus != entity.ContentStatusArchive {
		//报错
		return ErrInvalidContentStatusToPublish
	}

	//TODO:检查子内容是否合法
	contentData, err := contentdata.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		return err
	}
	subContentIds, err := contentData.SubContentIds(ctx)
	if err != nil {
		return err
	}
	_, contentList, err := da.GetContentDA().SearchContent(ctx, tx, da.ContentCondition{
		IDS: subContentIds,
	})
	if err != nil {
		return err
	}
	err = cm.checkPublishContentChildren(ctx, content, contentList)
	if err != nil {
		return err
	}

	return nil
}

func (cm *ContentModel) SearchContentByDynamoKey(ctx context.Context, tx *dbo.DBContext, condition da.DyKeyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error) {
	condition.OrgUserId = condition.Org + condition.Author
	key, objs, err := da.GetDyContentDA().SearchContentByKey(ctx, condition)
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
	contentWithDetails, err := cm.buildContentWithDetails(ctx, response, user)
	if err != nil {
		return "", nil, err
	}
	return key, contentWithDetails, nil
}

func (cm *ContentModel) searchContent(ctx context.Context, tx *dbo.DBContext, condition *da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	cachedContents := da.GetContentRedis().GetContentCacheBySearchCondition(ctx, condition)
	if cachedContents != nil {
		return cachedContents.Count, cachedContents.ContentList, nil
	}

	count, objs, err := da.GetContentDA().SearchContent(ctx, tx, *condition)
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, err
	}
	response := make([]*entity.ContentInfo, len(objs))
	for i := range objs {
		temp, err := contentdata.ConvertContentObj(ctx, objs[i])
		if err != nil {
			log.Error(ctx, "Can't parse contentdata, contentId: %v, error: %v", log.String("id", objs[i].ID), log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
			return 0, nil, err
		}
		response[i] = temp
	}
	contentWithDetails, err := cm.buildContentWithDetails(ctx, response, user)
	if err != nil {
		log.Error(ctx, "build content details failed", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, err
	}

	da.GetContentRedis().SaveContentCacheListBySearchCondition(ctx, condition, &da.ContentListWithKey{
		Count:       count,
		ContentList: contentWithDetails,
	})
	return count, contentWithDetails, nil
}

func (cm *ContentModel) searchContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition *da.CombineConditions, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	cachedContents := da.GetContentRedis().GetContentCacheBySearchCondition(ctx, condition)
	if cachedContents != nil {
		return cachedContents.Count, cachedContents.ContentList, nil
	}

	count, objs, err := da.GetContentDA().SearchContentUnSafe(ctx, tx, condition)
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, ErrReadContentFailed
	}
	response := make([]*entity.ContentInfo, len(objs))
	for i := range objs {
		temp, err := contentdata.ConvertContentObj(ctx, objs[i])
		if err != nil {
			log.Error(ctx, "can't parse contentdata", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
			return 0, nil, err
		}
		response[i] = temp
	}
	contentWithDetails, err := cm.buildContentWithDetails(ctx, response, user)
	if err != nil {
		log.Error(ctx, "build content details failed", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, err
	}
	da.GetContentRedis().SaveContentCacheListBySearchCondition(ctx, condition, &da.ContentListWithKey{
		Count:       count,
		ContentList: contentWithDetails,
	})

	return count, contentWithDetails, nil
}

func (cm *ContentModel) CreateContent(ctx context.Context, tx *dbo.DBContext, c entity.CreateContentRequest, operator *entity.Operator) (string, error) {
	//检查数据信息是否正确
	log.Info(ctx, "create content")
	err := cm.checkContentInfo(ctx, c, true)
	if err != nil {
		log.Warn(ctx, "check content failed", log.Err(err), log.String("uid", operator.UserID), log.Any("data", c))
		return "", err
	}

	//组装要创建的内容
	obj, err := cm.prepareCreateContentParams(ctx, c, operator)
	if err != nil {
		log.Warn(ctx, "prepare content failed", log.Err(err), log.String("uid", operator.UserID), log.Any("data", c))
		return "", err
	}

	//若要发布，则过滤状态
	if c.DoPublish {
		err = cm.preparePublishContent(ctx, tx, obj, operator)
		if err != nil {
			log.Error(ctx, "publish content failed", log.Err(err), log.String("uid", operator.UserID), log.Any("data", c))
			return "", err
		}
	}

	//添加内容
	pid, err := da.GetContentDA().CreateContent(ctx, tx, *obj)
	if err != nil {
		log.Error(ctx, "can't create contentdata", log.Err(err), log.String("uid", operator.UserID), log.Any("data", c))
		return "", err
	}

	return pid, nil
}

func (cm *ContentModel) UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, data entity.CreateContentRequest, user *entity.Operator) error {
	err := cm.checkContentInfo(ctx, data, false)
	if err != nil {
		return err
	}
	content, err := da.GetContentDA().GetContentById(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata on update contentdata", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID), log.Any("data", data))
		return err
	}

	content, err = cm.checkUpdateContent(ctx, tx, content, user)
	if err != nil {
		return err
	}
	//Check status
	//Check Authorization
	obj, err := cm.prepareUpdateContentParams(ctx, content, &data)
	if err != nil {
		log.Error(ctx, "can't prepare params contentdata on update contentdata", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID), log.Any("data", data))
		return err
	}

	//若要发布，则过滤状态
	if data.DoPublish {
		err = cm.preparePublishContent(ctx, tx, obj, user)
		if err != nil {
			log.Error(ctx, "publish content failed", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID), log.Any("data", data))
			return err
		}
	}

	//更新数据库
	err = da.GetContentDA().UpdateContent(ctx, tx, cid, *obj)
	if err != nil {
		log.Error(ctx, "update contentdata failed", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID), log.Any("data", data))
		return err
	}

	da.GetContentRedis().CleanContentCache(ctx, []string{cid})

	return nil
}

func (cm *ContentModel) UpdateContentPublishStatus(ctx context.Context, tx *dbo.DBContext, cid, reason, status string) error {
	content, err := da.GetContentDA().GetContentById(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't read contentdata on update contentdata", log.Err(err))
		return err
	}
	if content.ContentType.IsAsset() {
		return ErrInvalidContentType
	}

	content.PublishStatus = entity.NewContentPublishStatus(status)
	content.RejectReason = reason
	err = da.GetContentDA().UpdateContent(ctx, tx, cid, *content)
	if err != nil {
		log.Error(ctx, "update contentdata scope failed", log.Err(err))
		return ErrUpdateContentFailed
	}
	if status == entity.ContentStatusPublished && content.SourceID != "" {
		//处理source content
		err = cm.handleSourceContent(ctx, tx, content.ID, content.SourceID)
		if err != nil {
			return err
		}
	}

	da.GetContentRedis().CleanContentCache(ctx, []string{cid, content.SourceID})

	return nil
}

func (cm *ContentModel) UnlockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error {
	content, err := da.GetContentDA().GetContentById(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata for publishing", log.Err(err))
		return err
	}
	//TODO:检查权限
	//if content.LockedBy != user.UserID {
	//	return ErrNoAuth
	//}
	content.LockedBy = "-"
	return da.GetContentDA().UpdateContent(ctx, tx, cid, *content)
}
func (cm *ContentModel) LockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error) {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixContentLock, cid)
	if err != nil {
		return "", err
	}
	locker.Lock()
	defer locker.Unlock()

	content, err := da.GetContentDA().GetContentById(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return "", ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata for publishing", log.Err(err))
		return "", err
	}
	if content.ContentType.IsAsset() {
		return "", ErrInvalidContentType
	}

	if content.PublishStatus != entity.ContentStatusPublished {
		return "", ErrInvalidPublishStatus
	}

	//更新锁定状态
	if content.LockedBy != "" && content.LockedBy != "-" {
		return "", ErrContentAlreadyLocked
	}
	content.LockedBy = user.UserID
	err = da.GetContentDA().UpdateContent(ctx, tx, cid, *content)
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
func (cm *ContentModel) PublishContentBulk(ctx context.Context, tx *dbo.DBContext, ids []string, user *entity.Operator) error {
	updateIds := make([]string, 0)
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		_, contents, err := da.GetContentDA().SearchContent(ctx, tx, da.ContentCondition{
			IDS: ids,
		})
		if err != nil {
			log.Error(ctx, "can't read content on delete contentdata", log.Err(err), log.Strings("ids", ids), log.String("uid", user.UserID))
			return err
		}
		for i := range contents {
			if contents[i].ContentType.IsAsset() {
				log.Warn(ctx, "try to publish asset", log.Err(err), log.Strings("ids", ids), log.String("uid", user.UserID))
				continue
			}
			err = cm.doPublishContent(ctx, tx, contents[i], user)
			if err != nil {
				log.Error(ctx, "can't publish content", log.Err(err), log.Strings("ids", ids), log.String("uid", user.UserID))
				return err
			}
			updateIds = append(updateIds, contents[i].ID)
			if contents[i].SourceID != "" {
				updateIds = append(updateIds, contents[i].SourceID)
			}
		}
		return nil
	})

	da.GetContentRedis().CleanContentCache(ctx, updateIds)
	return err
}

func (cm *ContentModel) PublishContent(ctx context.Context, tx *dbo.DBContext, cid, scope string, user *entity.Operator) error {
	content, err := da.GetContentDA().GetContentById(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata for publishing", log.Err(err), log.String("cid", cid), log.String("scope", scope), log.String("uid", user.UserID))
		return err
	}
	if content.ContentType.IsAsset() {
		return ErrInvalidContentType
	}

	//发布
	if scope != "" {
		content.PublishScope = scope
	}

	err = cm.doPublishContent(ctx, tx, content, user)
	if err != nil {
		return err
	}

	da.GetContentRedis().CleanContentCache(ctx, []string{cid, content.SourceID})
	return nil
}

func (cm *ContentModel) DeleteContentBulk(ctx context.Context, tx *dbo.DBContext, ids []string, user *entity.Operator) error {
	deletedIds := make([]string, 0)
	deletedIds = append(deletedIds, ids...)
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		_, contents, err := da.GetContentDA().SearchContent(ctx, tx, da.ContentCondition{
			IDS: ids,
		})
		if err != nil {
			log.Error(ctx, "can't read content on delete contentdata", log.Err(err), log.Strings("ids", ids), log.String("uid", user.UserID))
			return err
		}
		for i := range contents {
			err = cm.doDeleteContent(ctx, tx, contents[i], user)
			if err != nil {
				return err
			}
			//record pending delete content id
			if contents[i].SourceID != "" {
				deletedIds = append(deletedIds, contents[i].SourceID)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	da.GetContentRedis().CleanContentCache(ctx, deletedIds)
	return nil
}

func (cm *ContentModel) doDeleteContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	if content.Author != user.UserID {
		return ErrNoAuth
	}
	if content.LockedBy != "-" && content.LockedBy != user.UserID {
		return ErrContentAlreadyLocked
	}

	obj := cm.prepareDeleteContentParams(ctx, content, content.PublishStatus)

	err := da.GetContentDA().UpdateContent(ctx, tx, content.ID, *obj)
	if err != nil {
		log.Error(ctx, "delete contentdata failed", log.Err(err), log.String("cid", content.ID), log.String("uid", user.UserID))
		return err
	}

	//解锁source content
	if content.SourceID != "" {
		err = cm.UnlockContent(ctx, tx, content.SourceID, user)
		if err != nil {
			log.Error(ctx, "unlock contentdata failed", log.Err(err), log.String("cid", content.SourceID), log.String("uid", user.UserID))
			return err
		}
	}
	return nil
}

func (cm *ContentModel) DeleteContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error {
	content, err := da.GetContentDA().GetContentById(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "content not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read content on delete content", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return err
	}

	err = cm.doDeleteContent(ctx, tx, content, user)
	if err != nil {
		return err
	}

	da.GetContentRedis().CleanContentCache(ctx, []string{cid, content.SourceID})
	return nil
}

func (cm *ContentModel) CloneContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error) {
	content, err := da.GetContentDA().GetContentById(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return "", ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata on update contentdata", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return "", err
	}
	if content.ContentType.IsAsset() {
		return "", ErrInvalidContentType
	}

	//检查是否有克隆权限
	err = cm.CheckContentAuthorization(ctx, tx, &entity.Content{
		ID:            content.ID,
		PublishScope:  content.PublishScope,
		PublishStatus: content.PublishStatus,
		Author:        content.Author,
		Org:           content.Org,
	}, user)
	if err != nil {
		log.Error(ctx, "no auth to read content for cloning", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return "", ErrCloneContentFailed
	}

	obj := cm.prepareCloneContentParams(ctx, content, user)

	id, err := da.GetContentDA().CreateContent(ctx, tx, *obj)
	if err != nil {
		log.Error(ctx, "clone contentdata failed", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return "", err
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
		content.PublishStatus == entity.ContentStatusHidden {
		return nil
	}

	if content.PublishStatus != entity.ContentStatusPublished {
		log.Error(ctx, "read unpublished content, userId: %v, contentId: %v", log.String("userID", user.UserID), log.String("contentID", content.ID))
		return ErrGetUnpublishedContent
	}
	//TODO: Check org scope

	return ErrGetUnauthorizedContent
}

func (cm *ContentModel) GetContentNameByID(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.ContentName, error) {
	cachedContent := da.GetContentRedis().GetContentCacheById(ctx, cid)
	if cachedContent != nil {
		return &entity.ContentName{
			ID:   cid,
			Name: cachedContent.Name,
		}, nil
	}
	obj, err := da.GetContentDA().GetContentById(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't read content", log.Err(err), log.String("cid", cid))
		return nil, err
	}
	return &entity.ContentName{
		ID:   cid,
		Name: obj.Name,
	}, nil
}

func (cm *ContentModel) GetContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error) {
	cachedContent := da.GetContentRedis().GetContentCacheById(ctx, cid)
	if cachedContent != nil {
		return cachedContent, nil
	}

	obj, err := da.GetContentDA().GetContentById(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return nil, ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Err(err))
		return nil, err
	}
	content, err := contentdata.ConvertContentObj(ctx, obj)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Err(err))
		return nil, ErrParseContentDataFailed
	}

	//补全相关内容
	contentData, err := contentdata.CreateContentData(ctx, obj.ContentType, obj.Data)
	if err != nil {
		return nil, err
	}
	err = contentData.PrepareResult(ctx)
	if err != nil {
		log.Error(ctx, "can't get contentdata for details", log.Err(err))
		return nil, ErrParseContentDataDetailsFailed
	}
	filledContentData, err := contentData.Marshal(ctx)
	if err != nil {
		log.Error(ctx, "can't marshal contentdata for details", log.Err(err))
		return nil, ErrParseContentDataDetailsFailed
	}
	content.Data = filledContentData

	contentWithDetails, err := cm.buildContentWithDetails(ctx, []*entity.ContentInfo{content}, user)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Err(err))
		return nil, ErrReadContentFailed
	}
	if len(contentWithDetails) < 1 {
		return &entity.ContentInfoWithDetails{
			ContentInfo: *content,
		}, nil
	}

	da.GetContentRedis().SaveContentCacheList(ctx, contentWithDetails)
	return contentWithDetails[0], nil
}

func (cm *ContentModel) GetContentNameByIdList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.ContentName, error) {
	if len(cids) < 1 {
		return nil, nil
	}
	resp := make([]*entity.ContentName, 0)

	nid, cachedContent := da.GetContentRedis().GetContentCacheByIdList(ctx, cids)
	for i := range cachedContent {
		resp = append(resp, &entity.ContentName{
			ID:   cachedContent[i].ID,
			Name: cachedContent[i].Name,
		})
	}
	if len(nid) < 1 {
		return resp, nil
	}

	_, data, err := da.GetContentDA().SearchContent(ctx, tx, da.ContentCondition{
		IDS: cids,
	})
	if err != nil {
		log.Error(ctx, "can't search content", log.Err(err), log.Strings("cids", cids))
		return nil, ErrReadContentFailed
	}
	for i := range data {
		resp = append(resp, &entity.ContentName{
			ID:   data[i].ID,
			Name: data[i].Name,
		})
	}
	return resp, nil
}

func (cm *ContentModel) GetContentByIdList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) ([]*entity.ContentInfoWithDetails, error) {
	if len(cids) < 1 {
		return nil, nil
	}

	nid, cachedContent := da.GetContentRedis().GetContentCacheByIdList(ctx, cids)
	if len(nid) < 1 {
		return cachedContent, nil
	}

	_, data, err := da.GetContentDA().SearchContent(ctx, tx, da.ContentCondition{
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
	contentWithDetails, err := cm.buildContentWithDetails(ctx, res, user)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Err(err))
		return nil, ErrReadContentFailed
	}

	da.GetContentRedis().SaveContentCacheList(ctx, contentWithDetails)

	return contentWithDetails, nil
}

func (cm *ContentModel) SearchUserContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	condition1 := condition
	condition2 := condition

	//condition1 private
	condition1.Author = user.UserID
	condition1.PublishStatus = cm.filterInvisiblePublishStatus(ctx, condition1.PublishStatus)

	//condition2 others
	//
	condition2.PublishStatus = []string{entity.ContentStatusPublished}

	//filter visible
	scopes := cm.listVisibleScopes(ctx, user)

	condition2.Scope = scopes

	combineCondition := &da.CombineConditions{
		SourceCondition: &condition1,
		TargetCondition: &condition2,
	}
	//where, params := combineCondition.GetConditions()
	//logger.WithContext(ctx).WithField("subject", "content").Infof("Combine condition: %#v, params: %#v", where, params)
	return cm.searchContentUnsafe(ctx, tx, combineCondition, user)
}

func (cm *ContentModel) SearchUserPrivateContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	condition.Author = user.UserID
	condition.PublishStatus = cm.filterInvisiblePublishStatus(ctx, condition.PublishStatus)

	return cm.searchContent(ctx, tx, &condition, user)
}

func (cm *ContentModel) ListPendingContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	condition.PublishStatus = []string{entity.ContentStatusPending}
	return cm.searchContent(ctx, tx, &condition, user)
}

func (cm *ContentModel) SearchContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	return cm.searchContent(ctx, tx, &condition, user)
}

func (cm *ContentModel) GetVisibleContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error) {
	var err error
	var contentData *entity.Content

	cachedContent := da.GetContentRedis().GetContentCacheById(ctx, cid)
	if cachedContent != nil {
		if cachedContent.LatestID == "" {
			return cachedContent, nil
		} else {
			latestCachedContent := da.GetContentRedis().GetContentCacheById(ctx, cachedContent.LatestID)
			if latestCachedContent != nil {
				return latestCachedContent, nil
			} else {
				contentData = &entity.Content{LatestID: cachedContent.LatestID}
			}
		}
	}

	if contentData == nil {
		contentData, err = da.GetContentDA().GetContentById(ctx, tx, cid)
		if err != nil {
			return nil, err
		}
	}

	if contentData.LatestID != "" {
		newContentData, err := da.GetContentDA().GetContentById(ctx, tx, contentData.LatestID)
		if err != nil {
			return nil, err
		}
		contentData = newContentData
	}

	content, err := contentdata.ConvertContentObj(ctx, contentData)
	if err != nil {
		return nil, err
	}
	contentWithDetails, err := cm.buildContentWithDetails(ctx, []*entity.ContentInfo{content}, user)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Err(err))
		return nil, ErrReadContentFailed
	}
	if len(contentWithDetails) < 1 {
		return &entity.ContentInfoWithDetails{
			ContentInfo: *content,
		}, nil
	}
	da.GetContentRedis().SaveContentCacheList(ctx, contentWithDetails)
	return contentWithDetails[0], nil
}

func (cm *ContentModel) filterInvisiblePublishStatus(ctx context.Context, status []string) []string {
	newStatus := make([]string, 0)
	for i := range status {
		if status[i] != entity.ContentStatusAttachment &&
			status[i] != entity.ContentStatusArchive &&
			status[i] != entity.ContentStatusHidden {
			newStatus = append(newStatus, status[i])
		}
	}
	return newStatus
}

func (cm *ContentModel) checkPublishContentChildren(ctx context.Context, c *entity.Content, children []*entity.Content) error {
	//TODO: To implement
	return nil
}

func (cm *ContentModel) buildContentWithDetails(ctx context.Context, contentList []*entity.ContentInfo, user *entity.Operator) ([]*entity.ContentInfoWithDetails, error) {
	orgName := ""
	orgProvider, err := external.GetOrganizationServiceProvider()
	if err != nil {
		log.Error(ctx, "can't get org provider", log.Err(err))
	} else {
		orgs, err := orgProvider.BatchGet(ctx, []string{user.OrgID})
		if err != nil || len(orgs) < 1 {
			log.Error(ctx, "can't get org info", log.Err(err))
		} else {
			orgName = orgs[0].Name
		}
	}

	programNameMap := make(map[string]string)
	subjectNameMap := make(map[string]string)
	developmentalNameMap := make(map[string]string)
	skillsNameMap := make(map[string]string)
	ageNameMap := make(map[string]string)
	gradeNameMap := make(map[string]string)

	programIds := make([]string, 0)
	subjectIds := make([]string, 0)
	developmentalIds := make([]string, 0)
	skillsIds := make([]string, 0)
	ageIds := make([]string, 0)
	gradeIds := make([]string, 0)

	for i := range contentList {
		programIds = append(programIds, contentList[i].Program...)
		subjectIds = append(subjectIds, contentList[i].Subject...)
		developmentalIds = append(developmentalIds, contentList[i].Developmental...)
		skillsIds = append(skillsIds, contentList[i].Skills...)
		ageIds = append(ageIds, contentList[i].Age...)
		gradeIds = append(gradeIds, contentList[i].Grade...)
	}

	//Program
	programProvider, err := external.GetProgramServiceProvider()
	if err != nil {
		log.Error(ctx, "can't get programProvider", log.Err(err))
	} else {
		programs, err := programProvider.BatchGet(ctx, programIds)
		if err != nil {
			log.Error(ctx, "can't get org info", log.Err(err))
		} else {
			for i := range programs {
				programNameMap[programs[i].ID] = programs[i].Name
			}
		}
	}

	//Subjects
	subjectsProvider, err := external.GetSubjectServiceProvider()
	if err != nil {
		log.Error(ctx, "can't get subjectsProvider", log.Err(err))
	} else {
		subjects, err := subjectsProvider.BatchGet(ctx, subjectIds)
		if err != nil {
			log.Error(ctx, "can't get subjects info", log.Err(err))
		} else {
			for i := range subjects {
				subjectNameMap[subjects[i].ID] = subjects[i].Name
			}
		}
	}

	//developmental
	developmentalsProvider, err := external.GetDevelopmentalServiceProvider()
	if err != nil {
		log.Error(ctx, "can't get developmentalsProvider", log.Err(err))
	} else {
		developmentals, err := developmentalsProvider.BatchGet(ctx, developmentalIds)
		if err != nil {
			log.Error(ctx, "can't get developmentals info", log.Err(err))
		} else {
			for i := range developmentals {
				developmentalNameMap[developmentals[i].ID] = developmentals[i].Name
			}
		}
	}

	//skill
	skillProvider, err := external.GetSkillServiceProvider()
	if err != nil {
		log.Error(ctx, "can't get skillProvider", log.Err(err))
	} else {
		skills, err := skillProvider.BatchGet(ctx, skillsIds)
		if err != nil {
			log.Error(ctx, "can't get skills info", log.Err(err))
		} else {
			for i := range skills {
				skillsNameMap[skills[i].ID] = skills[i].Name
			}
		}
	}

	//age
	ageProvider, err := external.GetAgeServiceProvider()
	if err != nil {
		log.Error(ctx, "can't get ageProvider", log.Err(err))
	} else {
		ages, err := ageProvider.BatchGet(ctx, ageIds)
		if err != nil {
			log.Error(ctx, "can't get age info", log.Err(err))
		} else {
			for i := range ages {
				ageNameMap[ages[i].ID] = ages[i].Name
			}
		}
	}

	//grade
	gradeProvider, err := external.GetGradeServiceProvider()
	if err != nil {
		log.Error(ctx, "can't get gradeProvider", log.Err(err))
	} else {
		grades, err := gradeProvider.BatchGet(ctx, gradeIds)
		if err != nil {
			log.Error(ctx, "can't get grade info", log.Err(err))
		} else {
			for i := range grades {
				gradeNameMap[grades[i].ID] = grades[i].Name
			}
		}
	}

	contentDetailsList := make([]*entity.ContentInfoWithDetails, len(contentList))
	for i := range contentList {
		programNames := make([]string, len(contentList[i].Program))
		subjectNames := make([]string, len(contentList[i].Subject))
		developmentalNames := make([]string, len(contentList[i].Developmental))
		skillsNames := make([]string, len(contentList[i].Skills))
		ageNames := make([]string, len(contentList[i].Age))
		gradeNames := make([]string, len(contentList[i].Grade))

		for j := range contentList[i].Program {
			programNames[j] = programNameMap[contentList[i].Program[j]]
		}
		for j := range contentList[i].Subject {
			subjectNames[j] = subjectNameMap[contentList[i].Subject[j]]
		}
		for j := range contentList[i].Developmental {
			developmentalNames[j] = developmentalNameMap[contentList[i].Developmental[j]]
		}
		for j := range contentList[i].Skills {
			skillsNames[j] = skillsNameMap[contentList[i].Skills[j]]
		}
		for j := range contentList[i].Age {
			ageNames[j] = ageNameMap[contentList[i].Age[j]]
		}
		for j := range contentList[i].Grade {
			gradeNames[j] = gradeNameMap[contentList[i].Grade[j]]
		}

		contentDetailsList[i] = &entity.ContentInfoWithDetails{
			ContentInfo:       *contentList[i],
			ContentTypeName:   contentList[i].ContentType.Name(),
			ProgramName:       programNames,
			SubjectName:       subjectNames,
			DevelopmentalName: developmentalNames,
			SkillsName:        skillsNames,
			AgeName:           ageNames,
			GradeName:         gradeNames,
			OrgName:           orgName,
		}
	}

	return contentDetailsList, nil
}

func (cm *ContentModel) listVisibleScopes(ctx context.Context, operator *entity.Operator) []string {
	return []string{operator.OrgID}
}

var (
	_contentModel     IContentModel
	_contentModelOnce sync.Once
)

func GetContentModel() IContentModel {
	_contentModelOnce.Do(func() {
		_contentModel = new(ContentModel)
	})
	return _contentModel
}
