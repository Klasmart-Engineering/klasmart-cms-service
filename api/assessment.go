package api

import (
	"github.com/dgrijalva/jwt-go"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary add assessments
// @Description add assessments
// @Tags assessments
// @ID addAssessment
// @Accept json
// @Produce json
// @Param assessment body entity.AddAssessmentArgs true "add assessment command"
// @Success 200 {object} entity.AddAssessmentResult
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments [post]
func (s *Server) addAssessment(c *gin.Context) {
	ctx := c.Request.Context()

	log.Debug(ctx, "add assessment jwt: call")
	body := struct {
		Token string `json:"token"`
	}{}
	if err := c.ShouldBind(&body); err != nil {
		log.Info(ctx, "add assessment jwt: bind failed",
			log.Err(err),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	args := entity.AddAssessmentArgs{}
	if _, err := jwt.ParseWithClaims(body.Token, &args, func(token *jwt.Token) (interface{}, error) {
		return config.Get().Assessment.AddAssessmentSecret, nil
	}); err != nil {
		log.Error(ctx, "add assessment jwt: parse with claims failed",
			log.Err(err),
			log.Any("token", body.Token),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	log.Debug(ctx, "add assessment jwt: fill args", log.Any("args", args), log.String("token", body.Token))
	newIDs, err := model.GetAssessmentModel().AddClassAndLive(ctx, s.getOperator(c), args)
	switch err {
	case nil:
		log.Debug(ctx, "add assessment jwt success",
			log.Any("args", args),
			log.Strings("new_id", newIDs),
		)
		c.JSON(http.StatusOK, entity.AddAssessmentResult{IDs: newIDs})
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	default:
		log.Error(ctx, "add assessment jwt: add failed",
			log.Err(err),
			log.Any("args", args),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary add assessments for test
// @Description add assessments for test
// @Tags assessments
// @ID addAssessmentForTest
// @Accept json
// @Produce json
// @Param assessment body entity.AddAssessmentArgs true "add assessment command"
// @Success 200 {object} entity.AddAssessmentResult
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments_for_test [post]
func (s *Server) addAssessmentForTest(c *gin.Context) {
	ctx := c.Request.Context()

	args := entity.AddAssessmentArgs{}
	if err := c.ShouldBind(&args); err != nil {
		log.Info(ctx, "add assessment: bind failed",
			log.Err(err),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	newIDs, err := model.GetAssessmentModel().AddClassAndLive(ctx, s.getOperator(c), args)
	switch err {
	case nil:
		c.JSON(http.StatusOK, entity.AddAssessmentResult{IDs: newIDs})
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	default:
		log.Error(ctx, "add assessment: add failed",
			log.Err(err),
			log.Any("args", args),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
