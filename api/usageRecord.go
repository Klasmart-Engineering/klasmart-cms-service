package api

import (
	"github.com/gin-gonic/gin"
)

type StudentUsageRecord struct {
	ClassType      string `json:"class_type"`
	RoomID         string `json:"room_id"`
	LessonMaterial struct {
		Url  string `json:"url"`
		Type string `json:"type" enums:"Iframe,Video,Audio,Image"`
	} `json:"lesson_material"`
	ContentType string `json:"h5p_type" enums:"h5p, audio, video, image, document"`
	ActionType  string `json:"action_type" enums:"view"`
	Timestamp   int64  `json:"timestamp"`
}

// @Summary student usage record
// @Description student usage record
// @Tags statistics
// @ID studentUsageRecord
// @Accept json
// @Produce json
// @Param milestone body StudentUsageRecord true "record usage"
// @Success 200 {object} string
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /student_usage_record/event [post]
func (s *Server) studentUsageRecordEvent(ctx *gin.Context) {
	// TODO:
	panic("wait for implement")
}
