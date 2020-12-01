package entity

import (
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"strings"
)

const (
	FolderItemTypeAll    ItemType = 0
	FolderItemTypeFolder ItemType = 1
	FolderItemTypeFile   ItemType = 2

	OwnerTypeOrganization OwnerType = 1
	OwnerTypeUser OwnerType = 2


	FolderFileTypeContent = "content"
	FolderFileTypeFolder = "folder"
	FolderFileTypeFolderItem = "item"

	//RootAssetsFolderName FolderPartition = "assets"
	//RootMaterialsAndPlansFolderName FolderPartition = "plans and materials"
	FolderPartitionAssets           FolderPartition = "assets"
	FolderPartitionMaterialAndPlans FolderPartition = "plans and materials"
)

type FolderPartition string
func NewFolderPartition(partition string) FolderPartition {
	switch partition {
	case "assets":
		return FolderPartitionAssets
	case "plans and materials":
		return FolderPartitionMaterialAndPlans
	}
	return FolderPartitionMaterialAndPlans
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
	OwnerType int       `json:"owner_type"`
	ParentID  string          `json:"parent_id"`
	Name      string          `json:"name"`
	Partition string `json:"partition"`

	Thumbnail string `json:"thumbnail"`
}

type MoveFolderRequest struct {
	ID             string          `json:"id"`
	OwnerType      int       `json:"owner_type"`
	Dist           string          `json:"dist"`
	Partition      string `json:"partition"`
	FolderFileType string          `json:"folder_file_type" enums:"content,folder"`
}

type UpdateFolderRequest struct {
	Name string `json:"name"`
	Thumbnail string `json:"thumbnail"`
}

type FolderIdWithFileType struct {
	ID string `json:"id"`
	FolderFileType string `json:"folder_file_type" enums:"content,folder"`
}

type MoveFolderIDBulkRequest struct {
	FolderInfo []FolderIdWithFileType `json:"folder_info"`
	OwnerType int `json:"owner_type"`
	Dist string `json:"dist"`
	Partition string `json:"partition"`
}

type CreateFolderItemRequest struct {
	//ID string `json:"id"`
	ParentFolderID string `json:"parent_folder_id"`
	//ItemType  ItemType  `json:"item_type"`
	Partition string `json:"partition"`
	Link      string          `json:"link"`
	OwnerType int `json:"owner_type"`
}

type Path string

func (p Path) ParentPath()string {
	if p == constant.FolderRootPath {
		return ""
	}
	return string(p)
}

func (p Path) Parents() []string {
	if p == constant.FolderRootPath {
		return nil
	}
	pairs := strings.Split(string(p), "/")
	ret := make([]string, len(pairs) - 1)
	for i := range ret {
		ret[i] = pairs[i + 1]
	}
	return ret
}

func (p Path) IsChild(f string) bool {
	parents := p.Parents()
	for i := range parents {
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

	ItemType ItemType `gorm:"type:int;NOT NULL" json:"item_type"`
	DirPath  Path     `gorm:"type:varchar(2048);NOT NULL;INDEX" json:"dir_path"`
	Partition string `gorm:"type:varchar(256);NOT NULL" json:"partition"`
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
func (f FolderItem) TableName() string{
	return "cms_folder_items"
}

func (f FolderItem) ChildrenPath() Path {
	//根目录情况
	if f.ID == constant.FolderRootPath {
		return NewPath(constant.FolderRootPath)
	}
	return NewPath(f.DirPath.ParentPath() + "/" + f.ID + "/")
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
	Partition string

	Path string
	Name string
	OrderBy string
	Pager   utils.Pager
}
