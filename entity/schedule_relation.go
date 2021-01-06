package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type ScheduleRelation struct {
	ID         string             `json:"id" gorm:"column:id;PRIMARY_KEY"`
	ScheduleID string             `json:"schedule_id" gorm:"column:schedule_id;type:varchar(100)"`
	RecordID   string             `json:"record_id" gorm:"column:record_id;type:varchar(100)"`
	RecordType ScheduleRecordType `json:"record_type" gorm:"column:record_type;type:varchar(100)"`
	GroupName  ScheduleGroupName  `json:"group_name" gorm:"column:group_name;type:varchar(100)"`
}

func (e ScheduleRelation) TableName() string {
	return constant.TableNameScheduleRelation
}

func (e ScheduleRelation) GetID() interface{} {
	return e.ID
}

type ScheduleRecordType string

const (
	ScheduleRecordTypeOrg     ScheduleRecordType = "org"
	ScheduleRecordTypeSchool  ScheduleRecordType = "school"
	ScheduleRecordTypeClass   ScheduleRecordType = "class"
	ScheduleRecordTypeTeacher ScheduleRecordType = "teacher"
	ScheduleRecordTypeStudent ScheduleRecordType = "student"
)

type ScheduleGroupName string

const (
	ScheduleGroupNameNone        ScheduleGroupName = "none"
	ScheduleGroupNameClassRoster ScheduleGroupName = "class_roster"
	ScheduleGroupNameParticipant ScheduleGroupName = "participant"
)
