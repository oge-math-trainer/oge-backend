package mail

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	stdmail "net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type SMTPConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

type SMTPMailer struct {
	config SMTPConfig
}

const smtpTimeout = 15 * time.Second

func NewSMTPMailer(config SMTPConfig) *SMTPMailer {
	config.Host = strings.TrimSpace(config.Host)
	config.Username = strings.TrimSpace(config.Username)
	config.Password = strings.TrimSpace(config.Password)
	config.FromEmail = strings.TrimSpace(config.FromEmail)
	config.FromName = strings.TrimSpace(config.FromName)
	return &SMTPMailer{config: config}
}

func (m *SMTPMailer) Configured() bool {
	return m != nil &&
		m.config.Host != "" &&
		m.config.Port > 0 &&
		m.config.FromEmail != ""
}

func (m *SMTPMailer) SendPasswordResetCode(ctx context.Context, email, code string, expiresAt time.Time) error {
	if !m.Configured() {
		return fmt.Errorf("smtp mailer is not configured")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	from := stdmail.Address{Name: m.config.FromName, Address: m.config.FromEmail}
	to := stdmail.Address{Address: email}
	subject := mime.QEncoding.Encode("utf-8", "Код восстановления пароля")
	body := fmt.Sprintf(
		"Ваш код для восстановления пароля: %s\n\nКод действует до %s.\nЕсли вы не запрашивали восстановление пароля, просто проигнорируйте это письмо.\n",
		code,
		expiresAt.Format("02.01.2006 15:04 MST"),
	)

	var msg bytes.Buffer
	msg.WriteString("From: " + from.String() + "\r\n")
	msg.WriteString("To: " + to.String() + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)

	addr := m.config.Host + ":" + strconv.Itoa(m.config.Port)
	var auth smtp.Auth
	if m.config.Username != "" || m.config.Password != "" {
		auth = smtp.PlainAuth("", m.config.Username, m.config.Password, m.config.Host)
	}
	return sendMail(ctx, addr, m.config.Host, auth, m.config.FromEmail, []string{email}, msg.Bytes())
}

func sendMail(ctx context.Context, addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, smtpTimeout)
	defer cancel()

	dialer := &net.Dialer{Timeout: smtpTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		config := &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: host,
		}
		if err := client.StartTLS(config); err != nil {
			return err
		}
	}
	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return err
			}
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}
