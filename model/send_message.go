package model

import (
	"context"
	"strings"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"

	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"
	"gopkg.in/gomail.v2"

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
		receivers[idx] = addMobilePrefix(mobile, w.MobilePrefix)
	}
	credential := common.NewCredential(w.SecretID, w.SecretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = w.EndPoint
	client, _ := sms.NewClient(credential, "", cpf)

	request := sms.NewSendSmsRequest()
	assignStringList(receivers, &request.PhoneNumberSet)
	request.Sign = &(w.Sign)
	request.SmsSdkAppid = &(w.SDKAppID)
	request.TemplateID = &(w.TemplateID)
	assignStringList([]string{msg, w.TemplateParamSet}, &request.TemplateParamSet)

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

func assignStringList(src []string, dst *[]*string) {
	for i := range src {
		*dst = append(*dst, &src[i])
	}
}

// add head to mobile, e.g. in China, +86
func addMobilePrefix(mobile, head string) string {
	sb := strings.Builder{}
	if !strings.HasPrefix(mobile, head) {
		sb.WriteString(head)
		sb.WriteString(mobile)
	}
	return sb.String()
}

type IEmailModel interface {
	SendEmail(ctx context.Context, recipient string, title string, html string, plain string) error
}

// AwsSesModel ses model
type AwsSesModel struct {
	Host          string `json:"host"`
	Port          int    `json:"port"`
	User          string `json:"user"`
	Password      string `json:"password"`
	SenderAddress string `json:"sender_address"`
	SenderName    string `json:"sender_name"`
}

func (ses AwsSesModel) SendEmail(ctx context.Context, recipient string, title string, html string, plain string) error {

	m := gomail.NewMessage()

	m.SetBody("text/html", html)

	m.AddAlternative("text/plain", plain)

	// Construct the message headers, including a Configuration Set and a Tag.
	m.SetHeaders(map[string][]string{
		"From":    {m.FormatAddress(ses.SenderAddress, ses.SenderName)},
		"To":      {recipient},
		"Subject": {title},
		// Comment or remove the next line if you are not using a configuration set
		//"X-SES-CONFIGURATION-SET": {ConfigSet},
		// Comment or remove the next line if you are not using custom tags
		"X-SES-MESSAGE-TAGS": {"genre=test,genre2=test2"},
	})

	// Send the email.
	d := gomail.NewDialer(ses.Host, ses.Port, ses.User, ses.Password)

	// Display an error message if something goes wrong; otherwise,
	// display a message confirming that the message was sent.
	if err := d.DialAndSend(m); err != nil {
		logger.WithContext(ctx).
			WithStacks().
			WithError(err).
			WithField("ses", ses).
			WithField("recipient", recipient).
			Debug("send mail failed")
		return err
	}
	return nil
}

var (
	_emailOnce  sync.Once
	_emailModel IEmailModel
)

// GetUserModel get user logic
func GetEmailModel() IEmailModel {
	_emailOnce.Do(func() {
		_emailModel = &AwsSesModel{
			Host:          "", // config.Get().PrivateConfig.Social.Email.Host,
			Port:          0,  // config.Get().PrivateConfig.Social.Email.Port,
			User:          "", // config.Get().PrivateConfig.Social.Email.UserName,
			Password:      "", // config.Get().PrivateConfig.Social.Email.Secret,
			SenderAddress: "", // config.Get().PrivateConfig.Social.Email.Address,
			SenderName:    "", //config.Get().PrivateConfig.Social.Email.SenderName,
		}
	})
	return _emailModel
}
