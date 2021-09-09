package api

import (
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"

	"github.com/dgrijalva/jwt-go"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		}
	}()
	op := s.getOperator(c)
	jwtToken := entity.JwtToken{}
	err = c.ShouldBindJSON(&jwtToken)
	if err != nil {
		log.Error(ctx, "invalid params", log.Err(err))
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

	err = model.GetReportModel().AddStudentUsageRecordTx(ctx, dbo.MustGetDB(ctx), op, &req.StudentUsageRecord)
	if err != nil {
		return
	}
	return
}
