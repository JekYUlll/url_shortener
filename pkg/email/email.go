package email

import (
	"fmt"
	"net/smtp"

	"github.com/jekyulll/url_shortener/config"
	mail "github.com/jordan-wright/email"
)

type EmailSend struct {
	addr    string
	myEail  string
	subject string
	auth    smtp.Auth
}

func NewEmailSend(cfg config.EmailConfig) (*EmailSend, error) {
	emailSend := &EmailSend{
		addr:    fmt.Sprintf("%s:%s", cfg.HostAddress, cfg.HostPort),
		auth:    smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.HostAddress),
		myEail:  cfg.Username,
		subject: cfg.Subject,
	}
	if err := emailSend.Send(cfg.TestMail, "test email"); err != nil {
		fmt.Printf("send email failed: %v\n", err)
		return nil, err
	}
	return emailSend, nil
}

func (e *EmailSend) Send(email string, emailCode string) error {
	instance := mail.NewEmail()
	instance.From = e.myEail
	instance.To = []string{email}
	instance.Subject = e.subject
	instance.Text = []byte(fmt.Sprintf("Your Verification code is: %s", emailCode))

	return instance.Send(e.addr, e.auth)
}
