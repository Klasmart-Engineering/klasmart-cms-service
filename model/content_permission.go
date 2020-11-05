package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"sync"
)

type QueryMode string
var(
	QueryModePending QueryMode   = "pending"
	QueryModePublished QueryMode = "published"
	QueryModePrivate QueryMode   = "private"
)

type IContentPermissionModel interface{
	CheckCreateContentPermission(ctx context.Context, data entity.CreateContentRequest, user *entity.Operator) (bool, error)
	CheckPublishContentsPermission(ctx context.Context, cids []string, user *entity.Operator) (bool, error)
	CheckGetContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error)

	CheckUpdateContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error)
	CheckLockContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error)
	CheckDeleteContentPermission(ctx context.Context, cids []string, user *entity.Operator) (bool, error)
	CheckQueryContentPermission(ctx context.Context, condition da.ContentCondition, mode QueryMode, user *entity.Operator) (bool, error)
}

type ContentPermissionModel struct{}

func (cpm *ContentPermissionModel) CheckCreateContentPermission(ctx context.Context, data entity.CreateContentRequest, user *entity.Operator) (bool, error) {
	//检查是否有添加权限
	hasPermission, err := cpm.checkCMSPermission(ctx, cpm.createPermissionName(ctx, data.ContentType), user)
	if err != nil{
		return false, err
	}
	//没有权限，直接返回
	if !hasPermission {
		return false, nil
	}
	return true, nil
}

func (cpm *ContentPermissionModel) CheckPublishContentsPermission(ctx context.Context, cids []string, user *entity.Operator) (bool, error) {
	contentList, err := da.GetContentDA().GetContentByIDList(ctx, dbo.MustGetDB(ctx), cids)
	if err != nil{
		log.Warn(ctx, "get content list failed", log.Strings("ids", cids), log.Err(err))
		return false, err
	}

	//排除自己的content
	othersContents := make([]*entity.Content, 0)
	for i := range contentList {
		if contentList[i].Org != user.OrgID {
			//若org_id不相等，则无权限
			log.Info(ctx, "publish content in other org", log.String("org_id", contentList[i].Org), log.String("user_org_id", user.OrgID))
			return false, nil
		}
		if contentList[i].Author != user.UserID{
			othersContents = append(othersContents, contentList[i])
		}
	}
	//若全是自己的content，则返回有权限
	if len(othersContents) == 0 {
		return true, nil
	}

	//检查是否有republished权限及该content是否为archive状态
	hasPermission, err := cpm.checkCMSPermission(ctx, []external.PermissionName{external.RepublishArchivedContent274}, user)
	if err != nil{
		return false, err
	}
	if !hasPermission {
		log.Info(ctx, "no republish author", log.String("user_id", user.UserID))
		return false, nil
	}

	//有republish archive权限
	for i := range othersContents {
		if othersContents[i].PublishStatus != entity.ContentStatusArchive {
			log.Info(ctx, "republish not archive content", log.String("cid", othersContents[i].ID))
			return false, nil
		}
	}
	return true, nil
}

func (cpm *ContentPermissionModel) CheckGetContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error) {
	content, err := da.GetContentDA().GetContentByID(ctx, dbo.MustGetDB(ctx), cid)
	if err != nil{
		log.Warn(ctx, "get content failed", log.String("id", cid), log.Err(err))
		return false, err
	}
	//排除其他机构
	if content.Org != user.OrgID {
		//若org_id不相等，则无权限
		log.Info(ctx, "view content in other org", log.String("cid", cid), log.String("user_org_id", user.OrgID))
		return false, nil
	}
	//若是自己的content，则可以获取
	if content.Author != user.UserID{
		return true, nil
	}
	//若不是自己的，未发布不能看
	if content.PublishStatus == entity.ContentStatusDraft ||
		content.PublishStatus == entity.ContentStatusRejected{
		log.Info(ctx, "view draft content", log.String("cid", cid), log.String("user_org_id", user.OrgID))
		return false, nil
	}

	//查看content的publish scope是否在自己的可见范围内
	switch content.PublishStatus {
	case entity.ContentStatusPublished:
		return cpm.checkContentScope(ctx, content, external.PublishedContentPage204, user)
	case entity.ContentStatusPending:
		return cpm.checkContentScope(ctx, content, external.PendingContentPage203, user)
	case entity.ContentStatusArchive:
		return cpm.checkContentScope(ctx, content, external.ArchivedContentPage205, user)
	}
	//权限可能不对，没有权限限制则允许
	return true, nil
}

func (cpm *ContentPermissionModel) CheckUpdateContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error) {
	content, err := da.GetContentDA().GetContentByID(ctx, dbo.MustGetDB(ctx), cid)
	if err != nil{
		log.Warn(ctx, "get content failed", log.String("id", cid), log.Err(err))
		return false, err
	}
	if content.ContentType == entity.ContentTypeAssets {
		log.Warn(ctx, "asset can't update", log.String("id", cid), log.Err(err))
		return false, nil
	}
	//排除其他机构
	if content.Org != user.OrgID {
		//若org_id不相等，则无权限
		log.Info(ctx, "update content in other org", log.String("cid", cid), log.String("user_org_id", user.OrgID))
		return false, nil
	}
	//若是自己的content，则可以修改
	if content.Author != user.UserID{
		return true, nil
	}
	//不是自己的，查看lock_by和修改权限
	if content.LockedBy != user.UserID {
		log.Info(ctx, "can't update content locked by others", log.String("cid", cid), log.String("user_id", user.UserID))
		return false, nil
	}

	//若被自己锁定，则查看权限
	return cpm.checkCMSPermission(ctx, cpm.editPermissionName(ctx, content.ContentType), user)
}

func (cpm *ContentPermissionModel) CheckLockContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error) {
	content, err := da.GetContentDA().GetContentByID(ctx, dbo.MustGetDB(ctx), cid)
	if err != nil{
		log.Warn(ctx, "get content failed", log.String("id", cid), log.Err(err))
		return false, err
	}
	if content.ContentType == entity.ContentTypeAssets {
		log.Warn(ctx, "asset can't update", log.String("id", cid), log.Err(err))
		return false, nil
	}
	if content.LockedBy != "" {
		log.Info(ctx, "can't lock content locked by others", log.String("cid", cid), log.String("user_id", user.UserID))
		return false, nil
	}
	//排除其他机构
	if content.Org != user.OrgID {
		//若org_id不相等，则无权限
		log.Info(ctx, "update content in other org", log.String("cid", cid), log.String("user_org_id", user.OrgID))
		return false, nil
	}
	//若是自己的content，则可以修改
	if content.Author != user.UserID{
		return true, nil
	}

	//若未锁定，则查看权限
	return cpm.checkCMSPermission(ctx, cpm.editPermissionName(ctx, content.ContentType), user)
}

func (cpm *ContentPermissionModel) CheckDeleteContentPermission(ctx context.Context, cids []string, user *entity.Operator) (bool, error) {
	contentList, err := da.GetContentDA().GetContentByIDList(ctx, dbo.MustGetDB(ctx), cids)
	if err != nil{
		log.Warn(ctx, "get content list failed", log.Strings("ids", cids), log.Err(err))
		return false, err
	}

	//排除自己的content
	othersContents := make([]*entity.Content, 0)
	for i := range contentList {
		if contentList[i].Org != user.OrgID {
			//若org_id不相等，则无权限
			log.Info(ctx, "publish content in other org", log.String("org_id", contentList[i].Org), log.String("user_org_id", user.OrgID))
			return false, nil
		}
		if contentList[i].Author != user.UserID{
			othersContents = append(othersContents, contentList[i])
		}
	}
	//若全是自己的content，则返回有权限
	if len(othersContents) == 0 {
		return true, nil
	}

	//判断archive和published
	hasArchive := false
	hasPublished := false
	for i := range othersContents {
		if othersContents[i].PublishStatus == entity.ContentStatusArchive {
			hasArchive = true
		}else if othersContents[i].PublishStatus == entity.ContentStatusPublished {
			hasPublished = true
		}else{
			log.Info(ctx, "only archive and published content of other's can be deleted", log.String("user_id", user.UserID), log.String("cid", othersContents[i].ID))
			return false, nil
		}
	}
	permissions := make([]external.PermissionName, 0)
	if hasArchive {
		permissions = append(permissions, external.DeleteArchivedContent275)
	}
	if hasPublished {
		permissions = append(permissions, external.ArchivePublishedContent273)
	}
	return cpm.checkCMSPermission(ctx, permissions, user)
}

func (cpm *ContentPermissionModel) CheckQueryContentPermission(ctx context.Context, condition da.ContentCondition, mode QueryMode, user *entity.Operator) (bool, error) {
	switch mode {
	case QueryModePending:
		permissions := []external.PermissionName{external.PendingContentPage203}
		return cpm.checkCMSPermission(ctx, permissions, user)
	case QueryModePublished:
		//published模式，若需要查看archive需要权限，查看assets需要权限
		permissions := []external.PermissionName{external.PublishedContentPage204}
		if containsStr(condition.PublishStatus, entity.ContentStatusArchive) {
			permissions = append(permissions, external.ArchivedContentPage205)
		}
		if containsInt(condition.ContentType, entity.ContentTypeAssets) {
			permissions = append(permissions, external.CreateContentPage201)
		}
		return cpm.checkCMSPermission(ctx, permissions, user)
	case QueryModePrivate:
		//private模式，若需要查看archive需要权限，查看assets需要权限
		permissions := make([]external.PermissionName, 0)
		if containsInt(condition.ContentType, entity.ContentTypeAssets) {
			permissions = append(permissions, external.CreateContentPage201)
		}
		if containsStr(condition.PublishStatus, entity.ContentStatusArchive) {
			permissions = append(permissions, external.ArchivedContentPage205)
		}
		return cpm.checkCMSPermission(ctx, permissions, user)
	}
	return false, nil
}

func (s *ContentPermissionModel) checkCMSPermission(ctx context.Context, permissions []external.PermissionName, op *entity.Operator) (bool, error) {
	if len(permissions) < 1 {
		return true, nil
	}
	for i := range permissions {
		hasPermission, err := external.GetPermissionServiceProvider().HasPermission(ctx, op, permissions[i])
		if err != nil{
			log.Error(ctx, "get permission failed", log.Err(err))
			return false, err
		}
		//有permission，直接返回
		if hasPermission {
			return true, nil
		}
	}

	log.Info(ctx, "No permission", log.String("userID", op.UserID))
	return false, nil
}
func (s *ContentPermissionModel) checkContentScope(ctx context.Context, content *entity.Content, permission external.PermissionName, op *entity.Operator) (bool, error) {
	orgs, err := external.GetPermissionServiceProvider().GetHasPermissionOrganizations(ctx, op, permission)
	if err != nil{
		log.Error(ctx, "get permission orgs failed", log.Err(err))
		return false, err
	}
	for i := range orgs{
		if orgs[i].ID == content.PublishScope {
			return true, nil
		}
	}
	return false, nil
}
func (s *ContentPermissionModel) createPermissionName(ctx context.Context, contentType entity.ContentType) []external.PermissionName{
	switch contentType {
	case entity.ContentTypeMaterial:
		return []external.PermissionName{external.CreateContentPage201, external.CreateLessonMaterial220}
	case entity.ContentTypeLesson:
		return []external.PermissionName{external.CreateContentPage201, external.CreateLessonPlan221}
	case entity.ContentTypeAssets:
		return []external.PermissionName{external.CreateContentPage201, external.CreateAsset320}
	}
	return []external.PermissionName{external.CreateContentPage201}
}

func (s *ContentPermissionModel) editPermissionName(ctx context.Context, contentType entity.ContentType) []external.PermissionName{
	switch contentType {
	case entity.ContentTypeMaterial:
		return []external.PermissionName{external.EditOrgPublishedContent235, external.EditLessonMaterialMetadataAndContent236}
	case entity.ContentTypeLesson:
		return []external.PermissionName{external.EditOrgPublishedContent235, external.EditLessonPlanContent238, external.EditLessonPlanMetadata237}
	}
	return []external.PermissionName{external.EditOrgPublishedContent235}
}

func (s *ContentPermissionModel) deletePermissionName(ctx context.Context, contentStatus entity.ContentPublishStatus) []external.PermissionName{
	switch contentStatus {
	case entity.ContentStatusArchive:
		return []external.PermissionName{external.EditOrgPublishedContent235, external.EditLessonMaterialMetadataAndContent236}
	case entity.ContentStatusPublished:
		return []external.PermissionName{external.EditOrgPublishedContent235, external.EditLessonPlanContent238, external.EditLessonPlanMetadata237}
	}
	return nil
}
func containsStr(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func containsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
var (
	_contentPermissionModel     IContentPermissionModel
	_contentPermissionModelOnce sync.Once
)

func GetContentPermissionModel() IContentPermissionModel {
	_contentPermissionModelOnce.Do(func() {
		_contentPermissionModel = new(ContentPermissionModel)
	})
	return _contentPermissionModel
}
