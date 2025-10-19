package notification

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

func (s *Handler) validateNotification(destination, raw string) error {
	var settings struct {
		URL, TelegramTokenPath, ChatID, DestinationEmail, SmtpServer, PasswordPath string
		Port                                                                       int
	}
	if json.Unmarshal([]byte(raw), &settings) != nil {
		return fmt.Errorf("%w: settings must contain JSON", entity.ErrInvalid)
	}
	switch strings.ToLower(destination) {
	case "webhook":
		if _, err := s.provider.Client.Validate(settings.URL); err != nil {
			return fmt.Errorf("%w: %v", entity.ErrInvalid, err)
		}
	case "telegram":
		return usecases.Required(settings.TelegramTokenPath, settings.ChatID)
	case "email":
		if settings.Port < 1 || settings.Port > 65535 {
			return entity.ErrInvalid
		}
		if strings.ContainsAny(settings.DestinationEmail, "\r\n") {
			return entity.ErrInvalid
		}
		if _, err := s.provider.Client.Validate("https://" + settings.SmtpServer + ":" + strconv.Itoa(settings.Port)); err != nil {
			return fmt.Errorf("%w: SMTP host is not allowed", entity.ErrInvalid)
		}
		return usecases.Required(settings.DestinationEmail, settings.SmtpServer, settings.PasswordPath)
	default:
		return nil
	}
	return nil
}
