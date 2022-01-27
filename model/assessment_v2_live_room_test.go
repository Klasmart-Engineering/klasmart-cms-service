package model

import (
	"context"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func TestGetTree(t *testing.T) {
	ctx := context.Background()
	roomInfo, err := getAssessmentLiveRoom().getRoomResultInfo(ctx, &entity.Operator{
		UserID: "",
		OrgID:  "6300b3c5-8936-497e-ba1f-d67164b59c65",
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6Ijk3ZTNmNTliLTQxZDQtNDJhMC1hZjIwLTY2M2MzY2IxMmM1YyIsImVtYWlsIjoidGVjaDFAeW9wbWFpbC5jb20iLCJleHAiOjE2Mzg1MjAyNzEsImlzcyI6ImtpZHNsb29wIn0.OYNsbVccSisKZ33r5ERCieE-d7lU_Zu2ldYg8lsPVDRA9PT-apoTG5Wz5DBzeEsM8ff0iZwK0po3nIUOMgLkytIrjv8Kl1BMHIX8Wj_voGoy4Ym7O4bUizm3m9BwLJ3q3QbNvT5PQq01MT_kUy2pMcHLqZ99DHKg1S73ufxv0E-cMfakeuwezoWajRKSa4t_rTbSCULZ2LITXrvCcti4ZOO9nWwe_J6Nzy0cfA3XwBMQA2VS5MrG0U-9Elzgrcwp91dxnr_agdBMeIwaW19tP02YSdB5klrRuML7cSp2rTeP9chYGKtyhsjkAmEJInxpVGVX6u4EBCZFxCYNqeOu3Q",
	}, "616e3154a9aa027b852f8297")

	if err != nil {
		t.Fatal(err)
	}
	t.Log((roomInfo))
}
