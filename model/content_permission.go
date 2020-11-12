package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
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
	CheckPublishContentsPermission(ctx context.Context, cids []string, scope string, user *entity.Operator) (bool, error)
	CheckGetContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error)

	CheckUpdateContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error)
	CheckLockContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error)
	CheckDeleteContentPermission(ctx context.Context, cids []string, user *entity.Operator) (bool, error)
	CheckQueryContentPermission(ctx context.Context, condition da.ContentCondition, mode QueryMode, user *entity.Operator) (bool, error)

	GetPermissionedOrgs(ctx context.Context, permission external.PermissionName, op *entity.Operator) ([]Entity, error)
}

type Entity struct {
	ID string
	Name string
}

type ContentPermissionModel struct{}

func (cpm *ContentPermissionModel) CheckCreateContentPermission(ctx context.Context, data entity.CreateContentRequest, user *entity.Operator) (bool, error) {
	//检查是否有添加权限
	hasPermission, err := cpm.checkHasPermission(ctx, cpm.createPermissionName(ctx, data.ContentType), user)
	if err != nil{
		return false, err
	}
	//没有权限，直接返回
	if !hasPermission {
		return false, nil
	}
	return true, nil
}

func (cpm *ContentPermissionModel) CheckPublishContentsPermission(ctx context.Context, cids []string, scope string, user *entity.Operator) (bool, error) {
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
		}else if scope != "" {
			//若作者是自己，且非republish，则查看权限
			ret, err := cpm.checkCMSPermission(ctx, scope, cpm.createPermissionName(ctx, contentList[i].ContentType), user)
			if err != nil{
				return false, err
			}
			if !ret {
				return false, nil
			}
		}
	}
	//若全是自己的content，则返回有权限
	if len(othersContents) == 0 {
		return true, nil
	}

	//检查是否有republished权限及该content是否为archive状态
	for i := range othersContents {
		if scope == "" {
			scope = othersContents[i].PublishScope
		}
		//有republish archive权限
		if othersContents[i].PublishStatus != entity.ContentStatusArchive {
			log.Info(ctx, "republish not archive content", log.String("cid", othersContents[i].ID))
			return false, nil
		}

		hasPermission, err := cpm.checkCMSPermission(ctx, scope , []external.PermissionName{external.RepublishArchivedContent274}, user)
		if err != nil{
			return false, err
		}
		if !hasPermission {
			log.Info(ctx, "no republish author", log.String("user_id", user.UserID))
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
	if content.Author == user.UserID{
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
	//不是自己的，查看lock_by和修改权限
	if content.LockedBy != "" && content.LockedBy != constant.LockedByNoBody && content.LockedBy != user.UserID {
		log.Info(ctx, "can't update content locked by others", log.String("cid", cid), log.String("user_id", user.UserID))
		return false, nil
	}
	if content.LockedBy == user.UserID {
		log.Info(ctx, "locked by user", log.String("cid", cid), log.String("user_id", user.UserID))
		return true, nil
	}
	//若是自己的content，则可以修改
	if content.Author == user.UserID{
		log.Info(ctx, "author edit", log.String("cid", cid), log.String("user_id", user.UserID))
		return true, nil
	}

	if content.PublishStatus != entity.ContentStatusPublished {
		log.Info(ctx, "update other's unpublished content", log.String("cid", cid), log.String("user_id", user.UserID))
		return false, nil
	}

	//若被自己锁定，则查看权限
	return cpm.checkCMSPermission(ctx, content.PublishScope, cpm.editPermissionName(ctx, content.ContentType), user)
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
	if content.LockedBy != "" && content.LockedBy != constant.LockedByNoBody && content.LockedBy != user.UserID {
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
	if content.Author == user.UserID || content.LockedBy == user.UserID{
		return true, nil
	}

	if content.PublishStatus != entity.ContentStatusPublished {
		log.Info(ctx, "update other's unpublished content", log.String("cid", cid), log.String("user_id", user.UserID))
		return false, nil
	}
	//若未锁定，则查看权限
	return cpm.checkCMSPermission(ctx, content.PublishScope, cpm.editPermissionName(ctx, content.ContentType), user)
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
	for i := range othersContents {
		permissions := make([]external.PermissionName, 0)
		if othersContents[i].PublishStatus == entity.ContentStatusArchive {
			permissions = append(permissions, external.DeleteArchivedContent275)
		}else if othersContents[i].PublishStatus == entity.ContentStatusPublished {
			permissions = append(permissions, external.ArchivePublishedContent273)
		}else{
			log.Info(ctx, "only archive and published content of other's can be deleted", log.String("user_id", user.UserID), log.String("cid", othersContents[i].ID))
			return false, nil
		}

		flag, err := cpm.checkCMSPermission(ctx, othersContents[i].PublishScope, permissions, user)
		if err != nil{
			return false, err
		}
		if !flag {
			return false, nil
		}
	}

	return true, nil
}

func (cpm *ContentPermissionModel) CheckQueryContentPermission(ctx context.Context, condition da.ContentCondition, mode QueryMode, user *entity.Operator) (bool, error) {
	permissions := make([]external.PermissionName, 0)
	switch mode {
	case QueryModePending:
		log.Info(ctx, "check query pending content permission", log.Any("user", user), log.Any("condition", condition), log.String("mode", string(mode)))
		permissions = []external.PermissionName{external.PendingContentPage203}
		for i := range condition.Scope {
			ret, err := cpm.checkCMSPermission(ctx, condition.Scope[i], permissions, user)
			if err != nil{
				return false, err
			}
			if !ret {
				return false, nil
			}
		}
		return true, nil

	case QueryModePublished:
		//published模式，若需要查看archive需要权限，查看assets需要权限
		log.Info(ctx, "check query published content permission", log.Any("user", user), log.Any("condition", condition), log.String("mode", string(mode)))
		permissions = []external.PermissionName{external.PublishedContentPage204}
		if utils.ContainsStr(condition.PublishStatus, entity.ContentStatusArchive) {
			permissions = append(permissions, external.ArchivedContentPage205)
		}
		if utils.ContainsInt(condition.ContentType, entity.ContentTypeAssets) {
			permissions = append(permissions, external.CreateContentPage201)
		}
		for i := range condition.Scope {
			ret, err := cpm.checkCMSPermission(ctx, condition.Scope[i], permissions, user)
			if err != nil{
				return false, err
			}
			if !ret {
				return false, nil
			}
		}
		return true, nil
	case QueryModePrivate:
		//private模式，若需要查看archive需要权限，查看assets需要权限
		log.Info(ctx, "check query private content permission", log.Any("user", user), log.Any("condition", condition), log.String("mode", string(mode)))
		permissions = make([]external.PermissionName, 0)
		if utils.ContainsInt(condition.ContentType, entity.ContentTypeAssets) {
			permissions = append(permissions, external.CreateAssetPage301)
		}
		if utils.ContainsStr(condition.PublishStatus, entity.ContentStatusArchive) {
			permissions = append(permissions, external.ArchivedContentPage205)
		}

		for i := range condition.Scope {
			ret, err := cpm.checkCMSPermission(ctx, condition.Scope[i], permissions, user)
			if err != nil{
				return false, err
			}
			if !ret {
				return false, nil
			}
		}
		return true, nil
	}
	return false, nil
}

func (s *ContentPermissionModel) GetPermissionedOrgs(ctx context.Context, permission external.PermissionName, op *entity.Operator) ([]Entity, error){
	schools, err := external.GetSchoolServiceProvider().GetByPermission(ctx, op, permission)
	if err != nil{
		log.Error(ctx, "get permission orgs failed", log.Err(err))
		return nil, err
	}
	entities := make([]Entity, 0)
	for i := range schools {
		entities = append(entities, Entity{
			ID:   schools[i].ID,
			Name: schools[i].Name,
		})
	}
	orgs, err := external.GetOrganizationServiceProvider().BatchGet(ctx, []string{op.OrgID})
	if err != nil || len(orgs) < 1{
		log.Error(ctx, "get org info failed", log.Err(err))
		return nil, err
	}
	orgName := ""
	if orgs[0].Valid {
		orgName = orgs[0].Name
	}
	entities = append(entities, Entity{
		ID:   op.OrgID,
		Name: orgName,
	})
	return entities, nil
}

func (s *ContentPermissionModel) checkCMSPermission(ctx context.Context, scope string, permissions []external.PermissionName, op *entity.Operator) (bool, error) {
	if len(permissions) < 1 {
		return true, nil
	}
	//查看org权限
	if scope == op.OrgID {
		//检查Permission权限
		for i := range permissions {
			hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, permissions[i])
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

	//检查学校权限
	for i := range permissions {
		hasPermission, err := external.GetPermissionServiceProvider().HasSchoolPermission(ctx, op.UserID, scope, permissions[i])
		if err != nil{
			log.Warn(ctx, "get school permission failed", log.Err(err))
			return false, err
		}
		if !hasPermission {
			return false, nil
		}
	}
	return false, nil
}


func (s *ContentPermissionModel) checkHasPermission(ctx context.Context, permissions []external.PermissionName, op *entity.Operator) (bool, error) {
	if len(permissions) < 1 {
		return true, nil
	}

	for i := range permissions {
		hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, permissions[i])
		if err != nil{
			log.Error(ctx, "get permission failed", log.Err(err))
			return false, err
		}
		//有permission，直接返回
		if hasPermission {
			return true, nil
		}
	}

	//检查学校权限
	for i := range permissions {
		schools, err := external.GetSchoolServiceProvider().GetByPermission(ctx, op, permissions[i])
		if err != nil{
			log.Error(ctx, "get school permission failed", log.Err(err))
			return false, err
		}
		if len(schools) > 0 {
			return true, nil
		}
	}
	return false, nil
}

func (s *ContentPermissionModel) checkContentScope(ctx context.Context, content *entity.Content, permission external.PermissionName, op *entity.Operator) (bool, error) {
	schools, err := external.GetSchoolServiceProvider().GetByPermission(ctx, op, permission)
	if err != nil{
		log.Error(ctx, "get permission orgs failed", log.Err(err))
		return false, err
	}
	orgs := []string{op.OrgID}
	for i := range schools {
		orgs = append(orgs, schools[i].ID)
	}
	log.Info(ctx, "user orgs with permission", log.Strings("orgs", orgs), log.String("permission", string(permission)), log.Any("user", op), log.Any("content", content))
	for i := range orgs{
		if orgs[i] == content.PublishScope {
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
