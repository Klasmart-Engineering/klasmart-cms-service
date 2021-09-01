package api

import (
	"github.com/gin-gonic/gin"
)

type StudentUsageRecord struct {
	Account          string `json:"account"`
	UserID           string `json:"user_id"`
	ClassType        string `json:"class_type"`
	RoomID           string `json:"room_id"`
	LessonMaterialID string `json:"lesson_material_id"`
	H5PType          string `json:"h5p_type"`
	ActionType       string `json:"action_type"`
	Timestamp        int64  `json:"timestamp"`
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
// @Router /statistics/student_usage_record [post]
func (s *Server) studentUsageRecord(ctx *gin.Context) {
	// TODO:
	panic("wait for implement")
}
