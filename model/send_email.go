package model

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gopkg.in/gomail.v2"
)

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
		log.Error(ctx, "SendEmail: DialAndSend failed", log.String("recipient", recipient), log.Any("ses", ses), log.Err(err))
		return err
	}
	return nil
}

var (
	_emailOnce  sync.Once
	_emailModel IEmailModel
)

// GetEmailModel get user logic
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
