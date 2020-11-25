package api

import (
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"github.com/gin-gonic/gin"
)

type LoginReq struct {
	Account  string `json:"account" form:"account"`
	Password string `json:"password" form:"password"`
	SmsCode  string `json:"sms_code" form:"sms_code"`
	WxCode   string `json:"code" form:"code"`
	State    string `json:"state" form:"state"`
}

// @ID userLogin
// @Summary login
// @Tags user
// @Description user login
// @Accept json
// @Produce json
// @Param outcome body LoginReq true "user login"
// @Success 200 {object} {"data": "ok"}
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /users/login [get]
func (s *Server) login(c *gin.Context) {
	var req LoginReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		log.Error(c.Request.Context(), "login:ShouldBindQuery failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	// TODO
	if req.WxCode != "" {
		c.SetCookie("access", "token", 0, "", "kidsloop.cn", true, true)
		c.JSON(http.StatusOK, "ok")
		return
	}
	if req.SmsCode != "" {
		c.SetCookie("access", "token", 0, "", "kidsloop.cn", true, true)
		c.JSON(http.StatusOK, "ok")
		return
	}

	if req.Password != "" && req.Account != "" {
		c.SetCookie("access", "token", 0, "", "kidsloop.cn", true, true)
		c.JSON(http.StatusOK, "ok")
		return
	}

	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
}

type RegisterReq struct {
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
// @Param outcome body RegisterReq true "user register"
// @Success 200 {object} {"data": "ok"}
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /users/register [post]
func (s *Server) register(c *gin.Context) {
	// TODO
}

type SendTemporaryCredentialReq struct {
	Mobile string `json:"mobile"`
	Email  string `json:"email"`
}

// @ID sendTemporaryCredential
// @Summary temporary credential
// @Tags user
// @Description send credential
// @Accept json
// @Produce json
// @Param outcome body SendTemporaryCredentialReq true "user register"
// @Success 200 {object} {"data": "ok"}
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /users/temp_code [post]
func (s *Server) sendTemporaryCredential(c *gin.Context) {
	// TODO
}

type ForgottenPwdReq struct {
}

// @ID forgottenPwd
// @Summary forget password
// @Tags user
// @Description forget password
// @Accept json
// @Produce json
// @Param outcome body ForgottenPwdReq true "login by new password and update password"
// @Success 200 {object} {"data": "ok"}
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /users/forgotten_pwd [post]
func (s *Server) forgottenPwd(c *gin.Context) {
	// TODO
}

type ResetPasswordReq struct {
}

// @ID resetPassword
// @Summary reset password
// @Tags user
// @Description reset password after login
// @Accept json
// @Produce json
// @Param outcome body ResetPasswordReq true "user reset password"
// @Success 200 {object} {"data": "ok"}
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /users/reset_password [post]
func (s *Server) resetPassword(c *gin.Context) {
	// TODO
}
