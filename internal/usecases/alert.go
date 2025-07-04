package usecases

import (
	"regexp"
	"strings"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
)

var TemplateVariables = []string{".id", ".status", ".condition_query", ".previous_execution", ".next_execution", ".repeat_interval", ".wait_time_before_alerting"}
var placeholder = regexp.MustCompile(`\{\s*(\.\w+)\s*\}`)

func RenderAlert(a entity.Alert) string {
	values := map[string]string{".id": a.ID, ".status": a.Status, ".condition_query": a.RawQuery, ".previous_execution": a.PreviousExecutionTime.Format("02/01/2006 15:04"), ".next_execution": a.NextExecutionTime.Format("02/01/2006 15:04"), ".repeat_interval": a.RepeatInterval.String(), ".wait_time_before_alerting": a.WaitTimeBeforeAlerting.String()}
	return placeholder.ReplaceAllStringFunc(a.Template, func(token string) string {
		key := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(token, "{"), "}"))
		if v, ok := values[key]; ok {
			return v
		}
		return token
	})
}

// EvaluateAlert returns whether a notification should be queued.
func EvaluateAlert(a *entity.Alert, now time.Time, firing bool, queryErr error) bool {
	a.PreviousExecutionTime = now
	a.NextExecutionTime = now.Add(10 * time.Second)
	if queryErr != nil {
		a.Status = "Error"
		a.PendingSince = nil
		return false
	}
	if !firing {
		a.Status = "Ok"
		a.PendingSince = nil
		a.LastNotifiedAt = nil
		return false
	}
	if a.PendingSince == nil {
		pending := now
		a.PendingSince = &pending
	}
	if now.Sub(*a.PendingSince) < time.Duration(a.WaitTimeBeforeAlerting) {
		a.Status = "Ok"
		return false
	}
	a.Status = "Firing"
	if a.LastNotifiedAt != nil && now.Sub(*a.LastNotifiedAt) < time.Duration(a.RepeatInterval) {
		return false
	}
	notified := now
	a.LastNotifiedAt = &notified
	return true
}
