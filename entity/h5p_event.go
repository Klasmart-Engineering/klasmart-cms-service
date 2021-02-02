package entity

import (
	"context"
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"time"
)

type CreateH5pEventRequest struct {
	VerbID              string `json:"verb_id"`
	PlayID              string `json:"play_id"`
	UserID              string `json:"user_id"`
	ScheduleID          string `json:"schedule_id"`
	LocalContentID      string `json:"local_content_id"`
	LocalLibraryName    string `json:"local_library_name"`
	LocalLibraryVersion string `json:"local_library_version"`
	MaterialID          string `json:"material_id"`
	EventTime           int64  `json:"time"`

	Extends map[string]interface{} `json:"extends"`

	//SubLibraryName   string `json:"sub_library_name"`
	//SubLibraryVersion string `json:"sub_library_version"`
	//SubContentID   string `json:"sub_content_id"`
	//TotalAmount    int    `json:"total_amount"`
	//CorrectAmount  int    `json:"correct_amount"`
}

func (c *CreateH5pEventRequest) ToEventObject(ctx context.Context, planID string, op *Operator) (*H5PEvent, error) {
	extends, err := json.Marshal(c.Extends)
	if err != nil {
		log.Warn(ctx, "Can't marshal request", log.Err(err))
		return nil, err
	}
	return &H5PEvent{
		VerbID:              c.VerbID,
		ScheduleID:          c.ScheduleID,
		LessonPlanID:        planID,
		MaterialID:          c.MaterialID,
		LocalLibraryName:    c.LocalLibraryName,
		LocalLibraryVersion: c.LocalLibraryVersion,
		LocalContentID:      c.LocalContentID,
		UserID:              op.UserID,
		PlayID:              c.PlayID,
		EventTime:           c.EventTime,
		Extends:             string(extends),
		CreateAt:            time.Now().Unix(),
	}, nil
}

type H5PEvent struct {
	ID                  string `gorm:"type:varchar(50);PRIMARY_KEY;column:id"`
	VerbID              string `gorm:"type:varchar(128);NOT NULL;column:verb_id"`
	ScheduleID          string `gorm:"type:varchar(50);NOT NULL;column:schedule_id"`
	PlayID              string `gorm:"type:varchar(50);NOT NULL; column:play_id"`
	LocalLibraryName    string `gorm:"type:varchar(50);NOT NULL;column:local_library_name"`
	LessonPlanID        string `gorm:"type:varchar(50);NOT NULL;column:lesson_plan_id"`
	LocalContentID      string `gorm:"type:varchar(50);NOT NULL;column:local_content_id"`
	MaterialID          string `gorm:"type:varchar(50);NOT NULL;column:material_id"`
	UserID              string `gorm:"type:varchar(50);NOT NULL;column:user_id"`
	EventTime           int64  `gorm:"type:bigint;NOT NULL;column:event_time"`
	LocalLibraryVersion string `gorm:"type:varchar(50);NOT NULL;column:local_library_version"`

	Extends  string `gorm:"type:json;NOT NULL;column:extends"`
	CreateAt int64  `gorm:"type:bigint;NOT NULL;column:create_at"`
}

func (H5PEvent) TableName() string {
	return "h5p_events"
}

type H5PEventsSortByTime []*H5PEvent

func (h H5PEventsSortByTime) Len() int {
	return len(h)
}

func (h H5PEventsSortByTime) Less(i, j int) bool {
	return h[i].EventTime < h[j].EventTime
}

func (h H5PEventsSortByTime) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}
