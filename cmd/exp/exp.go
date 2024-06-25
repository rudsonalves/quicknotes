package main

import (
	"github.com/rudsonalves/quicknotes/internal/mailer"
	"gopkg.in/gomail.v2"
)

func main1() {
	m := gomail.NewMessage()
	m.SetHeader("From", "quicknotes@quick.com")
	m.SetHeader("To", "alvestest67@gmail.com", "alvesdev67@gmail.com")
	// m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", "Bem Vindo!")
	m.SetBody("text/html", "<h1>Hello</h1> <p>Seja bem vindo ao <b>Quicknote</b>!</p>")
	// m.Attach("/home/Alex/lolcat.jpg")

	d := gomail.NewDialer("localhost", 1025, "", "")

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

func main2() {
	// testar envio de email
	msg := mailer.MailMessage{
		To:      []string{"alvesdev67@gmail.com", "alvestest67@gmail.com"},
		Subject: "Bem vindo2",
		Body:    []byte("Seja bem vindo ao Quicknotes!"),
		IsHtml:  false,
	}

	smtp := mailer.SMTPConfig{
		Host:     "localhost",
		Port:     1025,
		UserName: "",
		Password: "",
		From:     "quicknote@quick.com",
	}

	mailservice := mailer.NewSmtpMailService(smtp)
	mailservice.Send(msg)
}

func main() {
	main2()
}
