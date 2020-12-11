package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

type AddAuthedContentRequest struct {
	OrgId     string `json:"org_id"`
	ContentId string `json:"content_id"`
}

type BatchAddAuthedContentRequest struct {
	OrgId      string   `json:"org_id"`
	ContentIds []string `json:"content_ids"`
}

type DeleteAuthedContentRequest struct {
	OrgId     string `json:"org_id"`
	ContentId string `json:"content_id"`
}

type BatchDeleteAuthedContentRequest struct {
	OrgId      string   `json:"org_id"`
	ContentIds []string `json:"content_ids"`
}

type AuthedContentRecord struct {
	ID        string `json:"record_id" gorm:"type:varchar(50);PRIMARY_KEY;AUTO_INCREMENT"`
	OrgID     string `json:"org_id" gorm:"type:varchar(255);NOT NULL;column:org_id"`
	ContentID string `json:"content_id" gorm:"type:varchar(255);NOT NULL;column:content_id"`
	Creator   string `json:"creator" gorm:"type:varchar(255);NOT NULL;column:creator"`
	CreateAt  int64  `json:"create_at" gorm:"type:int;NOT NULL;column:create_at"`
	DeleteAt  int64  `json:"delete_at" gorm:"type:int;NOT NULL;column:delete_at"`
	Duration  int    `json:"duration" gorm:"type:int;NOT NULL;column:duration"`
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
	OrgIds     []string `json:"org_ids"`
	ContentIds []string `json:"content_ids"`
	Creator    []string `json:"creator"`
	Pager      utils.Pager
}
