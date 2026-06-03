package notification

import "log"

type Service interface {
	SendEmail(to string, subject string, body string) error
	SendSMS(phone string, message string) error
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) SendEmail(to string, subject string, body string) error {
	log.Printf("[email] to=%s subject=%q body=%q", to, subject, body)
	return nil
}

func (s *service) SendSMS(phone string, message string) error {
	log.Printf("[sms] phone=%s message=%q", phone, message)
	return nil
}
