package model

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	mutex "gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
)

var (
	ErrNoAuth = errors.New("no auth to operate")

	ErrNoContentData                     = errors.New("no content data")
	ErrInvalidResourceID                 = errors.New("invalid resource id")
	ErrInvalidContentData                = errors.New("invalid content data")
	ErrMarshalContentDataFailed          = errors.New("marshal content data failed")
	ErrInvalidPublishStatus              = errors.New("invalid publish status")
	ErrInvalidLockedContentPublishStatus = errors.New("invalid locked content publish status")
	ErrInvalidContentStatusToPublish     = errors.New("content status is invalid to publish")
	ErrNoContent                         = errors.New("no content")
	//ErrContentAlreadyLocked              = errors.New("content is already locked")
	ErrDeleteLessonInSchedule        = errors.New("can't delete lesson in schedule")
	ErrGetUnpublishedContent         = errors.New("unpublished content")
	ErrGetUnauthorizedContent        = errors.New("unauthorized content")
	ErrCloneContentFailed            = errors.New("clone content failed")
	ErrParseContentDataFailed        = errors.New("parse content data failed")
	ErrParseContentDataDetailsFailed = errors.New("parse content data details failed")
	ErrUpdateContentFailed           = errors.New("update contentdata into data access failed")
	ErrReadContentFailed             = errors.New("read content failed")
	ErrDeleteContentFailed           = errors.New("delete contentdata into data access failed")
	ErrInvalidVisibleScope           = errors.New("invalid visible scope")

	ErrSuggestTimeTooSmall = errors.New("suggest time is less than sub contents")

	ErrInvalidMaterialType    = errors.New("invalid material type")
	ErrInvalidSourceOrContent = errors.New("invalid content data source or content")

	ErrEmptyTeacherManual = errors.New("empty teacher manual")

	ErrBadRequest         = errors.New("bad request")
	ErrResourceNotFound   = errors.New("resource not found")
	ErrInvalidContentType = errors.New("invalid content type")
	ErrNoRejectReason     = errors.New("no reject reason")

	ErrInvalidSelectForm = errors.New("invalid select form")
	ErrUserNotFound      = errors.New("user not found in locked by")

	ErrMoveToDifferentPartition = errors.New("move to different parititon")
)

type ErrContentAlreadyLocked struct {
	LockedBy *external.User
}

func (e *ErrContentAlreadyLocked) Error() string {
	return "content is already locked"
}
func NewErrContentAlreadyLocked(ctx context.Context, lockedBy string, operator *entity.Operator) error {
	user, err := external.GetUserServiceProvider().Get(ctx, operator, lockedBy)
	if err != nil {
		return ErrUserNotFound
	}
	return &ErrContentAlreadyLocked{LockedBy: user}
}

type visiblePermission string

var (
	visiblePermissionPublished visiblePermission = "published"
	visiblePermissionPending   visiblePermission = "pending"
)

type IContentModel interface {
	CreateContent(ctx context.Context, tx *dbo.DBContext, c entity.CreateContentRequest, operator *entity.Operator) (string, error)
	UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, data entity.CreateContentRequest, user *entity.Operator) error
	PublishContent(ctx context.Context, tx *dbo.DBContext, cid string, scope []string, user *entity.Operator) error
	PublishContentWithAssets(ctx context.Context, tx *dbo.DBContext, cid string, scope []string, user *entity.Operator) error
	LockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error)
	DeleteContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error
	CloneContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error)

	PublishContentBulk(ctx context.Context, tx *dbo.DBContext, ids []string, user *entity.Operator) error
	DeleteContentBulk(ctx context.Context, tx *dbo.DBContext, ids []string, user *entity.Operator) error

	GetContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error)
	GetContentByIDList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) ([]*entity.ContentInfoWithDetails, error)
	GetLatestContentIDByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]string, error)
	GetPastContentIDByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error)
	GetRawContentByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.Content, error)

	GetContentNameByID(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.ContentName, error)
	GetContentNameByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.ContentName, error)
	GetContentSubContentsByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) ([]*entity.SubContentsWithName, error)
	GetContentsSubContentsMapByIDList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) (map[string][]*entity.SubContentsWithName, error)

	UpdateContentPublishStatus(ctx context.Context, tx *dbo.DBContext, cid string, reason []string, remark, status string) error
	CheckContentAuthorization(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error

	CountUserFolderContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, error)
	SearchUserPrivateFolderContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.FolderContentData, error)
	SearchUserFolderContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.FolderContentData, error)
	SearchUserContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)
	SearchUserPrivateContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)
	ListPendingContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)
	SearchContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)

	GetContentOutcomeByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error)
	GetVisibleContentOutcomeByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error)
	ContentDataCount(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.ContentStatisticsInfo, error)
	GetVisibleContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error)

	IsContentsOperatorByIDList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) (bool, error)
	ListVisibleScopes(ctx context.Context, permission visiblePermission, operator *entity.Operator) ([]string, error)

	UpdateContentPath(ctx context.Context, tx *dbo.DBContext, cid string, path string) error
	BatchReplaceContentPath(ctx context.Context, tx *dbo.DBContext, cids []string, oldPath, path string) error

	//For authed content
	SearchAuthedContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)
	CopyContent(ctx context.Context, tx *dbo.DBContext, cid string, deep bool, op *entity.Operator) (string, error)

	CreateContentData(ctx context.Context, contentType entity.ContentType, data string) (entity.ContentData, error)
	ConvertContentObj(ctx context.Context, tx *dbo.DBContext, obj *entity.Content, operator *entity.Operator) (*entity.ContentInfo, error)
	BatchConvertContentObj(ctx context.Context, tx *dbo.DBContext, objs []*entity.Content, operator *entity.Operator) ([]*entity.ContentInfo, error)

	PublishContentWithAssetsTx(ctx context.Context, cid string, scope []string, user *entity.Operator) error
	LockContentTx(ctx context.Context, cid string, user *entity.Operator) (string, error)
	CreateContentTx(ctx context.Context, c entity.CreateContentRequest, operator *entity.Operator) (string, error)
	CopyContentTx(ctx context.Context, cid string, deep bool, op *entity.Operator) (string, error)
	PublishContentBulkTx(ctx context.Context, ids []string, user *entity.Operator) error
	PublishContentTx(ctx context.Context, cid string, scope []string, user *entity.Operator) error
	DeleteContentBulkTx(ctx context.Context, ids []string, user *entity.Operator) error
	DeleteContentTx(ctx context.Context, cid string, user *entity.Operator) error
}

type ContentModel struct {
}

func (cm *ContentModel) handleSourceContent(ctx context.Context, tx *dbo.DBContext, contentID, sourceID string) error {
	sourceContent, err := da.GetContentDA().GetContentByID(ctx, tx, sourceID)
	if err != nil {
		return err
	}
	sourceContent.PublishStatus = entity.ContentStatusHidden
	sourceContent.LatestID = contentID
	//解锁source content
	//Unlock source content
	sourceContent.LockedBy = constant.LockedByNoBody
	err = da.GetContentDA().UpdateContent(ctx, tx, sourceID, *sourceContent)
	if err != nil {
		log.Error(ctx, "update source content failed", log.Err(err))
		return ErrUpdateContentFailed
	}

	err = folderModel.RemoveItemByLink(ctx, tx, entity.OwnerTypeOrganization, sourceContent.Org, entity.ContentLink(sourceContent.ID))
	if err != nil {
		log.Error(ctx, "remove old content folder item failed", log.Err(err), log.Any("content", sourceContent))
		return err
	}

	//更新所有latestID为sourceContent的Content
	//Update all sourceContent latestID fields
	_, oldContents, err := da.GetContentDA().SearchContent(ctx, tx, da.ContentCondition{
		LatestID: sourceContent.ID,
	})
	if err != nil {
		log.Error(ctx, "update old content failed", log.Err(err), log.String("SourceID", sourceContent.ID))
		return ErrUpdateContentFailed
	}

	oldContentIDs := make([]string, len(oldContents))
	for i := range oldContents {
		oldContents[i].LockedBy = constant.LockedByNoBody
		oldContents[i].PublishStatus = entity.ContentStatusHidden
		oldContents[i].LatestID = contentID
		err = da.GetContentDA().UpdateContent(ctx, tx, oldContents[i].ID, *oldContents[i])
		if err != nil {
			log.Error(ctx, "update old content failed", log.Err(err), log.String("OldID", oldContents[i].ID))
			return ErrUpdateContentFailed
		}

		oldContentIDs[i] = oldContents[i].ID
	}

	//TODO:For authed content => handle source for authed content list version => done
	err = GetAuthedContentRecordsModel().BatchUpdateVersion(ctx, tx, oldContentIDs, contentID)
	if err != nil {
		log.Error(ctx, "batch update authed content reocrds failed",
			log.Err(err),
			log.String("contentID", contentID),
			log.Strings("oldContentIDs", oldContentIDs))
		return err
	}

	return nil
}

func (cm *ContentModel) doPublishContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, scope []string, user *entity.Operator) error {
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

	//If scope changed, refresh visibility settings
	if scope != nil {
		err = cm.refreshContentVisibilitySettings(ctx, tx, content.ID, scope)
		if err != nil {
			log.Error(ctx, "refreshContentVisibilitySettings failed",
				log.Err(err),
				log.String("cid", content.ID),
				log.Strings("scope", scope),
				log.String("uid", user.UserID))
			return ErrUpdateContentFailed
		}
	}

	return nil
}

func (cm ContentModel) checkContentInfo(ctx context.Context, c entity.CreateContentRequest) error {
	if c.LessonType != "" {
		_, err := GetLessonTypeModel().GetByID(ctx, c.LessonType)
		if err != nil {
			log.Error(ctx, "lesson type invalid", log.Any("data", c), log.Err(err))
			return ErrInvalidContentData
		}
	}
	err := c.Validate()
	if err != nil {
		log.Error(ctx, "asset no need to check", log.Any("data", c), log.Err(err))
		return err
	}
	err = c.ContentType.Validate()
	if err != nil {
		log.Error(ctx, "content type invalid", log.Any("data", c), log.Err(err))
		return err
	}
	if len(c.Category) == 0 ||
		c.Program == "" {
		log.Error(ctx, "select form invalid", log.Any("data", c), log.Err(err))
		return ErrInvalidSelectForm
	}

	return nil
}

func (cm ContentModel) checkUpdateContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) (*entity.Content, error) {

	//若为asset，直接发布
	//if the content is assets, publish it immediate
	if content.ContentType.IsAsset() {
		log.Info(ctx, "asset no need to check", log.String("cid", content.ID))
		return content, nil
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
	//if content is published or pending, create a new content
	if content.PublishStatus != entity.ContentStatusDraft && content.PublishStatus != entity.ContentStatusRejected &&
		content.PublishStatus != entity.ContentStatusArchive {
		//报错
		//error
		log.Warn(ctx, "invalid content status", log.Any("content", content))
		return ErrInvalidContentStatusToPublish
	}

	contentData, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		log.Warn(ctx, "create content data failed", log.Any("contentData", contentData), log.Err(err))
		return err
	}
	subContentIDs := contentData.SubContentIDs(ctx)
	//No sub content, no need to check
	if len(subContentIDs) < 1 {
		return nil
	}

	_, contentList, err := da.GetContentDA().SearchContent(ctx, tx, da.ContentCondition{
		IDS: subContentIDs,
	})
	if err != nil {
		log.Error(ctx, "search content data failed", log.Any("IDS", subContentIDs), log.Err(err))
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
	response, err := cm.BatchConvertContentObj(ctx, tx, objs, user)
	if err != nil {
		log.Error(ctx, "Can't parse contentdata, contentID: %v, error: %v", log.Any("objs", objs), log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, err
	}

	contentWithDetails, err := cm.buildContentWithDetails(ctx, response, false, user)
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
	response, err := cm.BatchConvertContentObj(ctx, tx, objs, user)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, err
	}
	contentWithDetails, err := cm.buildContentWithDetails(ctx, response, false, user)
	if err != nil {
		log.Error(ctx, "build content details failed", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, err
	}

	return count, contentWithDetails, nil
}
func (cm *ContentModel) CreateContentTx(ctx context.Context, c entity.CreateContentRequest, operator *entity.Operator) (string, error) {
	cid, err := dbo.GetTransResult(ctx, func(ctx context.Context, tx *dbo.DBContext) (interface{}, error) {
		cid, err := cm.CreateContent(ctx, tx, c, operator)
		if err != nil {
			return "", err
		}
		return cid, nil
	})
	if cid == nil {
		return "", err
	}
	return cid.(string), err
}
func (cm *ContentModel) CreateContent(ctx context.Context, tx *dbo.DBContext, c entity.CreateContentRequest, operator *entity.Operator) (string, error) {
	//检查数据信息是否正确
	//valid the data
	c.Trim()

	log.Info(ctx, "create content")
	if c.ContentType.IsAsset() {
		// use operator's org id as asset publish scope, maybe not right...
		c.PublishScope = []string{operator.OrgID}
	}

	err := cm.checkContentInfo(ctx, c)
	if err != nil {
		log.Warn(ctx, "check content failed", log.Err(err), log.String("uid", operator.UserID), log.Any("data", c))
		return "", err
	}

	//组装要创建的内容
	//construct the new content structure
	obj, err := cm.prepareCreateContentParams(ctx, c, operator)
	if err != nil {
		log.Warn(ctx, "prepare content failed", log.Err(err), log.String("uid", operator.UserID), log.Any("data", c))
		return "", err
	}

	//添加内容
	//do insert content into database
	now := time.Now()
	obj.UpdateAt = now.Unix()
	obj.CreateAt = now.Unix()
	pid, err := da.GetContentDA().CreateContent(ctx, tx, *obj)
	if err != nil {
		log.Error(ctx, "can't create contentdata", log.Err(err), log.String("uid", operator.UserID), log.Any("data", c))
		return "", err
	}

	//Insert content properties
	err = cm.doCreateContentProperties(ctx, tx, entity.ContentProperties{
		ContentID:   pid,
		Program:     c.Program,
		Subject:     c.Subject,
		Category:    c.Category,
		SubCategory: c.SubCategory,
		Age:         c.Age,
		Grade:       c.Grade,
	}, false)
	if err != nil {
		log.Error(ctx, "doCreateContentProperties failed",
			log.Err(err),
			log.String("uid", operator.UserID),
			log.Any("data", c))
		return "", err
	}

	//Insert into visibility settings
	err = cm.insertContentVisibilitySettings(ctx, tx, pid, c.PublishScope)
	if err != nil {
		log.Error(ctx, "insertContentVisibilitySettings failed",
			log.Err(err),
			log.String("uid", operator.UserID),
			log.String("pid", pid),
			log.Any("data", c))
		return "", err
	}

	//Asset添加Folder
	//assets add to folder
	if c.ContentType.IsAsset() {
		err = GetFolderModel().AddOrUpdateOrgFolderItem(ctx, tx, entity.FolderPartitionAssets, constant.FolderRootPath, entity.ContentLink(pid), operator)
		if err != nil {
			log.Error(ctx, "can't create folder item", log.Err(err),
				log.String("link", entity.ContentLink(pid)),
				log.Any("data", c),
				log.Any("operator", operator))
			return "", err
		}
	}

	return pid, nil
}

func (cm *ContentModel) UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, data entity.CreateContentRequest, user *entity.Operator) error {
	data.Trim()
	if data.ContentType.IsAsset() {
		//Assets can't be updated
		return ErrInvalidContentType
	}

	err := cm.checkContentInfo(ctx, data)
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
	//do update data into database
	err = da.GetContentDA().UpdateContent(ctx, tx, cid, *obj)
	if err != nil {
		log.Error(ctx, "update contentdata failed", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID), log.Any("data", data))
		return err
	}
	//Insert content properties
	err = cm.doCreateContentProperties(ctx, tx, entity.ContentProperties{
		ContentID:   cid,
		Program:     data.Program,
		Subject:     data.Subject,
		Category:    data.Category,
		SubCategory: data.SubCategory,
		Age:         data.Age,
		Grade:       data.Grade,
	}, true)
	if err != nil {
		log.Error(ctx, "doCreateContentProperties failed",
			log.Err(err),
			log.String("uid", user.UserID),
			log.Any("data", data))
		return err
	}

	//若已发布，不能修改publishScope
	//If it is already published, can't update publish scope
	if content.PublishStatus == entity.ContentStatusDraft ||
		content.PublishStatus == entity.ContentStatusRejected {
		err = cm.refreshContentVisibilitySettings(ctx, tx, content.ID, data.PublishScope)
		if err != nil {
			log.Error(ctx, "refreshContentVisibilitySettings failed",
				log.Err(err),
				log.String("cid", cid),
				log.String("uid", user.UserID),
				log.Any("data", data))
			return err
		}
	}

	da.GetContentRedis().CleanContentCache(ctx, []string{cid})
	return nil
}

func (cm *ContentModel) UpdateContentPath(ctx context.Context, tx *dbo.DBContext, cid string, path string) error {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't read content on update path", log.Err(err))
		return err
	}
	content.DirPath = path
	err = da.GetContentDA().UpdateContent(ctx, tx, cid, *content)
	if err != nil {
		log.Error(ctx, "update content path failed", log.Err(err))
		return ErrUpdateContentFailed
	}

	return nil
}

func (cm *ContentModel) BatchReplaceContentPath(ctx context.Context, tx *dbo.DBContext, cids []string, oldPath, path string) error {
	err := da.GetContentDA().BatchReplaceContentPath(ctx, tx, cids, oldPath, path)
	if err != nil {
		log.Error(ctx, "update content path failed", log.Err(err))
		return ErrUpdateContentFailed
	}

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
	operator := &entity.Operator{
		UserID: content.Author,
		OrgID:  content.Org,
	}

	//更新content的path
	err = cm.checkAndUpdateContentPath(ctx, tx, content, operator)
	if err != nil {
		return err
	}

	rejectReason := strings.Join(reason, constant.StringArraySeparator)
	content.RejectReason = rejectReason
	content.Remark = remark
	err = da.GetContentDA().UpdateContent(ctx, tx, cid, *content)
	if err != nil {
		log.Error(ctx, "update contentdata scope failed", log.Err(err))
		return ErrUpdateContentFailed
	}
	//更新Folder信息
	//update folder info
	err = GetFolderModel().AddOrUpdateOrgFolderItem(ctx, tx, entity.FolderPartitionMaterialAndPlans, content.DirPath, entity.ContentLink(content.ID), operator)
	if err != nil {
		return err
	}

	if status == entity.ContentStatusPublished && content.SourceID != "" {
		//处理source content
		//handle with source content
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
	content.LockedBy = constant.LockedByNoBody
	return da.GetContentDA().UpdateContent(ctx, tx, cid, *content)
}

func (cm *ContentModel) LockContentTx(ctx context.Context, cid string, user *entity.Operator) (string, error) {
	ncid, err := dbo.GetTransResult(ctx, func(ctx context.Context, tx *dbo.DBContext) (interface{}, error) {
		ncid, err := cm.LockContent(ctx, tx, cid, user)
		if err != nil {
			return nil, err
		}
		return ncid, nil
	})
	if ncid == nil {
		return "", err
	}
	return ncid.(string), err
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
	//if it is locked by current user, return cloned content id
	if content.LockedBy == user.UserID {
		_, data, err := da.GetContentDA().SearchContent(ctx, tx, da.ContentCondition{
			SourceID: cid,
		})
		if err != nil {
			log.Error(ctx, "search source content failed", log.String("cid", cid))
			return "", err
		}
		if len(data) < 1 {
			//被自己锁定且找不到content
			//if locked by current user, but content is not found, panic
			log.Info(ctx, "no content in source content", log.String("cid", cid))
			return "", ErrNoContent
		}
		if data[0].PublishStatus != entity.ContentStatusRejected && data[0].PublishStatus != entity.ContentStatusDraft {
			log.Info(ctx, "invalid locked content status", log.String("lock cid", data[0].ID), log.String("status", string(data[0].PublishStatus)), log.String("cid", cid))
			return "", ErrInvalidLockedContentPublishStatus
		}
		//找到data
		//find the data
		return data[0].ID, nil
	}

	//更新锁定状态
	//update lock status
	if content.LockedBy != "" && content.LockedBy != constant.LockedByNoBody {
		return "", NewErrContentAlreadyLocked(ctx, content.LockedBy, user)
	}
	content.LockedBy = user.UserID
	//content.Author = user.UserID
	err = da.GetContentDA().UpdateContent(ctx, tx, cid, *content)
	if err != nil {
		return "", err
	}
	//克隆Content
	//clone content
	ccid, err := cm.CloneContent(ctx, tx, content.ID, user)
	if err != nil {
		return "", err
	}
	da.GetContentRedis().CleanContentCache(ctx, []string{cid, content.ID})

	//Update with new content
	return ccid, nil

}
func (cm *ContentModel) PublishContentBulkTx(ctx context.Context, ids []string, user *entity.Operator) error {
	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := cm.PublishContentBulk(ctx, tx, ids, user)
		if err != nil {
			return err
		}
		return nil
	})
}
func (cm *ContentModel) PublishContentBulk(ctx context.Context, tx *dbo.DBContext, ids []string, user *entity.Operator) error {
	updateIDs := make([]string, 0)
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
		err = cm.doPublishContent(ctx, tx, contents[i], nil, user)
		if err != nil {
			log.Error(ctx, "can't publish content", log.Err(err), log.Strings("ids", ids), log.String("uid", user.UserID))
			return err
		}
		updateIDs = append(updateIDs, contents[i].ID)
		if contents[i].SourceID != "" {
			updateIDs = append(updateIDs, contents[i].SourceID)
		}
	}
	da.GetContentRedis().CleanContentCache(ctx, updateIDs)
	return err
}

//TODO:For authed content => implement search auth content => done
func (cm *ContentModel) SearchAuthedContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	//set condition with authed flag
	condition.AuthedOrgID = entity.NullStrings{
		Strings: []string{user.OrgID, constant.ShareToAll},
		Valid:   true,
	}
	condition.PublishStatus = []string{entity.ContentStatusPublished}
	return cm.searchContent(ctx, tx, &condition, user)
}
func (cm *ContentModel) CopyContentTx(ctx context.Context, cid string, deep bool, op *entity.Operator) (string, error) {
	id, err := dbo.GetTransResult(ctx, func(ctx context.Context, tx *dbo.DBContext) (interface{}, error) {
		cid, err := cm.CopyContent(ctx, tx, cid, deep, op)
		if err != nil {
			return "", err
		}
		return cid, nil
	})
	if id == nil {
		return "", err
	}
	return id.(string), err
}

//TODO:For authed content => implement copy content => done
func (cm *ContentModel) CopyContent(ctx context.Context, tx *dbo.DBContext, cid string, deep bool, op *entity.Operator) (string, error) {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", op.UserID))
		return "", ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read content on copy", log.Err(err), log.String("cid", cid), log.String("uid", op.UserID))
		return "", err
	}
	if content.ContentType.IsAsset() {
		return "", ErrInvalidContentType
	}

	//if deep copy & content is lesson plan copy sub contents
	if deep && content.ContentType == entity.ContentTypePlan {
		//get sub contents in plan
		cd, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
		if err != nil {
			return "", err
		}
		materialIDs := cd.SubContentIDs(ctx)
		//深度拷贝
		//copy sub contents & get id map
		materialMap, err := cm.copyContentList(ctx, tx, materialIDs, op)
		//replace subcontent ids
		cd.ReplaceContentIDs(ctx, materialMap)

		newData, err := cd.Marshal(ctx)
		if err != nil {
			return "", err
		}
		content.Data = newData
	}

	id, err := cm.doCopyContent(ctx, tx, content, op)
	if err != nil {
		return "", err
	}

	return id, nil
}
func (cm *ContentModel) copyContentList(ctx context.Context, tx *dbo.DBContext, cids []string, op *entity.Operator) (map[string]string, error) {
	contentList, err := da.GetContentDA().GetContentByIDList(ctx, tx, cids)
	if err != nil {
		log.Error(ctx, "can't read content on copy", log.Err(err), log.Strings("cids", cids), log.String("uid", op.UserID))
		return nil, err
	}
	if len(contentList) != len(cids) {
		log.Warn(ctx, "copy content list contains invalid content",
			log.Err(err),
			log.Strings("cids", cids),
			log.Any("contentList", contentList))
		return nil, ErrNoContent
	}
	ret := make(map[string]string)
	for i := range contentList {
		copyedID, err := cm.doCopyContent(ctx, tx, contentList[i], op)
		if err != nil {
			return nil, err
		}
		ret[contentList[i].ID] = copyedID
	}
	return ret, nil
}
func (cm *ContentModel) doCopyContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, op *entity.Operator) (string, error) {
	//检查是否有克隆权限
	//check if user have copy permission
	err := cm.CheckContentAuthorization(ctx, tx, &entity.Content{
		ID:            content.ID,
		PublishStatus: content.PublishStatus,
		Author:        content.Author,
		Org:           content.Org,
	}, op)
	if err != nil {
		log.Error(ctx, "no auth to read content for cloning",
			log.Err(err),
			log.String("cid", content.ID),
			log.String("uid", op.UserID))
		return "", ErrCloneContentFailed
	}

	obj := cm.prepareCopyContentParams(ctx, content, op)

	now := time.Now()
	obj.UpdateAt = now.Unix()
	obj.CreateAt = now.Unix()
	id, err := da.GetContentDA().CreateContent(ctx, tx, *obj)
	if err != nil {
		log.Error(ctx, "clone contentdata failed",
			log.Err(err),
			log.String("cid", content.ID),
			log.String("uid", op.UserID))
		return "", err
	}
	//load content properties
	contentProperties, err := cm.getContentProperties(ctx, content.ID)
	if err != nil {
		log.Warn(ctx, "getContentProperties failed",
			log.Err(err), log.String("cid", content.ID))
		return "", ErrInvalidContentData
	}
	contentProperties.ContentID = id
	err = cm.doCreateContentProperties(ctx, tx, *contentProperties, false)
	if err != nil {
		log.Warn(ctx, "doCreateContentProperties failed",
			log.Err(err),
			log.Any("contentProperties", contentProperties),
			log.String("cid", content.ID))
		return "", err
	}

	contentVisibilitySettings, err := cm.getContentVisibilitySettings(ctx, content.ID)
	if err != nil {
		log.Warn(ctx, "getContentVisibilitySettings failed",
			log.Err(err), log.String("cid", content.ID))
		return "", ErrInvalidContentData
	}
	err = cm.insertContentVisibilitySettings(ctx, tx, id, contentVisibilitySettings.VisibilitySettings)
	if err != nil {
		log.Warn(ctx, "insertContentVisibilitySettings failed",
			log.Err(err), log.String("cid", content.ID),
			log.Strings("VisibilitySettings", contentVisibilitySettings.VisibilitySettings))
		return "", ErrInvalidContentData
	}

	return id, nil
}

func (cm *ContentModel) PublishContentTx(ctx context.Context, cid string, scope []string, user *entity.Operator) error {
	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := cm.PublishContent(ctx, tx, cid, scope, user)
		return err
	})
}
func (cm *ContentModel) PublishContent(ctx context.Context, tx *dbo.DBContext, cid string, scope []string, user *entity.Operator) error {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata for publishing", log.Err(err), log.String("cid", cid), log.Strings("scope", scope), log.String("uid", user.UserID))
		return err
	}
	if content.ContentType.IsAsset() {
		return ErrInvalidContentType
	}

	err = cm.doPublishContent(ctx, tx, content, scope, user)
	if err != nil {
		return err
	}

	da.GetContentRedis().CleanContentCache(ctx, []string{cid, content.SourceID})
	return nil
}

func (cm *ContentModel) validatePublishContentWithAssets(ctx context.Context, content *entity.Content, user *entity.Operator) error {
	if content.ContentType != entity.ContentTypeMaterial && content.ContentType != entity.ContentTypePlan {
		return ErrInvalidContentType
	}
	//查看data
	cd, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		log.Warn(ctx, "create content data failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}
	switch data := cd.(type) {
	case *MaterialData:
		if data.InputSource != entity.MaterialInputSourceDisk {
			log.Warn(ctx, "invalid material type", log.Err(err), log.String("uid", user.UserID), log.Any("data", data))
			return ErrInvalidMaterialType
		}
	case *LessonData:
		if len(data.TeacherManualBatch) < 1 {
			log.Warn(ctx, "empty teacher manual batch", log.Err(err), log.String("uid", user.UserID), log.Any("data", data))
			return ErrEmptyTeacherManual
		}

	default:
		log.Warn(ctx, "validate content data with content type 2 failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}

	return nil
}

func (cm *ContentModel) prepareForPublishMaterialsAssets(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	//创建data对象
	//create content data object
	cd, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		log.Warn(ctx, "create content data failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}
	//解析data的fileType
	err = cd.PrepareSave(ctx, entity.ExtraDataInRequest{})
	materialData, ok := cd.(*MaterialData)
	if !ok {
		log.Warn(ctx, "asset content data type failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}

	//load content properties
	contentProperties, err := cm.getContentProperties(ctx, content.ID)
	if err != nil {
		log.Warn(ctx, "getContentProperties failed",
			log.Err(err), log.String("cid", content.ID))
		return ErrInvalidContentData
	}

	//创建assets data对象，并解析
	//create assets data object, and parse it
	assetsData := new(AssetsData)
	assetsData.Source = materialData.Source
	assetsDataJSON, err := assetsData.Marshal(ctx)
	if err != nil {
		log.Warn(ctx, "marshal assets data failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrMarshalContentDataFailed
	}
	//创建assets
	//Create assets
	req := entity.CreateContentRequest{
		ContentType: entity.ContentTypeAssets,
		Name:        content.Name,
		Keywords:    utils.StringToStringArray(ctx, content.Keywords),
		Description: content.Description,
		Thumbnail:   content.Thumbnail,
		SuggestTime: content.SuggestTime,
		Data:        assetsDataJSON,

		Program:     contentProperties.Program,
		Subject:     contentProperties.Subject,
		Category:    contentProperties.Category,
		SubCategory: contentProperties.SubCategory,
		Age:         contentProperties.Age,
		Grade:       contentProperties.Grade,
	}
	_, err = cm.CreateContent(ctx, tx, req, user)
	if err != nil {
		log.Warn(ctx, "create assets failed", log.Err(err), log.String("uid", user.UserID), log.Any("req", req))
		return err
	}

	//更新content状态
	//update content status
	materialData.InputSource = entity.MaterialInputSourceDisk
	d, err := materialData.Marshal(ctx)
	if err != nil {
		log.Warn(ctx, "marshal content data failed", log.Err(err), log.String("uid", user.UserID), log.Any("content", content))
		return err
	}

	content.Data = d
	return nil
}

func (cm *ContentModel) prepareForPublishPlansAssets(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	//创建data对象
	//create content data object
	cd, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		log.Warn(ctx, "create content data failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}
	//解析data的fileType
	//parse data for fileType
	err = cd.PrepareSave(ctx, entity.ExtraDataInRequest{})
	lessonData, ok := cd.(*LessonData)
	if !ok {
		log.Warn(ctx, "asset content data type failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}

	//创建assets data对象，并解析
	//create assets data object, and parse it
	for i := range lessonData.TeacherManualBatch {
		assetsData := new(AssetsData)
		assetsData.Source = SourceID(lessonData.TeacherManualBatch[i].ID)
		assetsDataJSON, err := assetsData.Marshal(ctx)
		if err != nil {
			log.Warn(ctx, "marshal assets data failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
			return ErrMarshalContentDataFailed
		}
		//创建assets
		req := entity.CreateContentRequest{
			ContentType: entity.ContentTypeAssets,
			Name:        content.Name,
			Keywords:    append(utils.StringToStringArray(ctx, content.Keywords), constant.TeacherManualAssetsKeyword),
			Description: content.Description,
			Thumbnail:   "",
			SuggestTime: 0,
			Data:        assetsDataJSON,
		}
		pid, err := cm.CreateContent(ctx, tx, req, user)
		if err != nil {
			log.Warn(ctx, "create assets failed", log.Err(err), log.String("uid", user.UserID), log.Any("req", req))
			return err
		}
		err = cm.copyContentProperties(ctx, tx, content.ID, pid)
		if err != nil {
			log.Warn(ctx, "copyContentProperties failed",
				log.Err(err),
				log.String("uid", user.UserID),
				log.Any("content", content),
				log.Any("req", req))
			return err
		}
	}
	return nil
}

func (cm *ContentModel) PublishContentWithAssetsTx(ctx context.Context, cid string, scope []string, user *entity.Operator) error {
	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		return cm.PublishContentWithAssets(ctx, tx, cid, scope, user)
	})
}
func (cm *ContentModel) PublishContentWithAssets(ctx context.Context, tx *dbo.DBContext, cid string, scope []string, user *entity.Operator) error {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read content data for publishing", log.Err(err), log.String("cid", cid), log.Strings("scope", scope), log.String("uid", user.UserID))
		return err
	}
	switch content.ContentType {
	case entity.ContentTypeMaterial:
		err = cm.publishMaterialWithAssets(ctx, tx, content, scope, user)
		if err != nil {
			return err
		}
	case entity.ContentTypePlan:
		err = cm.doPublishPlanWithAssets(ctx, tx, content, scope, user)
		if err != nil {
			return err
		}
	default:
		log.Warn(ctx, "content invalid",
			log.String("cid", cid),
			log.Any("content", content))
		return ErrInvalidContentType
	}

	da.GetContentRedis().CleanContentCache(ctx, []string{content.ID, content.SourceID})
	return nil
}
func (cm *ContentModel) DeleteContentBulkTx(ctx context.Context, ids []string, user *entity.Operator) error {
	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := cm.DeleteContentBulk(ctx, tx, ids, user)
		if err != nil {
			return err
		}
		return nil
	})
}
func (cm *ContentModel) DeleteContentBulk(ctx context.Context, tx *dbo.DBContext, ids []string, user *entity.Operator) error {
	deletedIDs := make([]string, 0)
	deletedIDs = append(deletedIDs, ids...)
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
			deletedIDs = append(deletedIDs, contents[i].SourceID)
		}
	}
	da.GetContentRedis().CleanContentCache(ctx, deletedIDs)
	return nil
}

func (cm *ContentModel) publishMaterialWithAssets(ctx context.Context, tx *dbo.DBContext, content *entity.Content, scope []string, user *entity.Operator) error {
	err := cm.validatePublishContentWithAssets(ctx, content, user)
	if err != nil {
		log.Error(ctx, "validate for publishing failed", log.Err(err), log.String("cid", content.ID), log.Strings("scope", scope), log.String("uid", user.UserID))
		return err
	}

	//准备发布（1.创建assets，2.修改contentdata）
	//preparing to publish (1.create assets 2.update content data)
	err = cm.prepareForPublishMaterialsAssets(ctx, tx, content, user)
	if err != nil {
		return err
	}

	//发布
	//do publish
	err = cm.doPublishContent(ctx, tx, content, scope, user)
	if err != nil {
		return err
	}

	return nil
}

func (cm *ContentModel) doPublishPlanWithAssets(ctx context.Context, tx *dbo.DBContext, content *entity.Content, scope []string, user *entity.Operator) error {
	err := cm.validatePublishContentWithAssets(ctx, content, user)
	if err != nil {
		log.Error(ctx, "validate for publishing failed", log.Err(err), log.String("cid", content.ID), log.Strings("scope", scope), log.String("uid", user.UserID))
		return err
	}
	//load content properties
	contentProperties, err := cm.getContentProperties(ctx, content.ID)
	if err != nil {
		log.Warn(ctx, "getContentProperties failed",
			log.Err(err), log.String("cid", content.ID))
		return ErrInvalidContentData
	}

	contentVisibilitySettings, err := cm.getContentVisibilitySettings(ctx, content.ID)
	if err != nil {
		log.Warn(ctx, "getContentVisibilitySettings failed",
			log.Err(err), log.String("cid", content.ID))
		return ErrInvalidContentData
	}

	//create content data object
	cd, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		log.Warn(ctx, "create content data failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}

	//parse data for lesson plan
	lessonData, ok := cd.(*LessonData)
	if !ok {
		log.Warn(ctx, "asset content data type failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}
	log.Info(ctx,
		"publish plans with teacher manuals",
		log.Any("content", content),
		log.Any("data", lessonData.TeacherManualBatch))
	//create assets data object, and parse it
	for i := range lessonData.TeacherManualBatch {
		assetsData := new(AssetsData)
		assetsData.Source = SourceID(lessonData.TeacherManualBatch[i].ID)
		assetsDataJSON, err := assetsData.Marshal(ctx)
		if err != nil {
			log.Warn(ctx, "marshal assets data failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
			return ErrMarshalContentDataFailed
		}
		//创建assets
		req := entity.CreateContentRequest{
			ContentType:  entity.ContentTypeAssets,
			Name:         content.Name,
			Keywords:     append(utils.StringToStringArray(ctx, content.Keywords), constant.TeacherManualAssetsKeyword),
			Description:  content.Description,
			Thumbnail:    "",
			SuggestTime:  0,
			Program:      contentProperties.Program,
			Subject:      contentProperties.Subject,
			Category:     contentProperties.Category,
			SubCategory:  contentProperties.SubCategory,
			Age:          contentProperties.Age,
			Grade:        contentProperties.Grade,
			PublishScope: contentVisibilitySettings.VisibilitySettings,
			Data:         assetsDataJSON,
		}
		_, err = cm.CreateContent(ctx, tx, req, user)
		if err != nil {
			log.Warn(ctx, "create assets failed", log.Err(err), log.String("uid", user.UserID), log.Any("req", req))
			return err
		}
	}
	//do publish
	err = cm.doPublishContent(ctx, tx, content, scope, user)
	if err != nil {
		return err
	}
	return nil
}

func (cm *ContentModel) publishPlanWithAssets(ctx context.Context, tx *dbo.DBContext, content *entity.Content, scope []string, user *entity.Operator) error {
	err := cm.validatePublishContentWithAssets(ctx, content, user)
	if err != nil {
		log.Error(ctx, "validate for publishing failed", log.Err(err), log.String("cid", content.ID), log.Strings("scope", scope), log.String("uid", user.UserID))
		return err
	}

	//准备发布（1.创建assets，2.修改contentdata）
	//preparing to publish (1.create assets 2.update content data)
	err = cm.prepareForPublishPlansAssets(ctx, tx, content, user)
	if err != nil {
		return err
	}

	//发布
	//do publish
	err = cm.doPublishContent(ctx, tx, content, scope, user)
	if err != nil {
		return err
	}

	return nil
}
func (cm *ContentModel) checkDeleteContent(ctx context.Context, content *entity.Content) error {
	if content.PublishStatus == entity.ContentStatusArchive && content.ContentType == entity.ContentTypePlan {
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
	if content.LockedBy != constant.LockedByNoBody && content.LockedBy != user.UserID {
		return NewErrContentAlreadyLocked(ctx, content.LockedBy, user)
	}
	if content.PublishStatus == entity.ContentStatusPublished && content.LockedBy != constant.LockedByNoBody {
		return NewErrContentAlreadyLocked(ctx, content.LockedBy, user)
	}

	err := cm.checkDeleteContent(ctx, content)
	if err != nil {
		log.Error(ctx, "check delete content failed", log.Err(err), log.String("cid", content.ID), log.String("uid", user.UserID))
		return err
	}

	obj := cm.prepareDeleteContentParams(ctx, content, content.PublishStatus, user)

	err = da.GetContentDA().UpdateContent(ctx, tx, content.ID, *obj)
	if err != nil {
		log.Error(ctx, "delete contentdata failed", log.Err(err), log.String("cid", content.ID), log.String("uid", user.UserID))
		return err
	}

	//folder中删除
	//temp remove folder
	err = GetFolderModel().RemoveItemByLink(ctx, tx, entity.OwnerTypeOrganization, content.Org, entity.ContentLink(content.ID))
	if err != nil {
		log.Error(ctx, "remove content folder item failed", log.Err(err), log.Any("content", content))
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
func (cm *ContentModel) DeleteContentTx(ctx context.Context, cid string, user *entity.Operator) error {
	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := cm.DeleteContent(ctx, tx, cid, user)
		if err != nil {
			return err
		}
		return nil
	})
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
	//check permission
	err = cm.CheckContentAuthorization(ctx, tx, &entity.Content{
		ID:            content.ID,
		PublishStatus: content.PublishStatus,
		Author:        content.Author,
		Org:           content.Org,
	}, user)
	if err != nil {
		log.Error(ctx, "no auth to read content for cloning", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return "", ErrCloneContentFailed
	}

	obj := cm.prepareCloneContentParams(ctx, content, user)

	now := time.Now()
	obj.UpdateAt = now.Unix()
	obj.CreateAt = now.Unix()
	id, err := da.GetContentDA().CreateContent(ctx, tx, *obj)
	if err != nil {
		log.Error(ctx, "clone contentdata failed", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return "", err
	}
	//load content properties
	contentProperties, err := cm.getContentProperties(ctx, cid)
	if err != nil {
		log.Warn(ctx, "getContentProperties failed",
			log.Err(err), log.String("cid", cid))
		return "", ErrInvalidContentData
	}
	log.Info(ctx, "getContentProperties",
		log.Any("contentProperties", contentProperties),
		log.String("cid", cid))

	contentProperties.ContentID = id
	err = cm.doCreateContentProperties(ctx, tx, *contentProperties, false)
	if err != nil {
		log.Warn(ctx, "doCreateContentProperties failed",
			log.Err(err),
			log.Any("contentProperties", contentProperties),
			log.String("cid", cid))
		return "", err
	}

	contentVisibilitySettings, err := cm.getContentVisibilitySettings(ctx, cid)
	if err != nil {
		log.Warn(ctx, "getContentVisibilitySettings failed",
			log.Err(err), log.String("cid", cid))
		return "", ErrInvalidContentData
	}

	log.Info(ctx, "getContentVisibilitySettings",
		log.Any("contentVisibilitySettings", contentVisibilitySettings),
		log.String("cid", cid))

	err = cm.insertContentVisibilitySettings(ctx, tx, id, contentVisibilitySettings.VisibilitySettings)
	if err != nil {
		log.Warn(ctx, "insertContentVisibilitySettings failed",
			log.Err(err), log.String("cid", cid),
			log.Strings("VisibilitySettings", contentVisibilitySettings.VisibilitySettings))
		return "", ErrInvalidContentData
	}
	da.GetContentRedis().CleanContentCache(ctx, []string{id, obj.ID})
	return id, nil
}

func (cm *ContentModel) CheckContentAuthorization(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	//if user.UserID == content.Author {
	//	return nil
	//}

	if content.PublishStatus == entity.ContentStatusAttachment ||
		content.PublishStatus == entity.ContentStatusHidden {
		return nil
	}

	if content.PublishStatus != entity.ContentStatusPublished {
		log.Error(ctx, "read unpublished content, userID: %v, contentID: %v", log.String("userID", user.UserID), log.String("contentID", content.ID))
		return ErrGetUnpublishedContent
	}
	return nil
}

func (cm *ContentModel) GetContentsSubContentsMapByIDList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) (map[string][]*entity.SubContentsWithName, error) {
	objs, err := da.GetContentDA().GetContentByIDList(ctx, tx, cids)
	if err != nil {
		log.Error(ctx, "can't read content", log.Err(err), log.Strings("cids", cids))
		return nil, err
	}
	contentInfoMap := make(map[string][]*entity.SubContentsWithName)
	for _, obj := range objs {
		cd, err := cm.CreateContentData(ctx, obj.ContentType, obj.Data)
		if err != nil {
			log.Error(ctx, "can't unmarshal contentdata", log.Err(err), log.Any("content", obj))
			return nil, err
		}

		switch v := cd.(type) {
		case *LessonData:
			//存在子内容，则返回子内容
			//if the content contains sub contents, return sub contents
			content, err := cm.ConvertContentObj(ctx, obj, user)
			if err != nil {
				log.Error(ctx, "can't parse contentdata", log.Err(err))
				return nil, ErrParseContentDataFailed
			}
			err = v.PrepareVersion(ctx)
			if err != nil {
				log.Error(ctx, "can't prepare version for sub contents", log.Err(err), log.Any("content", content))
				return nil, err
			}
			err = v.PrepareResult(ctx, content, user)
			if err != nil {
				log.Error(ctx, "can't get sub contents", log.Err(err), log.Any("content", content))
				return nil, err
			}
			ret := make([]*entity.SubContentsWithName, 0)
			v.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
				if l.Material != nil {
					cd0, err := cm.CreateContentData(ctx, l.Material.ContentType, l.Material.Data)
					if err != nil {
						log.Error(ctx, "can't parse sub content data",
							log.Err(err),
							log.Any("lesson", l),
							log.Any("subContent", l.Material))
						return
					}
					ret = append(ret, &entity.SubContentsWithName{
						ID:   l.Material.ID,
						Name: l.Material.Name,
						Data: cd0,
					})
				}
			})
			contentInfoMap[obj.ID] = ret
		case *MaterialData:
			//若不存在子内容，则返回当前内容
			//if sub contents is not exists, return current content
			ret := []*entity.SubContentsWithName{
				{
					ID:   obj.ID,
					Name: obj.Name,
					Data: v,
				},
			}
			contentInfoMap[obj.ID] = ret
		case *AssetsData:
			//若不存在子内容，则返回当前内容
			//if sub contents is not exists, return current content
			ret := []*entity.SubContentsWithName{
				{
					ID:   obj.ID,
					Name: obj.Name,
					Data: v,
				},
			}
			contentInfoMap[obj.ID] = ret
		}
	}

	return contentInfoMap, nil
}

func (cm *ContentModel) GetContentSubContentsByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) ([]*entity.SubContentsWithName, error) {
	obj, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't read content", log.Err(err), log.String("cid", cid))
		return nil, err
	}
	//获取最新数据
	//fetch newest data
	if obj.LatestID != "" {
		obj, err = da.GetContentDA().GetContentByID(ctx, tx, obj.LatestID)
		if err != nil {
			log.Error(ctx, "can't read content", log.Err(err), log.String("cid", cid))
			return nil, err
		}
	}

	cd, err := cm.CreateContentData(ctx, obj.ContentType, obj.Data)
	if err != nil {
		log.Error(ctx, "can't unmarshal contentdata", log.Err(err), log.Any("content", obj))
		return nil, err
	}

	switch v := cd.(type) {
	case *LessonData:
		//存在子内容，则返回子内容
		//if the content contains sub contents, return sub contents
		content, err := cm.ConvertContentObj(ctx, obj, user)
		if err != nil {
			log.Error(ctx, "can't parse contentdata", log.Err(err))
			return nil, ErrParseContentDataFailed
		}
		err = v.PrepareVersion(ctx)
		if err != nil {
			log.Error(ctx, "can't prepare version for sub contents", log.Err(err), log.Any("content", content))
			return nil, err
		}
		err = v.PrepareResult(ctx, content, user)
		if err != nil {
			log.Error(ctx, "can't get sub contents", log.Err(err), log.Any("content", content))
			return nil, err
		}
		ret := make([]*entity.SubContentsWithName, 0)
		v.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
			if l.Material != nil {
				cd0, err := cm.CreateContentData(ctx, l.Material.ContentType, l.Material.Data)
				if err != nil {
					log.Error(ctx, "can't parse sub content data",
						log.Err(err),
						log.Any("lesson", l),
						log.Any("subContent", l.Material))
					return
				}
				ret = append(ret, &entity.SubContentsWithName{
					ID:   l.Material.ID,
					Name: l.Material.Name,
					Data: cd0,
				})
			}
		})
		return ret, nil
	case *MaterialData:
		//若不存在子内容，则返回当前内容
		//if sub contents is not exists, return current content
		ret := []*entity.SubContentsWithName{
			{
				ID:   cid,
				Name: obj.Name,
				Data: v,
			},
		}
		return ret, nil
	case *AssetsData:
		//若不存在子内容，则返回当前内容
		//if sub contents is not exists, return current content
		ret := []*entity.SubContentsWithName{
			{
				ID:   cid,
				Name: obj.Name,
				Data: v,
			},
		}
		return ret, nil
	}

	return nil, ErrInvalidContentData
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

func (cm *ContentModel) GetRawContentByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.Content, error) {
	obj, err := da.GetContentDA().GetContentByIDList(ctx, tx, cids)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.Strings("cids", cids))
		return nil, ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Err(err), log.Strings("cids", cids))
		return nil, err
	}
	return obj, nil
}

func (cm *ContentModel) GetContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error) {
	content := da.GetContentRedis().GetContentCacheByID(ctx, cid)
	if content == nil {
		log.Info(ctx, "Not cached content by id")
		obj, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
		if err == dbo.ErrRecordNotFound {
			log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
			return nil, ErrNoContent
		}
		if err != nil {
			log.Error(ctx, "can't read contentdata", log.Err(err))
			return nil, err
		}
		content, err = cm.ConvertContentObj(ctx, obj, user)
		if err != nil {
			log.Error(ctx, "can't parse contentdata", log.Err(err))
			return nil, ErrParseContentDataFailed
		}
		da.GetContentRedis().SaveContentCache(ctx, content)
	}
	log.Info(ctx, "pre fill content details", log.Any("content", content))
	return cm.fillContentDetails(ctx, content, user)
}

func (cm *ContentModel) fillContentDetails(ctx context.Context, content *entity.ContentInfo, user *entity.Operator) (*entity.ContentInfoWithDetails, error) {
	//补全相关内容
	//fill related data
	contentData, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		return nil, err
	}

	//TODO:For authed content => prepare result for latest version content => done
	err = contentData.PrepareVersion(ctx)
	if err != nil {
		log.Error(ctx, "can't update contentdata version for details", log.Err(err))
		return nil, ErrParseContentDataDetailsFailed
	}
	err = contentData.PrepareResult(ctx, content, user)
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

	contentWithDetails, err := cm.buildContentWithDetails(ctx, []*entity.ContentInfo{content}, true, user)
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

	nid, cachedContent := da.GetContentRedis().GetContentCacheByIDList(ctx, cids)
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

	data, err := da.GetContentDA().GetContentByIDList(ctx, tx, nid)
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

func (cm *ContentModel) GetPastContentIDByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error) {
	data, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't get content", log.Err(err), log.String("cid", cid))
		return nil, ErrReadContentFailed
	}

	latestID := data.LatestID
	if data.LatestID == "" {
		latestID = data.ID
	}

	_, res, err := da.GetContentDA().SearchContent(ctx, tx, da.ContentCondition{
		LatestID: latestID,
	})
	if err != nil {
		log.Error(ctx, "can't search content", log.Err(err), log.String("latest id", data.LatestID))
		return nil, ErrReadContentFailed
	}
	resp := make([]string, len(res)+1)
	for i := range res {
		resp[i] = res[i].ID
	}
	resp[len(res)] = data.ID
	return utils.SliceDeduplication(resp), nil
}

func (cm *ContentModel) GetLatestContentIDByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]string, error) {
	if len(cids) < 1 {
		return nil, nil
	}
	resp := make([]string, len(cids))
	data, err := da.GetContentDA().GetContentByIDList(ctx, tx, cids)
	if err != nil {
		log.Error(ctx, "can't search content", log.Err(err), log.Strings("cids", cids))
		return nil, ErrReadContentFailed
	}
	for i := range data {
		if data[i].LatestID != "" {
			resp[i] = data[i].LatestID
		} else {
			resp[i] = data[i].ID
		}
	}
	return resp, nil
}

func (cm *ContentModel) IsContentsOperatorByIDList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) (bool, error) {
	data, err := da.GetContentDA().GetContentByIDList(ctx, tx, cids)
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Strings("cids", cids), log.Err(err))
		return false, ErrReadContentFailed
	}
	for i := range data {
		if data[i].Author != user.UserID {
			return false, nil
		}
	}
	return true, nil
}

func (cm *ContentModel) GetContentByIDList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) ([]*entity.ContentInfoWithDetails, error) {
	if len(cids) < 1 {
		return nil, nil
	}

	nid, cachedContent := da.GetContentRedis().GetContentCacheByIDList(ctx, cids)
	//全在缓存中
	//all cached
	if len(nid) < 1 {
		contentWithDetails, err := cm.buildContentWithDetails(ctx, cachedContent, true, user)
		if err != nil {
			log.Error(ctx, "can't parse contentdata", log.Strings("cids", cids), log.Err(err))
			return nil, ErrReadContentFailed
		}
		return contentWithDetails, nil
	}

	data, err := da.GetContentDA().GetContentByIDList(ctx, tx, nid)
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Strings("cids", cids), log.Err(err))
		return nil, ErrReadContentFailed
	}
	res, err := cm.BatchConvertContentObj(ctx, tx, data, user)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Strings("cids", cids), log.Any("data", data), log.Err(err))
		return nil, ErrReadContentFailed
	}
	res = append(res, cachedContent...)
	contentWithDetails, err := cm.buildContentWithDetails(ctx, res, true, user)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Strings("cids", cids), log.Err(err))
		return nil, ErrReadContentFailed
	}

	da.GetContentRedis().SaveContentCacheList(ctx, res)

	return contentWithDetails, nil
}

func (cm *ContentModel) SearchUserPrivateFolderContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.FolderContentData, error) {
	//构造个人查询条件
	//construct query condition for private search
	condition.Author = user.UserID
	condition.PublishStatus = cm.filterInvisiblePublishStatus(ctx, condition.PublishStatus)
	scope, err := cm.listAllScopes(ctx, user)
	if err != nil {
		return 0, nil, err
	}
	if len(scope) == 0 {
		log.Info(ctx, "no valid scope", log.Strings("scopes", scope), log.Any("user", user))
		scope = []string{constant.NoSearchItem}
	}
	err = cm.filterRootPath(ctx, &condition, entity.OwnerTypeOrganization, user)
	if err != nil {
		return 0, nil, err
	}
	condition.VisibilitySettings = scope
	searchUserIDs := cm.getRelatedUserID(ctx, condition.Name, user)
	condition.JoinUserIDList = searchUserIDs
	//cm.addUserCondition(ctx, &condition, user)

	//生成folder condition
	folderCondition := cm.buildFolderCondition(ctx, condition, searchUserIDs, user)

	log.Info(ctx, "search folder content", log.Any("condition", condition), log.Any("folderCondition", folderCondition), log.String("uid", user.UserID))
	count, objs, err := da.GetContentDA().SearchFolderContent(ctx, tx, condition, *folderCondition)
	if err != nil {
		log.Error(ctx, "can't read folder content", log.Err(err), log.Any("condition", condition), log.Any("folderCondition", folderCondition), log.String("uid", user.UserID))
		return 0, nil, ErrReadContentFailed
	}
	ret := cm.convertFolderContent(ctx, objs, user)
	cm.fillFolderContent(ctx, ret, user)
	return count, ret, nil
}

func (cm *ContentModel) CountUserFolderContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, error) {
	searchUserIDs := cm.getRelatedUserID(ctx, condition.Name, user)
	err := cm.filterRootPath(ctx, &condition, entity.OwnerTypeOrganization, user)
	if err != nil {
		log.Warn(ctx, "filterRootPath failed", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, err
	}
	combineCondition, err := cm.buildUserContentCondition(ctx, tx, condition, searchUserIDs, user)
	if err != nil {
		log.Warn(ctx, "buildUserContentCondition failed", log.Err(err), log.Any("condition", condition), log.Any("searchUserIDs", searchUserIDs), log.String("uid", user.UserID))
		return 0, err
	}
	folderCondition := cm.buildFolderCondition(ctx, condition, searchUserIDs, user)

	log.Info(ctx, "count folder content", log.Any("combineCondition", combineCondition), log.Any("folderCondition", folderCondition), log.String("uid", user.UserID))
	total, err := da.GetContentDA().CountFolderContentUnsafe(ctx, tx, *combineCondition, *folderCondition)
	if err != nil {
		log.Error(ctx, "can't read folder content", log.Err(err), log.Any("combineCondition", combineCondition), log.Any("folderCondition", folderCondition), log.String("uid", user.UserID))
		return 0, ErrReadContentFailed
	}
	return total, nil
}
func (cm *ContentModel) SearchUserFolderContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.FolderContentData, error) {
	searchUserIDs := cm.getRelatedUserID(ctx, condition.Name, user)
	err := cm.filterRootPath(ctx, &condition, entity.OwnerTypeOrganization, user)
	if err != nil {
		log.Warn(ctx, "filterRootPath failed", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, err
	}
	combineCondition, err := cm.buildUserContentCondition(ctx, tx, condition, searchUserIDs, user)
	if err != nil {
		log.Warn(ctx, "buildUserContentCondition failed", log.Err(err), log.Any("condition", condition), log.Any("searchUserIDs", searchUserIDs), log.String("uid", user.UserID))
		return 0, nil, err
	}
	folderCondition := cm.buildFolderCondition(ctx, condition, searchUserIDs, user)

	log.Info(ctx, "search folder content", log.Any("combineCondition", combineCondition), log.Any("folderCondition", folderCondition), log.String("uid", user.UserID))
	count, objs, err := da.GetContentDA().SearchFolderContentUnsafe(ctx, tx, *combineCondition, *folderCondition)
	if err != nil {
		log.Error(ctx, "can't read folder content", log.Err(err), log.Any("combineCondition", combineCondition), log.Any("folderCondition", folderCondition), log.String("uid", user.UserID))
		return 0, nil, ErrReadContentFailed
	}
	ret := cm.convertFolderContent(ctx, objs, user)
	cm.fillFolderContent(ctx, ret, user)

	return count, ret, nil
}
func (cm *ContentModel) SearchUserContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	//where, params := combineCondition.GetConditions()
	//logger.WithContext(ctx).WithField("subject", "content").Infof("Combine condition: %#v, params: %#v", where, params)
	searchUserIDs := cm.getRelatedUserID(ctx, condition.Name, user)
	combineCondition, err := cm.buildUserContentCondition(ctx, tx, condition, searchUserIDs, user)
	if err != nil {
		return 0, nil, err
	}
	return cm.searchContentUnsafe(ctx, tx, combineCondition, user)
}

func (cm *ContentModel) SearchUserPrivateContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	condition.Author = user.UserID
	condition.PublishStatus = cm.filterInvisiblePublishStatus(ctx, condition.PublishStatus)
	scope, err := cm.listAllScopes(ctx, user)
	if err != nil {
		return 0, nil, err
	}
	if len(scope) == 0 {
		log.Info(ctx, "no valid scope", log.Strings("scopes", scope), log.Any("user", user))
		scope = []string{constant.NoSearchItem}
	}
	condition.VisibilitySettings = scope

	cm.addUserCondition(ctx, &condition, user)
	return cm.searchContent(ctx, tx, &condition, user)
}

func (cm *ContentModel) ListPendingContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	condition.PublishStatus = []string{entity.ContentStatusPending}
	scope, err := cm.ListVisibleScopes(ctx, visiblePermissionPending, user)
	if err != nil {
		return 0, nil, err
	}
	if len(scope) == 0 {
		log.Info(ctx, "no valid private scope", log.Strings("scopes", scope), log.Any("user", user))
		scope = []string{constant.NoSearchItem}
	}
	condition.VisibilitySettings = scope

	cm.addUserCondition(ctx, &condition, user)
	return cm.searchContent(ctx, tx, &condition, user)
}

func (cm *ContentModel) addUserCondition(ctx context.Context, condition *da.ContentCondition, user *entity.Operator) {
	condition.JoinUserIDList = cm.getRelatedUserID(ctx, condition.Name, user)
}

func (cm *ContentModel) getRelatedUserID(ctx context.Context, keyword string, user *entity.Operator) []string {
	if keyword == "" {
		return nil
	}
	users, err := external.GetUserServiceProvider().Query(ctx, user, user.OrgID, keyword)
	if err != nil {
		log.Warn(ctx, "get user info failed", log.Err(err), log.String("keyword", keyword), log.Any("user", user))
		return nil
	}
	if len(users) < 1 {
		log.Info(ctx, "user info not found in keywords", log.Err(err), log.String("keyword", keyword), log.String("userID", user.UserID), log.String("orgID", user.OrgID))
		return nil
	}
	ids := make([]string, len(users))
	for i := range users {
		ids[i] = users[i].ID
	}
	return ids
}

func (cm *ContentModel) SearchContent(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	condition.PublishStatus = cm.filterInvisiblePublishStatus(ctx, condition.PublishStatus)
	cm.addUserCondition(ctx, &condition, user)
	return cm.searchContent(ctx, tx, &condition, user)
}

func (cm *ContentModel) refreshContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, cid string, scope []string) error {
	alreadyScopes, err := da.GetContentDA().GetContentVisibilitySettings(ctx, tx, cid)
	if err != nil {
		log.Error(ctx,
			"GetContentVisibilitySettings failed",
			log.Err(err),
			log.String("cid", cid),
			log.Strings("scope", scope))
		return err
	}
	pendingAddScopes := cm.checkDiff(ctx, alreadyScopes, scope)
	pendingDeleteScopes := cm.checkDiff(ctx, scope, alreadyScopes)

	err = da.GetContentDA().BatchCreateContentVisibilitySettings(ctx, tx, cid, pendingAddScopes)
	if err != nil {
		log.Error(ctx,
			"BatchCreateContentVisibilitySettings failed",
			log.Err(err),
			log.String("cid", cid),
			log.Strings("alreadyScopes", alreadyScopes),
			log.Strings("scope", scope),
			log.Strings("pendingAddScopes", pendingAddScopes))
		return err
	}
	err = da.GetContentDA().BatchDeleteContentVisibilitySettings(ctx, tx, cid, pendingDeleteScopes)
	if err != nil {
		log.Error(ctx,
			"BatchDeleteContentVisibilitySettings failed",
			log.Err(err),
			log.String("cid", cid),
			log.Strings("alreadyScopes", alreadyScopes),
			log.Strings("scope", scope),
			log.Strings("pendingDeleteScopes", pendingDeleteScopes))
		return err
	}
	return nil
}

func (cm *ContentModel) checkDiff(ctx context.Context, base []string, compare []string) []string {
	sourceMap := make(map[string]bool)
	ret := make([]string, 0)
	for i := range base {
		sourceMap[base[i]] = true
	}
	for i := range compare {
		_, exist := sourceMap[compare[i]]
		if !exist {
			ret = append(ret, compare[i])
		}
	}
	return ret
}

func (cm *ContentModel) insertContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, cid string, scope []string) error {
	//Insert visibility settings
	err := da.GetContentDA().BatchCreateContentVisibilitySettings(ctx, tx, cid, scope)
	if err != nil {
		log.Error(ctx,
			"BatchCreateContentVisibilitySettings failed",
			log.Err(err),
			log.Strings("scope", scope),
			log.String("cid", cid))
		return err
	}
	return nil
}

func (cm *ContentModel) getVisibleContentOutcomeByIDs(ctx context.Context, tx *dbo.DBContext, cids []string) ([]string, error) {
	//get latest content ids
	newCids, err := cm.GetLatestContentIDByIDList(ctx, tx, cids)
	if err != nil {
		log.Error(ctx, "can't get content latest id", log.Err(err), log.Strings("cids", cids))
		return nil, err
	}
	//get content objects
	contents, err := da.GetContentDA().GetContentByIDList(ctx, tx, newCids)
	if err != nil {
		log.Error(ctx, "can't read content", log.Err(err), log.Strings("newCids", newCids))
		return nil, err
	}

	//collect outcomes
	ret := make([]string, 0)
	for i := range contents {
		if contents[i].Outcomes == "" {
			continue
		}
		outcomes := strings.Split(contents[i].Outcomes, constant.StringArraySeparator)
		for i := range outcomes {
			if outcomes[i] != "" {
				ret = append(ret, outcomes[i])
			}
		}
	}
	ret = utils.SliceDeduplication(ret)
	return ret, nil
}

func (cm *ContentModel) GetVisibleContentOutcomeByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error) {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't get content", log.Err(err), log.String("cid", cid))
		return nil, err
	}
	if content.LatestID != "" {
		content, err = da.GetContentDA().GetContentByID(ctx, tx, content.LatestID)
		if err != nil {
			log.Error(ctx, "can't get latest content", log.Err(err), log.String("cid", cid))
			return nil, err
		}
	}
	ret := make([]string, 0)
	//if content is a plan, collect outcomes from materials
	if content.ContentType == entity.ContentTypePlan {
		contentData, err := cm.CreateContentData(ctx, entity.ContentTypePlan, content.Data)
		if err != nil {
			log.Error(ctx, "can't parse content data",
				log.Err(err),
				log.String("cid", cid),
				log.String("data", content.Data))
			return nil, err
		}
		contentIDList := contentData.(*LessonData).SubContentIDs(ctx)
		outcomes, err := cm.getVisibleContentOutcomeByIDs(ctx, tx, contentIDList)
		if err != nil {
			log.Error(ctx, "can't get outcomes from materials",
				log.Err(err),
				log.Strings("contentIDList", contentIDList))
			return nil, err
		}
		ret = append(ret, outcomes...)
	}

	if content.Outcomes == "" {
		return ret, nil
	}
	outcomes := strings.Split(content.Outcomes, constant.StringArraySeparator)
	for i := range outcomes {
		if outcomes[i] != "" {
			ret = append(ret, outcomes[i])
		}
	}
	ret = utils.SliceDeduplication(ret)

	return ret, nil
}

func (cm *ContentModel) GetContentOutcomeByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error) {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't get content", log.Err(err), log.String("cid", cid))
		return nil, err
	}
	ret := cm.parseContentOutcomes(ctx, content)
	//if content is a lesson, add materials outcomes
	if content.ContentType == entity.ContentTypePlan {
		data, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
		if err != nil {
			log.Error(ctx, "parse content data failed",
				log.Err(err),
				log.Any("content", content))
			return nil, err
		}
		materialIDs := data.SubContentIDs(ctx)
		if len(materialIDs) > 0 {
			materials, err := da.GetContentDA().GetContentByIDList(ctx, tx, materialIDs)
			if err != nil {
				log.Error(ctx, "parse content data failed",
					log.Err(err),
					log.Any("content", content))
				return nil, err
			}

			for i := range materials {
				outcomes := cm.parseContentOutcomes(ctx, materials[i])
				if len(outcomes) > 0 {
					ret = append(ret, outcomes...)
				}
			}
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
	if content.ContentType != entity.ContentTypePlan {
		log.Error(ctx, "invalid content type", log.Err(err), log.String("cid", cid), log.Int("contentType", int(content.ContentType)))
		return nil, ErrInvalidContentType
	}

	cd, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		log.Error(ctx, "can't parse content data", log.Err(err), log.String("cid", cid), log.Int("contentType", int(content.ContentType)), log.String("data", content.Data))
		return nil, err
	}
	subContentIDs := cd.SubContentIDs(ctx)
	_, subContents, err := da.GetContentDA().SearchContent(ctx, tx, da.ContentCondition{
		IDS: subContentIDs,
	})
	if err != nil {
		log.Error(ctx, "search data failed", log.Err(err), log.String("cid", cid),
			log.Int("contentType", int(content.ContentType)),
			log.String("data", content.Data),
			log.Strings("subContentIDs", subContentIDs))
		return nil, err
	}

	identityOutComes := make(map[string]bool)
	outcomesCount := 0
	for i := range subContents {
		if subContents[i].Outcomes == "" {
			continue
		}
		subOutcomes := strings.Split(subContents[i].Outcomes, constant.StringArraySeparator)
		for j := range subOutcomes {
			_, ok := identityOutComes[subOutcomes[j]]
			if !ok {
				outcomesCount++
			}
			identityOutComes[subOutcomes[j]] = true
		}
	}
	return &entity.ContentStatisticsInfo{
		SubContentCount: len(subContentIDs),
		OutcomesCount:   outcomesCount,
	}, nil
}

func (cm *ContentModel) GetVisibleContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error) {
	var err error
	log.Info(ctx, "GetVisibleContentByID")
	cachedContent := da.GetContentRedis().GetContentCacheByID(ctx, cid)
	if cachedContent != nil {
		if cachedContent.LatestID == "" {
			log.Info(ctx, "Cached latest content")
			return cm.fillContentDetails(ctx, cachedContent, user)
		} else {
			log.Info(ctx, "Cached content not latest")
			return cm.GetContentByID(ctx, tx, cachedContent.LatestID, user)
		}
	}
	log.Info(ctx, "Not cached latest content")

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

func (cm *ContentModel) parseContentOutcomes(ctx context.Context, content *entity.Content) []string {
	if content.Outcomes == "" {
		return nil
	}
	outcomes := strings.Split(content.Outcomes, constant.StringArraySeparator)
	ret := make([]string, 0)
	for i := range outcomes {
		if outcomes[i] != "" {
			ret = append(ret, outcomes[i])
		}
	}
	return ret
}

func (cm *ContentModel) filterRootPath(ctx context.Context, condition *da.ContentCondition, ownerType entity.OwnerType, operator *entity.Operator) error {
	if condition.DirPath != "" {
		return nil
	}

	//isAssets := false
	//for i := range condition.ContentType {
	//	if entity.NewContentType(condition.ContentType[i]).IsAsset() {
	//		isAssets = true
	//		break
	//	}
	//}
	//
	//if isAssets {
	//	root, err := GetFolderModel().GetRootFolder(ctx, entity.RootAssetsFolderName, ownerType, operator)
	//	if err != nil{
	//		log.Error(ctx, "can't get root folder", log.Err(err), log.Any("ownerType", ownerType), log.Any("operator", operator), log.String("partition", string(entity.RootAssetsFolderName)))
	//		return err
	//	}
	//	condition.DirPath = string(root.ChildrenPath())
	//	return nil
	//}
	//
	//root, err := GetFolderModel().GetRootFolder(ctx, entity.RootMaterialsAndPlansFolderName, ownerType, operator)
	//if err != nil{
	//	log.Error(ctx, "can't get root folder", log.Err(err), log.Any("ownerType", ownerType), log.Any("operator", operator), log.String("partition", string(entity.RootMaterialsAndPlansFolderName)))
	//	return err
	//}
	//condition.DirPath = string(root.ChildrenPath())
	condition.DirPath = constant.FolderRootPath
	return nil
}

func (cm *ContentModel) filterPublishedPublishStatus(ctx context.Context, status []string) []string {
	newStatus := make([]string, 0)
	for i := range status {
		if status[i] == entity.ContentStatusPublished ||
			status[i] == entity.ContentStatusArchive {
			newStatus = append(newStatus, status[i])
		}
	}
	if len(newStatus) < 1 {
		return []string{
			entity.ContentStatusPublished,
		}
	}
	return newStatus
}

func (cm *ContentModel) buildUserContentCondition(ctx context.Context, tx *dbo.DBContext, condition da.ContentCondition, searchUserIDs []string, user *entity.Operator) (*da.CombineConditions, error) {
	condition1 := condition
	condition2 := condition

	//condition1 private
	condition1.Author = user.UserID
	condition1.PublishStatus = cm.filterInvisiblePublishStatus(ctx, condition1.PublishStatus)

	scope, err := cm.listAllScopes(ctx, user)
	if err != nil {
		return nil, err
	}
	if len(scope) == 0 {
		log.Info(ctx, "no valid private scope", log.Strings("scopes", scope), log.Any("user", user))
		scope = []string{constant.NoSearchItem}
	}
	condition1.VisibilitySettings = scope
	//condition2 others

	condition2.PublishStatus = cm.filterPublishedPublishStatus(ctx, condition2.PublishStatus)

	//filter visible
	if len(condition.ContentType) == 1 && condition.ContentType[0] == entity.ContentTypeAssets {
		condition2.VisibilitySettings = []string{user.OrgID}
	} else {
		scopes, err := cm.ListVisibleScopes(ctx, visiblePermissionPublished, user)
		if err != nil {
			return nil, err
		}
		if len(scopes) == 0 {
			log.Info(ctx, "no valid scope", log.Strings("scopes", scopes), log.Any("user", user))
			scopes = []string{constant.NoSearchItem}
		}
		condition2.VisibilitySettings = scopes
	}

	//condition2.Scope = scopes
	condition1.JoinUserIDList = searchUserIDs
	condition2.JoinUserIDList = searchUserIDs
	combineCondition := &da.CombineConditions{
		SourceCondition: &condition1,
		TargetCondition: &condition2,
	}
	return combineCondition, nil
}

func (cm *ContentModel) checkPublishContentChildren(ctx context.Context, c *entity.Content, children []*entity.Content) error {
	//TODO: To implement, check publish scope
	for i := range children {
		if children[i].PublishStatus != entity.ContentStatusPublished &&
			children[i].PublishStatus != entity.ContentStatusHidden {
			log.Warn(ctx, "check children status failed", log.Any("content", children[i]))
			return ErrInvalidPublishStatus
		}
	}
	//TODO:For authed content => update check for authed content list

	return nil
}

func (cm *ContentModel) buildVisibilitySettingsMap(ctx context.Context, contentList []*entity.ContentInfo) (map[string][]string, error) {
	contentIDs := make([]string, len(contentList))
	for i := range contentList {
		contentIDs[i] = contentList[i].ID
	}

	visibilitySettings, err := da.GetContentDA().SearchContentVisibilitySettings(ctx, dbo.MustGetDB(ctx), &da.ContentVisibilitySettingsCondition{
		ContentIDs: contentIDs,
	})
	if err != nil {
		log.Error(ctx, "GetContentVisibilitySettings failed",
			log.Err(err),
			log.Any("contentList", contentList))
		return nil, err
	}

	ret := make(map[string][]string)
	for i := range visibilitySettings {
		ret[visibilitySettings[i].ContentID] = append(ret[visibilitySettings[i].ContentID], visibilitySettings[i].VisibilitySetting)
	}
	return ret, nil
}

func (cm *ContentModel) doCreateContentProperties(ctx context.Context, tx *dbo.DBContext, c entity.ContentProperties, refresh bool) error {
	propertyDA := da.GetContentPropertyDA()

	if refresh {
		//delete old properties
		err := propertyDA.CleanByContentID(ctx, tx, c.ContentID)
		if err != nil {
			log.Error(ctx, "CleanByContentID failed",
				log.Err(err),
				log.String("ContentID", c.ContentID),
				log.Any("data", c))
			return err
		}
	}
	size := len(c.Grade) + len(c.Subject) + len(c.Age) + len(c.Category) + len(c.SubCategory)
	properties := make([]*entity.ContentProperty, size+1)
	index := 0
	for i := range c.Grade {
		properties[index] = &entity.ContentProperty{
			PropertyType: entity.ContentPropertyTypeGrade,
			ContentID:    c.ContentID,
			PropertyID:   c.Grade[i],
			Sequence:     i,
		}
		index++
	}
	for i := range c.Subject {
		properties[index] = &entity.ContentProperty{
			PropertyType: entity.ContentPropertyTypeSubject,
			ContentID:    c.ContentID,
			PropertyID:   c.Subject[i],
			Sequence:     i,
		}
		index++
	}
	for i := range c.Age {
		properties[index] = &entity.ContentProperty{
			PropertyType: entity.ContentPropertyTypeAge,
			ContentID:    c.ContentID,
			PropertyID:   c.Age[i],
			Sequence:     i,
		}
		index++
	}
	for i := range c.Category {
		properties[index] = &entity.ContentProperty{
			PropertyType: entity.ContentPropertyTypeCategory,
			ContentID:    c.ContentID,
			PropertyID:   c.Category[i],
			Sequence:     i,
		}
		index++
	}
	for i := range c.SubCategory {
		properties[index] = &entity.ContentProperty{
			PropertyType: entity.ContentPropertyTypeSubCategory,
			ContentID:    c.ContentID,
			PropertyID:   c.SubCategory[i],
			Sequence:     i,
		}
		index++
	}
	properties[index] = &entity.ContentProperty{
		PropertyType: entity.ContentPropertyTypeProgram,
		ContentID:    c.ContentID,
		PropertyID:   c.Program,
		Sequence:     0,
	}

	err := propertyDA.BatchAdd(ctx, tx, properties)
	if err != nil {
		log.Error(ctx, "BatchAdd failed",
			log.Err(err),
			log.String("ContentID", c.ContentID),
			log.Any("properties", properties),
			log.Any("data", c))
		return err
	}
	return nil
}

func (cm *ContentModel) copyContentProperties(ctx context.Context, tx *dbo.DBContext, cid string, newCid string) error {
	propertyDA := da.GetContentPropertyDA()

	contentProperties, err := propertyDA.BatchGetByContentIDList(ctx, tx, []string{cid})
	if err != nil {
		log.Error(ctx, "BatchGetByContentIDList failed",
			log.Err(err),
			log.String("cid", cid),
			log.String("newCid", newCid))
		return err
	}

	for i := range contentProperties {
		contentProperties[i].ContentID = newCid
	}
	err = propertyDA.BatchAdd(ctx, tx, contentProperties)
	if err != nil {
		log.Error(ctx, "BatchAdd failed",
			log.Err(err),
			log.String("ContentID", newCid),
			log.Any("contentProperties", contentProperties))
		return err
	}
	return nil
}
func (cm *ContentModel) buildContentWithDetails(ctx context.Context, contentList []*entity.ContentInfo, outComes bool, user *entity.Operator) ([]*entity.ContentInfoWithDetails, error) {
	orgName := ""
	orgProvider := external.GetOrganizationServiceProvider()
	orgs, err := orgProvider.BatchGet(ctx, user, []string{user.OrgID})
	if err != nil || len(orgs) < 1 {
		log.Error(ctx, "can't get org info", log.Err(err))
	} else {
		if orgs[0].Valid {
			orgName = orgs[0].Name
		} else {
			log.Warn(ctx, "invalid value", log.String("org_id", user.OrgID))
		}
	}

	programNameMap := make(map[string]string)
	subjectNameMap := make(map[string]string)
	developmentalNameMap := make(map[string]string)
	skillsNameMap := make(map[string]string)
	ageNameMap := make(map[string]string)
	gradeNameMap := make(map[string]string)
	publishScopeNameMap := make(map[string]string)
	lessonTypeNameMap := make(map[string]string)
	userNameMap := make(map[string]string)

	programIDs := make([]string, 0)
	subjectIDs := make([]string, 0)
	developmentalIDs := make([]string, 0)
	skillsIDs := make([]string, 0)
	ageIDs := make([]string, 0)
	gradeIDs := make([]string, 0)
	scopeIDs := make([]string, 0)
	lessonTypeIDs := make([]string, 0)
	userIDs := make([]string, 0)

	visibilitySettingsMap, err := cm.buildVisibilitySettingsMap(ctx, contentList)
	if err != nil {
		log.Error(ctx, "buildVisibilitySettingsMap failed",
			log.Err(err),
			log.Any("contentList", contentList))
		return nil, err
	}

	for i := range contentList {
		if contentList[i].Program != "" {
			programIDs = append(programIDs, contentList[i].Program)
		}
		subjectIDs = append(subjectIDs, contentList[i].Subject...)
		developmentalIDs = append(developmentalIDs, contentList[i].Category...)
		skillsIDs = append(skillsIDs, contentList[i].SubCategory...)
		ageIDs = append(ageIDs, contentList[i].Age...)
		gradeIDs = append(gradeIDs, contentList[i].Grade...)

		scopeIDs = append(scopeIDs, visibilitySettingsMap[contentList[i].ID]...)
		lessonTypeIDs = append(lessonTypeIDs, contentList[i].LessonType)
		userIDs = append(userIDs, contentList[i].Author)
		userIDs = append(userIDs, contentList[i].Creator)
	}

	//Users
	users, err := external.GetUserServiceProvider().BatchGet(ctx, user, userIDs)
	if err != nil {
		log.Error(ctx, "can't get user info", log.Err(err), log.Strings("ids", userIDs))
	} else {
		for i := range users {
			if !users[i].Valid {
				log.Warn(ctx, "user not exists, may be deleted", log.String("id", userIDs[i]))
				continue
			}

			userNameMap[users[i].ID] = users[i].Name
		}
	}

	//LessonType
	lessonTypes, err := GetLessonTypeModel().Query(ctx, &da.LessonTypeCondition{
		IDs: entity.NullStrings{
			Strings: lessonTypeIDs,
			Valid:   len(lessonTypeIDs) != 0,
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
	programs, err := external.GetProgramServiceProvider().BatchGet(ctx, user, programIDs)
	if err != nil {
		log.Error(ctx, "can't get programs", log.Err(err), log.Strings("ids", programIDs))
	} else {
		for i := range programs {
			programNameMap[programs[i].ID] = programs[i].Name
		}
	}

	//Subjects
	subjects, err := external.GetSubjectServiceProvider().BatchGet(ctx, user, subjectIDs)
	if err != nil {
		log.Error(ctx, "can't get subjects info", log.Err(err))
	} else {
		for i := range subjects {
			subjectNameMap[subjects[i].ID] = subjects[i].Name
		}
	}

	//developmental
	developmentals, err := external.GetCategoryServiceProvider().BatchGet(ctx, user, developmentalIDs)
	if err != nil {
		log.Error(ctx, "can't get developmentals info", log.Err(err))
	} else {
		for i := range developmentals {
			developmentalNameMap[developmentals[i].ID] = developmentals[i].Name
		}
	}

	//scope
	//TODO:change to get org name
	publishScopeNameList, err := external.GetOrganizationServiceProvider().GetNameByOrganizationOrSchool(ctx, user, scopeIDs)
	if err != nil {
		log.Error(ctx, "can't get publish scope info", log.Strings("scope", scopeIDs), log.Err(err))
	} else {
		for i := range scopeIDs {
			publishScopeNameMap[scopeIDs[i]] = publishScopeNameList[i]
		}
	}

	//skill
	skills, err := external.GetSubCategoryServiceProvider().BatchGet(ctx, user, skillsIDs)
	if err != nil {
		log.Error(ctx, "can't get skills info", log.Strings("skillsIDs", skillsIDs), log.Err(err))
	} else {
		for i := range skills {
			skillsNameMap[skills[i].ID] = skills[i].Name
		}
	}

	//age
	ages, err := external.GetAgeServiceProvider().BatchGet(ctx, user, ageIDs)
	if err != nil {
		log.Error(ctx, "can't get age info", log.Strings("ageIDs", ageIDs), log.Err(err))
	} else {
		for i := range ages {
			ageNameMap[ages[i].ID] = ages[i].Name
		}
	}

	//grade
	grades, err := external.GetGradeServiceProvider().BatchGet(ctx, user, gradeIDs)
	if err != nil {
		log.Error(ctx, "can't get grade info", log.Strings("gradeIDs", gradeIDs), log.Err(err))
	} else {
		for i := range grades {
			gradeNameMap[grades[i].ID] = grades[i].Name
		}
	}

	//Outcomes
	outcomeIDs := make([]string, 0)
	for i := range contentList {
		outcomeIDs = append(outcomeIDs, contentList[i].Outcomes...)
	}
	//outcomeEntities, err := GetOutcomeModel().GetLatestOutcomesByIDs(ctx, dbo.MustGetDB(ctx), outcomeIDs, user)
	//if err != nil {
	//	log.Error(ctx, "get latest outcomes entity failed", log.Err(err), log.Strings("outcome list", outcomeIDs), log.String("uid", user.UserID))
	//}
	//outcomeMaps := make(map[string]*entity.Outcome, len(outcomeEntities))
	//for i := range outcomeEntities {
	//	outcomeMaps[outcomeEntities[i].ID] = outcomeEntities[i]
	//}

	contentDetailsList := make([]*entity.ContentInfoWithDetails, len(contentList))
	for i := range contentList {
		subjectNames := make([]string, len(contentList[i].Subject))
		developmentalNames := make([]string, len(contentList[i].Category))
		skillsNames := make([]string, len(contentList[i].SubCategory))
		ageNames := make([]string, len(contentList[i].Age))
		gradeNames := make([]string, len(contentList[i].Grade))

		for j := range contentList[i].Subject {
			subjectNames[j] = subjectNameMap[contentList[i].Subject[j]]
		}
		for j := range contentList[i].Category {
			developmentalNames[j] = developmentalNameMap[contentList[i].Category[j]]
		}
		for j := range contentList[i].SubCategory {
			skillsNames[j] = skillsNameMap[contentList[i].SubCategory[j]]
		}
		for j := range contentList[i].Age {
			ageNames[j] = ageNameMap[contentList[i].Age[j]]
		}
		for j := range contentList[i].Grade {
			gradeNames[j] = gradeNameMap[contentList[i].Grade[j]]
		}

		outcomeEntities := make([]*entity.Outcome, 0)
		if outComes {
			outcomeEntities, err = GetOutcomeModel().GetLatestOutcomesByIDs(ctx, user, dbo.MustGetDB(ctx), contentList[i].Outcomes)
			if err != nil {
				log.Error(ctx, "get latest outcomes entity failed", log.Err(err), log.Strings("outcome list", contentList[i].Outcomes), log.String("uid", user.UserID))
			}
		}
		contentList[i].PublishScope = visibilitySettingsMap[contentList[i].ID]
		publishScopeNames := make([]string, len(contentList[i].PublishScope))
		log.Info(ctx, "get publish scope names",
			log.Strings("contentList[i].PublishScope", contentList[i].PublishScope),
			log.Any("visibilitySettingsMap", visibilitySettingsMap),
			log.String("contentList[i].ID", contentList[i].ID),
			log.Strings("visibilitySettingsMap[contentList[i].ID]", visibilitySettingsMap[contentList[i].ID]))
		for j := range contentList[i].PublishScope {
			publishScopeNames[j] = publishScopeNameMap[contentList[i].PublishScope[j]]
		}
		contentList[i].AuthorName = userNameMap[contentList[i].Author]
		contentDetailsList[i] = &entity.ContentInfoWithDetails{
			ContentInfo:      *contentList[i],
			ContentTypeName:  contentList[i].ContentType.Name(),
			ProgramName:      programNameMap[contentList[i].Program],
			SubjectName:      subjectNames,
			CategoryName:     developmentalNames,
			SubCategoryName:  skillsNames,
			AgeName:          ageNames,
			GradeName:        gradeNames,
			LessonTypeName:   lessonTypeNameMap[contentList[i].LessonType],
			PublishScopeName: publishScopeNames,
			OrgName:          orgName,
			//AuthorName:        userNameMap[contentList[i].Author],
			CreatorName:     userNameMap[contentList[i].Creator],
			OutcomeEntities: outcomeEntities,
			IsMine:          contentList[i].Author == user.UserID,
		}
	}

	return contentDetailsList, nil
}

func (cm *ContentModel) getOutcomes(ctx context.Context, pickIDs []string, user *entity.Operator) []*entity.Outcome {
	outcomeEntities, err := GetOutcomeModel().GetLatestOutcomesByIDs(ctx, user, dbo.MustGetDB(ctx), pickIDs)
	if err != nil {
		log.Error(ctx, "get latest outcomes entity failed", log.Err(err), log.Strings("outcome list", pickIDs), log.String("uid", user.UserID))
	}
	return outcomeEntities
}

func (cm *ContentModel) ListVisibleScopes(ctx context.Context, permission visiblePermission, operator *entity.Operator) ([]string, error) {
	//TODO:添加scope
	p := external.PublishedContentPage204
	if permission == visiblePermissionPending {
		p = external.PendingContentPage203
	}
	ret := make([]string, 0)

	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, p)
	if err != nil {
		log.Warn(ctx, "can't get schools from org", log.Err(err))
		return nil, err
	}
	if hasPermission {
		schools, err := external.GetSchoolServiceProvider().GetByOrganizationID(ctx, operator, operator.OrgID)
		if err != nil {
			log.Warn(ctx, "GetByOrganizationID failed", log.Err(err), log.Any("operator", operator))
			return nil, err
		}
		ret = append(ret, operator.OrgID)
		for i := range schools {
			ret = append(ret, schools[i].ID)
		}
		return ret, nil
	}

	schools, err := external.GetSchoolServiceProvider().GetByPermission(ctx, operator, p)
	if err != nil {
		log.Warn(ctx, "can't get schools from org", log.Err(err))
		return nil, err
	}
	for i := range schools {
		ret = append(ret, schools[i].ID)
	}

	if len(ret) == 0 {
		return ret, ErrInvalidVisibleScope
	}
	ret = append(ret, operator.OrgID)

	return ret, nil
}
func (cm *ContentModel) listAllScopes(ctx context.Context, operator *entity.Operator) ([]string, error) {
	schools, err := external.GetSchoolServiceProvider().GetByOrganizationID(ctx, operator, operator.OrgID)
	if err != nil {
		log.Warn(ctx, "can't get schools from org", log.Err(err))
		return nil, err
	}
	ret := []string{operator.OrgID}
	for i := range schools {
		ret = append(ret, schools[i].ID)
	}

	return ret, nil
}

func (cm *ContentModel) getContentVisibilitySettings(ctx context.Context, cid string) (*entity.ContentVisibilitySettings, error) {
	contentVisibilitySettings, err := da.GetContentDA().GetContentVisibilitySettings(ctx, dbo.MustGetDB(ctx), cid)
	if err != nil {
		log.Error(ctx, "GetContentVisibilitySettings",
			log.Err(err),
			log.String("id", cid))
		return nil, err
	}
	res := &entity.ContentVisibilitySettings{
		ContentID:          cid,
		VisibilitySettings: contentVisibilitySettings,
	}
	log.Info(ctx, "ContentVisibilitySettings",
		log.Strings("visibilitySettings", contentVisibilitySettings),
		log.String("id", cid))
	return res, nil
}
func (cm *ContentModel) getContentProperties(ctx context.Context, cid string) (*entity.ContentProperties, error) {
	contentProperties, err := da.GetContentPropertyDA().BatchGetByContentIDList(ctx, dbo.MustGetDB(ctx), []string{cid})
	if err != nil {
		log.Error(ctx, "BatchGetByContentIDList",
			log.Err(err),
			log.String("id", cid))
		return nil, err
	}

	subjects := make([]string, 0)
	categories := make([]string, 0)
	subCategories := make([]string, 0)
	ages := make([]string, 0)
	grades := make([]string, 0)
	program := ""

	for i := range contentProperties {
		switch contentProperties[i].PropertyType {
		case entity.ContentPropertyTypeProgram:
			program = contentProperties[i].PropertyID
		case entity.ContentPropertyTypeSubject:
			subjects = append(subjects, contentProperties[i].PropertyID)
		case entity.ContentPropertyTypeCategory:
			categories = append(categories, contentProperties[i].PropertyID)
		case entity.ContentPropertyTypeAge:
			ages = append(ages, contentProperties[i].PropertyID)
		case entity.ContentPropertyTypeGrade:
			grades = append(grades, contentProperties[i].PropertyID)
		case entity.ContentPropertyTypeSubCategory:
			subCategories = append(subCategories, contentProperties[i].PropertyID)
		}
	}
	return &entity.ContentProperties{
		ContentID:   cid,
		Program:     program,
		Subject:     subjects,
		Category:    categories,
		SubCategory: subCategories,
		Age:         ages,
		Grade:       grades,
	}, nil
}

func (cm *ContentModel) buildFolderCondition(ctx context.Context, condition da.ContentCondition, searchUserIDs []string, user *entity.Operator) *da.FolderCondition {
	dirPath := condition.DirPath
	isAssets := false
	disableFolder := true
	for i := range condition.ContentType {
		if entity.NewContentType(condition.ContentType[i]).IsAsset() {
			isAssets = true
			continue
		}
		if entity.NewContentType(condition.ContentType[i]) == entity.AliasContentTypeFolder {
			disableFolder = false
		}
	}
	partition := entity.FolderPartitionMaterialAndPlans
	if isAssets {
		partition = entity.FolderPartitionAssets
	}

	folderCondition := &da.FolderCondition{
		OwnerType:    int(entity.OwnerTypeOrganization),
		ItemType:     int(entity.FolderItemTypeFolder),
		Owner:        user.OrgID,
		NameLike:     condition.Name,
		Name:         condition.ContentName,
		ExactDirPath: dirPath,
		//Editors:      searchUserIDs,
		Partition: partition,
		Disable:   disableFolder,
	}
	return folderCondition
}

func (cm *ContentModel) fillFolderContent(ctx context.Context, objs []*entity.FolderContentData, user *entity.Operator) {
	authorIDs := make([]string, len(objs))
	for i := range objs {
		authorIDs[i] = objs[i].Author
	}

	users, err := external.GetUserServiceProvider().BatchGet(ctx, user, authorIDs)
	if err != nil {
		log.Warn(ctx, "get user info failed", log.Err(err), log.Any("objs", objs))
	}
	authorMap := make(map[string]string)
	for i := range users {
		if users[i].Valid {
			authorMap[users[i].ID] = users[i].Name
		}
	}
	for i := range objs {
		objs[i].AuthorName = authorMap[objs[i].Author]
		objs[i].ContentTypeName = objs[i].ContentType.Name()
	}
}

func (cm *ContentModel) convertFolderContent(ctx context.Context, objs []*entity.FolderContent, user *entity.Operator) []*entity.FolderContentData {
	ret := make([]*entity.FolderContentData, len(objs))
	for i := range objs {
		ret[i] = &entity.FolderContentData{
			ID:              objs[i].ID,
			ContentName:     objs[i].ContentName,
			ContentType:     objs[i].ContentType,
			Description:     objs[i].Description,
			Keywords:        strings.Split(objs[i].Keywords, constant.StringArraySeparator),
			Author:          objs[i].Author,
			ItemsCount:      objs[i].ItemsCount,
			PublishStatus:   objs[i].PublishStatus,
			Thumbnail:       objs[i].Thumbnail,
			Data:            objs[i].Data,
			AuthorName:      objs[i].Author,
			DirPath:         objs[i].DirPath,
			ContentTypeName: objs[i].ContentTypeName,
			CreateAt:        objs[i].CreateAt,
			UpdateAt:        objs[i].UpdateAt,
		}
	}
	return ret
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
