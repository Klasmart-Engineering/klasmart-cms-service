package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
)

func (s *Server) addScheduleFeedback(c *gin.Context) {

}

func (s *Server) verifyScheduleFeedback(ctx context.Context, input *entity.ScheduleFeedbackAddInput) (ResponseLabel, error) {
	if strings.TrimSpace(input.UserID) == "" {
		log.Info(ctx, "user id is required", log.Any("input", input))
		return GeneralUnknown, constant.ErrInvalidArgs
	}
	if strings.TrimSpace(input.ScheduleID) == "" {
		log.Info(ctx, "schedule id is required", log.Any("input", input))
		return GeneralUnknown, constant.ErrInvalidArgs
	}
	if strings.TrimSpace(input.AssignmentUrl) == "" {
		log.Info(ctx, "assignment url is required", log.Any("input", input))
		return GeneralUnknown, constant.ErrInvalidArgs
	}
	return "", nil
}
