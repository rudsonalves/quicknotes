package mailer

import "gopkg.in/gomail.v2"

type SMTPConfig struct {
	From     string
	Host     string
	Port     int
	UserName string
	Password string
}

type smtpMailService struct {
	from   string
	dialer *gomail.Dialer
}

func NewSMTPMailService(cfg SMTPConfig) MailService {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.UserName, cfg.Password)
	return &smtpMailService{from: cfg.From, dialer: dialer}
}

func (ss *smtpMailService) Send(msg MailMessage) error {
	m := gomail.NewMessage()
	m.SetHeader("From", ss.from)
	m.SetHeader("To", msg.To...)
	// m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", msg.Subject)
	if msg.IsHtml {
		m.SetBody("text/html", string(msg.Body))
	} else {
		m.SetBody("text/plain", string(msg.Body))
	}
	// m.Attach("/home/Alex/lolcat.jpg")

	return ss.dialer.DialAndSend(m)
}
