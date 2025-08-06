package service

import (
	"fmt"
	"net/smtp"
	"strconv"

	"github.com/NOTMKW/JWT/internal/config"
)

type EmailService struct {
	config *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{config: cfg}
}

func (e *EmailService) SendMFACode(email, code string) error {
	auth := smtp.PlainAuth("", e.config.SMTPUsername, e.config.SMTPPassword, e.config.SMTPHost)

	subject := "Your MFA Code"
	body := fmt.Sprintf("Your verification code is %s \n\n This Code will expire in 5 minutes.", code)

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", email, subject, body))

	addr := e.config.SMTPHost + ":" + strconv.Itoa(e.config.SMTPPort)

	return smtp.SendMail(addr, auth, e.config.FromEmail, []string{email}, msg)
}
