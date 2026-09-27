// Package mailer sends transactional email (2FA codes). In dev, when no SMTP
// host is configured, codes are logged to stdout instead of being sent.
package mailer

import (
	"fmt"
	"log"
	"net/smtp"
)

// Mailer sends plaintext emails.
type Mailer interface {
	Send(to, subject, body string) error
}

// Config for the SMTP mailer.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// New returns an SMTP mailer, or a stdout logger mailer when Host is empty.
func New(cfg Config) Mailer {
	if cfg.Host == "" {
		log.Printf("mailer: SMTP_HOST not set — emails (incl. 2FA codes) will be logged to stdout")
		return &logMailer{from: cfg.From}
	}
	return &smtpMailer{cfg: cfg}
}

type logMailer struct{ from string }

func (m *logMailer) Send(to, subject, body string) error {
	log.Printf("mailer(dev): to=%s subject=%q\n%s", to, subject, body)
	return nil
}

type smtpMailer struct{ cfg Config }

func (m *smtpMailer) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.cfg.From, to, subject, body)

	var auth smtp.Auth
	if m.cfg.User != "" {
		auth = smtp.PlainAuth("", m.cfg.User, m.cfg.Password, m.cfg.Host)
	}
	return smtp.SendMail(addr, auth, m.cfg.From, []string{to}, []byte(msg))
}
