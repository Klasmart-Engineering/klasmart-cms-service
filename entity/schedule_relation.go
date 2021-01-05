package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type ScheduleRelation struct {
	ID         string             `json:"id" gorm:"column:id;PRIMARY_KEY"`
	RecordID   string             `json:"record_id" gorm:"column:record_id;type:varchar(100)"`
	RecordType ScheduleRecordType `json:"record_type" gorm:"column:record_type;type:varchar(100)"`
	GroupName  ScheduleGroupName  `json:"group_name" gorm:"column:group_name;type:varchar(100)"`
	CreateID   string             `gorm:"column:create_id;type:varchar(100)"`
	UpdateID   string             `gorm:"column:update_id;type:varchar(100)"`
	DeleteID   string             `gorm:"column:delete_id;type:varchar(100)"`
	CreateAt   int64              `gorm:"column:create_at;type:bigint"`
	UpdateAt   int64              `gorm:"column:update_at;type:bigint"`
	DeleteAt   int64              `gorm:"column:delete_at;type:bigint"`
}

func (e ScheduleRelation) TableName() string {
	return constant.TableNameScheduleRelation
}

func (e ScheduleRelation) GetID() interface{} {
	return e.ID
}

type ScheduleRecordType string

const (
	ScheduleRecordTypeOrg    ScheduleRecordType = "org"
	ScheduleRecordTypeSchool ScheduleRecordType = "school"
	ScheduleRecordTypeClass  ScheduleRecordType = "class"
	ScheduleRecordTypeUser   ScheduleRecordType = "teacher"
)

type ScheduleGroupName string

const (
	ScheduleGroupNameNone        ScheduleGroupName = "none"
	ScheduleGroupNameClassRoster ScheduleGroupName = "class_roster"
	ScheduleGroupNameParticipant ScheduleGroupName = "participant"
)
