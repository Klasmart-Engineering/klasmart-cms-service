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
	VisibilitySettingsTypeMySchools         VisibilitySettingsType = 1
	VisibilitySettingsTypeAllSchools        VisibilitySettingsType = 2
	VisibilitySettingsTypeOnlyOrg           VisibilitySettingsType = 4
	VisibilitySettingsTypeOrgWithMySchools  VisibilitySettingsType = 5
	VisibilitySettingsTypeOrgWithAllSchools VisibilitySettingsType = 6
)

const (
	OwnerTypeUser   OwnerType = 1
	OwnerTypeOthers OwnerType = 2
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

type VisibilitySettingsType int
type OwnerType int

type IContentPermissionTable interface {
	//GetCreatePermissionSets Get permission set list when a user want to create a content
	GetCreatePermissionSets(ctx context.Context, req ContentProfile) (IPermissionSet, error)

	//GetEditPermissionSets Get permission set list when a user want to edit a content
	GetEditPermissionSets(ctx context.Context, req ContentProfile) (IPermissionSet, error)

	//GetPublishPermissionSets Get permission set list when a user want to publish contents
	//ContentProfile list indicates the batch of Content profiles, permission set list must
	//contains all content profiles
	GetPublishPermissionSets(ctx context.Context, req []*ContentProfile) (IPermissionSet, error)

	//GetViewPermissionSets Get permission set list when a user want to view contents
	GetViewPermissionSets(ctx context.Context, req []*ContentProfile) (IPermissionSet, error)

	//GetRemovePermissionSets Get permission set list when a user want to remove contents
	GetRemovePermissionSets(ctx context.Context, req []*ContentProfile) (IPermissionSet, error)

	//GetApprovePermissionSets Get permission set list when a user want to approve contents
	GetApprovePermissionSets(ctx context.Context, req []*ContentProfile) (IPermissionSet, error)

	//GetRejectPermissionSets Get permission set list when a user want to reject contents
	GetRejectPermissionSets(ctx context.Context, req []*ContentProfile) (IPermissionSet, error)

	//BatchGetContentPermissions Get permissions of contents, return with a map, the key of the map is the content id,
	//The value of the map is content permission
	BatchGetContentPermissions(ctx context.Context, operator *entity.Operator, reqs []*ContentEntityProfile) (map[string]entity.ContentPermission, error)
}

type IPermissionSet interface {
	HasPermission(ctx context.Context, operator *entity.Operator) error
}

type PermissionSet struct {
	Permissions []external.PermissionName `json:"permissions"`
}

// The AND of a set of permissions have access if and only if all of its permissions have access
func (p PermissionSet) HasPermission(ctx context.Context, operator *entity.Operator) error {
	if len(p.Permissions) < 1 {
		return nil
	}

	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, operator, p.Permissions)
	if err != nil {
		log.Warn(ctx, "HasOrganizationPermissions failed",
			log.Any("operator", operator),
			log.Any("permissions", p.Permissions),
			log.Err(err))
		return err
	}

	for _, v := range p.Permissions {
		if val, ok := hasPermission[v]; ok && val {
			continue
		}

		return ErrHasNoPermission
	}

	return nil
}

type PermissionSets []*PermissionSet

// The OR of a set of permissionSet have no permission if and only if all of its permissionSet have no permission
func (p PermissionSets) HasPermission(ctx context.Context, operator *entity.Operator) error {
	if len(p) < 1 {
		return nil
	}

	permissions := p.getAllPermissions()

	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, operator, permissions)
	if err != nil {
		log.Warn(ctx, "HasOrganizationPermissions failed",
			log.Any("operator", operator),
			log.Any("permissions", permissions),
			log.Err(err))
		return err
	}

	for i := range p {
		err = nil
		for _, v := range p[i].Permissions {
			if val, ok := hasPermission[v]; !ok || !val {
				err = ErrHasNoPermission
				break
			}
		}

		if err == nil {
			return nil
		}
	}

	return ErrHasNoPermission
}

func (p PermissionSets) getAllPermissions() []external.PermissionName {
	result := make([]external.PermissionName, 0)
	set := make(map[external.PermissionName]struct{})

	for i := range p {
		for _, v := range p[i].Permissions {
			if _, ok := set[v]; !ok {
				set[v] = struct{}{}
				result = append(result, v)
			}
		}
	}

	return result
}

type PermissionSetsList [][]*PermissionSet

// The AND of a set of permissionSets have access if and only if all of its permissionSets have access
func (p PermissionSetsList) HasPermission(ctx context.Context, operator *entity.Operator) error {
	if len(p) < 1 {
		return nil
	}

	permissions := p.getAllPermissions()

	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, operator, permissions)
	if err != nil {
		log.Warn(ctx, "HasOrganizationPermissions failed",
			log.Any("operator", operator),
			log.Any("permissions", permissions),
			log.Err(err))
		return err
	}

	for _, permissionSets := range p {
		for _, permissionSet := range permissionSets {
			err = nil
			for _, permission := range permissionSet.Permissions {
				if val, ok := hasPermission[permission]; !ok || !val {
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

func (p PermissionSetsList) getAllPermissions() []external.PermissionName {
	result := make([]external.PermissionName, 0)
	set := make(map[external.PermissionName]struct{})

	for _, permissionSets := range p {
		for _, permissionSet := range permissionSets {
			for _, v := range permissionSet.Permissions {
				if _, ok := set[v]; !ok {
					set[v] = struct{}{}
					result = append(result, v)
				}
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

type ContentEntityProfile struct {
	ContentProfile
	ID string `json:"id"`
}

// Implement TextUnmarshaler in order to support Unmarshal when key of map is a struct
func (s *ContentProfile) UnmarshalText(text []byte) error {
	type x ContentProfile
	return json.Unmarshal(text, (*x)(s))
}

type ContentPermissionTable struct {
	createPermissionDict  map[ContentProfile][]*PermissionSet
	editPermissionDict    map[ContentProfile][]*PermissionSet
	publishPermissionDict map[ContentProfile][]*PermissionSet
	viewPermissionDict    map[ContentProfile][]*PermissionSet
	removePermissionDict  map[ContentProfile][]*PermissionSet
	approvePermissionDict map[ContentProfile][]*PermissionSet
	rejectPermissionDict  map[ContentProfile][]*PermissionSet
}

var (
	_contentPermissionTable     IContentPermissionTable
	_contentPermissionTableOnce sync.Once
)

func NewContentPermissionTable() IContentPermissionTable {
	_contentPermissionTableOnce.Do(func() {
		contentPermissionTable := &ContentPermissionTable{}
		contentPermissionTable.loadContentPermissionDict()

		_contentPermissionTable = contentPermissionTable
	})
	return _contentPermissionTable
}

func (c *ContentPermissionTable) loadContentPermissionDict() {
	permissionDict := struct {
		CreatePermissionDict  map[ContentProfile][]*PermissionSet `json:"createPermissionDict"`
		EditPermissionDict    map[ContentProfile][]*PermissionSet `json:"editPermissionDict"`
		PublishPermissionDict map[ContentProfile][]*PermissionSet `json:"publishPermissionDict"`
		ViewPermissionDict    map[ContentProfile][]*PermissionSet `json:"viewPermissionDict"`
		RemovePermissionDict  map[ContentProfile][]*PermissionSet `json:"removePermissionDict"`
		ApprovePermissionDict map[ContentProfile][]*PermissionSet `json:"approvePermissionDict"`
		RejectPermissionDict  map[ContentProfile][]*PermissionSet `json:"rejectPermissionDict"`
	}{}

	err := json.Unmarshal(constant.ContentPermissionJsonData, &permissionDict)
	if err != nil {
		log.Error(context.TODO(), "invalid content permission data", log.Err(err))
		panic(err)
	}

	c.createPermissionDict = permissionDict.CreatePermissionDict
	c.editPermissionDict = permissionDict.EditPermissionDict
	c.publishPermissionDict = permissionDict.PublishPermissionDict
	c.viewPermissionDict = permissionDict.ViewPermissionDict
	c.removePermissionDict = permissionDict.RemovePermissionDict
	c.approvePermissionDict = permissionDict.ApprovePermissionDict
	c.rejectPermissionDict = permissionDict.RejectPermissionDict
}

func (c *ContentPermissionTable) GetCreatePermissionSets(ctx context.Context, req ContentProfile) (IPermissionSet, error) {
	if v, ok := c.createPermissionDict[req]; ok {
		return PermissionSets(v), nil
	}

	log.Warn(context.TODO(), "undefined permission",
		log.Any("ContentType", req),
		log.Err(ErrUndefinedPermission))
	return nil, ErrUndefinedPermission
}

func (c *ContentPermissionTable) GetEditPermissionSets(ctx context.Context, req ContentProfile) (IPermissionSet, error) {
	if v, ok := c.editPermissionDict[req]; ok {
		return PermissionSets(v), nil
	}

	log.Warn(context.TODO(), "undefined permission",
		log.Any("ContentType", req),
		log.Err(ErrUndefinedPermission))
	return nil, ErrUndefinedPermission
}

func (c *ContentPermissionTable) GetPublishPermissionSets(ctx context.Context, req []*ContentProfile) (IPermissionSet, error) {
	set := make(map[ContentProfile]struct{})
	permissionSetsList := make([][]*PermissionSet, 0)

	for i := range req {
		key := *req[i]
		if _, ok := set[key]; !ok {
			if _, ok := c.publishPermissionDict[key]; !ok {
				log.Warn(context.TODO(), "undefined permission",
					log.Any("ContentType", key),
					log.Err(ErrUndefinedPermission))
				return nil, ErrUndefinedPermission
			}

			set[key] = struct{}{}
			permissionSetsList = append(permissionSetsList, c.publishPermissionDict[key])
		}
	}

	return PermissionSetsList(permissionSetsList), nil
}

func (c *ContentPermissionTable) GetViewPermissionSets(ctx context.Context, req []*ContentProfile) (IPermissionSet, error) {
	set := make(map[ContentProfile]struct{})
	permissionSetsList := make([][]*PermissionSet, 0)

	for i := range req {
		key := *req[i]
		if _, ok := set[key]; !ok {
			if _, ok := c.viewPermissionDict[key]; !ok {
				log.Warn(context.TODO(), "undefined permission",
					log.Any("ContentType", key),
					log.Err(ErrUndefinedPermission))
				return nil, ErrUndefinedPermission
			}

			set[key] = struct{}{}
			permissionSetsList = append(permissionSetsList, c.viewPermissionDict[key])
		}
	}

	return PermissionSetsList(permissionSetsList), nil
}

func (c *ContentPermissionTable) GetRemovePermissionSets(ctx context.Context, req []*ContentProfile) (IPermissionSet, error) {
	set := make(map[ContentProfile]struct{})
	permissionSetsList := make([][]*PermissionSet, 0)

	for i := range req {
		key := *req[i]
		if _, ok := set[key]; !ok {
			if _, ok := c.removePermissionDict[key]; !ok {
				log.Warn(context.TODO(), "undefined permission",
					log.Any("ContentType", key),
					log.Err(ErrUndefinedPermission))
				return nil, ErrUndefinedPermission
			}

			set[key] = struct{}{}
			permissionSetsList = append(permissionSetsList, c.removePermissionDict[key])
		}
	}

	return PermissionSetsList(permissionSetsList), nil
}

func (c *ContentPermissionTable) GetApprovePermissionSets(ctx context.Context, req []*ContentProfile) (IPermissionSet, error) {
	set := make(map[ContentProfile]struct{})
	permissionSetsList := make([][]*PermissionSet, 0)

	for i := range req {
		key := *req[i]
		if _, ok := set[key]; !ok {
			if _, ok := c.approvePermissionDict[key]; !ok {
				log.Warn(context.TODO(), "undefined permission",
					log.Any("ContentType", key),
					log.Err(ErrUndefinedPermission))
				return nil, ErrUndefinedPermission
			}

			set[key] = struct{}{}
			permissionSetsList = append(permissionSetsList, c.approvePermissionDict[key])
		}
	}

	return PermissionSetsList(permissionSetsList), nil
}

func (c *ContentPermissionTable) GetRejectPermissionSets(ctx context.Context, req []*ContentProfile) (IPermissionSet, error) {
	set := make(map[ContentProfile]struct{})
	permissionSetsList := make([][]*PermissionSet, 0)

	for i := range req {
		key := *req[i]
		if _, ok := set[key]; !ok {
			if _, ok := c.rejectPermissionDict[key]; !ok {
				log.Warn(context.TODO(), "undefined permission",
					log.Any("ContentType", key),
					log.Err(ErrUndefinedPermission))
				return nil, ErrUndefinedPermission
			}

			set[key] = struct{}{}
			permissionSetsList = append(permissionSetsList, c.rejectPermissionDict[key])
		}
	}

	return PermissionSetsList(permissionSetsList), nil
}

type contentPermission struct {
	allowEdit      bool
	allowDelete    bool
	allowApprove   bool
	allowReject    bool
	allowRepublish bool
}

//BatchGetContentPermissions Get permissions of contents, return with a map, the key of the map is the content id,
//The value of the map is content permission
func (c *ContentPermissionTable) BatchGetContentPermissions(ctx context.Context, operator *entity.Operator, req []*ContentEntityProfile) (map[string]entity.ContentPermission, error) {
	contentPermissionDict, err := c.getOperatorContentPermission(ctx, operator)
	if err != nil {
		log.Warn(ctx, "getOperatorContentPermission failed",
			log.Any("operator", operator),
			log.Err(err))
		return nil, err
	}

	result := make(map[string]entity.ContentPermission, len(req))
	for _, r := range req {
		pm := contentPermissionDict[r.ContentProfile]
		result[r.ID] = entity.ContentPermission{
			ID:             r.ID,
			AllowEdit:      pm.allowEdit,
			AllowDelete:    pm.allowDelete,
			AllowApprove:   pm.allowApprove,
			AllowReject:    pm.allowReject,
			AllowRepublish: pm.allowRepublish,
		}
	}

	return result, nil
}

func (c *ContentPermissionTable) getOperatorContentPermission(ctx context.Context, operator *entity.Operator) (map[ContentProfile]*contentPermission, error) {
	result := make(map[ContentProfile]*contentPermission)

	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, operator, allContentPermissionSet)
	if err != nil {
		log.Warn(ctx, "HasOrganizationPermissions failed",
			log.Any("operator", operator),
			log.Any("permissions", allContentPermissionSet),
			log.Err(err))
		return nil, err
	}

	c.getOperatorContentEditPermission(result, hasPermission)
	c.getOperatorContentRemovePermission(result, hasPermission)
	c.getOperatorContentRepublishPermission(result, hasPermission)
	c.getOperatorContentApprovePermission(result, hasPermission)
	c.getOperatorContentRejectPermission(result, hasPermission)

	return result, nil
}

func (c *ContentPermissionTable) getOperatorContentEditPermission(result map[ContentProfile]*contentPermission, hasPermission map[external.PermissionName]bool) {
	for k, v := range c.editPermissionDict {
		if _, ok := result[k]; !ok {
			result[k] = &contentPermission{}
		}

		allowEdit := true
		for i := range v {
			allowEdit = true
			for _, v := range v[i].Permissions {
				if val, ok := hasPermission[v]; !ok || !val {
					allowEdit = false
					break
				}
			}

			if allowEdit == true {
				break
			}
		}
		result[k].allowEdit = allowEdit
	}
}

func (c *ContentPermissionTable) getOperatorContentRepublishPermission(result map[ContentProfile]*contentPermission, hasPermission map[external.PermissionName]bool) {
	for k, v := range c.publishPermissionDict {
		if _, ok := result[k]; !ok {
			result[k] = &contentPermission{}
		}

		allowRepublish := true
		for i := range v {
			allowRepublish = true
			for _, v := range v[i].Permissions {
				if val, ok := hasPermission[v]; !ok || !val {
					allowRepublish = false
					break
				}
			}

			if allowRepublish == true {
				break
			}
		}
		result[k].allowRepublish = allowRepublish
	}
}

func (c *ContentPermissionTable) getOperatorContentRemovePermission(result map[ContentProfile]*contentPermission, hasPermission map[external.PermissionName]bool) {
	for k, v := range c.removePermissionDict {
		if _, ok := result[k]; !ok {
			result[k] = &contentPermission{}
		}

		allowDelete := true
		for i := range v {
			allowDelete = true
			for _, v := range v[i].Permissions {
				if val, ok := hasPermission[v]; !ok || !val {
					allowDelete = false
					break
				}
			}

			if allowDelete == true {
				break
			}
		}
		result[k].allowDelete = allowDelete
	}
}

func (c *ContentPermissionTable) getOperatorContentApprovePermission(result map[ContentProfile]*contentPermission, hasPermission map[external.PermissionName]bool) {
	for k, v := range c.approvePermissionDict {
		if _, ok := result[k]; !ok {
			result[k] = &contentPermission{}
		}

		allowApprove := true
		for i := range v {
			allowApprove = true
			for _, v := range v[i].Permissions {
				if val, ok := hasPermission[v]; !ok || !val {
					allowApprove = false
					break
				}
			}

			if allowApprove == true {
				break
			}
		}
		result[k].allowApprove = allowApprove
	}
}

func (c *ContentPermissionTable) getOperatorContentRejectPermission(result map[ContentProfile]*contentPermission, hasPermission map[external.PermissionName]bool) {
	for k, v := range c.rejectPermissionDict {
		if _, ok := result[k]; !ok {
			result[k] = &contentPermission{}
		}

		allowReject := true
		for i := range v {
			allowReject = true
			for _, v := range v[i].Permissions {
				if val, ok := hasPermission[v]; !ok || !val {
					allowReject = false
					break
				}
			}

			if allowReject == true {
				break
			}
		}
		result[k].allowReject = allowReject
	}
}
