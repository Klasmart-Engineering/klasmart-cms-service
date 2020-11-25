package entity

import (
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"strings"
)

const (
	FolderItemTypeAll    ItemType = 0
	FolderItemTypeFolder ItemType = 1
	FolderItemTypeFile   ItemType = 2

	OwnerTypeOrganization OwnerType = 1
	OwnerTypeUser OwnerType = 2


	FileTypeContent = "content"

	RootAssetsFolderName FolderPartition = "assets"
	RootMaterialsAndPlansFolderName FolderPartition = "plans and materials"
)

type FolderPartition string

func NewFolderPartition(partition string) FolderPartition{
	switch partition {
	case "assets":
		return RootAssetsFolderName
	case "plans and materials":
		return RootMaterialsAndPlansFolderName
	}
	return RootMaterialsAndPlansFolderName
}

func (f FolderPartition) Valid() bool{
	if f == RootAssetsFolderName || f == RootMaterialsAndPlansFolderName {
		return true
	}
	return false
}
func (f FolderPartition) Path() string{
	return "/" + string(f)
}

type OwnerType int

func (o OwnerType) Valid() bool {
	if o == OwnerTypeOrganization || o == OwnerTypeUser {
		return true
	}
	return false
}
func (o OwnerType) Owner(operator *Operator) string {
	if o == OwnerTypeUser {
		return operator.UserID
	}
	return operator.OrgID
}
func NewOwnerType(num int) OwnerType{
	switch num {
	case int(OwnerTypeOrganization):
		return OwnerTypeOrganization
	case int(OwnerTypeUser):
		return OwnerTypeUser
	}
	return OwnerTypeOrganization
}

type ItemType int
func (o ItemType) Valid() bool {
	if o == FolderItemTypeFolder || o == FolderItemTypeFile {
		return true
	}
	return false
}

func (o ItemType) ValidExcludeFolder() bool {
	if o == FolderItemTypeFile {
		return true
	}
	return false
}

func (o ItemType) IsFolder() bool {
	if o == FolderItemTypeFolder {
		return true
	}
	return false
}

func NewItemType(num int) ItemType{
	switch num {
	case int(FolderItemTypeFile):
		return FolderItemTypeFile
	case int(FolderItemTypeFolder):
		return FolderItemTypeFolder
	}
	return FolderItemTypeFolder
}

type CreateFolderRequest struct {
	OwnerType OwnerType `json:"owner_type"`
	ParentID  string    `json:"parent_id"`
	Name      string    `json:"name"`

	Thumbnail string `json:"thumbnail"`
}

type MoveFolderRequest struct {
	Dist string `json:"dist"`
}

type UpdateFolderRequest struct {
	Name string `json:"name"`
	Thumbnail string `json:"thumbnail"`
}

type MoveFolderIDBulk struct {
	IDs []string `json:"ids"`
	Dist string `json:"dist"`
}

type CreateFolderItemRequest struct {
	FolderID  string    `json:"folder_id"`
	//ItemType  ItemType  `json:"item_type"`
	Link      string    `json:"link"`
}

type Path string

func (p Path) ParentPath()string {
	if p == "/" {
		return ""
	}
	return string(p)
}

func (p Path) Parents() []string {
	return strings.Split(string(p), "/")
}

func (p Path) IsChild(f string) bool {
	parents := p.Parents()
	for i := range parents {
		if parents[i] == "" {
			continue
		}
		if parents[i] == f {
			return true
		}
	}
	return false
}

func NewPath(p string) Path{
	return Path(p)
}

type FolderItem struct {
	ID        string    `gorm:"type:varchar(50);PRIMARY_KEY" json:"id"`
	OwnerType OwnerType `gorm:"type:int;NOT NULL" json:"owner_type"`
	Owner     string    `gorm:"type:varchar(50);NOT NULL" json:"owner"`
	ParentID  string    `gorm:"type:varchar(50)" json:"parent_id"`
	Link      string    `gorm:"type:varchar(50)" json:"link"`

	ItemType ItemType `gorm:"type:int;NOT NULL"`
	DirPath  Path     `gorm:"type:varchar(2048);NOT NULL;INDEX" json:"dir_path"`
	Name     string   `gorm:"type:varchar(256);NOT NULL" json:"name"`

	Thumbnail string	`gorm:"type:text" json:"thumbnail"`
	Creator string 	`gorm:"type:varchar(50)" json:"creator"`

	ItemsCount int `gorm:"type:int" json:"items_count"`
	Editor string 	`gorm:"type:varchar(50);NOT NULL" json:"editor"`
	//VisibilitySetting string	`gorm:"type:varchar(50)" json:"visibility_setting"`

	CreateAt int64 `gorm:"type:bigint;NOT NULL;column:create_at" json:"create_at"`
	UpdateAt int64 `gorm:"type:bigint;NOT NULL;column:update_at" json:"update_at"`
	DeleteAt int64 `gorm:"type:bigint;column:delete_at" json:"-"`
}

func (f FolderItem) ChildrenPath() Path {
	return NewPath(f.DirPath.ParentPath() + "/" + f.ID)
}

type FolderItemInfo struct {
	FolderItem
	Items []*FolderItem `json:"items"`
}

type SearchFolderCondition struct {
	IDs       []string
	OwnerType OwnerType
	Owner     string
	ItemType  ItemType
	ParentID  string
	Link      string

	Path string
	Name string
	OrderBy string
	Pager   utils.Pager
}
