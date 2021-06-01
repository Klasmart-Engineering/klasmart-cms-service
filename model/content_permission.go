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
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
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
	CheckQueryContentPermission(ctx context.Context, condition *entity.ContentConditionRequest, user *entity.Operator) (bool, error)

	CheckReviewContentPermission(ctx context.Context, isApprove bool, cids []string, user *entity.Operator) (bool, error)

	GetPermissionOrgs(ctx context.Context, permission external.PermissionName, op *entity.Operator) ([]entity.OrganizationOrSchool, error)
}

type ContentPermissionMySchoolModel struct {
}

func (c *ContentPermissionMySchoolModel) CheckCreateContentPermission(ctx context.Context, data entity.CreateContentRequest, user *entity.Operator) (bool, error) {
	schoolsInfo, err := GetContentFilterModel().QueryUserSchools(ctx, user)
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

	err = GetContentPermissionChecker().HasPermissionWithLogicalAnd(ctx, user, ContentPermissionModeCreate, []*ContentProfile{&profile})
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

	err = GetContentPermissionChecker().HasPermissionWithLogicalAnd(ctx, user, ContentPermissionModePublish, profiles)
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

	err = GetContentPermissionChecker().HasPermissionWithLogicalAnd(ctx, user, ContentPermissionModePublish, profiles)
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
	hasAuth, err := c.checkAuthedContent(ctx, cid, user)
	if err != nil {
		log.Error(ctx, "checkAuthedContent failed",
			log.Err(err),
			log.String("cid", cid),
			log.Any("user", user))
		return false, err
	}
	if hasAuth {
		return true, nil
	}

	profiles, err := c.buildViewContentProfileByID(ctx, cid, user)
	if err != nil {
		log.Error(ctx, "buildContentProfileByIDs result",
			log.Err(err),
			log.String("cid", cid),
			log.Any("user", user))
		return false, err
	}

	err = GetContentPermissionChecker().HasPermissionWithLogicalOr(ctx, user, ContentPermissionModeView, profiles)
	if err != nil {
		log.Error(ctx, "No permission",
			log.Err(err),
			log.Any("profiles", profiles),
			log.Any("user", user))
		return false, nil
	}
	return true, nil
}
func (c *ContentPermissionMySchoolModel) checkAuthedContent(ctx context.Context, cid string, user *entity.Operator) (bool, error) {
	records, err := GetAuthedContentRecordsModel().QueryRecordsList(ctx, dbo.MustGetDB(ctx), entity.SearchAuthedContentRequest{
		OrgIDs:     []string{user.OrgID},
		ContentIDs: []string{cid},
	}, user)
	if err != nil {
		log.Error(ctx, "No permission",
			log.Err(err),
			log.String("cid", cid),
			log.Any("user", user))
		return false, err
	}
	if len(records) > 0 {
		return true, nil
	}
	return false, nil
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

	err = GetContentPermissionChecker().HasPermissionWithLogicalAnd(ctx, user, ContentPermissionModeEdit, profiles)
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

	err = GetContentPermissionChecker().HasPermissionWithLogicalAnd(ctx, user, ContentPermissionModeRemove, profiles)
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

	var permissionSetList PermissionSetList
	if isApprove {
		permissionSetList, err = GetContentPermissionChecker().GetPermissionSetList(ctx, ContentPermissionModeApprove, profiles)
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
		permissionSetList, err = GetContentPermissionChecker().GetPermissionSetList(ctx, ContentPermissionModeReject, profiles)
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

	err = permissionSetList.HasPermissionWithLogicalAnd(ctx, user)
	if err != nil {
		log.Error(ctx, "No permission",
			log.Err(err),
			log.Any("profiles", profiles),
			log.Any("user", user))
		return false, nil
	}
	return true, nil
}

func (c *ContentPermissionMySchoolModel) CheckQueryContentPermission(ctx context.Context, condition *entity.ContentConditionRequest, user *entity.Operator) (bool, error) {
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

	err = GetContentPermissionChecker().HasPermissionWithLogicalAnd(ctx, user, ContentPermissionModeView, contentProfiles)
	if err != nil {
		log.Error(ctx, "No permission",
			log.Err(err),
			log.Any("profiles", contentProfiles),
			log.Any("user", user))
		return false, nil
	}
	return true, nil
}

func (c *ContentPermissionMySchoolModel) buildViewContentProfileByID(ctx context.Context, cid string, user *entity.Operator) ([]*ContentProfile, error) {
	contentList, err := GetContentModel().GetRawContentByIDListWithVisibilitySettings(ctx, dbo.MustGetDB(ctx), []string{cid})
	if err != nil {
		log.Error(ctx, "GetContentByIDList failed",
			log.Err(err),
			log.String("cid", cid),
			log.Any("user", user))
		return nil, err
	}
	log.Debug(ctx, "GetRawContentByIDListWithVisibilitySettings result",
		log.Any("contentList", contentList),
		log.String("cid", cid),
		log.Any("user", user))
	if len(contentList) < 1 {
		log.Warn(ctx, "content list is nil",
			log.String("cid", cid),
			log.Any("contentList", contentList),
			log.Any("user", user))
		return nil, ErrEmptyContentList
	}
	content := contentList[0]
	profiles, err := c.buildViewContentProfiles(ctx, content, user)
	if err != nil {
		log.Error(ctx, "buildContentProfiles failed",
			log.Err(err),
			log.Any("contentList", contentList),
			log.Any("user", user))
		return nil, err
	}
	log.Debug(ctx, "buildContentProfiles result",
		log.Any("contentList", contentList),
		log.String("cid", cid),
		log.Any("profiles", profiles))

	return profiles, nil
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

func (c *ContentPermissionMySchoolModel) buildByConditionContentProfiles(ctx context.Context, condition *entity.ContentConditionRequest, user *entity.Operator) ([]*ContentProfile, error) {
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
	publishAssetsStatus := condition.PublishStatus
	if len(publishStatus) == 0 {
		publishStatus = []string{
			entity.ContentStatusPublished,
			entity.ContentStatusDraft,
			entity.ContentStatusPending,
			entity.ContentStatusRejected,
			entity.ContentStatusAttachment,
			entity.ContentStatusHidden,
			entity.ContentStatusArchive}
		publishAssetsStatus = []string{
			entity.ContentStatusPublished,
		}
	}
	visibilitySettings := VisibilitySettingsTypeOrgWithAllSchools

	if len(condition.VisibilitySettings) != 0 {
		schoolsInfo, err := GetContentFilterModel().QueryUserSchools(ctx, user)
		if err != nil {
			log.Error(ctx, "getVisibilitySettingsType failed",
				log.Err(err),
				log.Any("data", condition),
				log.Any("user", user))
			return nil, err
		}
		visibilitySetting, err := c.getVisibilitySettingsType(ctx, condition.VisibilitySettings, schoolsInfo, user)
		if err != nil {
			log.Error(ctx, "getVisibilitySettingsType failed",
				log.Err(err),
				log.Any("condition.VisibilitySettings", condition.VisibilitySettings),
				log.Any("user", user))
			return nil, err
		}
		visibilitySettings = visibilitySetting
	}
	author := OwnerTypeOthers
	if condition.Author != "" {
		author = c.getOwnerType(ctx, condition.Author, user)
	}

	contentProfiles := make([]*ContentProfile, 0)
	for i := range contentTypes {
		tempPublishStatus := publishStatus
		if contentTypes[i] == entity.ContentTypeAssets {
			tempPublishStatus = publishAssetsStatus
		}
		for j := range tempPublishStatus {
			contentProfiles = append(contentProfiles, &ContentProfile{
				ContentType:        entity.ContentType(contentTypes[i]),
				Status:             entity.ContentPublishStatus(publishStatus[j]),
				VisibilitySettings: visibilitySettings,
				Owner:              author,
			})
		}
	}
	return contentProfiles, nil
}
func (c *ContentPermissionMySchoolModel) buildContentProfiles(ctx context.Context, content []*entity.ContentWithVisibilitySettings, user *entity.Operator) ([]*ContentProfile, error) {
	profiles := make([]*ContentProfile, 0)

	schoolsInfo, err := GetContentFilterModel().QueryUserSchools(ctx, user)
	if err != nil {
		log.Error(ctx, "getVisibilitySettingsType failed",
			log.Err(err),
			log.Any("content", content),
			log.Any("user", user))
		return nil, err
	}

	for i := range content {
		log.Debug(ctx, "getVisibilitySettingsType",
			log.Any("schoolsInfo", schoolsInfo),
			log.Any("content", content))
		visibilitySettingType, err := c.getVisibilitySettingsType(ctx, content[i].VisibilitySettings, schoolsInfo, user)
		if err != nil {
			log.Error(ctx, "getVisibilitySettingsType failed",
				log.Err(err),
				log.Any("content", content))
			return nil, err
		}
		log.Debug(ctx, "getVisibilitySettingsType result",
			log.Any("schoolsInfo", schoolsInfo),
			log.Any("visibilitySettingType", visibilitySettingType),
			log.Any("content", content))
		profiles = append(profiles, &ContentProfile{
			ContentType:        content[i].ContentType,
			Status:             content[i].PublishStatus,
			VisibilitySettings: visibilitySettingType,
			Owner:              c.getOwnerType(ctx, content[i].Author, user),
		})
	}
	return profiles, nil
}

func (c *ContentPermissionMySchoolModel) buildViewContentProfiles(ctx context.Context, content *entity.ContentWithVisibilitySettings, user *entity.Operator) ([]*ContentProfile, error) {
	profiles := make([]*ContentProfile, 0)

	schoolsInfo, err := GetContentFilterModel().QueryUserSchools(ctx, user)
	if err != nil {
		log.Error(ctx, "getVisibilitySettingsType failed",
			log.Err(err),
			log.Any("content", content),
			log.Any("user", user))
		return nil, err
	}

	visibilitySettingType := make([]VisibilitySettingsType, 0)

	log.Debug(ctx, "buildViewContentProfiles.getViewVisibilitySettingsType",
		log.Any("schoolsInfo", schoolsInfo),
		log.Any("content", content))
	vsType, err := c.getViewVisibilitySettingsType(ctx, content.VisibilitySettings, schoolsInfo, user)
	if err != nil {
		log.Error(ctx, "getViewVisibilitySettingsType failed",
			log.Err(err),
			log.Any("content", content))
		return nil, err
	}
	visibilitySettingType = append(visibilitySettingType, vsType...)

	log.Debug(ctx, "buildViewContentProfiles.getVisibilitySettingsType result",
		log.Any("schoolsInfo", schoolsInfo),
		log.Any("visibilitySettingType", visibilitySettingType),
		log.Any("content", content))
	for j := range visibilitySettingType {
		profiles = append(profiles, &ContentProfile{
			ContentType:        content.ContentType,
			Status:             content.PublishStatus,
			VisibilitySettings: visibilitySettingType[j],
			Owner:              c.getOwnerType(ctx, content.Author, user),
		})
	}
	return profiles, nil
}

func (c *ContentPermissionMySchoolModel) getVisibilitySettingsType(ctx context.Context, visibilitySettings []string, schoolInfo *contentFilterUserSchoolInfo, user *entity.Operator) (VisibilitySettingsType, error) {
	containsOrg := false
	containsOtherSchools := false
	containsSchools := false
	for i := range visibilitySettings {
		if visibilitySettings[i] == user.OrgID {
			//contains org
			containsOrg = true
		} else {
			containsSchools = true
			if !utils.ContainsStr(schoolInfo.MySchool, visibilitySettings[i]) {
				if utils.ContainsStr(schoolInfo.AllSchool, visibilitySettings[i]) {
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

func (c *ContentPermissionMySchoolModel) getViewVisibilitySettingsType(ctx context.Context, visibilitySettings []string, schoolInfo *contentFilterUserSchoolInfo, user *entity.Operator) ([]VisibilitySettingsType, error) {
	containsOrg := false
	containsOtherSchools := false
	containsSchools := false
	containsMySchool := false
	for i := range visibilitySettings {
		if visibilitySettings[i] == user.OrgID {
			//contains org
			containsOrg = true
		} else {
			containsSchools = true
			if !utils.ContainsStr(schoolInfo.MySchool, visibilitySettings[i]) {
				if utils.ContainsStr(schoolInfo.AllSchool, visibilitySettings[i]) {
					//contains other schools in org
					containsOtherSchools = true
				} else {
					log.Warn(ctx, "visibility setting is not in all schools",
						log.Strings("visibilitySettings", visibilitySettings),
						log.Any("mySchool", schoolInfo.MySchool),
						log.Any("allSchool", schoolInfo.AllSchool),
						log.Any("user", user))
					return []VisibilitySettingsType{VisibilitySettingsTypeAllSchools}, ErrInvalidVisibilitySetting
				}
			} else {
				containsMySchool = true
			}
		}
	}
	log.Info(ctx, "visibility settings check result",
		log.Strings("visibilitySettings", visibilitySettings),
		log.Bool("containsOrg", containsOrg),
		log.Bool("containsMySchool", containsMySchool),
		log.Bool("containsOtherSchools", containsOtherSchools),
		log.Bool("containsSchools", containsSchools))

	res := make([]VisibilitySettingsType, 0)

	//contains org
	if containsOrg {
		//contains other schools
		if containsOtherSchools {
			res = append(res, VisibilitySettingsTypeOrgWithAllSchools)
		}
		if !containsSchools {
			//only contains org
			res = append(res, VisibilitySettingsTypeOnlyOrg)
		}
		res = append(res, VisibilitySettingsTypeOrgWithMySchools)
	}

	//contains other schools but org
	if containsMySchool {
		res = append(res, VisibilitySettingsTypeMySchools)
	}
	if containsOtherSchools {
		res = append(res, VisibilitySettingsTypeAllSchools)
	}

	log.Info(ctx, "build visibility settings result",
		log.Strings("visibilitySettings", visibilitySettings),
		log.Bool("containsOrg", containsOrg),
		log.Bool("containsMySchool", containsMySchool),
		log.Bool("containsOtherSchools", containsOtherSchools),
		log.Bool("containsSchools", containsSchools),
		log.Any("res", res))
	//only contains my schools
	return res, nil
}

func (s *ContentPermissionMySchoolModel) GetPermissionOrgs(ctx context.Context, permission external.PermissionName, op *entity.Operator) ([]entity.OrganizationOrSchool, error) {
	schools, err := external.GetSchoolServiceProvider().GetByPermission(ctx, op, permission)
	if err != nil {
		log.Error(ctx, "get permission orgs failed", log.Err(err))
		return nil, err
	}
	entities := make([]entity.OrganizationOrSchool, 0)
	for i := range schools {
		entities = append(entities, entity.OrganizationOrSchool{
			ID:   schools[i].ID,
			Name: schools[i].Name,
		})
	}
	orgs, err := external.GetOrganizationServiceProvider().GetByPermission(ctx, op, permission)
	if err != nil || len(orgs) < 1 {
		log.Error(ctx, "get org info failed", log.Err(err))
		return nil, err
	}
	for i := range orgs {
		if orgs[i].ID == op.OrgID {
			entities = append(entities, entity.OrganizationOrSchool{
				ID:   op.OrgID,
				Name: orgs[0].Name,
			})
		}
	}

	return entities, nil
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
