package model

import (
	"context"
	"errors"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	ErrNoAvailableVisibilitySettings = errors.New("no available visibilitySettings")
)

var (
	publishedContentFilter = PermissionList{
		ViewOrgPermission:       external.ViewOrgPublished215,
		ViewMySchoolPermission:  external.ViewMySchoolPublished218,
		ViewAllSchoolPermission: external.ViewAllSchoolsPublished227,
		ViewMyPermission:        external.ViewMyPublished214,
	}
	pendingContentFilter = PermissionList{
		ViewOrgPermission:       external.ViewOrgPending213,
		ViewMySchoolPermission:  external.ViewMySchoolPending225,
		ViewAllSchoolPermission: external.ViewAllSchoolsPending228,
		ViewMyPermission:        external.ViewMyPending212,
	}
	archivedContentFilter = PermissionList{
		ViewOrgPermission:       external.ViewOrgArchived217,
		ViewMySchoolPermission:  external.ViewMySchoolArchived226,
		ViewAllSchoolPermission: external.ViewAllSchoolsArchived229,
		ViewMyPermission:        external.ViewMyArchived216,
	}
)

type IContentFilterModel interface {
	FilterPublishContent(ctx context.Context, c *entity.ContentConditionRequest, user *entity.Operator) error
	FilterPendingContent(ctx context.Context, c *entity.ContentConditionRequest, user *entity.Operator) error
	FilterArchivedContent(ctx context.Context, c *entity.ContentConditionRequest, user *entity.Operator) error
	QueryUserSchools(ctx context.Context, user *entity.Operator) (*contentFilterUserSchoolInfo, error)
}

type PermissionList struct {
	ViewOrgPermission       external.PermissionName
	ViewMySchoolPermission  external.PermissionName
	ViewAllSchoolPermission external.PermissionName
	ViewMyPermission        external.PermissionName
}

func (p PermissionList) Array() []external.PermissionName {
	return []external.PermissionName{
		p.ViewOrgPermission,
		p.ViewMySchoolPermission,
		p.ViewAllSchoolPermission,
		p.ViewMyPermission,
	}
}

type contentFilterUserSchoolInfo struct {
	MySchool  []string
	AllSchool []string
}
type ContentFilterModel struct {
}

func (cf *ContentFilterModel) FilterPublishContent(ctx context.Context, c *entity.ContentConditionRequest, user *entity.Operator) error {
	log.Debug(ctx, "FilterPublishContent",
		log.Any("condition", c),
		log.Any("user", user))
	return cf.doFilterContent(ctx, c, publishedContentFilter, true, user)
}
func (cf *ContentFilterModel) FilterPendingContent(ctx context.Context, c *entity.ContentConditionRequest, user *entity.Operator) error {
	//check view_my_pending_212 permission for condition1 to fill PublishedQueryMode
	log.Debug(ctx, "FilterPendingContent",
		log.Any("condition", c),
		log.Any("user", user))

	return cf.doFilterContent(ctx, c, pendingContentFilter, false, user)
}
func (cf *ContentFilterModel) FilterArchivedContent(ctx context.Context, c *entity.ContentConditionRequest, user *entity.Operator) error {
	//check view_my_archived_216 permission to fill PublishedQueryMode
	log.Debug(ctx, "FilterArchivedContent",
		log.Any("condition", c),
		log.Any("user", user))
	return cf.doFilterContent(ctx, c, archivedContentFilter, true, user)
}

func (cf *ContentFilterModel) doFilterContent(ctx context.Context, c *entity.ContentConditionRequest,
	permissionList PermissionList, allowQueryAll bool, user *entity.Operator) error {
	//check view_my_published_214 permission to fill PublishedQueryMode
	//1.fill organizations & schools into visibility settings
	schoolsInfo, err := cf.QueryUserSchools(ctx, user)
	if err != nil {
		log.Error(ctx, "querySchools failed",
			log.Err(err),
			log.Any("user", user))
		return err
	}
	log.Debug(ctx, "querySchools result",
		log.Any("schoolsInfo", schoolsInfo),
		log.Any("permissionList", permissionList),
		log.Any("user", user))
	permissionMap, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, user, permissionList.Array())
	if err != nil {
		log.Error(ctx, "GetPermissionServiceProvider failed",
			log.Err(err),
			log.Any("permissionSet", permissionList.Array()),
			log.Any("user", user))
		return err
	}
	allowVisibilitySettings := make([]string, 0)
	if permissionMap[permissionList.ViewOrgPermission] {
		allowVisibilitySettings = append(allowVisibilitySettings, user.OrgID)
	}
	if permissionMap[permissionList.ViewMySchoolPermission] {
		allowVisibilitySettings = append(allowVisibilitySettings, schoolsInfo.MySchool...)
	}
	if permissionMap[permissionList.ViewAllSchoolPermission] {
		allowVisibilitySettings = append(allowVisibilitySettings, schoolsInfo.AllSchool...)
	}
	log.Debug(ctx, "HasOrganizationPermissions result",
		log.Any("allowVisibilitySettings", allowVisibilitySettings),
		log.Any("permissionMap", permissionMap),
		log.Any("user", user))
	//deduplicate
	allowVisibilitySettings = utils.SliceDeduplication(allowVisibilitySettings)
	log.Debug(ctx, "After SliceDeduplication result",
		log.Any("allowVisibilitySettings", allowVisibilitySettings),
		log.Any("permissionMap", permissionMap),
		log.Any("user", user))

	//filter user visibility settings
	c.VisibilitySettings = cf.filterVisibilitySettings(ctx, c.VisibilitySettings, allowVisibilitySettings)
	log.Debug(ctx, "filterVisibilitySettings result",
		log.Strings("c.VisibilitySettings", c.VisibilitySettings))

	//2. fill PublishedQueryMode
	if len(c.VisibilitySettings) < 1 && !permissionMap[external.ViewMyPublished214] {
		//if no available vs & no permission to view my own
		log.Warn(ctx, "ErrNoAvailableVisibilitySettings",
			log.Any("user", user),
			log.Any("permissionMap", permissionMap),
			log.Any("permissionList", permissionList),
			log.Any("Condition", c))
		return ErrNoAvailableVisibilitySettings
	}

	c.PublishedQueryMode = entity.PublishedQueryModeOnlyOthers
	if allowQueryAll {
		if permissionMap[permissionList.ViewMyPermission] && len(c.VisibilitySettings) > 0 {
			//user can search all
			c.PublishedQueryMode = entity.PublishedQueryModeAll
		} else if permissionMap[permissionList.ViewMyPermission] {
			c.PublishedQueryMode = entity.PublishedQueryModeOnlyOwner
		} else if len(c.VisibilitySettings) > 0 {
			c.PublishedQueryMode = entity.PublishedQueryModeOnlyOthers
		} else {
			c.PublishedQueryMode = entity.PublishedQueryModeNone
		}
	}

	log.Debug(ctx, "filter result",
		log.Strings("c.VisibilitySettings", c.VisibilitySettings),
		log.String("c.PublishedQueryMode", string(c.PublishedQueryMode)),
		log.Any("permissionMap", permissionMap))
	return nil
}

func (cf *ContentFilterModel) filterVisibilitySettings(ctx context.Context, queryVisibilitySettings, allowVisibilitySettings []string) []string {
	if len(queryVisibilitySettings) == 0 {
		return allowVisibilitySettings
	}
	res := make([]string, 0)
	for i := range queryVisibilitySettings {
		for j := range allowVisibilitySettings {
			//if queryVisibilitySettings is in allowVisibilitySettings, insert into result
			if queryVisibilitySettings[i] == allowVisibilitySettings[j] {
				res = append(res, queryVisibilitySettings[i])
				break
			}
		}
	}
	return res
}

func (cf *ContentFilterModel) QueryUserSchools(ctx context.Context, user *entity.Operator) (*contentFilterUserSchoolInfo, error) {
	//todo: complete it
	schools, err := external.GetSchoolServiceProvider().GetByOrganizationID(ctx, user, user.OrgID)
	if err != nil {
		log.Error(ctx, "GetByOrganizationID failed",
			log.Err(err),
			log.Any("user", user))
		return nil, err
	}
	mySchools, err := external.GetSchoolServiceProvider().GetByOperator(ctx, user)
	if err != nil {
		log.Error(ctx, "GetByOperator failed",
			log.Err(err),
			log.Any("user", user))
		return nil, err
	}
	schoolInfo := &contentFilterUserSchoolInfo{
		MySchool:  make([]string, len(mySchools)),
		AllSchool: make([]string, len(schools)),
	}
	for i := range schools {
		schoolInfo.AllSchool[i] = schools[i].ID
	}
	for i := range mySchools {
		schoolInfo.MySchool[i] = mySchools[i].ID
	}

	return schoolInfo, nil
}

var (
	_contentFilterModel     IContentFilterModel
	_contentFilterModelOnce sync.Once
)

func GetContentFilterModel() IContentFilterModel {
	_contentFilterModelOnce.Do(func() {
		_contentFilterModel = new(ContentFilterModel)
	})
	return _contentFilterModel
}
