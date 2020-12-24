package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

type AddAuthedContentRequest struct {
	OrgID        string `json:"org_id"`
	FromFolderID string `json:"from_folder_id"`
	ContentID    string `json:"content_id"`
}

type BatchAddAuthedContentRequest struct {
	OrgID      string   `json:"org_id"`
	FolderID   string   `json:"folder_id"`
	ContentIDs []string `json:"content_ids"`
}

type DeleteAuthedContentRequest struct {
	OrgID     string `json:"org_id"`
	ContentID string `json:"content_id"`
}

type BatchDeleteAuthedContentRequest struct {
	OrgID      string   `json:"org_id"`
	ContentIDs []string `json:"content_ids"`
}

type BatchDeleteAuthedContentByOrgsRequest struct {
	OrgIDs     []string `json:"org_ids"`
	FolderIDs []string `json:"folder_i_ds"`
	ContentIDs []string `json:"content_ids"`
}

type AuthedContentRecord struct {
	ID           string `json:"record_id" gorm:"type:varchar(50);PRIMARY_KEY;AUTO_INCREMENT"`
	OrgID        string `json:"org_id" gorm:"type:varchar(50);NOT NULL;column:org_id"`
	FromFolderID string `json:"from_folder_id" gorm:"type:varchar(50);column:from_folder_id"`
	ContentID    string `json:"content_id" gorm:"type:varchar(50);NOT NULL;column:content_id"`
	Creator      string `json:"creator" gorm:"type:varchar(50);NOT NULL;column:creator"`
	CreateAt     int64  `json:"create_at" gorm:"type:int;NOT NULL;column:create_at"`
	DeleteAt     int64  `json:"delete_at" gorm:"type:int;NOT NULL;DEFAULT: 0;column:delete_at"`
	Duration     int    `json:"duration" gorm:"type:int;NOT NULL;DEFAULT: 0;column:duration"`
}

func (AuthedContentRecord) TableName() string {
	return "cms_authed_contents"
}

type AuthedContentRecordInfo struct {
	AuthedContentRecord
	ContentInfoWithDetails
}

type AuthedContentRecordInfoResponse struct {
	Total int                        `json:"total"`
	List  []*AuthedContentRecordInfo `json:"list"`
}

type AuthedOrgList struct {
	Total int                 `json:"total"`
	List  []*OrganizationInfo `json:"orgs"`
}

type OrganizationInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SearchAuthedContentRequest struct {
	IDs        []string `json:"ids"`
	OrgIDs     []string `json:"org_ids"`
	ContentIDs []string `json:"content_ids"`
	Creator    []string `json:"creator"`
	Pager      *utils.Pager
}

type SharedFolderRecord struct {
	ID       string `json:"record_id" gorm:"type:varchar(50);PRIMARY_KEY;AUTO_INCREMENT"`
	FolderID string `json:"folder_id" gorm:"type:varchar(255);NOT NULL;column:folder_id"`
	OrgID    string `json:"org_id" gorm:"type:varchar(255);NOT NULL;column:org_id"`
	Creator  string `json:"creator" gorm:"type:varchar(255);NOT NULL;column:creator"`
	CreateAt int64  `json:"create_at" gorm:"type:int;NOT NULL;column:create_at"`
	UpdateAt int64  `json:"update_at" gorm:"type:int;NOT NULL;column:update_at"`
	DeleteAt int64  `json:"delete_at" gorm:"type:int;NOT NULL;DEFAULT: 0;column:delete_at"`
}

func (SharedFolderRecord) TableName() string {
	return "cms_shared_folders"
}
