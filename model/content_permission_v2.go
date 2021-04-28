//content_permission_v2 is for new version of cms permission check
//add my_schools permission check
package model

import (
	"context"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"sync"
)

var (
	ErrInvalidVisibilitySetting = errors.New("invalid visibility settings")
)
type IContentPermissionMySchoolModel interface {
	CheckCreateContentPermission(ctx context.Context, data entity.CreateContentRequest, user *entity.Operator) (bool, error)
	CheckPublishContentsPermission(ctx context.Context, cid string, scopes []string, user *entity.Operator) (bool, error)
	CheckRepublishContentsPermission(ctx context.Context, cids []string, user *entity.Operator) (bool, error)

	CheckGetContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error)

	CheckUpdateContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error)
	CheckDeleteContentPermission(ctx context.Context, cids []string, user *entity.Operator) (bool, error)
	CheckQueryContentPermission(ctx context.Context, condition da.ContentCondition, user *entity.Operator) (bool, error)
}

type ContentPermissionMySchoolModel struct{
	mySchools []*external.School
	allSchools []*external.School
}
func (c *ContentPermissionMySchoolModel) CheckCreateContentPermission(ctx context.Context, data entity.CreateContentRequest, user *entity.Operator) (bool, error){
	err := c.querySchools(ctx, user)
	if err != nil {
		log.Error(ctx, "getVisibilitySettingsType failed",
			log.Err(err),
			log.Any("data", data),
			log.Any("user", user))
		return false, err
	}
	visibilitySettingType, err := c.getVisibilitySettingsType(ctx, data.PublishScope, user)
	if err != nil {
		log.Error(ctx, "getVisibilitySettingsType failed",
			log.Err(err),
			log.Any("data", data),
			log.Any("user", user))
		return false, err
	}

	profile := ContentProfile{
		ContentType:        data.ContentType,
		Status:             entity.ContentStatusDraft,
		VisibilitySettings: visibilitySettingType,
		Owner:              OwnerTypeUser,
	}
	permissionSetList, err := NewContentPermissionTable().GetCreatePermissionSets(ctx, profile)
	if err != nil {
		log.Error(ctx, "GetCreatePermissionSets failed",
			log.Err(err),
			log.Any("profile", profile),
			log.Any("user", user))
		return false, err
	}
	err = permissionSetList.HasPermission(ctx, user)
	if err != nil {
		log.Error(ctx, "No permission",
			log.Err(err),
			log.Any("profile", profile),
			log.Any("user", user))
		return false, nil
	}
	return true, nil
}

func (c *ContentPermissionMySchoolModel) CheckRepublishContentsPermission(ctx context.Context, cids []string, user *entity.Operator) (bool, error){
	contentList, err := GetContentModel().GetRawContentByIDListWithVisibilitySettings(ctx, dbo.MustGetDB(ctx), cids)
	if err != nil {
		log.Error(ctx, "GetContentByIDList failed",
			log.Err(err),
			log.Strings("cids", cids),
			log.Any("user", user))
		return false, err
	}
	profiles, err := c.buildContentProfiles(ctx, contentList, user)
	if err != nil {
		return false, err
	}
	permissionSetList, err := NewContentPermissionTable().GetPublishPermissionSets(ctx, profiles)
	if err != nil {
		return false, err
	}

	err = permissionSetList.HasPermission(ctx, user)
	if err != nil {
		log.Error(ctx, "No permission",
			log.Err(err),
			log.Any("profiles", profiles),
			log.Any("user", user))
		return false, nil
	}
	return true, nil
}

func (c *ContentPermissionMySchoolModel) CheckPublishContentsPermission(ctx context.Context, cid string, scopes []string, user *entity.Operator) (bool, error){
	contentList, err := GetContentModel().GetRawContentByIDListWithVisibilitySettings(ctx, dbo.MustGetDB(ctx), []string{cid})
	if err != nil {
		log.Error(ctx, "GetContentByIDList failed",
			log.Err(err),
			log.String("cid", cid),
			log.Any("user", user))
		return false, err
	}
	for i := range contentList {
		contentList[i].VisibilitySettings = scopes
	}
	profiles, err := c.buildContentProfiles(ctx, contentList, user)
	if err != nil {
		return false, err
	}
	permissionSetList, err := NewContentPermissionTable().GetPublishPermissionSets(ctx, profiles)
	if err != nil {
		return false, err
	}

	err = permissionSetList.HasPermission(ctx, user)
	if err != nil {
		log.Error(ctx, "No permission",
			log.Err(err),
			log.Any("profiles", profiles),
			log.Any("user", user))
		return false, nil
	}
	return true, nil
}

func (c *ContentPermissionMySchoolModel) CheckGetContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error){
	contentList, err := GetContentModel().GetRawContentByIDListWithVisibilitySettings(ctx, dbo.MustGetDB(ctx), []string{cid})
	if err != nil {
		log.Error(ctx, "GetContentByIDList failed",
			log.Err(err),
			log.String("cid", cid),
			log.Any("user", user))
		return false, err
	}
	if len(contentList) < 1 {
		log.Warn(ctx, "content list is nil",
			log.String("cid", cid),
			log.Any("contentList", contentList),
			log.Any("user", user))
		return false, err
	}

	profiles, err := c.buildContentProfiles(ctx, contentList, user)
	if err != nil {
		return false, err
	}

	permissionSetList, err := NewContentPermissionTable().GetViewPermissionSets(ctx, profiles)
	if err != nil {
		return false, err
	}
	err = permissionSetList.HasPermission(ctx, user)
	if err != nil {
		log.Error(ctx, "No permission",
			log.Err(err),
			log.Any("profiles", profiles),
			log.Any("user", user))
		return false, nil
	}
	return true, nil
}

func (c *ContentPermissionMySchoolModel) CheckUpdateContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error){
	contentList, err := GetContentModel().GetRawContentByIDListWithVisibilitySettings(ctx, dbo.MustGetDB(ctx), []string{cid})
	if err != nil {
		log.Error(ctx, "GetContentByIDList failed",
			log.Err(err),
			log.String("cid", cid),
			log.Any("user", user))
		return false, err
	}
	if len(contentList) < 1 {
		log.Warn(ctx, "content list is nil",
			log.String("cid", cid),
			log.Any("contentList", contentList),
			log.Any("user", user))
		return false, err
	}

	profiles, err := c.buildContentProfiles(ctx, contentList, user)
	if err != nil {
		return false, err
	}

	permissionSetList, err := NewContentPermissionTable().GetEditPermissionSets(ctx, *profiles[0])
	if err != nil {
		return false, err
	}
	err = permissionSetList.HasPermission(ctx, user)
	if err != nil {
		log.Error(ctx, "No permission",
			log.Err(err),
			log.Any("profiles", profiles),
			log.Any("user", user))
		return false, nil
	}
	return true, nil
}
func (c *ContentPermissionMySchoolModel) CheckDeleteContentPermission(ctx context.Context, cids []string, user *entity.Operator) (bool, error){
	contentList, err := GetContentModel().GetRawContentByIDListWithVisibilitySettings(ctx, dbo.MustGetDB(ctx), cids)
	if err != nil {
		log.Error(ctx, "GetContentByIDList failed",
			log.Err(err),
			log.Strings("cids", cids),
			log.Any("user", user))
		return false, err
	}
	profiles, err := c.buildContentProfiles(ctx, contentList, user)
	if err != nil {
		return false, err
	}
	permissionSetList, err := NewContentPermissionTable().GetRemovePermissionSets(ctx, profiles)
	if err != nil {
		return false, err
	}
	err = permissionSetList.HasPermission(ctx, user)
	if err != nil {
		log.Error(ctx, "No permission",
			log.Err(err),
			log.Any("profiles", profiles),
			log.Any("user", user))
		return false, nil
	}
	return true, nil
}
func (c *ContentPermissionMySchoolModel) CheckQueryContentPermission(ctx context.Context, condition da.ContentCondition, user *entity.Operator) (bool, error){
	contentProfiles, err := c.buildByConditionContentProfiles(ctx, condition, user)
	if err != nil {
		return false, err
	}
	permissionSetList, err := NewContentPermissionTable().GetViewPermissionSets(ctx, contentProfiles)
	if err != nil {
		return false, err
	}
	err = permissionSetList.HasPermission(ctx, user)
	if err != nil {
		log.Error(ctx, "No permission",
			log.Err(err),
			log.Any("profiles", contentProfiles),
			log.Any("user", user))
		return false, nil
	}
	return true, nil
}

//func (c *ContentPermissionMySchoolModel) checkPermissionSet(ctx context.Context, permissionSetList []*IPermissionSet, user *entity.Operator) (bool, error){
//	for i := range permissionSetList {
//		err := permissionSetList[i].HasPermission(ctx, *user)
//		if err == nil {
//			return true, nil
//		}
//		if err != ErrHasNoPermission {
//			log.Error(ctx, "HasPermission failed",
//				log.Err(err),
//				log.Any("permissionSetList", permissionSetList),
//				log.Any("user", user))
//		}
//	}
//	log.Info(ctx, "Has no permission",
//		log.Any("permissionSetList", permissionSetList),
//		log.Any("user", user))
//	return false, nil
//}
func (c *ContentPermissionMySchoolModel) buildByConditionContentProfiles(ctx context.Context, condition da.ContentCondition, user *entity.Operator) ([]*ContentProfile, error) {
	contentTypes := condition.ContentType
	if len(contentTypes) == 0 {
		contentTypes = []int{entity.ContentTypePlan, entity.ContentTypeMaterial, entity.ContentTypeAssets}
	}
	publishStatus := condition.PublishStatus
	if len(publishStatus) == 0 {
		publishStatus = []string{
			entity.ContentStatusPublished,
			entity.ContentStatusDraft,
			entity.ContentStatusPending,
			entity.ContentStatusRejected,
			entity.ContentStatusAttachment,
			entity.ContentStatusHidden,
			entity.ContentStatusArchive,}
	}
	visibilitySettings := VisibilitySettingsTypeContainsOrg

	if len(condition.VisibilitySettings) == 0 {
		err := c.querySchools(ctx, user)
		if err != nil {
			log.Error(ctx, "getVisibilitySettingsType failed",
				log.Err(err),
				log.Any("data", condition),
				log.Any("user", user))
			return nil, err
		}
		visibilitySetting, err := c.getVisibilitySettingsType(ctx, condition.VisibilitySettings, user)
		log.Error(ctx, "getVisibilitySettingsType failed",
			log.Err(err),
			log.Any("condition.VisibilitySettings", condition.VisibilitySettings),
			log.Any("user", user))
		visibilitySettings = visibilitySetting
		return nil, err
	}
	author := OwnerTypeOthers
	if len(condition.Author) != 0 {
		author = c.getOwnerType(ctx, condition.Author, user)
	}

	contentProfiles := make([]*ContentProfile, len(contentTypes) * len(publishStatus))
	for i := range contentTypes {
		for j := range publishStatus {
			contentProfiles[j + i * len(contentTypes)] = &ContentProfile{
				ContentType:        entity.ContentType(contentTypes[i]),
				Status:             entity.ContentPublishStatus(publishStatus[j]),
				VisibilitySettings: visibilitySettings,
				Owner:              author,
			}
		}
	}
	return contentProfiles, nil
}
func (c *ContentPermissionMySchoolModel) buildContentProfiles(ctx context.Context, content []*entity.ContentWithVisibilitySettings, user *entity.Operator) ([]*ContentProfile, error){
	profiles := make([]*ContentProfile, len(content))

	err := c.querySchools(ctx, user)
	if err != nil {
		log.Error(ctx, "getVisibilitySettingsType failed",
			log.Err(err),
			log.Any("content", content),
			log.Any("user", user))
		return nil, err
	}

	for i := range content {
		visibilitySettingType, err := c.getVisibilitySettingsType(ctx, content[i].VisibilitySettings, user)
		if err != nil {
			log.Error(ctx, "getVisibilitySettingsType failed",
				log.Err(err),
				log.Any("content", content))
			return nil, err
		}
		profiles[i] = &ContentProfile{
			ContentType:        content[i].ContentType,
			Status:             content[i].PublishStatus,
			VisibilitySettings: visibilitySettingType,
			Owner:              c.getOwnerType(ctx, content[i].Author, user),
		}
	}
	return profiles, nil
}

func (c *ContentPermissionMySchoolModel) querySchools(ctx context.Context, user *entity.Operator) error{
	//todo: complete it
	schools, err := external.GetSchoolServiceProvider().GetByOrganizationID(ctx, user, user.OrgID)
	if err != nil {
		log.Error(ctx, "GetByOrganizationID failed",
			log.Err(err),
			log.Any("user", user))
		return err
	}
	mySchools, err := external.GetSchoolServiceProvider().GetByOperator(ctx, user)
	if err != nil {
		log.Error(ctx, "GetByOperator failed",
			log.Err(err),
			log.Any("user", user))
		return err
	}
	c.mySchools = mySchools
	c.allSchools = schools
	return nil
}

func (c *ContentPermissionMySchoolModel) getVisibilitySettingsType(ctx context.Context, visibilitySettings []string, user *entity.Operator) (VisibilitySettingsType, error){
	for i := range visibilitySettings {
		if visibilitySettings[i] == user.OrgID {
			return VisibilitySettingsTypeContainsOrg, nil
		}
		if !c.checkInSchools(ctx, visibilitySettings[i], c.mySchools) {
			if c.checkInSchools(ctx, visibilitySettings[i], c.allSchools) {
				return VisibilitySettingsTypeAllSchools, nil
			}else{
				log.Warn(ctx, "visibility setting is not in all schools",
					log.Strings("visibilitySettings", visibilitySettings),
					log.Any("mySchool", c.mySchools),
					log.Any("allSchool", c.allSchools),
					log.Any("user", user))
				return VisibilitySettingsTypeContainsOrg, ErrInvalidVisibilitySetting
			}
		}
	}
	return VisibilitySettingsTypeMySchools, nil
}

func (c *ContentPermissionMySchoolModel) checkInSchools(ctx context.Context, visibilitySetting string, schools []*external.School) bool{
	for i := range schools {
		if schools[i].ID == visibilitySetting {
			return true
		}
	}
	return false
}

func (c *ContentPermissionMySchoolModel) getOwnerType(ctx context.Context, owner string, user *entity.Operator) OwnerType{
	if owner == user.UserID {
		return OwnerTypeUser
	}
	return OwnerTypeOthers
}

var (
	_contentPermissionMySchoolModel     IContentPermissionMySchoolModel
	_contentPermissionMySchoolModelOnce sync.Once
)

func GetContentPermissionMySchoolModel() IContentPermissionMySchoolModel {
	_contentPermissionMySchoolModelOnce.Do(func() {
		_contentPermissionMySchoolModel = new(ContentPermissionMySchoolModel)
	})
	return _contentPermissionMySchoolModel
}
