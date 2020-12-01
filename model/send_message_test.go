package model

import (
	"context"
	"testing"
)

func TestTencentSMS_SendSms(t *testing.T) {
	Setup()
	err := GetSMSSender().SendSms(context.Background(), []string{"15221776376"}, "174337")
	if err != nil {
		t.Fatal(err)
	}
}

func TestAwsSesModel_SendEmail(t *testing.T) {
	GetEmailModel().SendEmail(context.Background(), "", "", "", "")
	Setup()
}
