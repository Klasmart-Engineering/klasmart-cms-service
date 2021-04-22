package model

import (
	"context"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

var (
	ErrHasNoPermission = errors.New("has no permission")
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

type PermissionSet struct {
	Permissions []external.PermissionName
}

//HasPermission check whether the user has the permission of PermissionSet
func (p PermissionSet) HasPermission(ctx context.Context, operator entity.Operator) error {
	return ErrHasNoPermission
}

type ContentProfile struct {
	ContentType        entity.ContentType          //ContentType using entity.Content type
	Status             entity.ContentPublishStatus //ContentStatus using entity.ContentPublishStatus type
	VisibilitySettings VisibilitySettingsType
	Owner              OwnerType
}

type IContentPermissionTable interface {
	//GetCreatePermissionSets Get permission set list when a user want to create a content
	GetCreatePermissionSets(ctx context.Context, req ContentProfile) ([]*PermissionSet, error)
	//GetEditPermissionSets Get permission set list when a user want to edit a content
	GetEditPermissionSets(ctx context.Context, req ContentProfile) ([]*PermissionSet, error)
	//GetPublishPermissionSets Get permission set list when a user want to publish contents
	//ContentProfile list indicates the batch of Content profiles, permission set list must
	//contains all content profiles
	GetPublishPermissionSets(ctx context.Context, req []*ContentProfile) ([]*PermissionSet, error)
	//GetViewPermissionSets Get permission set list when a user want to view contents
	GetViewPermissionSets(ctx context.Context, req []*ContentProfile) ([]*PermissionSet, error)
	//GetRemovePermissionSets Get permission set list when a user want to remove contents
	GetRemovePermissionSets(ctx context.Context, req []*ContentProfile) ([]*PermissionSet, error)
}
