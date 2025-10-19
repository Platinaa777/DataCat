package usecases

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Platinaa777/DataCat/internal/entity"
)

func (s *Service) validateConnection(raw string) error {
	var cfg struct{ ServerURL, ClusterURL string }
	if json.Unmarshal([]byte(raw), &cfg) != nil {
		return fmt.Errorf("%w: connectionString must contain JSON", entity.ErrInvalid)
	}
	u := cfg.ServerURL
	if u == "" {
		u = cfg.ClusterURL
	}
	if _, err := s.Endpoints.Validate(u); err != nil {
		return fmt.Errorf("%w: %v", entity.ErrInvalid, err)
	}
	return nil
}

func (s *Service) validateNotification(destination, raw string) error {
	var settings struct {
		URL, TelegramTokenPath, ChatID, DestinationEmail, SmtpServer, PasswordPath string
		Port                                                                       int
	}
	if json.Unmarshal([]byte(raw), &settings) != nil {
		return fmt.Errorf("%w: settings must contain JSON", entity.ErrInvalid)
	}
	switch strings.ToLower(destination) {
	case "webhook":
		if _, err := s.Endpoints.Validate(settings.URL); err != nil {
			return fmt.Errorf("%w: %v", entity.ErrInvalid, err)
		}
	case "telegram":
		return Required(settings.TelegramTokenPath, settings.ChatID)
	case "email":
		if settings.Port < 1 || settings.Port > 65535 {
			return entity.ErrInvalid
		}
		if strings.ContainsAny(settings.DestinationEmail, "\r\n") {
			return entity.ErrInvalid
		}
		if _, err := s.Endpoints.Validate("https://" + settings.SmtpServer + ":" + strconv.Itoa(settings.Port)); err != nil {
			return fmt.Errorf("%w: SMTP host is not allowed", entity.ErrInvalid)
		}
		return Required(settings.DestinationEmail, settings.SmtpServer, settings.PasswordPath)
	default:
		return nil
	}
	return nil
}
