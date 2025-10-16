package usecases

import (
	"errors"
	"testing"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
)

func TestAlertPendingRepeatRecovery(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	a := entity.Alert{ID: "a", Template: "{ .status }: {.condition_query}", RawQuery: "up == 0", WaitTimeBeforeAlerting: entity.Duration(30 * time.Second), RepeatInterval: entity.Duration(time.Minute)}
	if EvaluateAlert(&a, now, true, nil) || a.Status != "Ok" {
		t.Fatal("fired before wait interval")
	}
	if !EvaluateAlert(&a, now.Add(30*time.Second), true, nil) || a.Status != "Firing" {
		t.Fatal("did not fire after wait interval")
	}
	if got := RenderAlert(a); got != "Firing: up == 0" {
		t.Fatal(got)
	}
	if EvaluateAlert(&a, now.Add(45*time.Second), true, nil) {
		t.Fatal("repeat interval ignored")
	}
	if !EvaluateAlert(&a, now.Add(90*time.Second), true, nil) {
		t.Fatal("repeat notification missing")
	}
	if EvaluateAlert(&a, now.Add(100*time.Second), false, errors.New("upstream failed")) || a.Status != "Error" {
		t.Fatal("upstream failure was treated as recovery")
	}
	if EvaluateAlert(&a, now.Add(110*time.Second), false, nil) || a.Status != "Ok" || a.LastNotifiedAt != nil || a.PendingSince != nil {
		t.Fatal("recovery did not reset alert")
	}
}
