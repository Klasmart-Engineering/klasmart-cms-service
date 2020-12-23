package api

import (
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	AuthTo   string `json:"auth_to" form:"auth_to"`
	AuthCode string `json:"auth_code" form:"auth_code"`
	AuthType string `json:"auth_type" from:"auth_type"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// @ID userLogin
// @Summary login
// @Tags user
// @Description user login
// @Accept json
// @Produce json
// @Param outcome body LoginRequest true "user login"
// @Success 200	{object} LoginResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 401 {object} UnAuthorizedResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /users/login [post]
func (s *Server) login(c *gin.Context) {
	ctx := c.Request.Context()
	var req LoginRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Warn(ctx, "login:ShouldBindQuery failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	if req.AuthType != constant.LoginByPassword && req.AuthType != constant.LoginByCode {
		log.Warn(ctx, "login:param illegal", log.Any("req", req))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	user, err := model.GetUserModel().GetUserByAccount(ctx, req.AuthTo)
	if err == constant.ErrRecordNotFound {
		log.Warn(ctx, "login:GetUserByAccount", log.Any("req", req))
		c.JSON(http.StatusUnauthorized, L(GeneralUnAuthorized))
		return
	}
	if err != nil {
		log.Error(ctx, "login:GetUserByAccount", log.Any("req", req), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	var pass bool
	if req.AuthType == constant.LoginByCode {
		pass, err = model.VerifyCode(ctx, req.AuthTo, req.AuthCode)
		if err != nil {
			log.Error(ctx, "login:GetUserByAccount", log.Any("req", req), log.Err(err))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
	}
	if req.AuthType == constant.LoginByPassword {
		pass = model.VerifySecretWithSalt(ctx, req.AuthCode, user.Secret, user.Salt)
	}
	if !pass {
		log.Warn(ctx, "login:not pass", log.Bool("pass", pass), log.Any("req", req))
		c.JSON(http.StatusUnauthorized, L(GeneralUnknown))
		return
	}

	token, err := model.GetTokenFromUser(ctx, user)
	if err != nil {
		log.Error(ctx, "login:GetUserByAccount", log.Any("req", req), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}

	c.JSON(http.StatusOK, LoginResponse{token})
	//c.SetCookie("access", token, 0, "/", config.Get().KidsloopCNLoginConfig.CookieDomain, true, true)
	//c.Status(http.StatusOK)
}

type RegisterRequest struct {
	Account  string `json:"account" form:"account"`     // 当前是电话号码
	AuthCode string `json:"auth_code" form:"auth_code"` // 验证码
	Password string `json:"password" form:"password"`   // 密码
	ActType  string `json:"act_type" form:"act_type"`   // 注册类型
}

// @ID userRegister
// @Summary register
// @Tags user
// @Description user register
// @Accept json
// @Produce json
// @Param outcome body RegisterRequest true "user register"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 401 {object} UnAuthorizedResponse
// @Failure 409 {object} ConflictResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /users/register [post]
func (s *Server) register(c *gin.Context) {
	ctx := c.Request.Context()
	var req RegisterRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Warn(ctx, "register:ShouldBindQuery failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	pass, err := model.VerifyCode(ctx, req.Account, req.AuthCode)
	if err == constant.ErrUnAuthorized {
		log.Warn(ctx, "register: VerifyCode failed", log.Any("req", req))
		c.JSON(http.StatusUnauthorized, L(GeneralUnAuthorized))
		return
	}

	if err != nil {
		log.Error(ctx, "register:VerifyCode failed", log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !pass {
		log.Warn(ctx, "register: not pass", log.Any("req", req))
		c.JSON(http.StatusUnauthorized, L(GeneralUnAuthorized))
		return
	}
	user, err := model.GetUserModel().RegisterUser(ctx, req.Account, req.Password, req.ActType)
	if err == constant.ErrDuplicateRecord {
		log.Warn(ctx, "register:RegisterUser conflict", log.Err(err))
		c.JSON(http.StatusConflict, L(GeneralUnknown))
		return
	}
	if err != nil {
		log.Error(ctx, "register:RegisterUser failed", log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}

	token, err := model.GetTokenFromUser(ctx, user)
	if err != nil {
		log.Error(ctx, "register:Token failed", log.Any("req", req), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}

	c.JSON(http.StatusOK, LoginResponse{token})
	//c.SetCookie("access", token, 0, "/", config.Get().KidsloopCNLoginConfig.CookieDomain, true, true)
	//c.Status(http.StatusOK)
}

type SendCodeRequest struct {
	Mobile string `json:"mobile" form:"mobile"`
	Email  string `json:"email" form:"email"`
}

// @ID sendCode
// @Summary send verify code
// @Tags user
// @Description send verify code or uri
// @Accept json
// @Produce json
// @Param outcome body SendCodeRequest true "send verify code"
// @Success 200
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /users/send_code [post]
func (s *Server) sendCode(c *gin.Context) {
	var req SendCodeRequest
	ctx := c.Request.Context()
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Warn(ctx, "verification:ShouldBindJSON failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	if req.Mobile != "" {
		code, err := model.GetBubbleMachine(req.Mobile).Launch(ctx)
		if err != nil {
			log.Error(ctx, "sendCode: launch failed", log.Err(err))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		err = model.GetSMSSender().SendSms(ctx, []string{req.Mobile}, code)
		if err != nil {
			log.Error(ctx, "sendCode: SendSms failed", log.Err(err))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		c.Status(http.StatusOK)
		return
	}

	if req.Email != "" {
		code, err := model.GetBubbleMachine(req.Email).Launch(ctx)
		if err != nil {
			log.Error(ctx, "sendCode: launch failed", log.Err(err))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		// TODO: uri
		err = model.GetEmailModel().SendEmail(ctx, req.Email, "", "", code)
		if err != nil {
			log.Error(ctx, "sendCode: SendSms failed", log.Err(err))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		c.JSON(http.StatusOK, "ok")
		return
	}

	log.Warn(c.Request.Context(), "verification:param wrong", log.Any("req", req))
	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	return
}

// @ID inviteNotify
// @Summary invite notify
// @Tags user
// @Description send verify code or uri
// @Accept json
// @Produce json
// @Param outcome body SendCodeRequest true "send verify code"
// @Success 200
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /users/send_code [post]
func (s *Server) inviteNotify(c *gin.Context) {
	var req SendCodeRequest
	ctx := c.Request.Context()
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Warn(ctx, "verification:ShouldBindJSON failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	if req.Mobile != "" {
		code := config.Get().KidsloopCNLoginConfig.InviteNotify
		err = model.GetSMSSender().SendSms(ctx, []string{req.Mobile}, code)
		if err != nil {
			log.Error(ctx, "sendCode: SendSms failed", log.Err(err))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		c.Status(http.StatusOK)
		return
	}

	if req.Email != "" {
		// TODO: text
		code := config.Get().KidsloopCNLoginConfig.InviteNotify
		// TODO: uri
		err = model.GetEmailModel().SendEmail(ctx, req.Email, "", "", code)
		if err != nil {
			log.Error(ctx, "sendCode: SendSms failed", log.Err(err))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		c.Status(http.StatusOK)
		return
	}

	log.Warn(c.Request.Context(), "verification:param wrong", log.Any("req", req))
	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	return
}

type ForgottenPasswordRequest struct {
	AuthTo   string `json:"auth_to" form:"auth_to"`
	AuthCode string `json:"auth_code" form:"auth_code"`
	Password string `json:"password" form:"password"`
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
	ctx := c.Request.Context()
	var req ForgottenPasswordRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Warn(ctx, "forgottenPassword: ShouldBindJson failed", log.Any("req", req), log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	pass, err := model.VerifyCode(ctx, req.AuthTo, req.AuthCode)
	if err != nil {
		log.Error(ctx, "forgottenPassword:VerifyCode failed", log.Any("req", req), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !pass {
		log.Warn(ctx, "forgottenPassword:VerifyCode failed", log.Any("req", req), log.Err(err))
		c.JSON(http.StatusUnauthorized, L(GeneralUnAuthorized))
		return
	}
	user, err := model.GetUserModel().UpdateAccountPassword(ctx, req.AuthTo, req.Password)
	if err != nil {
		log.Error(ctx, "forgottenPassword:UpdateAccountPassword failed", log.Any("req", req), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	token, err := model.GetTokenFromUser(ctx, user)
	if err != nil {
		log.Error(ctx, "forgottenPassword:GetTokenFromUser failed", log.Any("req", req), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}

	c.JSON(http.StatusOK, LoginResponse{token})
}

type ResetPasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
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
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var req ResetPasswordRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Warn(ctx, "resetPassword: ShouldBindJSON failed")
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	err = model.GetUserModel().ResetUserPassword(ctx, op.UserID, req.OldPassword, req.NewPassword)
	if err != nil {
		log.Error(ctx, "resetPassword:ResetUserPassword failed", log.Any("req", req), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.Status(http.StatusOK)
}

type CheckAccountResponse struct {
	Status string `json:"status"`
}

// @ID checkAccount
// @Summary checkAccount
// @Tags user
// @Description check account register
// @Accept json
// @Produce json
// @Param account query string true "account"
// @Success 200 {object} CheckAccountResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 409
// @Failure 500 {object} InternalServerErrorResponse
// @Router /users/check_account [get]
func (s *Server) checkAccount(c *gin.Context) {
	ctx := c.Request.Context()
	var req struct {
		Account string `json:"account" form:"account"`
	}
	err := c.ShouldBindQuery(&req)
	if err != nil || req.Account == "" {
		log.Warn(ctx, "checkAccount:ShouldBindQuery failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	user, err := model.GetUserModel().GetUserByAccount(ctx, req.Account)
	if err == constant.ErrRecordNotFound {
		c.JSON(http.StatusOK, CheckAccountResponse{constant.AccountUnregister})
		return
	}
	if user != nil {
		c.JSON(http.StatusOK, CheckAccountResponse{constant.AccountExist})
		return
	}

	log.Error(ctx, "checkAccount: error", log.String("account", req.Account), log.Err(err))
	c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
}
