package notifications

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/extapi/httpclient"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

type Sender struct {
	Client  *httpclient.Client
	Service *usecases.Service
}
type settings struct {
	URL               string
	TelegramTokenPath string
	ChatID            string
	DestinationEmail  string
	SmtpServer        string
	Port              int
	PasswordPath      string
}

func (s *Sender) Send(ctx context.Context, ch entity.NotificationChannel, message string) error {
	var cfg settings
	if err := json.Unmarshal([]byte(ch.Settings), &cfg); err != nil {
		return fmt.Errorf("invalid notification settings")
	}
	switch strings.ToLower(ch.DestinationName) {
	case "webhook":
		return s.Client.Do(ctx, "POST", cfg.URL, message, nil, nil)
	case "telegram":
		token, err := s.Service.Secret(ctx, ch.NamespaceID, cfg.TelegramTokenPath)
		if err != nil {
			return err
		}
		if token == "" || strings.ContainsAny(token, "/?#\r\n") {
			return fmt.Errorf("invalid Telegram token")
		}
		var response struct {
			OK bool `json:"ok"`
		}
		if err = s.Client.Do(ctx, "POST", "https://api.telegram.org/bot"+token+"/sendMessage", map[string]string{"chat_id": cfg.ChatID, "text": message}, nil, &response); err != nil {
			return err
		}
		if !response.OK {
			return fmt.Errorf("Telegram rejected notification")
		}
		return nil
	case "email":
		return s.email(ctx, ch.NamespaceID, cfg, message)
	default:
		p, _, err := usecases.Read[entity.Plugin](ctx, s.Service.Store, usecases.Global, "plugin", ch.DestinationName)
		if err != nil {
			return err
		}
		if !p.IsActive {
			return fmt.Errorf("notification plugin is disabled")
		}
		return s.Client.Do(ctx, "POST", p.Endpoint, map[string]any{"version": 1, "operation": "notifications.send", "namespaceId": ch.NamespaceID, "settings": json.RawMessage(ch.Settings), "input": map[string]string{"message": message}}, nil, nil)
	}
}
func (s *Sender) email(ctx context.Context, ns string, cfg settings, message string) error {
	if strings.ContainsAny(cfg.DestinationEmail, "\r\n") {
		return fmt.Errorf("invalid email address")
	}
	if _, err := mail.ParseAddress(cfg.DestinationEmail); err != nil {
		return fmt.Errorf("invalid email address")
	}
	address := net.JoinHostPort(cfg.SmtpServer, strconv.Itoa(cfg.Port))
	if _, err := s.Client.Validate("https://" + address); err != nil {
		return err
	}
	password, err := s.Service.Secret(ctx, ns, cfg.PasswordPath)
	if err != nil {
		return err
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	connection, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("connect SMTP server failed")
	}
	defer connection.Close()
	deadline := time.Now().Add(15 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	connection.SetDeadline(deadline)
	var transport net.Conn = connection
	if cfg.Port == 465 {
		secured := tls.Client(connection, &tls.Config{ServerName: cfg.SmtpServer, MinVersion: tls.VersionTLS12})
		if err = secured.HandshakeContext(ctx); err != nil {
			return fmt.Errorf("SMTP TLS handshake failed")
		}
		transport = secured
	}
	client, err := smtp.NewClient(transport, cfg.SmtpServer)
	if err != nil {
		return fmt.Errorf("start SMTP session failed")
	}
	defer client.Close()
	if cfg.Port != 465 {
		if err = client.StartTLS(&tls.Config{ServerName: cfg.SmtpServer, MinVersion: tls.VersionTLS12}); err != nil {
			return fmt.Errorf("SMTP STARTTLS required")
		}
	}
	if err = client.Auth(smtp.PlainAuth("", cfg.DestinationEmail, password, cfg.SmtpServer)); err != nil {
		return fmt.Errorf("SMTP authentication failed")
	}
	if err = client.Mail(cfg.DestinationEmail); err != nil {
		return fmt.Errorf("SMTP sender rejected")
	}
	if err = client.Rcpt(cfg.DestinationEmail); err != nil {
		return fmt.Errorf("SMTP recipient rejected")
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP data rejected")
	}
	content := "From: " + cfg.DestinationEmail + "\r\nTo: " + cfg.DestinationEmail + "\r\nSubject: DataCat alert\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + message
	if _, err = w.Write([]byte(content)); err != nil {
		return fmt.Errorf("write SMTP message failed")
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("SMTP message rejected")
	}
	return client.Quit()
}
