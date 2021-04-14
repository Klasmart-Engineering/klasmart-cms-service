package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type ScheduleRelation struct {
	ID           string               `json:"id" gorm:"column:id;PRIMARY_KEY"`
	ScheduleID   string               `json:"schedule_id" gorm:"column:schedule_id;type:varchar(100)"`
	RelationID   string               `json:"relation_id" gorm:"column:relation_id;type:varchar(100)"`
	RelationType ScheduleRelationType `json:"relation_type" gorm:"column:relation_type;type:varchar(100)"`
}

func (e ScheduleRelation) TableName() string {
	return constant.TableNameScheduleRelation
}

func (e ScheduleRelation) GetID() interface{} {
	return e.ID
}

type ScheduleRelationType string

const (
	ScheduleRelationTypeInvalid            ScheduleRelationType = "Invalid"
	ScheduleRelationTypeOrg                ScheduleRelationType = "org"
	ScheduleRelationTypeSchool             ScheduleRelationType = "school"
	ScheduleRelationTypeClassRosterClass   ScheduleRelationType = "class_roster_class"
	ScheduleRelationTypeParticipantClass   ScheduleRelationType = "participant_class"
	ScheduleRelationTypeClassRosterTeacher ScheduleRelationType = "class_roster_teacher"
	ScheduleRelationTypeClassRosterStudent ScheduleRelationType = "class_roster_student"
	ScheduleRelationTypeParticipantTeacher ScheduleRelationType = "participant_teacher"
	ScheduleRelationTypeParticipantStudent ScheduleRelationType = "participant_student"
)

func (s ScheduleRelationType) String() string {
	return string(s)
}

type ScheduleUserRelation struct {
	Teachers []*ScheduleShortInfo
	Students []*ScheduleShortInfo
}
