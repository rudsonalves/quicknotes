package mailer

import (
	"fmt"
	"strings"
)

type consoleMailService struct {
	from string
}

func NewConsoleMailService(from string) MailService {
	return &consoleMailService{from: from}
}

// Send implements MailService.
func (cs *consoleMailService) Send(msg MailMessage) error {
	fmt.Printf("\nFrom: %s\nTo: [%s]\nSubject: %s\nBody: %s\n\n",
		cs.from, strings.Join(msg.To, ", "), msg.Subject, msg.Body,
	)

	return nil
}
