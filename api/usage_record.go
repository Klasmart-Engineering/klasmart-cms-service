package api

import (
	"net/http"
	"strings"

	"github.com/KL-Engineering/dbo"

	"github.com/KL-Engineering/kidsloop-cms-service/model"

	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/golang-jwt/jwt"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/gin-gonic/gin"
)

// @Summary student usage record
// @Description student usage record
// @Tags usageRecord
// @ID studentUsageRecord
// @Accept json
// @Produce json
// @Param milestone body entity.JwtToken true "record usage"
// @Success 200 {object} string
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /student_usage_record/event [post]
func (s *Server) addStudentUsageRecordEvent(c *gin.Context) {
	ctx := c.Request.Context()

	log.Debug(ctx, "add student usage record")

	var err error
	defer func() {
		if err == nil {
			return
		}
		switch err {
		case constant.ErrInvalidArgs:
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		default:
			s.defaultErrorHandler(c, err)
		}
	}()
	op := s.getOperator(c)
	jwtToken := entity.JwtToken{}
	err = c.ShouldBindJSON(&jwtToken)
	if err != nil {
		log.Warn(ctx, "invalid params", log.Err(err))
		err = constant.ErrInvalidArgs
		return
	}

	log.Info(ctx, "get event token",
		log.String("log type", "report"),
		log.String("token", jwtToken.Token),
		log.String("step", "REPORT step1"))

	req := entity.StudentUsageRecordInJwt{}
	_, err = jwt.ParseWithClaims(jwtToken.Token, &req, func(t *jwt.Token) (interface{}, error) {
		// return config.Get().Report.PublicKey, nil

		// temp solution, ops is offline
		return config.Get().Assessment.AddAssessmentSecret, nil
	})
	log.Info(ctx, "/student_usage_record/event", log.Any("req", req))
	if err != nil {
		log.Warn(ctx, "jwt.ParseWithClaims failed", log.Err(err), log.Any("jwtToken", jwtToken))
		err = constant.ErrInvalidArgs
		return
	}
	req.ContentType = strings.ToLower(strings.TrimSpace(req.ContentType))

	if req.StudentUsageRecord.RoomID == "" {
		log.Warn(ctx, "room_id is required", log.Any("req", req))
		err = constant.ErrInvalidArgs
		return
	}
	if len(req.Students) <= 0 {
		log.Warn(ctx, "students is required", log.Any("req", req))
		err = constant.ErrInvalidArgs
		return
	}

	err = model.GetReportModel().AddStudentUsageRecordTx(ctx, dbo.MustGetDB(ctx), op, &req.StudentUsageRecord)
	if err != nil {
		return
	}
	return
}
