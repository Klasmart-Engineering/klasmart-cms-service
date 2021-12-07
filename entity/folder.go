package entity

import (
	"strings"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

const (
	FolderItemTypeAll    ItemType = 0
	FolderItemTypeFolder ItemType = 1
	FolderItemTypeFile   ItemType = 2

	OwnerTypeOrganization OwnerType = 1
	OwnerTypeUser         OwnerType = 2

	FolderFileTypeContent FolderFileType = "content"
	FolderFileTypeFolder  FolderFileType = "folder"

	//RootAssetsFolderName FolderPartition = "assets"
	//RootMaterialsAndPlansFolderName FolderPartition = "plans and materials"
	FolderPartitionAssets           FolderPartition = "assets"
	FolderPartitionMaterialAndPlans FolderPartition = "plans and materials"
)

type FolderFileType string

func NewFolderFileType(fileType string) FolderFileType {
	switch fileType {
	case "content":
		return FolderFileTypeContent
	case "folder":
		return FolderFileTypeFolder
	}
	return FolderFileTypeContent
}
func (f FolderFileType) Valid() bool {
	if f == FolderFileTypeContent || f == FolderFileTypeFolder {
		return true
	}
	return false
}

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

func (f FolderPartition) Valid() bool {
	if f == FolderPartitionMaterialAndPlans || f == FolderPartitionAssets {
		return true
	}
	return false
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
func NewOwnerType(num int) OwnerType {
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

func (o ItemType) ValidSearch() bool {
	if o == FolderItemTypeFolder || o == FolderItemTypeFile ||
		o == FolderItemTypeAll {
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

func NewItemType(num int) ItemType {
	switch num {
	case int(FolderItemTypeFile):
		return FolderItemTypeFile
	case int(FolderItemTypeFolder):
		return FolderItemTypeFolder
	}
	return FolderItemTypeFolder
}

type CreateFolderRequest struct {
	OwnerType   OwnerType       `json:"owner_type"`
	ParentID    string          `json:"parent_id"`
	Name        string          `json:"name"`
	Partition   FolderPartition `json:"partition"`
	Description string          `json:"description"`
	Keywords    []string        `json:"keywords"`

	Thumbnail string `json:"thumbnail"`
}

type MoveFolderRequest struct {
	ID             string          `json:"id"`
	OwnerType      OwnerType       `json:"owner_type"`
	Dist           string          `json:"dist"`
	Partition      FolderPartition `json:"partition"`
	FolderFileType FolderFileType  `json:"folder_file_type" enums:"content,folder"`
}

type UpdateFolderRequest struct {
	Name        string   `json:"name"`
	Thumbnail   string   `json:"thumbnail"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

type ShareFoldersRequest struct {
	FolderIDs []string `json:"folder_ids"`
	OrgIDs    []string `json:"org_ids"`
}

type FolderShareRecord struct {
	FolderID string              `json:"folder_id"`
	Orgs     []*OrganizationInfo `json:"orgs"`
}
type FolderShareRecords struct {
	Data []*FolderShareRecord `json:"data"`
}

type ShareFoldersDeleteAddOrgList struct {
	DeleteOrgs []string
	AddOrgs    []string
}
type ReplaceLinkedContentPathRequest struct {
	FromFolder *FolderItem
	DistFolder *FolderItem
	ContentIDs []string
}

type HandleSharedFolderAndAuthedContentRequest struct {
	SharedFolderPendingOrgsMap map[string]*ShareFoldersDeleteAddOrgList
	ShareFolders               []*FolderItem
	//ContentFolderMap           *ContentFolderMap
}

type ContentFolderMap struct {
	ContentFolderMap map[string][]string
	AllContentIDs    []string
}

type FolderIdWithFileType struct {
	ID             string         `json:"id"`
	FolderFileType FolderFileType `json:"folder_file_type" enums:"content,folder"`
}

type MoveFolderIDBulkRequest struct {
	FolderInfo []FolderIdWithFileType `json:"folder_info"`
	OwnerType  OwnerType              `json:"owner_type"`
	Dist       string                 `json:"dist"`
	Partition  FolderPartition        `json:"partition"`
}

type RemoveItemBulk struct {
	FolderIDs []string `json:"folder_ids"`
}

type CreateFolderItemRequest struct {
	//ID string `json:"id"`
	ParentFolderID string `json:"parent_folder_id"`
	//ItemType  ItemType  `json:"item_type"`
	Partition FolderPartition `json:"partition"`
	Link      string          `json:"link"`
	OwnerType OwnerType       `json:"owner_type"`
}

type FolderItemsCount struct {
	ID       string
	Classify string
	DirPath  string
	Count    int64
}

type UpdateFolderItemsCountRequest struct {
	ID    string
	Count int
}

type FolderDescendantItemsAndPath struct {
	DescendantFoldersMap  map[string][]*FolderItem
	FolderChildrenPathMap map[string]Path // for example: c is located /a/b/  map[c]Path = /a/b/c
	FolderDirPathMap      map[string]Path // for example: c is located /a/b/  map[c]Path = /a/b
}

type Path string

func (p Path) AsParentPath() string {
	if p == constant.FolderRootPath {
		return ""
	}
	return string(p)
}
func (p Path) Parent() string {
	if p == constant.FolderRootPath {
		return constant.FolderRootPath
	}
	pairs := strings.Split(string(p), constant.FolderPathSeparator)
	return pairs[len(pairs)-1]
}

func (p Path) Parents() []string {
	if p == constant.FolderRootPath {
		return nil
	}
	pairs := strings.Split(string(p), constant.FolderPathSeparator)
	ret := make([]string, len(pairs)-1)
	for i := range ret {
		ret[i] = pairs[i+1]
	}
	return ret
}
func (p Path) String() string {
	return string(p)
}
func (p Path) Equal(p0 Path) bool {
	return string(p) == string(p0)
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

func NewPath(p string) Path {
	return Path(p)
}

type FolderItem struct {
	ID        string    `gorm:"type:varchar(50);PRIMARY_KEY" json:"id"`
	OwnerType OwnerType `gorm:"type:int;NOT NULL" json:"owner_type"`
	Owner     string    `gorm:"type:varchar(50);NOT NULL" json:"owner"`
	ParentID  string    `gorm:"type:varchar(50)" json:"parent_id"`
	Link      string    `gorm:"type:varchar(50)" json:"link"`

	ItemType  ItemType        `gorm:"type:int;NOT NULL" json:"item_type"`
	DirPath   Path            `gorm:"type:varchar(2048);NOT NULL;INDEX" json:"dir_path"`
	Partition FolderPartition `gorm:"type:varchar(256);NOT NULL" json:"partition"`
	Name      string          `gorm:"type:varchar(256);NOT NULL" json:"name"`

	Description string `gorm:"type:text;NOT NULL" json:"description"`
	Keywords    string `gorm:"type:text;NOT NULL" json:"keywords"`

	Thumbnail string `gorm:"type:text" json:"thumbnail"`
	Creator   string `gorm:"type:varchar(50)" json:"creator"`

	ItemsCount int    `gorm:"type:int" json:"items_count"`
	Editor     string `gorm:"type:varchar(50);NOT NULL" json:"editor"`

	Extra string `gorm:"type:text;NULL" json:"extra"`
	//VisibilitySetting string	`gorm:"type:varchar(50)" json:"visibility_setting"`

	CreateAt int64 `gorm:"type:bigint;NOT NULL;column:create_at" json:"create_at"`
	UpdateAt int64 `gorm:"type:bigint;NOT NULL;column:update_at" json:"update_at"`
	DeleteAt int64 `gorm:"type:bigint;column:delete_at" json:"-"`
}

func (f FolderItem) TableName() string {
	return "cms_folder_items"
}

func (f FolderItem) ChildrenPath() Path {
	//根目录情况
	//Root path case
	if f.ID == constant.FolderRootPath {
		return NewPath(constant.FolderRootPath)
	}
	return NewPath(f.DirPath.AsParentPath() + constant.FolderPathSeparator + f.ID)
}

type FolderItemInfo struct {
	ID        string    `json:"id"`
	OwnerType OwnerType `json:"owner_type"`
	Owner     string    `json:"owner"`
	ParentID  string    `json:"parent_id"`
	Link      string    `json:"link"`

	ItemType  ItemType        `json:"item_type"`
	DirPath   Path            `json:"dir_path"`
	Partition FolderPartition `json:"partition"`
	Name      string          `json:"name"`

	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`

	Thumbnail   string `json:"thumbnail"`
	Creator     string `json:"creator"`
	CreatorName string `json:"creator_name"`
	EditorName  string `json:"editor_name"`

	ItemsCount int    `json:"items_count"`
	Editor     string `json:"editor"`

	CreateAt int64 `json:"create_at"`
	UpdateAt int64 `json:"update_at"`

	Available int               `json:"available"`
	Items     []*FolderItemInfo `json:"items"`
}

type SearchFolderCondition struct {
	IDs       []string
	OwnerType OwnerType
	Owner     string
	ItemType  ItemType
	ParentID  string
	Link      string
	Partition string

	Path    string
	Name    string
	OrderBy string
	Pager   utils.Pager
}
