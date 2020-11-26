package api

import (
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	AuthTo   string `json:"auth_to" form:"auth_to"`
	AuthCode string `json:"auth_code" form:"auth_code"`
	AuthType int    `json:"auth_type" from:"auth_type"`
}

// @ID userLogin
// @Summary login
// @Tags user
// @Description user login
// @Accept json
// @Produce json
// @Param outcome body LoginRequest true "user login"
// @Success 200
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /users/login [get]
func (s *Server) login(c *gin.Context) {
	var req LoginRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		log.Error(c.Request.Context(), "login:ShouldBindQuery failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	// TODO

	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
}

type RegisterRequest struct {
	Mobile string `json:"mobile" form:"mobile"`
	Email  string `json:"email" form:"email"`
	Name   string `json:"name" form:"name"`
	Avatar string `json:"avatar" form:"avatar"`
	Gender string `json:"gender" form:"gender"`
}

// @ID userRegister
// @Summary register
// @Tags user
// @Description user register
// @Accept json
// @Produce json
// @Param outcome body RegisterRequest true "user register"
// @Success 200
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /users/register [post]
func (s *Server) register(c *gin.Context) {
	// TODO
}

type VerificationRequest struct {
	Mobile string `json:"mobile" form:"mobile"`
	Email  string `json:"email" form:"email"`
	State  string `json:"state" form:"state"`
}

// @ID verification
// @Summary send verify code
// @Tags user
// @Description send verify code or uri
// @Accept json
// @Produce json
// @Param outcome body VerificationRequest true "send verify code"
// @Success 200
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /users/verification [post]
func (s *Server) verification(c *gin.Context) {
	var req VerificationRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Error(c.Request.Context(), "verification:ShouldBindJSON failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	if req.Email == "" && req.Mobile == "" {
		log.Warn(c.Request.Context(), "verification:param wrong", log.Any("req", req))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	// TODO
}

type ForgottenPasswordRequest struct {
}

// @ID forgottenPassword
// @Summary forget password
// @Tags user
// @Description forget password
// @Accept json
// @Produce json
// @Param outcome body ForgottenPasswordRequest true "login by new password and update password"
// @Success 200
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /users/forgotten_pwd [post]
func (s *Server) forgottenPassword(c *gin.Context) {
	// TODO
}

type ResetPasswordRequest struct {
}

// @ID resetPassword
// @Summary reset password
// @Tags user
// @Description reset password after login
// @Accept json
// @Produce json
// @Param outcome body ResetPasswordRequest true "user reset password"
// @Success 200
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /users/reset_password [post]
func (s *Server) resetPassword(c *gin.Context) {
	// TODO
}
