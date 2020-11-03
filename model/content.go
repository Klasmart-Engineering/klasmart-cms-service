package model

import (
	"context"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
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

var (
	ErrNoSuchURL           = errors.New("no such url")
	ErrRequestItemIsNil    = errors.New("request item is nil")
	ErrNoAuth              = errors.New("no auth to operate")
	ErrCreateContentFailed = errors.New("create contentdata into data access failed")

	ErrNoContentData                     = errors.New("no content data")
	ErrInvalidResourceId                 = errors.New("invalid resource id")
	ErrInvalidContentData                = errors.New("invalid content data")
	ErrMarshalContentDataFailed          = errors.New("marshal content data failed")
	ErrInvalidPublishStatus              = errors.New("invalid publish status")
	ErrInvalidLockedContentPublishStatus = errors.New("invalid locked content publish status")
	ErrInvalidContentStatusToPublish     = errors.New("content status is invalid to publish")
	ErrNoContent                         = errors.New("no content")
	ErrContentAlreadyLocked              = errors.New("content is already locked")
	ErrDeleteLessonInSchedule            = errors.New("can't delete lesson in schedule")
	ErrGetUnpublishedContent             = errors.New("unpublished content")
	ErrGetUnauthorizedContent            = errors.New("unauthorized content")
	ErrCloneContentFailed                = errors.New("clone content failed")
	ErrParseContentDataFailed            = errors.New("parse content data failed")
	ErrParseContentDataDetailsFailed     = errors.New("parse content data details failed")
	ErrUpdateContentFailed               = errors.New("update contentdata into data access failed")
	ErrReadContentFailed                 = errors.New("read content failed")
	ErrDeleteContentFailed               = errors.New("delete contentdata into data access failed")

	ErrInvalidMaterialType = errors.New("invalid material type")

	ErrBadRequest         = errors.New("bad request")
	ErrResourceNotFound   = errors.New("resource not found")
	ErrInvalidContentType = errors.New("invalid content type")
	ErrNoRejectReason     = errors.New("no reject reason")

	ErrInvalidSelectForm = errors.New("invalid select form")
)

type IContentModel interface {
	CreateContent(ctx context.Context, tx *dbo.DBContext, c entity.CreateContentRequest, operator *entity.Operator) (string, error)
	UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, data entity.CreateContentRequest, user *entity.Operator) error
	PublishContent(ctx context.Context, tx *dbo.DBContext, cid, scope string, user *entity.Operator) error
	PublishContentWithAssets(ctx context.Context, tx *dbo.DBContext, cid, scope string, user *entity.Operator) error
	LockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error)
	DeleteContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error
	CloneContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error)

	PublishContentBulk(ctx context.Context, tx *dbo.DBContext, ids []string, user *entity.Operator) error
	DeleteContentBulk(ctx context.Context, tx *dbo.DBContext, ids []string, user *entity.Operator) error

	GetContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error)
	GetContentByIdList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) ([]*entity.ContentInfoWithDetails, error)
	GetContentNameByID(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.ContentName, error)
	GetContentNameByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.ContentName, error)
	GetContentSubContentsByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]*entity.SubContentsWithName, error)

	UpdateContentPublishStatus(ctx context.Context, tx *dbo.DBContext, cid string, reason []string, remark, status string) error
	CheckContentAuthorization(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error

	SearchUserContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)
	SearchUserPrivateContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)
	ListPendingContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)
	SearchContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)
	GetContentOutcomeByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error)
	ContentDataCount(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.ContentStatisticsInfo, error)
	GetVisibleContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error)
}

type ContentModel struct {
}

func (cm *ContentModel) handleSourceContent(ctx context.Context, tx *dbo.DBContext, contentId, sourceId string) error {
	sourceContent, err := da.GetContentDA().GetContentByID(ctx, tx, sourceId)
	if err != nil {
		return err
	}
	sourceContent.PublishStatus = entity.ContentStatusHidden
	sourceContent.LatestID = contentId
	//解锁source content
	sourceContent.LockedBy = constant.LockedByNoBody
	err = da.GetContentDA().UpdateContent(ctx, tx, sourceId, *sourceContent)
	if err != nil {
		log.Error(ctx, "update source content failed", log.Err(err))
		return ErrUpdateContentFailed
	}

	//更新所有latestID为sourceContent的Content
	_, oldContents, err := da.GetContentDA().SearchContent(ctx, tx, da.ContentCondition{
		LatestID: sourceContent.ID,
	})
	if err != nil {
		log.Error(ctx, "update old content failed", log.Err(err), log.String("SourceID", sourceContent.ID))
		return ErrUpdateContentFailed
	}
	for i := range oldContents {
		oldContents[i].LockedBy = constant.LockedByNoBody
		oldContents[i].PublishStatus = entity.ContentStatusHidden
		oldContents[i].LatestID = contentId
		err = da.GetContentDA().UpdateContent(ctx, tx, oldContents[i].ID, *oldContents[i])
		if err != nil {
			log.Error(ctx, "update old content failed", log.Err(err), log.String("OldID", oldContents[i].ID))
			return ErrUpdateContentFailed
		}
	}

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
	if c.LessonType != "" {
		_, err := GetLessonTypeModel().GetByID(ctx, c.LessonType)
		if err != nil {
			log.Error(ctx, "lesson type invalid", log.Any("data", c), log.Err(err))
			return err
		}
	}
	err := c.Validate()
	if err != nil {
		log.Error(ctx, "asset no need to check", log.Any("data", c), log.Bool("created", created), log.Err(err))
		return err
	}
	err = c.ContentType.Validate()
	if err != nil {
		log.Error(ctx, "content type invalid", log.Any("data", c), log.Bool("created", created), log.Err(err))
		return err
	}
	if created {
		if len(c.Developmental) == 0 ||
			len(c.Program) == 0 {
			log.Error(ctx, "select form invalid", log.Any("data", c), log.Bool("created", created), log.Err(err))
			return ErrInvalidSelectForm
		}
	}
	return nil
}

func (cm ContentModel) checkUpdateContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) (*entity.Content, error) {

	//若为asset，直接发布
	if content.ContentType.IsAsset() {
		log.Info(ctx, "asset no need to check", log.String("cid", content.ID))
		return content, nil
	}

	//TODO:maybe wrong
	if content.Author != user.UserID {
		return nil, ErrNoAuth
	}
	if content.PublishStatus == entity.ContentStatusPending ||
		content.PublishStatus == entity.ContentStatusArchive ||
		content.PublishStatus == entity.ContentStatusHidden ||
		content.PublishStatus == entity.ContentStatusAttachment ||
		content.PublishStatus == entity.ContentStatusPublished {
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
	subContentIds := contentData.SubContentIds(ctx)
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

func (cm *ContentModel) searchContent(ctx context.Context, tx *dbo.DBContext, condition *da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	log.Debug(ctx, "search content ", log.Any("condition", condition), log.String("uid", user.UserID))

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

	return count, contentWithDetails, nil
}

func (cm *ContentModel) searchContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition *da.CombineConditions, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
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

	return count, contentWithDetails, nil
}

func (cm *ContentModel) CreateContent(ctx context.Context, tx *dbo.DBContext, c entity.CreateContentRequest, operator *entity.Operator) (string, error) {
	//检查数据信息是否正确
	log.Info(ctx, "create content")
	if c.ContentType.IsAsset() {
		provider := external.GetPublishScopeProvider()
		c.PublishScope = provider.DefaultPublishScope(ctx)
	}

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

	//添加内容
	pid, err := da.GetContentDA().CreateContent(ctx, tx, *obj)
	if err != nil {
		log.Error(ctx, "can't create contentdata", log.Err(err), log.String("uid", operator.UserID), log.Any("data", c))
		return "", err
	}

	return pid, nil
}

func (cm *ContentModel) UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, data entity.CreateContentRequest, user *entity.Operator) error {
	if data.ContentType.IsAsset() {
		//Assets can't be updated
		return ErrInvalidContentType
	}

	err := cm.checkContentInfo(ctx, data, false)
	if err != nil {
		return err
	}
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
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
		log.Error(ctx, "check update content failed", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID), log.Any("data", data))
		return err
	}
	//Check status
	//Check Authorization
	obj, err := cm.prepareUpdateContentParams(ctx, content, &data)
	if err != nil {
		log.Error(ctx, "can't prepare params contentdata on update contentdata", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID), log.Any("data", data))
		return err
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

func (cm *ContentModel) UpdateContentPublishStatus(ctx context.Context, tx *dbo.DBContext, cid string, reason []string, remark, status string) error {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't read contentdata on update contentdata", log.Err(err))
		return err
	}
	if content.ContentType.IsAsset() {
		return ErrInvalidContentType
	}

	content.PublishStatus = entity.NewContentPublishStatus(status)
	if status == entity.ContentStatusRejected && len(reason) < 1 && remark == "" {
		return ErrNoRejectReason
	}

	rejectReason := strings.Join(reason, ",")
	content.RejectReason = rejectReason
	content.Remark = remark
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
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
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
	content.LockedBy = constant.LockedByNoBody
	return da.GetContentDA().UpdateContent(ctx, tx, cid, *content)
}
func (cm *ContentModel) LockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error) {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixContentLock, cid)
	if err != nil {
		return "", err
	}
	locker.Lock()
	defer locker.Unlock()

	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return "", ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata for publishing", log.Err(err))
		return "", err
	}
	if content.ContentType.IsAsset() {
		log.Info(ctx, "asset handle", log.String("cid", cid))
		return "", ErrInvalidContentType
	}

	if content.PublishStatus != entity.ContentStatusPublished {
		log.Info(ctx, "invalid publish status", log.String("cid", cid))
		return "", ErrInvalidPublishStatus
	}

	//被自己锁住，则返回锁定id
	if content.LockedBy == user.UserID {
		_, data, err := da.GetContentDA().SearchContent(ctx, tx, da.ContentCondition{
			SourceID: cid,
		})
		if err != nil {
			log.Info(ctx, "search source content failed", log.String("cid", cid))
			return "", err
		}
		if len(data) < 1 {
			//被自己锁定且找不到content
			log.Info(ctx, "no content in source content", log.String("cid", cid))
			return "", ErrNoContent
		}
		if data[0].PublishStatus != entity.ContentStatusRejected && data[0].PublishStatus != entity.ContentStatusDraft {
			log.Info(ctx, "invalid locked content status", log.String("lock cid", data[0].ID), log.String("status", string(data[0].PublishStatus)), log.String("cid", cid))
			return "", ErrInvalidLockedContentPublishStatus
		}
		//找到data
		return data[0].ID, nil
	}

	//更新锁定状态
	if content.LockedBy != "" && content.LockedBy != constant.LockedByNoBody {
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
	da.GetContentRedis().CleanContentCache(ctx, []string{cid, content.ID})

	//Update with new content
	return ccid, nil

}
func (cm *ContentModel) PublishContentBulk(ctx context.Context, tx *dbo.DBContext, ids []string, user *entity.Operator) error {
	updateIds := make([]string, 0)
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
	da.GetContentRedis().CleanContentCache(ctx, updateIds)
	return err
}

func (cm *ContentModel) PublishContent(ctx context.Context, tx *dbo.DBContext, cid, scope string, user *entity.Operator) error {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
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

func (cm *ContentModel) validatePublishContentWithAssets(ctx context.Context, content *entity.Content, user *entity.Operator) error {
	if content.ContentType != entity.ContentTypeMaterial {
		return ErrInvalidContentType
	}
	//查看data
	cd, err := contentdata.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		log.Warn(ctx, "create content data failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}
	materialData, ok := cd.(*contentdata.MaterialData)
	if !ok {
		log.Warn(ctx, "validate content data with content type 2 failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}
	if materialData.InputSource != entity.MaterialInputSourceDisk {
		log.Warn(ctx, "invalid material type", log.Err(err), log.String("uid", user.UserID), log.Any("materialData", materialData))
		return ErrInvalidMaterialType
	}
	return nil
}

func (cm *ContentModel) prepareForPublishAssets(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	//创建data对象
	cd, err := contentdata.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		log.Warn(ctx, "create content data failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}
	//解析data的fileType
	cd.PrepareSave(ctx)
	materialData, ok := cd.(*contentdata.MaterialData)
	if !ok {
		log.Warn(ctx, "asset content data type failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}

	//创建assets data对象，并解析
	assetsData := new(contentdata.AssetsData)
	assetsData.Source = materialData.Source
	assetsDataJSON, err := assetsData.Marshal(ctx)
	if !ok {
		log.Warn(ctx, "marshal assets data failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrMarshalContentDataFailed
	}
	//创建assets
	req := entity.CreateContentRequest{
		ContentType:   entity.ContentTypeAssets,
		Name:          content.Name,
		Program:       content.Program,
		Subject:       stringToStringArray(ctx, content.Subject),
		Developmental: stringToStringArray(ctx, content.Developmental),
		Skills:        stringToStringArray(ctx, content.Skills),
		Age:           stringToStringArray(ctx, content.Age),
		Grade:         stringToStringArray(ctx, content.Grade),
		Keywords:      stringToStringArray(ctx, content.Keywords),
		Description:   content.Description,
		Thumbnail:     content.Thumbnail,
		SuggestTime:   content.SuggestTime,
		Data:          assetsDataJSON,
	}
	_, err = cm.CreateContent(ctx, tx, req, user)
	if err != nil {
		log.Warn(ctx, "create assets failed", log.Err(err), log.String("uid", user.UserID), log.Any("req", req))
		return err
	}

	//更新content状态
	materialData.InputSource = entity.MaterialInputSourceDisk
	d, err := materialData.Marshal(ctx)
	if err != nil {
		log.Warn(ctx, "marshal content data failed", log.Err(err), log.String("uid", user.UserID), log.Any("content", content))
		return err
	}

	content.Data = d
	return nil
}
func (cm *ContentModel) PublishContentWithAssets(ctx context.Context, tx *dbo.DBContext, cid, scope string, user *entity.Operator) error {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read content data for publishing", log.Err(err), log.String("cid", cid), log.String("scope", scope), log.String("uid", user.UserID))
		return err
	}
	err = cm.validatePublishContentWithAssets(ctx, content, user)
	if err != nil {
		log.Error(ctx, "validate for publishing failed", log.Err(err), log.String("cid", cid), log.String("scope", scope), log.String("uid", user.UserID))
		return err
	}

	//修改发布状态
	if scope != "" {
		content.PublishScope = scope
	}

	//准备发布（1.创建assets，2.修改contentdata）
	err = cm.prepareForPublishAssets(ctx, tx, content, user)
	if err != nil {
		return err
	}

	//发布
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
	if err != nil {
		return err
	}
	da.GetContentRedis().CleanContentCache(ctx, deletedIds)
	return nil
}

func (cm *ContentModel) checkDeleteContent(ctx context.Context, content *entity.Content) error {
	if content.PublishStatus == entity.ContentStatusArchive && content.ContentType == entity.ContentTypeLesson {
		exist, err := GetScheduleModel().ExistScheduleByLessonPlanID(ctx, content.ID)
		if err != nil {
			return err
		}
		if exist {
			return ErrDeleteLessonInSchedule
		}
	}
	return nil
}

func (cm *ContentModel) doDeleteContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	if content.Author != user.UserID {
		return ErrNoAuth
	}
	if content.LockedBy != constant.LockedByNoBody && content.LockedBy != user.UserID {
		return ErrContentAlreadyLocked
	}
	if content.PublishStatus == entity.ContentStatusPublished && content.LockedBy != constant.LockedByNoBody {
		return ErrContentAlreadyLocked
	}

	err := cm.checkDeleteContent(ctx, content)
	if err != nil {
		log.Error(ctx, "check delete content failed", log.Err(err), log.String("cid", content.ID), log.String("uid", user.UserID))
		return err
	}

	obj := cm.prepareDeleteContentParams(ctx, content, content.PublishStatus)

	err = da.GetContentDA().UpdateContent(ctx, tx, content.ID, *obj)
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
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
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
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
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

	da.GetContentRedis().CleanContentCache(ctx, []string{id, obj.ID})
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

func (cm *ContentModel) GetContentSubContentsByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]*entity.SubContentsWithName, error) {
	obj, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't read content", log.Err(err), log.String("cid", cid))
		return nil, err
	}
	cd, err := contentdata.CreateContentData(ctx, obj.ContentType, obj.Data)
	if err != nil {
		log.Error(ctx, "can't unmarshal contentdata", log.Err(err), log.Any("content", obj))
		return nil, err
	}
	ids := cd.SubContentIds(ctx)
	//若不存在子内容，则返回当前内容
	if obj.ContentType == entity.ContentTypeMaterial {
		ret := []*entity.SubContentsWithName{
			{
				ID:   cid,
				Name: obj.Name,
				Data: cd,
			},
		}
		return ret, nil
	}

	//存在子内容，则返回子内容
	subContents, err := da.GetContentDA().GetContentByIDList(ctx, tx, ids)
	if err != nil {
		log.Error(ctx, "can't get sub contents", log.Err(err), log.Strings("ids", ids))
		return nil, err
	}
	subContentMap := make(map[string]*entity.Content, len(subContents))
	for i := range subContents {
		subContentMap[subContents[i].ID] = subContents[i]
	}

	ret := make([]*entity.SubContentsWithName, 0)
	for i := range ids {
		subContent, ok := subContentMap[ids[i]]
		if !ok {
			continue
		}
		cd, err := contentdata.CreateContentData(ctx, subContent.ContentType, subContent.Data)
		if err != nil {
			log.Error(ctx, "can't parse sub content data", log.Err(err), log.Any("subContent", subContentMap[ids[i]]))
			return nil, err
		}
		ret = append(ret, &entity.SubContentsWithName{
			ID:   ids[i],
			Name: subContent.Name,
			Data: cd,
		})
	}

	return ret, nil
}

func (cm *ContentModel) GetContentNameByID(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.ContentName, error) {
	cachedContent := da.GetContentRedis().GetContentCacheByID(ctx, cid)
	if cachedContent != nil {
		return &entity.ContentName{
			ID:          cid,
			Name:        cachedContent.Name,
			ContentType: cachedContent.ContentType,
		}, nil
	}
	obj, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't read content", log.Err(err), log.String("cid", cid))
		return nil, err
	}
	return &entity.ContentName{
		ID:          cid,
		Name:        obj.Name,
		ContentType: obj.ContentType,
	}, nil
}

func (cm *ContentModel) GetContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error) {
	content := da.GetContentRedis().GetContentCacheByID(ctx, cid)

	if content == nil {
		obj, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
		if err == dbo.ErrRecordNotFound {
			log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
			return nil, ErrNoContent
		}
		if err != nil {
			log.Error(ctx, "can't read contentdata", log.Err(err))
			return nil, err
		}
		content, err = contentdata.ConvertContentObj(ctx, obj)
		if err != nil {
			log.Error(ctx, "can't parse contentdata", log.Err(err))
			return nil, ErrParseContentDataFailed
		}
		da.GetContentRedis().SaveContentCache(ctx, content)
	}

	//补全相关内容
	contentData, err := contentdata.CreateContentData(ctx, content.ContentType, content.Data)
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

	return contentWithDetails[0], nil
}

func (cm *ContentModel) GetContentNameByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.ContentName, error) {
	if len(cids) < 1 {
		return nil, nil
	}
	resp := make([]*entity.ContentName, 0)

	nid, cachedContent := da.GetContentRedis().GetContentCacheByIdList(ctx, cids)
	for i := range cachedContent {
		resp = append(resp, &entity.ContentName{
			ID:          cachedContent[i].ID,
			Name:        cachedContent[i].Name,
			ContentType: cachedContent[i].ContentType,
		})
	}
	if len(nid) < 1 {
		return resp, nil
	}

	data, err := da.GetContentDA().GetContentByIDList(ctx, tx, cids)
	if err != nil {
		log.Error(ctx, "can't search content", log.Err(err), log.Strings("cids", cids))
		return nil, ErrReadContentFailed
	}
	for i := range data {
		resp = append(resp, &entity.ContentName{
			ID:          data[i].ID,
			Name:        data[i].Name,
			ContentType: data[i].ContentType,
		})
	}
	return resp, nil
}

func (cm *ContentModel) GetContentByIdList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) ([]*entity.ContentInfoWithDetails, error) {
	if len(cids) < 1 {
		return nil, nil
	}

	nid, cachedContent := da.GetContentRedis().GetContentCacheByIdList(ctx, cids)
	//全在缓存中
	if len(nid) < 1 {
		contentWithDetails, err := cm.buildContentWithDetails(ctx, cachedContent, user)
		if err != nil {
			log.Error(ctx, "can't parse contentdata", log.Err(err))
			return nil, ErrReadContentFailed
		}
		return contentWithDetails, nil
	}

	data, err := da.GetContentDA().GetContentByIDList(ctx, tx, nid)
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
	res = append(res, cachedContent...)
	contentWithDetails, err := cm.buildContentWithDetails(ctx, res, user)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Err(err))
		return nil, ErrReadContentFailed
	}

	da.GetContentRedis().SaveContentCacheList(ctx, res)

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
	condition.PublishStatus = cm.filterInvisiblePublishStatus(ctx, condition.PublishStatus)
	return cm.searchContent(ctx, tx, &condition, user)
}

func (cm *ContentModel) GetContentOutcomeByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error) {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't get content", log.Err(err), log.String("cid", cid))
		return nil, err
	}
	if content.Outcomes == "" {
		return nil, nil
	}
	outcomes := strings.Split(content.Outcomes, ",")
	ret := make([]string, 0)
	for i := range outcomes {
		if outcomes[i] != "" {
			ret = append(ret, outcomes[i])
		}
	}

	return ret, nil
}

func (cm *ContentModel) ContentDataCount(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.ContentStatisticsInfo, error) {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't get content", log.Err(err), log.String("cid", cid))
		return nil, err
	}
	if content.ContentType != entity.ContentTypeLesson {
		log.Error(ctx, "invalid content type", log.Err(err), log.String("cid", cid), log.Int("contentType", int(content.ContentType)))
		return nil, ErrInvalidContentType
	}

	cd, err := contentdata.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		log.Error(ctx, "can't parse content data", log.Err(err), log.String("cid", cid), log.Int("contentType", int(content.ContentType)), log.String("data", content.Data))
		return nil, err
	}
	subContentIds := cd.SubContentIds(ctx)
	_, subContents, err := da.GetContentDA().SearchContent(ctx, tx, da.ContentCondition{
		IDS: subContentIds,
	})

	identityOutComes := make(map[string]bool)
	outcomesCount := 0
	for i := range subContents {
		if subContents[i].Outcomes == "" {
			continue
		}
		subOutcomes := strings.Split(subContents[i].Outcomes, ",")
		for j := range subOutcomes {
			_, ok := identityOutComes[subOutcomes[j]]
			if !ok {
				outcomesCount++
			}
			identityOutComes[subOutcomes[j]] = true
		}
	}
	return &entity.ContentStatisticsInfo{
		SubContentCount: len(subContentIds),
		OutcomesCount:   outcomesCount,
	}, nil
}

func (cm *ContentModel) GetVisibleContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error) {
	var err error
	cachedContent := da.GetContentRedis().GetContentCacheByID(ctx, cid)
	if cachedContent != nil {
		if cachedContent.LatestID == "" {
			return cm.GetContentByID(ctx, tx, cachedContent.ID, user)
		} else {
			return cm.GetContentByID(ctx, tx, cachedContent.LatestID, user)
		}
	}

	contentObj, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return nil, ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Err(err))
		return nil, err
	}
	if contentObj.LatestID == "" {
		return cm.GetContentByID(ctx, tx, cid, user)
	}
	return cm.GetContentByID(ctx, tx, contentObj.LatestID, user)
}

func (cm *ContentModel) filterInvisiblePublishStatus(ctx context.Context, status []string) []string {
	newStatus := make([]string, 0)
	if len(status) < 1 {
		return []string{
			entity.ContentStatusPublished,
			entity.ContentStatusPending,
			entity.ContentStatusDraft,
			entity.ContentStatusRejected,
		}
	}

	for i := range status {
		if status[i] != entity.ContentStatusAttachment &&
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
	orgProvider := external.GetOrganizationServiceProvider()
	orgs, err := orgProvider.BatchGet(ctx, []string{user.OrgID})
	if err != nil || len(orgs) < 1 {
		log.Error(ctx, "can't get org info", log.Err(err))
	} else {
		orgName = orgs[0].Name
	}

	programNameMap := make(map[string]string)
	subjectNameMap := make(map[string]string)
	developmentalNameMap := make(map[string]string)
	skillsNameMap := make(map[string]string)
	ageNameMap := make(map[string]string)
	gradeNameMap := make(map[string]string)
	publishScopeNameMap := make(map[string]string)
	lessonTypeNameMap := make(map[string]string)

	programIds := make([]string, 0)
	subjectIds := make([]string, 0)
	developmentalIds := make([]string, 0)
	skillsIds := make([]string, 0)
	ageIds := make([]string, 0)
	gradeIds := make([]string, 0)
	scopeIds := make([]string, 0)
	lessonTypeIds := make([]string, 0)

	for i := range contentList {
		programIds = append(programIds, contentList[i].Program)
		subjectIds = append(subjectIds, contentList[i].Subject...)
		developmentalIds = append(developmentalIds, contentList[i].Developmental...)
		skillsIds = append(skillsIds, contentList[i].Skills...)
		ageIds = append(ageIds, contentList[i].Age...)
		gradeIds = append(gradeIds, contentList[i].Grade...)

		scopeIds = append(scopeIds, contentList[i].PublishScope)
		lessonTypeIds = append(lessonTypeIds, contentList[i].LessonType)
	}

	//LessonType
	lessonTypes, err := GetLessonTypeModel().Query(ctx, &da.LessonTypeCondition{
		IDs: entity.NullStrings{
			Strings: lessonTypeIds,
			Valid:   len(lessonTypeIds) != 0,
		},
	})
	if err != nil {
		log.Error(ctx, "can't get lesson type info", log.Err(err))
	} else {
		for i := range lessonTypes {
			lessonTypeNameMap[lessonTypes[i].ID] = lessonTypes[i].Name
		}
	}

	//Program
	programs, err := GetProgramModel().Query(ctx, &da.ProgramCondition{
		IDs: entity.NullStrings{
			Strings: programIds,
			Valid:   len(programIds) != 0,
		},
	})
	if err != nil {
		log.Error(ctx, "can't get org info", log.Err(err))
	} else {
		for i := range programs {
			programNameMap[programs[i].ID] = programs[i].Name
		}
	}

	//Subjects
	subjects, err := GetSubjectModel().Query(ctx, &da.SubjectCondition{
		IDs: entity.NullStrings{
			Strings: subjectIds,
			Valid:   len(subjectIds) != 0,
		},
	})
	if err != nil {
		log.Error(ctx, "can't get subjects info", log.Err(err))
	} else {
		for i := range subjects {
			subjectNameMap[subjects[i].ID] = subjects[i].Name
		}
	}

	//developmental
	developmentals, err := GetDevelopmentalModel().Query(ctx, &da.DevelopmentalCondition{
		IDs: entity.NullStrings{
			Strings: developmentalIds,
			Valid:   len(developmentalIds) != 0,
		},
	})

	if err != nil {
		log.Error(ctx, "can't get developmentals info", log.Err(err))
	} else {
		for i := range developmentals {
			developmentalNameMap[developmentals[i].ID] = developmentals[i].Name
		}
	}

	//scope
	publishScopeProvider := external.GetPublishScopeProvider()
	publishScopes, err := publishScopeProvider.BatchGet(ctx, scopeIds)
	if err != nil {
		log.Error(ctx, "can't get publish scope info", log.Err(err))
	} else {
		for i := range publishScopes {
			publishScopeNameMap[publishScopes[i].ID] = publishScopes[i].Name
		}
	}

	//skill
	skills, err := GetSkillModel().Query(ctx, &da.SkillCondition{
		IDs: entity.NullStrings{
			Strings: skillsIds,
			Valid:   len(skillsIds) != 0,
		},
	})
	if err != nil {
		log.Error(ctx, "can't get skills info", log.Err(err))
	} else {
		for i := range skills {
			skillsNameMap[skills[i].ID] = skills[i].Name
		}
	}

	//age
	ages, err := GetAgeModel().Query(ctx, &da.AgeCondition{
		IDs: entity.NullStrings{
			Strings: ageIds,
			Valid:   len(ageIds) != 0,
		},
	})
	if err != nil {
		log.Error(ctx, "can't get age info", log.Err(err))
	} else {
		for i := range ages {
			ageNameMap[ages[i].ID] = ages[i].Name
		}
	}

	//grade
	grades, err := GetGradeModel().Query(ctx, &da.GradeCondition{
		IDs: entity.NullStrings{
			Strings: gradeIds,
			Valid:   len(gradeIds) != 0,
		},
	})
	if err != nil {
		log.Error(ctx, "can't get grade info", log.Err(err))
	} else {
		for i := range grades {
			gradeNameMap[grades[i].ID] = grades[i].Name
		}
	}

	//Outcomes
	outcomeIds := make([]string, 0)
	for i := range contentList {
		outcomeIds = append(outcomeIds, contentList[i].Outcomes...)
	}
	outcomeEntities, err := GetOutcomeModel().GetLatestOutcomesByIDs(ctx, dbo.MustGetDB(ctx), outcomeIds, user)
	if err != nil {
		log.Error(ctx, "get latest outcomes entity failed", log.Err(err), log.Strings("outcome list", outcomeIds), log.String("uid", user.UserID))
	}
	outcomeMaps := make(map[string]*entity.Outcome, len(outcomeEntities))
	for i := range outcomeEntities {
		outcomeMaps[outcomeEntities[i].ID] = outcomeEntities[i]
	}

	contentDetailsList := make([]*entity.ContentInfoWithDetails, len(contentList))
	for i := range contentList {
		subjectNames := make([]string, len(contentList[i].Subject))
		developmentalNames := make([]string, len(contentList[i].Developmental))
		skillsNames := make([]string, len(contentList[i].Skills))
		ageNames := make([]string, len(contentList[i].Age))
		gradeNames := make([]string, len(contentList[i].Grade))

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
			ProgramName:       programNameMap[contentList[i].Program],
			SubjectName:       subjectNames,
			DevelopmentalName: developmentalNames,
			SkillsName:        skillsNames,
			AgeName:           ageNames,
			GradeName:         gradeNames,
			LessonTypeName:    lessonTypeNameMap[contentList[i].LessonType],
			PublishScopeName:  publishScopeNameMap[contentList[i].PublishScope],
			OrgName:           orgName,
			OutcomeEntities:   cm.pickOutcomes(ctx, contentList[i].Outcomes, outcomeMaps),
		}
	}

	return contentDetailsList, nil
}

func (cm *ContentModel) pickOutcomes(ctx context.Context, pickIds []string, outcomeMaps map[string]*entity.Outcome) []*entity.Outcome {
	ret := make([]*entity.Outcome, 0)
	for i := range pickIds {
		outcome, ok := outcomeMaps[pickIds[i]]
		if ok {
			ret = append(ret, outcome)
		}
	}
	return ret
}

func (cm *ContentModel) listVisibleScopes(ctx context.Context, operator *entity.Operator) []string {
	return []string{operator.OrgID}
}

func stringToStringArray(ctx context.Context, str string) []string {
	if str == "" {
		return nil
	}
	return strings.Split(str, ",")
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
