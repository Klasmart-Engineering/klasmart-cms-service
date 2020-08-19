package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"
	"gitlab.badanamu.com.cn/calmisland/dbo"
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
	PublishContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error

	LockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error
	UnlockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error

	GetContentById(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error)
	GetContentByIdList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.ContentInfo, error)

	DeleteContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error
	CloneContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error)

	UpdateContentPublishStatus(ctx context.Context, tx *dbo.DBContext, cid, reason, status string) error
	CheckContentAuthorization(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error

	SearchUserContent(ctx context.Context, tx *dbo.DBContext, condition da.DyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error)
	SearchUserPrivateContent(ctx context.Context, tx *dbo.DBContext, condition da.DyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error)
	ListPendingContent(ctx context.Context, tx *dbo.DBContext, condition da.DyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error)
	SearchContent(ctx context.Context, tx *dbo.DBContext, condition da.DyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error)

	GetResourceContentById(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error)
}

type ContentModel struct {
}

func (cm *ContentModel) doPublishContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	//TODO:Maybe wrong
	if user.OrgID == content.Org && user.Role != "teacher" {
		content.PublishScope = getContentPublishScope(ctx, user)
		content.PublishStatus = entity.ContentStatusPublished
	} else {
		//TODO:更新发布状态
		//nextApprover, err := klcr.Get().SubmitApproval(ctx, tx, user.ID, user.OrgID, user.Role.String(), string(user.OrgCategory), int(content.Id), int(scope), "kidsloop")
		//if err != nil {
		//	logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Create publish contentdata approval failed, error: %v", err)
		//	return constant.ErrPublishContentApplicationFailed
		//}

		content.PublishScope = getContentPublishScope(ctx, user)
		content.PublishStatus = entity.ContentStatusPending
	}

	err := da.GetDyContentDA().UpdateContent(ctx, content.ID, *content)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Update lesson plan failed, error: %v", err)
		return ErrUpdateContentFailed
	}

	return nil
}

func (cm ContentModel) checkContentInfo(ctx context.Context, c entity.CreateContentRequest, created bool) error {
	logger.WithContext(ctx).WithField("subject", "content").Infof("checkContentEntity entity, data: %#v, created: %v", c, created)
	//TODO:Check age, category...
	return nil
}

func (cm ContentModel) checkUpdateContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) (*entity.Content, error) {
	if content.Author != user.UserID {
		return nil, ErrNoAuth
	}
	if content.PublishStatus == entity.ContentStatusPending {
		return nil, ErrInvalidPublishStatus
	}
	if content.PublishStatus == entity.ContentStatusPublished {
		//CloneContent content
		cid, err := cm.CloneContent(ctx, tx, content.ID, user)
		if err != nil {
			return nil, err
		}
		newContent, err := da.GetDyContentDA().GetContentById(ctx, cid)
		if err != nil {
			return nil, err
		}
		//Update with new content
		return newContent, nil
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
	return nil
}

func (cm *ContentModel) searchContent(ctx context.Context, tx *dbo.DBContext, condition *da.DyContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error) {
	key, objs, err := da.GetDyContentDA().SearchContent(ctx, condition)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't read contentdata, error: %v", err)
		return "", nil, ErrReadContentFailed
	}
	logger.WithContext(ctx).WithField("subject", "contentdata").Infof("Read count: %v,  data: %v", key, objs)
	response := make([]*entity.ContentInfo, len(objs))
	for i := range objs {
		temp, err := contentdata.ConvertContentObj(ctx, objs[i])
		if err != nil {
			logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't parse contentdata, contentId: %v, error: %v", objs[i].ID, err)
			return "", nil, err
		}
		response[i] = temp
	}
	logger.WithContext(ctx).WithField("subject", "contentdata").Infof("Content with contentdata: %v", key, response)
	contentWithDetails, err := buildContentWithDetails(ctx, response, user)
	if err != nil {
		return "", nil, err
	}
	logger.WithContext(ctx).WithField("subject", "contentdata").Infof("Content with details: %v", key, contentWithDetails)
	return key, contentWithDetails, nil
}

func (cm *ContentModel) searchContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition *da.DyCombineContentCondition, user *entity.Operator) (string, []*entity.ContentInfoWithDetails, error) {
	key, objs, err := da.GetDyContentDA().SearchContent(ctx, condition)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't read contentdata, error: %v", err)
		return "", nil, ErrReadContentFailed
	}
	response := make([]*entity.ContentInfo, len(objs))
	for i := range objs {
		temp, err := contentdata.ConvertContentObj(ctx, objs[i])
		if err != nil {
			logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't parse contentdata, contentId: %v, error: %v", objs[i].ID, err)
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

	//添加内容
	pid, err := da.GetDyContentDA().CreateContent(ctx, *obj)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't create contentdata, error: %v", err)
		return "", ErrCreateContentFailed
	}

	//添加版本信息
	err = GetContentVersionModel().AddContentVersion(ctx, tx, "", pid)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't add version info, error: %v", err)
		return "", ErrCreateContentFailed
	}

	if c.DoPublish{
		err = cm.PublishContent(ctx, tx, pid, operator)
		if err != nil {
			logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Publish content failed, error: %v", err)
			return pid, err
		}
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
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't read contentdata on update contentdata, error: %v", err)
		return ErrNoContentData
	}
	logger.WithContext(ctx).WithField("subject", "contentdata").Infof("Get contentdata data: %#v", content)

	content, err = cm.checkUpdateContent(ctx, tx, content, user)
	if err != nil {
		return err
	}
	//Check status
	//Check Authorization
	obj, err := cm.prepareUpdateContentParams(ctx, content, &data)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't prepare params contentdata on update contentdata, error: %v", err)
		return err
	}

	err = da.GetDyContentDA().UpdateContent(ctx, cid, *obj)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Update contentdata failed, error: %v", err)
		return ErrUpdateContentFailed
	}

	if data.DoPublish{
		err = cm.PublishContent(ctx, tx, cid, user)
		if err != nil {
			logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Publish content failed, error: %v", err)
			return err
		}
	}
	return nil
}

func (cm *ContentModel) UpdateContentPublishStatus(ctx context.Context, tx *dbo.DBContext, cid, reason, status string) error {
	content, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't read contentdata on update contentdata, error: %v", err)
		return ErrReadContentFailed
	}

	content.PublishStatus = entity.NewContentPublishStatus(status)
	content.RejectReason = reason
	err = da.GetDyContentDA().UpdateContent(ctx, cid, *content)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Update contentdata scope failed, error: %v", err)
		return ErrUpdateContentFailed
	}
	return nil
}

func (cm *ContentModel) UnlockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error{
	content, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "course").Warnf("Can't read contentdata for publishing, error: %v", err)
		return ErrNoContent
	}
	//TODO:检查权限

	if content.LockedBy != user.UserID {
		return ErrNoAuth
	}
	now := time.Now()
	content.LockedBy = "-"
	content.UpdatedAt = &now
	return da.GetDyContentDA().UpdateContent(ctx, cid, *content)
}
func (cm *ContentModel) LockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error{
	locker, err := mutex.NewLock(ctx, "content", "lock", cid)
	if err != nil{
		return err
	}
	locker.Lock()
	defer locker.Unlock()

	content, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "course").Warnf("Can't read contentdata for publishing, error: %v", err)
		return ErrNoContent
	}
	//TODO:检查权限

	if content.LockedBy != "" && content.LockedBy != "-" {
		return ErrContentAlreadyLocked
	}
	now := time.Now()
	content.LockedBy = cid
	content.UpdatedAt = &now
	return da.GetDyContentDA().UpdateContent(ctx, cid, *content)
}

func (cm *ContentModel) PublishContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error {
	content, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "course").Warnf("Can't read contentdata for publishing, error: %v", err)
		return ErrNoContent
	}

	err = cm.checkPublishContent(ctx, tx, content, user)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "course").Warnf("Check content scope & sub content scope failed, error: %v", err)
		return err
	}

	//发布
	err = cm.doPublishContent(ctx, tx, content, user)
	if err != nil {
		return err
	}
	return nil
}

func (cm *ContentModel) DeleteContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error {
	content, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't read contentdata on delete contentdata, error: %v", err)
		return ErrReadContentFailed
	}

	if content.Author != user.UserID {
		return ErrNoAuth
	}


	obj, err := cm.prepareDeleteContentParams(ctx, content, content.PublishStatus)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Prepare contentdata params failed, error: %v", err)
		return ErrDeleteContentFailed
	}

	err = da.GetDyContentDA().UpdateContent(ctx, cid, *obj)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Delete contentdata failed, error: %v", err)
		return ErrDeleteContentFailed
	}

	return nil
}

func (cm *ContentModel) CloneContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error){
	content, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't read contentdata on update contentdata, error: %v", err)
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
		logger.WithContext(ctx).WithField("subject", "course").Warnf("No auth to read content for cloning, error: %v", err)
		return "", ErrCloneContentFailed
	}

	obj := cm.prepareCloneContentParams(ctx, content, user)

	id, err := da.GetDyContentDA().CreateContent(ctx, *obj)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("clone contentdata failed, error: %v", err)
		return "", ErrUpdateContentFailed
	}

	//记录版本信息
	err = GetContentVersionModel().AddContentVersion(ctx, tx, cid, id)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't record content version info, error: %v", err)
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
		content.PublishStatus == entity.ContentStatusHidden{
		return nil
	}

	if content.PublishStatus != entity.ContentStatusPublished {
		logger.WithContext(ctx).WithField("subject", "content").Warnf("Read unpublished content, userId: %v, contentId: %v", user.UserID, content.ID)
		return ErrGetUnpublishedContent
	}
	//TODO: Check org scope

	return ErrGetUnauthorizedContent
}

func (cm *ContentModel) GetContentById(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error) {
	obj, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't read contentdata, error: %v", err)
		return nil, ErrNoContent
	}
	content, err := contentdata.ConvertContentObj(ctx, obj)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't parse contentdata, error: %v", err)
		return nil, ErrParseContentDataFailed
	}
	logger.WithContext(ctx).WithField("subject", "content").Infof("Get content by id, Content extra: %#v", content.Extra)

	//补全相关内容
	err = content.Data.PrepareResult(ctx)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't get contentdata for details, error: %v", err)
		return nil, ErrParseContentDataDetailsFailed
	}

	contentWithDetails, err := buildContentWithDetails(ctx, []*entity.ContentInfo{content}, user)
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't parse contentdata, error: %v", err)
		return nil, ErrReadContentFailed
	}
	if len(contentWithDetails) < 1 {
		return &entity.ContentInfoWithDetails{
			ContentInfo: *content,
		}, nil
	}
	return contentWithDetails[0], nil
}

func (cm *ContentModel) GetContentByIdList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.ContentInfo, error) {
	if len(cids) < 1 {
		return nil, nil
	}
	_, data, err := da.GetDyContentDA().SearchContent(ctx, &da.DyContentCondition{
		IDS: cids,
	})
	if err != nil {
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't read contentdata, error: %v", err)
		return nil, ErrReadContentFailed
	}
	res := make([]*entity.ContentInfo, len(data))
	for i := range data {
		temp, err := contentdata.ConvertContentObj(ctx, data[i])
		if err != nil {
			logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't parse contentdata, id: %v, error: %v", data[i].ID, err)
			return nil, ErrReadContentFailed
		}
		res[i] = temp
	}
	return res, nil
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

func (cm *ContentModel) GetResourceContentById(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error){
	contentData, err := da.GetDyContentDA().GetContentById(ctx, cid)
	if err != nil{
		return nil, err
	}
	if contentData.SourceId != cid {
		newContentData, err := da.GetDyContentDA().GetContentById(ctx, contentData.Subject)
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
		logger.WithContext(ctx).WithField("subject", "contentdata").Warnf("Can't parse contentdata, error: %v", err)
		return nil, ErrReadContentFailed
	}
	if len(contentWithDetails) < 1 {
		return &entity.ContentInfoWithDetails{
			ContentInfo: *content,
		}, nil
	}
	return contentWithDetails[0], nil
}

func buildContentWithDetails(ctx context.Context, contentList []*entity.ContentInfo, user *entity.Operator) ([]*entity.ContentInfoWithDetails, error) {
	return nil, nil
}

func getContentPublishScope(ctx context.Context, operator *entity.Operator) string {
	return ""
}

func listVisibleScopes(ctx context.Context, operator *entity.Operator) []string {
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