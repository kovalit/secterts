// Package mailer sends transactional email (2FA codes). In dev, when no SMTP
// host is configured, codes are logged to stdout instead of being sent.
package mailer

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"

	gomail "gopkg.in/mail.v2"
)

// Mailer sends plaintext emails.
type Mailer interface {
	Send(to, subject, body string) error
}

// Config for the SMTP mailer.
type Config struct {
	Host        string
	Port        int
	User        string
	Password    string
	InsecureTLS bool
	From        string
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
	if to == "" {
		return fmt.Errorf("mailer: recipient is required")
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.cfg.From)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)

	dialer := gomail.NewDialer(m.cfg.Host, m.cfg.Port, m.cfg.User, m.cfg.Password)
	dialer.Timeout = 15 * time.Second
	dialer.TLSConfig = &tls.Config{
		MinVersion:         tls.VersionTLS12,
		ServerName:         m.cfg.Host,
		InsecureSkipVerify: m.cfg.InsecureTLS, // explicitly opt-in for local SMTP only
	}

	if err := dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("mailer: send via %s:%d: %w", m.cfg.Host, m.cfg.Port, err)
	}
	return nil
}
