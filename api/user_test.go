package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/KL-Engineering/kidsloop-cms-service/constant"
)

func TestUserLogin(t *testing.T) {
	req := LoginRequest{
		AuthTo:   phone,
		AuthCode: "Bada123456",
		AuthType: constant.LoginByPassword,
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	res := DoHttp(http.MethodPost, prefix+"/users/login", string(data))
	fmt.Println(res)
}

// 15026743257
// 15221776376
var phone = "+8615221776376"

func TestUserRegister(t *testing.T) {
	req := RegisterRequest{
		Account:  phone,
		AuthCode: "861782",
		Password: "Bada123456",
		ActType:  constant.AccountPhone,
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
	res := DoHttp(http.MethodPost, prefix+"/users/register", string(data))
	fmt.Println(res)
}

func TestUserSendCode(t *testing.T) {
	req := SendCodeRequest{
		Mobile: phone,
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
	res := DoHttp(http.MethodPost, prefix+"/users/send_code", string(data))
	fmt.Println(res)
}

func TestUserForgetPassword(t *testing.T) {

	req := ForgottenPasswordRequest{
		AuthTo:   phone,
		AuthCode: "972031",
		Password: "Bada123456",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
	res := DoHttp(http.MethodPost, prefix+"/users/forgotten_pwd", string(data))
	fmt.Println(res)
}

func TestUserResetPassword(t *testing.T) {
	req := ResetPasswordRequest{
		OldPassword: "Bada1234",
		NewPassword: "12345",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
	res := DoHttp(http.MethodPost, prefix+"/users/reset_password", string(data))
	fmt.Println(res)
}

func TestUserCheckAccount(t *testing.T) {
	res := DoHttp(http.MethodGet, prefix+"/users/check_account?account="+"15026743257", "")
	fmt.Println(res)
}
