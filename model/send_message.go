package model

import (
	"context"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20190711"
)

type TencentSMS struct {
	SDKAppID         string
	SecretID         string
	SecretKey        string
	EndPoint         string
	Sign             string
	TemplateID       string
	TemplateParamSet string
	MobilePrefix     string
}

func GetSMSSender() *TencentSMS {
	sender := &TencentSMS{
		SDKAppID:         config.Get().TencentConfig.Sms.SDKAppID,
		SecretID:         config.Get().TencentConfig.Sms.SecretID,
		SecretKey:        config.Get().TencentConfig.Sms.SecretKey,
		EndPoint:         config.Get().TencentConfig.Sms.EndPoint,
		Sign:             config.Get().TencentConfig.Sms.Sign,
		TemplateID:       config.Get().TencentConfig.Sms.TemplateID,
		TemplateParamSet: config.Get().TencentConfig.Sms.TemplateParamSet,
		MobilePrefix:     config.Get().TencentConfig.Sms.MobilePrefix,
	}
	return sender
}

func (w *TencentSMS) SendSms(ctx context.Context, receivers []string, msg string) (err error) {
	for idx, mobile := range receivers {
		receivers[idx] = w.addMobilePrefix(mobile, w.MobilePrefix)
	}
	credential := common.NewCredential(w.SecretID, w.SecretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = w.EndPoint
	client, _ := sms.NewClient(credential, "", cpf)

	request := sms.NewSendSmsRequest()
	w.assignStringList(receivers, &request.PhoneNumberSet)
	request.Sign = &(w.Sign)
	request.SmsSdkAppid = &(w.SDKAppID)
	request.TemplateID = &(w.TemplateID)
	w.assignStringList([]string{msg, w.TemplateParamSet}, &request.TemplateParamSet)

	resp, err := client.SendSms(request)
	if err != nil {
		log.Error(ctx, "SendSms: failed", log.Strings("receives", receivers), log.String("msg", msg), log.Err(err))
		return
	}

	for _, r := range resp.Response.SendStatusSet {
		if *r.Fee == 0 {
			log.Warn(ctx, "SendSms: response error", log.String("failed", *(r.Message)))
		}
	}

	return
}

func (w *TencentSMS) assignStringList(src []string, dst *[]*string) {
	for i := range src {
		*dst = append(*dst, &src[i])
	}
}

// add head to mobile, e.g. in China, +86
func (w *TencentSMS) addMobilePrefix(mobile, head string) string {
	sb := strings.Builder{}
	if !strings.HasPrefix(mobile, head) {
		sb.WriteString(head)
		sb.WriteString(mobile)
	}
	return sb.String()
}
