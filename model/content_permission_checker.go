package model

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

var (
	ErrHasNoPermission     = errors.New("has no permission")
	ErrUndefinedPermission = errors.New("undefined permission")
)

const (
	VisibilitySettingsTypeMySchools         VisibilitySettingsType = "my schools"
	VisibilitySettingsTypeAllSchools        VisibilitySettingsType = "all schools"
	VisibilitySettingsTypeOnlyOrg           VisibilitySettingsType = "only org"
	VisibilitySettingsTypeOrgWithMySchools  VisibilitySettingsType = "org and my schools"
	VisibilitySettingsTypeOrgWithAllSchools VisibilitySettingsType = "org and all schools"
)

const (
	OwnerTypeUser   OwnerType = "user"
	OwnerTypeOthers OwnerType = "others"
)

const (
	ContentPermissionModeCreate  ContentPermissionMode = "create"
	ContentPermissionModeEdit    ContentPermissionMode = "edit"
	ContentPermissionModePublish ContentPermissionMode = "publish"
	ContentPermissionModeView    ContentPermissionMode = "view"
	ContentPermissionModeRemove  ContentPermissionMode = "remove"
	ContentPermissionModeApprove ContentPermissionMode = "approve"
	ContentPermissionModeReject  ContentPermissionMode = "reject"
)

// Permissions includes edit, remove, publish, approve, reject action. Update synchronously with the permission definition
var allContentPermissionSet []external.PermissionName = []external.PermissionName{
	external.DeleteAsset340,
	external.CreateLessonMaterial220,
	external.CreateLessonPlan221,
	external.CreateMySchoolsContent223,
	external.CreateAllSchoolsContent224,
	external.EditMyPublishedContent234,
	external.EditOrgPublishedContent235,
	external.EditMySchoolsPublished247,
	external.EditAllSchoolsPublished249,
	external.EditLessonMaterialMetadataAndContent236,
	external.EditLessonPlanMetadata237,
	external.EditLessonPlanContent238,
	external.EditMyUnpublishedContent230,
	external.RemoveOrgPublishedContent254,
	external.RemoveMySchoolsPublished242,
	external.RemoveAllSchoolsPublished245,
	external.DeleteOrgArchivedContent253,
	external.DeleteMySchoolsArchived243,
	external.DeleteAllSchoolsArchived246,
	external.DeleteMyUnpublishedContent240,
	external.DeleteMyPending251,
	external.DeleteOrgPendingContent252,
	external.DeleteMySchoolsPending241,
	external.DeleteAllSchoolsPending244,
	external.ApprovePendingContent271,
	external.RejectPendingContent272,
	external.RepublishArchivedContent274,
	external.FullContentMmanagement294,
}

type VisibilitySettingsType string
type OwnerType string
type ContentPermissionMode string

type ContentPermissionChecker interface {
	// HasPermission check operator has permission of all content profiles for mode or not
	HasPermissionWithLogicalAnd(ctx context.Context, operator *entity.Operator, mode ContentPermissionMode, req []*ContentProfile) error

	// HasPermission check operator has permission of any content profile for mode or not
	HasPermissionWithLogicalOr(ctx context.Context, operator *entity.Operator, mode ContentPermissionMode, req []*ContentProfile) error

	// BatchGetContentPermissions Get permissions of contents, return with a map, the key of the map is the content id,
	// The value of the map is content permission
	BatchGetContentPermission(ctx context.Context, operator *entity.Operator, reqs []*ContentEntityProfile) (map[string]entity.ContentPermission, error)

	// Get permission list of all content profiles for mode
	GetPermissionSetList(ctx context.Context, mode ContentPermissionMode, req []*ContentProfile) (PermissionSetList, error)
}

type contentPermissionChecker struct {
	contentPermissionDict map[ContentPermissionMode]map[ContentProfile][]*PermissionSet
}

var (
	_contentPermissionChecker     ContentPermissionChecker
	_contentPermissionCheckerOnce sync.Once
)

func GetContentPermissionChecker() ContentPermissionChecker {
	_contentPermissionCheckerOnce.Do(func() {
		contentPermissionChecker := &contentPermissionChecker{}
		contentPermissionChecker.loadContentPermissionDict()

		_contentPermissionChecker = contentPermissionChecker
	})
	return _contentPermissionChecker
}

func (c *contentPermissionChecker) loadContentPermissionDict() {
	var dict map[ContentPermissionMode]map[ContentProfile][]*PermissionSet

	err := json.Unmarshal(constant.ContentPermissionJsonData, &dict)
	if err != nil {
		log.Panic(context.TODO(), "invalid content permission data", log.Err(err))
	}

	c.contentPermissionDict = dict
}

func (c *contentPermissionChecker) HasPermissionWithLogicalAnd(ctx context.Context, operator *entity.Operator, mode ContentPermissionMode, req []*ContentProfile) error {
	if len(req) < 1 {
		log.Error(ctx, "undefined permission",
			log.Any("mode", mode),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Err(ErrUndefinedPermission))
		return ErrUndefinedPermission
	}

	permissionSetList, err := c.GetPermissionSetList(ctx, mode, req)
	if err != nil {
		return err
	}

	return permissionSetList.HasPermissionWithLogicalAnd(ctx, operator)
}

func (c *contentPermissionChecker) HasPermissionWithLogicalOr(ctx context.Context, operator *entity.Operator, mode ContentPermissionMode, req []*ContentProfile) error {
	if len(req) < 1 {
		log.Error(ctx, "undefined permission",
			log.Any("mode", mode),
			log.Any("operator", operator),
			log.Any("req", req),
			log.Err(ErrUndefinedPermission))
		return ErrUndefinedPermission
	}

	permissionSetList, err := c.GetPermissionSetList(ctx, mode, req)
	if err != nil {
		return err
	}

	return permissionSetList.HasPermissionWithLogicalOr(ctx, operator)
}

func (c *contentPermissionChecker) BatchGetContentPermission(ctx context.Context, operator *entity.Operator, req []*ContentEntityProfile) (map[string]entity.ContentPermission, error) {
	contentPermissionDict, err := c.getOperatorContentPermission(ctx, operator)
	if err != nil {
		log.Error(ctx, "getOperatorContentPermission failed",
			log.Any("operator", operator),
			log.Err(err))
		return nil, err
	}

	result := make(map[string]entity.ContentPermission, len(req))
	for _, r := range req {
		pm := contentPermissionDict[r.ContentProfile]
		result[r.ID] = entity.ContentPermission{
			ID:             r.ID,
			AllowEdit:      pm[ContentPermissionModeEdit],
			AllowDelete:    pm[ContentPermissionModeRemove],
			AllowApprove:   pm[ContentPermissionModeApprove],
			AllowReject:    pm[ContentPermissionModeReject],
			AllowRepublish: pm[ContentPermissionModePublish],
		}
	}

	return result, nil
}

func (c *contentPermissionChecker) getOperatorContentPermission(ctx context.Context, operator *entity.Operator) (map[ContentProfile]map[ContentPermissionMode]bool, error) {
	result := make(map[ContentProfile]map[ContentPermissionMode]bool)

	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, operator, allContentPermissionSet)
	if err != nil {
		log.Error(ctx, "HasOrganizationPermissions failed",
			log.Any("operator", operator),
			log.Any("permissions", allContentPermissionSet),
			log.Err(err))
		return nil, err
	}

	c.getOperatorContentPermissionByMode(ContentPermissionModeEdit, result, hasPermission)
	c.getOperatorContentPermissionByMode(ContentPermissionModePublish, result, hasPermission)
	c.getOperatorContentPermissionByMode(ContentPermissionModeApprove, result, hasPermission)
	c.getOperatorContentPermissionByMode(ContentPermissionModeReject, result, hasPermission)
	c.getOperatorContentPermissionByMode(ContentPermissionModeRemove, result, hasPermission)

	return result, nil
}

func (c *contentPermissionChecker) getOperatorContentPermissionByMode(mode ContentPermissionMode, result map[ContentProfile]map[ContentPermissionMode]bool, hasPermission map[external.PermissionName]bool) {
	for k, v := range c.contentPermissionDict[mode] {
		if _, ok := result[k]; !ok {
			result[k] = make(map[ContentPermissionMode]bool)
		}

		allow := true
		for i := range v {
			allow = true
			for _, v := range v[i].Permissions {
				if val, ok := hasPermission[v]; !ok || !val {
					allow = false
					break
				}
			}

			if allow == true {
				break
			}
		}
		result[k][mode] = allow
	}
}

type PermissionSet struct {
	Permissions []external.PermissionName `json:"permissions"`
}

type PermissionSetList [][]*PermissionSet

func (c *contentPermissionChecker) GetPermissionSetList(ctx context.Context, mode ContentPermissionMode, req []*ContentProfile) (PermissionSetList, error) {
	var dict map[ContentProfile][]*PermissionSet
	dict, ok := c.contentPermissionDict[mode]
	if !ok {
		log.Error(ctx, "undefined permission",
			log.Any("mode", mode),
			log.Any("req", req),
			log.Any("dict", c.contentPermissionDict),
			log.Err(ErrUndefinedPermission))
		return nil, ErrUndefinedPermission
	}

	set := make(map[ContentProfile]struct{})
	permissionSetList := make([][]*PermissionSet, 0)

	for i := range req {
		if req[i] == nil {
			continue
		}

		key := *req[i]
		if _, ok := set[key]; !ok {
			if _, ok := dict[key]; !ok {
				log.Error(ctx, "undefined permission",
					log.Any("contentProfile", key),
					log.Any("mode", mode),
					log.Any("permissionDict", dict),
					log.Err(ErrUndefinedPermission))
				return nil, ErrUndefinedPermission
			}

			set[key] = struct{}{}
			permissionSetList = append(permissionSetList, dict[key])
		}
	}

	return PermissionSetList(permissionSetList), nil
}

// The AND of a set of external.PermissionNamehave access if and only if all of its permissions have access
// The OR of a set of permissionSet have access if and only if all of its permissions have access
// The AND of a set of []*permissionSet have access if and only if all of its permissionSets have access
func (p PermissionSetList) HasPermissionWithLogicalAnd(ctx context.Context, operator *entity.Operator) error {
	if len(p) < 1 {
		return nil
	}

	permissions := p.getAllPermissions()

	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, operator, permissions)
	if err != nil {
		log.Error(ctx, "HasOrganizationPermissions failed",
			log.Any("operator", operator),
			log.Any("permissions", permissions),
			log.Err(err))
		return err
	}

	for _, pmSetList := range p {
		for _, pmSet := range pmSetList {
			err = nil
			for _, pm := range pmSet.Permissions {
				if val, ok := hasPermission[pm]; !ok || !val {
					err = ErrHasNoPermission
					break
				}
			}

			if err == nil {
				break
			}
		}

		if err != nil {
			return ErrHasNoPermission
		}
	}

	return nil
}

// The AND of a set of external.PermissionNamehave access if and only if all of its permissions have access
// The OR of a set of permissionSet have access if and only if all of its permissions have access
// The OR of a set of []*permissionSet have access if any permissionSets have access
func (p PermissionSetList) HasPermissionWithLogicalOr(ctx context.Context, operator *entity.Operator) error {
	if len(p) < 1 {
		return nil
	}

	permissions := p.getAllPermissions()

	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, operator, permissions)
	if err != nil {
		log.Error(ctx, "HasOrganizationPermissions failed",
			log.Any("operator", operator),
			log.Any("permissions", permissions),
			log.Err(err))
		return err
	}

	for _, pmSetList := range p {
		for _, pmSet := range pmSetList {
			err = nil
			for _, pm := range pmSet.Permissions {
				if val, ok := hasPermission[pm]; !ok || !val {
					err = ErrHasNoPermission
					break
				}
			}

			if err == nil {
				return nil
			}
		}
	}

	return ErrHasNoPermission
}

func (p PermissionSetList) getAllPermissions() []external.PermissionName {
	result := make([]external.PermissionName, 0)

	for _, pmSetList := range p {
		for _, pmSet := range pmSetList {
			for _, pm := range pmSet.Permissions {
				result = append(result, pm)
			}
		}
	}

	return result
}

type ContentProfile struct {
	ContentType        entity.ContentType          `json:"contentType"` //ContentType using entity.Content type
	Status             entity.ContentPublishStatus `json:"status"`      //ContentStatus using entity.ContentPublishStatus type
	VisibilitySettings VisibilitySettingsType      `json:"visibilitySettings"`
	Owner              OwnerType                   `json:"owner"`
}

// Implement TextUnmarshaler in order to support Unmarshal when key of map is a struct
func (s *ContentProfile) UnmarshalText(text []byte) error {
	type x ContentProfile
	return json.Unmarshal(text, (*x)(s))
}

type ContentEntityProfile struct {
	ContentProfile
	ID string `json:"id"`
}
