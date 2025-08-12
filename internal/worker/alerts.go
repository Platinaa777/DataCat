package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/extapi/notifications"
	"github.com/Platinaa777/DataCat/internal/extapi/telemetry"
	"github.com/Platinaa777/DataCat/internal/repository/postgres"
	"github.com/Platinaa777/DataCat/internal/usecases"
	"github.com/jackc/pgx/v5"
)

type Worker struct {
	Store     *postgres.Store
	Service   *usecases.Service
	Telemetry *telemetry.Provider
	Sender    *notifications.Sender
	Log       *slog.Logger
}

func (w *Worker) Run(ctx context.Context) {
	ticks := 0
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ticks++
			if ticks%3600 == 0 {
				if _, err := w.Store.Pool().Exec(ctx, "DELETE FROM datacat.metric_batches WHERE received_at<now()-interval '7 days'"); err != nil {
					w.Log.WarnContext(ctx, "metric retention failed")
				}
				if _, err := w.Store.Pool().Exec(ctx, "DELETE FROM datacat.outbox WHERE delivered_at<now()-interval '30 days'"); err != nil {
					w.Log.WarnContext(ctx, "outbox retention failed")
				}
			}
			for i := 0; i < 16; i++ {
				processed, err := w.alert(ctx)
				if err != nil {
					w.Log.ErrorContext(ctx, "alert worker failed", "error", err)
					break
				}
				if !processed {
					break
				}
			}
			for i := 0; i < 32; i++ {
				processed, err := w.deliver(ctx)
				if err != nil {
					w.Log.ErrorContext(ctx, "outbox worker failed", "error", err)
					break
				}
				if !processed {
					break
				}
			}
		}
	}
}
func (w *Worker) alert(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	tx, err := w.Store.Pool().Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(context.Background())
	var ns, id string
	var data []byte
	err = tx.QueryRow(ctx, `SELECT namespace,id,data FROM datacat.resources WHERE kind='alert' AND (data->>'nextExecutionTime')::timestamptz<=now() ORDER BY data->>'nextExecutionTime',id FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(&ns, &id, &data)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var alert entity.Alert
	if err = json.Unmarshal(data, &alert); err != nil {
		return false, err
	}
	d, queryErr := w.Service.DataSource(ctx, ns, alert.DataSource.ID)
	firing := false
	if queryErr == nil {
		query, err := w.Service.ExpandQuery(ctx, ns, "", alert.RawQuery)
		queryErr = err
		if queryErr == nil {
			var points []entity.MetricPoint
			if strings.EqualFold(d.Type, "Prometheus") {
				points, queryErr = w.Telemetry.Instant(ctx, d, query)
			} else {
				points, queryErr = w.customMetrics(ctx, d, query)
			}
			for _, p := range points {
				if p.Value > 0 {
					firing = true
				}
			}
		}
	}
	if queryErr != nil {
		w.Log.WarnContext(ctx, "alert query failed", "alert_id", id, "namespace", ns)
	}
	if usecases.EvaluateAlert(&alert, w.Service.Now(), firing, queryErr) {
		channels, err := usecases.List[entity.NotificationChannel](ctx, w.Store, ns, "notification-channel")
		if err != nil {
			return false, err
		}
		for _, ch := range channels {
			if ch.GroupID == alert.NotificationChannelGroup.ID {
				if _, err = tx.Exec(ctx, `INSERT INTO datacat.outbox(namespace,channel_id,message) VALUES($1,$2,$3)`, ns, fmt.Sprint(ch.ID), usecases.RenderAlert(alert)); err != nil {
					return false, err
				}
			}
		}
	}
	data, err = json.Marshal(alert)
	if err != nil {
		return false, err
	}
	if _, err = tx.Exec(ctx, `UPDATE datacat.resources SET data=$3,version=version+1,updated_at=now() WHERE namespace=$1 AND kind='alert' AND id=$2`, ns, id, data); err != nil {
		return false, err
	}
	return true, tx.Commit(ctx)
}
func (w *Worker) customMetrics(ctx context.Context, d entity.DataSource, query string) ([]entity.MetricPoint, error) {
	p, _, err := usecases.Read[entity.Plugin](ctx, w.Store, usecases.Global, "plugin", d.Type)
	if err != nil {
		return nil, err
	}
	if !p.IsActive {
		return nil, fmt.Errorf("plugin disabled")
	}
	var points []entity.MetricPoint
	err = w.Telemetry.Client.Do(ctx, "POST", p.Endpoint, map[string]any{"version": 1, "operation": "metrics.query", "namespaceId": d.NamespaceID, "connectionSettings": json.RawMessage(d.ConnectionString), "settings": p.Settings, "input": map[string]string{"query": query}}, nil, &points)
	return points, err
}
func (w *Worker) deliver(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	tx, err := w.Store.Pool().Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(context.Background())
	var id int64
	var ns, channel, message string
	var attempts int
	err = tx.QueryRow(ctx, `SELECT id,namespace,channel_id,message,attempts FROM datacat.outbox WHERE delivered_at IS NULL AND attempts<8 AND available_at<=now() ORDER BY id FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(&id, &ns, &channel, &message, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	ch, sendErr := w.Service.Channel(ctx, ns, channel)
	if sendErr == nil {
		sendErr = w.Sender.Send(ctx, ch, message)
	}
	if sendErr == nil {
		_, err = tx.Exec(ctx, `UPDATE datacat.outbox SET delivered_at=now(),last_error=NULL WHERE id=$1`, id)
	} else {
		w.Log.WarnContext(ctx, "notification delivery failed", "outbox_id", id, "attempt", attempts+1)
		delay := time.Duration(1<<uint(min(attempts, 7))) * time.Second
		_, err = tx.Exec(ctx, `UPDATE datacat.outbox SET attempts=attempts+1,available_at=$2,last_error='delivery failed' WHERE id=$1`, id, time.Now().UTC().Add(delay))
	}
	if err != nil {
		return false, err
	}
	return true, tx.Commit(ctx)
}
