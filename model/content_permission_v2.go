//content_permission_v2 is for new version of cms permission check
//add my_schools permission check
package model

import (
	"context"
	"errors"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var (
	ErrInvalidVisibilitySetting = errors.New("invalid visibility settings")
	ErrEmptyContentList         = errors.New("content list is nil")
)

type IContentPermissionMySchoolModel interface {
	CheckCreateContentPermission(ctx context.Context, data entity.CreateContentRequest, user *entity.Operator) (bool, error)
	CheckPublishContentsPermission(ctx context.Context, cid string, scopes []string, user *entity.Operator) (bool, error)
	CheckRepublishContentsPermission(ctx context.Context, cids []string, user *entity.Operator) (bool, error)

	CheckGetContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error)

	CheckUpdateContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error)
	CheckDeleteContentPermission(ctx context.Context, cids []string, user *entity.Operator) (bool, error)
	CheckQueryContentPermission(ctx context.Context, condition entity.ContentConditionRequest, user *entity.Operator) (bool, error)

	CheckReviewContentPermission(ctx context.Context, isApprove bool, cids []string, user *entity.Operator) (bool, error)
}

type ContentPermissionMySchoolModel struct {
}

func (c *ContentPermissionMySchoolModel) CheckCreateContentPermission(ctx context.Context, data entity.CreateContentRequest, user *entity.Operator) (bool, error) {
	schoolsInfo, err := querySchools(ctx, user)
	if err != nil {
		log.Error(ctx, "querySchools failed",
			log.Err(err),
			log.Any("data", data),
			log.Any("user", user))
		return false, err
	}
	log.Debug(ctx, "querySchools result",
		log.Any("schoolsInfo", schoolsInfo),
		log.Any("data", data),
		log.Any("user", user))

	visibilitySettingType, err := c.getVisibilitySettingsType(ctx, data.PublishScope, schoolsInfo, user)
	if err != nil {
		log.Error(ctx, "getVisibilitySettingsType failed",
			log.Err(err),
			log.Any("data", data),
			log.Any("user", user))
		return false, err
	}
	log.Debug(ctx, "visibilitySettingType result",
		log.Any("visibilitySettingType", visibilitySettingType))

	profile := ContentProfile{
		ContentType:        data.ContentType,
		Status:             entity.ContentStatusDraft,
		VisibilitySettings: visibilitySettingType,
		Owner:              OwnerTypeUser,
	}
	permissionSetList, err := NewContentPermissionTable().GetCreatePermissionSets(ctx, profile)
	if err == ErrUndefinedPermission {
		log.Error(ctx, "ErrUndefinedPermission",
			log.Err(err),
			log.Any("contentProfiles", profile),
			log.Any("user", user))
		return false, nil
	}
	if err != nil {
		log.Error(ctx, "GetCreatePermissionSets failed",
			log.Err(err),
			log.Any("profile", profile),
			log.Any("user", user))
		return false, err
	}

	log.Debug(ctx, "GetCreatePermissionSets result",
		log.Any("permissionSetList", permissionSetList))

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

func (c *ContentPermissionMySchoolModel) CheckRepublishContentsPermission(ctx context.Context, cids []string, user *entity.Operator) (bool, error) {
	profiles, err := c.buildContentProfileByIDs(ctx, cids, user)
	if err != nil {
		log.Debug(ctx, "buildContentProfileByIDs result",
			log.Strings("cids", cids),
			log.Any("user", user))
		return false, err
	}

	log.Debug(ctx, "buildContentProfiles result",
		log.Any("profiles", profiles),
		log.Strings("cids", cids))
	permissionSetList, err := NewContentPermissionTable().GetPublishPermissionSets(ctx, profiles)
	if err == ErrUndefinedPermission {
		log.Error(ctx, "ErrUndefinedPermission",
			log.Err(err),
			log.Any("contentProfiles", profiles),
			log.Any("user", user))
		return false, nil
	}
	if err != nil {
		return false, err
	}
	log.Debug(ctx, "GetPublishPermissionSets result",
		log.Any("permissionSetList", permissionSetList))
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

func (c *ContentPermissionMySchoolModel) CheckPublishContentsPermission(ctx context.Context, cid string, scopes []string, user *entity.Operator) (bool, error) {
	profiles, err := c.buildContentProfileByIDs(ctx, []string{cid}, user)
	if err != nil {
		log.Debug(ctx, "buildContentProfileByIDs result",
			log.String("cid", cid),
			log.Any("user", user))
		return false, err
	}

	log.Debug(ctx, "buildContentProfiles result",
		log.Any("profiles", profiles))
	permissionSetList, err := NewContentPermissionTable().GetPublishPermissionSets(ctx, profiles)
	if err == ErrUndefinedPermission {
		log.Error(ctx, "ErrUndefinedPermission",
			log.Err(err),
			log.Any("contentProfiles", profiles),
			log.Any("user", user))
		return false, nil
	}
	if err != nil {
		return false, err
	}
	log.Debug(ctx, "GetPublishPermissionSets result",
		log.Any("profiles", profiles),
		log.Any("permissionSetList", permissionSetList))

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

func (c *ContentPermissionMySchoolModel) CheckGetContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error) {
	profiles, err := c.buildContentProfileByIDs(ctx, []string{cid}, user)
	if err != nil {
		log.Debug(ctx, "buildContentProfileByIDs result",
			log.String("cid", cid),
			log.Any("user", user))
		return false, err
	}

	permissionSetList, err := NewContentPermissionTable().GetViewPermissionSets(ctx, profiles)
	if err == ErrUndefinedPermission {
		log.Error(ctx, "ErrUndefinedPermission",
			log.Err(err),
			log.Any("contentProfiles", profiles),
			log.Any("user", user))
		return false, nil
	}
	if err != nil {
		return false, err
	}
	log.Debug(ctx, "GetViewPermissionSets result",
		log.Any("permissionSetList", permissionSetList))

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

func (c *ContentPermissionMySchoolModel) CheckUpdateContentPermission(ctx context.Context, cid string, user *entity.Operator) (bool, error) {
	profiles, err := c.buildContentProfileByIDs(ctx, []string{cid}, user)
	if err != nil {
		log.Debug(ctx, "buildContentProfileByIDs result",
			log.String("cid", cid),
			log.Any("user", user))
		return false, err
	}

	log.Debug(ctx, "buildContentProfiles result",
		log.Any("profiles", profiles),
		log.String("cid", cid))

	permissionSetList, err := NewContentPermissionTable().GetEditPermissionSets(ctx, *profiles[0])
	if err == ErrUndefinedPermission {
		log.Error(ctx, "ErrUndefinedPermission",
			log.Err(err),
			log.Any("contentProfiles", profiles),
			log.Any("user", user))
		return false, nil
	}
	if err != nil {
		return false, err
	}
	log.Debug(ctx, "GetEditPermissionSets result",
		log.Any("profiles", profiles),
		log.Any("permissionSetList", permissionSetList))

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
func (c *ContentPermissionMySchoolModel) CheckDeleteContentPermission(ctx context.Context, cids []string, user *entity.Operator) (bool, error) {
	profiles, err := c.buildContentProfileByIDs(ctx, cids, user)
	if err != nil {
		log.Debug(ctx, "buildContentProfileByIDs result",
			log.Strings("cids", cids),
			log.Any("user", user))
		return false, err
	}
	log.Debug(ctx, "buildContentProfiles result",
		log.Strings("cids", cids),
		log.Any("profiles", profiles))

	permissionSetList, err := NewContentPermissionTable().GetRemovePermissionSets(ctx, profiles)
	if err == ErrUndefinedPermission {
		log.Error(ctx, "ErrUndefinedPermission",
			log.Err(err),
			log.Any("contentProfiles", profiles),
			log.Any("user", user))
		return false, nil
	}
	if err != nil {
		return false, err
	}
	log.Debug(ctx, "GetRemovePermissionSets result",
		log.Any("permissionSetList", permissionSetList),
		log.Any("profiles", profiles))

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

func (c *ContentPermissionMySchoolModel) CheckReviewContentPermission(ctx context.Context, isApprove bool, cids []string, user *entity.Operator) (bool, error) {
	profiles, err := c.buildContentProfileByIDs(ctx, cids, user)
	if err != nil {
		log.Debug(ctx, "buildContentProfileByIDs result",
			log.Strings("cids", cids),
			log.Any("user", user))
		return false, err
	}
	log.Debug(ctx, "buildContentProfiles result",
		log.Any("profiles", profiles),
		log.Strings("cids", cids))

	var permissionSetList IPermissionSet
	if isApprove {
		permissionSetList, err = NewContentPermissionTable().GetApprovePermissionSets(ctx, profiles)
		if err == ErrUndefinedPermission {
			log.Error(ctx, "ErrUndefinedPermission",
				log.Err(err),
				log.Any("contentProfiles", profiles),
				log.Any("user", user))
			return false, nil
		}
		if err != nil {
			return false, err
		}
	} else {
		permissionSetList, err = NewContentPermissionTable().GetRejectPermissionSets(ctx, profiles)
		if err == ErrUndefinedPermission {
			log.Error(ctx, "ErrUndefinedPermission",
				log.Err(err),
				log.Any("contentProfiles", profiles),
				log.Any("user", user))
			return false, nil
		}
		if err != nil {
			return false, err
		}
	}
	log.Debug(ctx, "GetReviewPermissionSets result",
		log.Any("permissionSetList", permissionSetList))

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

func (c *ContentPermissionMySchoolModel) CheckQueryContentPermission(ctx context.Context, condition entity.ContentConditionRequest, user *entity.Operator) (bool, error) {
	//if condition is self, set as user id
	if condition.Author == constant.Self {
		condition.Author = user.UserID
	}

	contentProfiles, err := c.buildByConditionContentProfiles(ctx, condition, user)
	if err != nil {
		log.Error(ctx, "buildByConditionContentProfiles Failed",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("user", user))
		return false, err
	}

	log.Debug(ctx, "buildByConditionContentProfiles result",
		log.Any("condition", condition),
		log.Any("contentProfiles", contentProfiles),
		log.Any("user", user))

	permissionSetList, err := NewContentPermissionTable().GetViewPermissionSets(ctx, contentProfiles)
	if err == ErrUndefinedPermission {
		log.Error(ctx, "ErrUndefinedPermission",
			log.Err(err),
			log.Any("contentProfiles", contentProfiles),
			log.Any("user", user))
		return false, nil
	}
	if err != nil {
		return false, err
	}
	log.Debug(ctx, "GetViewPermissionSets result",
		log.Any("permissionSetList", permissionSetList),
		log.Any("contentProfiles", contentProfiles),
		log.Any("user", user))
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

func (c *ContentPermissionMySchoolModel) buildContentProfileByIDs(ctx context.Context, cids []string, user *entity.Operator) ([]*ContentProfile, error) {
	contentList, err := GetContentModel().GetRawContentByIDListWithVisibilitySettings(ctx, dbo.MustGetDB(ctx), cids)
	if err != nil {
		log.Error(ctx, "GetContentByIDList failed",
			log.Err(err),
			log.Strings("cid", cids),
			log.Any("user", user))
		return nil, err
	}
	log.Debug(ctx, "GetRawContentByIDListWithVisibilitySettings result",
		log.Any("contentList", contentList),
		log.Strings("cid", cids),
		log.Any("user", user))
	if len(contentList) < 1 {
		log.Warn(ctx, "content list is nil",
			log.Strings("cid", cids),
			log.Any("contentList", contentList),
			log.Any("user", user))
		return nil, ErrEmptyContentList
	}
	profiles, err := c.buildContentProfiles(ctx, contentList, user)
	if err != nil {
		log.Error(ctx, "buildContentProfiles failed",
			log.Err(err),
			log.Any("contentList", contentList),
			log.Any("user", user))
		return nil, err
	}
	log.Debug(ctx, "buildContentProfiles result",
		log.Any("contentList", contentList),
		log.Strings("cid", cids),
		log.Any("profiles", profiles))

	return profiles, nil
}

func (c *ContentPermissionMySchoolModel) buildByConditionContentProfiles(ctx context.Context, condition entity.ContentConditionRequest, user *entity.Operator) ([]*ContentProfile, error) {
	contentTypes := make([]int, 0)
	for i := range condition.ContentType {
		ct := entity.NewContentType(condition.ContentType[i])
		if ct.Validate() == nil {
			contentTypes = append(contentTypes, condition.ContentType[i])
		}
	}

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
			entity.ContentStatusArchive}
	}
	visibilitySettings := VisibilitySettingsTypeOrgWithAllSchools

	if len(condition.VisibilitySettings) != 0 {
		schoolsInfo, err := querySchools(ctx, user)
		if err != nil {
			log.Error(ctx, "getVisibilitySettingsType failed",
				log.Err(err),
				log.Any("data", condition),
				log.Any("user", user))
			return nil, err
		}
		visibilitySetting, err := c.getVisibilitySettingsType(ctx, condition.VisibilitySettings, schoolsInfo, user)
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

	contentProfiles := make([]*ContentProfile, len(contentTypes)*len(publishStatus))
	for i := range contentTypes {
		for j := range publishStatus {
			contentProfiles[j+i*len(publishStatus)] = &ContentProfile{
				ContentType:        entity.ContentType(contentTypes[i]),
				Status:             entity.ContentPublishStatus(publishStatus[j]),
				VisibilitySettings: visibilitySettings,
				Owner:              author,
			}
		}
	}
	return contentProfiles, nil
}
func (c *ContentPermissionMySchoolModel) buildContentProfiles(ctx context.Context, content []*entity.ContentWithVisibilitySettings, user *entity.Operator) ([]*ContentProfile, error) {
	profiles := make([]*ContentProfile, len(content))

	schoolsInfo, err := querySchools(ctx, user)
	if err != nil {
		log.Error(ctx, "getVisibilitySettingsType failed",
			log.Err(err),
			log.Any("content", content),
			log.Any("user", user))
		return nil, err
	}

	for i := range content {
		visibilitySettingType, err := c.getVisibilitySettingsType(ctx, content[i].VisibilitySettings, schoolsInfo, user)
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

func (c *ContentPermissionMySchoolModel) getVisibilitySettingsType(ctx context.Context, visibilitySettings []string, schoolInfo *SchoolInfo, user *entity.Operator) (VisibilitySettingsType, error) {
	containsOrg := false
	containsOtherSchools := false
	containsSchools := false
	for i := range visibilitySettings {
		if visibilitySettings[i] == user.OrgID {
			//contains org
			containsOrg = true
		} else {
			containsSchools = true
		}
		if !c.checkInSchools(ctx, visibilitySettings[i], schoolInfo.MySchool) {
			if c.checkInSchools(ctx, visibilitySettings[i], schoolInfo.AllSchool) {
				//contains other schools in org
				containsOtherSchools = true
			} else {
				log.Warn(ctx, "visibility setting is not in all schools",
					log.Strings("visibilitySettings", visibilitySettings),
					log.Any("mySchool", schoolInfo.MySchool),
					log.Any("allSchool", schoolInfo.AllSchool),
					log.Any("user", user))
				return VisibilitySettingsTypeAllSchools, ErrInvalidVisibilitySetting
			}
		}
	}
	log.Info(ctx, "visibility settings check result",
		log.Strings("visibilitySettings", visibilitySettings),
		log.Bool("containsOrg", containsOrg),
		log.Bool("containsOtherSchools", containsOtherSchools),
		log.Bool("containsSchools", containsSchools))

	//contains org
	if containsOrg {
		//contains other schools
		if containsOtherSchools {
			return VisibilitySettingsTypeOrgWithAllSchools, nil
		}
		if !containsSchools {
			//only contains org
			return VisibilitySettingsTypeOnlyOrg, nil
		}
		return VisibilitySettingsTypeOrgWithMySchools, nil
	}

	//contains other schools but org
	if containsOtherSchools {
		return VisibilitySettingsTypeAllSchools, nil
	}
	//only contains my schools
	return VisibilitySettingsTypeMySchools, nil
}

func (c *ContentPermissionMySchoolModel) checkInSchools(ctx context.Context, visibilitySetting string, schools []string) bool {
	for i := range schools {
		if schools[i] == visibilitySetting {
			return true
		}
	}
	return false
}

func (c *ContentPermissionMySchoolModel) getOwnerType(ctx context.Context, owner string, user *entity.Operator) OwnerType {
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
