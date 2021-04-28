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
	VisibilitySettingsTypeMySchools   VisibilitySettingsType = 1
	VisibilitySettingsTypeAllSchools  VisibilitySettingsType = 2
	VisibilitySettingsTypeContainsOrg VisibilitySettingsType = 3

	OwnerTypeUser   OwnerType = 1
	OwnerTypeOthers OwnerType = 2
)

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
