// Package legacy imports the final .NET database schema without changing it.
package legacy

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/pkg/secrets"
	"github.com/Platinaa777/DataCat/internal/repository/postgres"
	"github.com/Platinaa777/DataCat/internal/usecases"
	"github.com/Platinaa777/DataCat/internal/usecases/contract"
	"github.com/jackc/pgx/v5"
)

const defaultNamespace = "00000000-0000-0000-0000-000000000000"

type row map[string]json.RawMessage

func (r row) text(key string) string {
	var s string
	if json.Unmarshal(r[key], &s) == nil {
		return s
	}
	if string(r[key]) == "null" {
		return ""
	}
	return string(r[key])
}
func (r row) number(key string) int64 { n, _ := strconv.ParseInt(r.text(key), 10, 64); return n }
func (r row) timestamp(key string) time.Time {
	s := r.text(key)
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04:05.999999", "2006-01-02T15:04:05"} {
		if t, e := time.Parse(layout, s); e == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}
func (r row) namespace() string {
	if ns := r.text("namespace_id"); ns != "" {
		return ns
	}
	return defaultNamespace
}
func (r row) tags() []string {
	out := []string{}
	json.Unmarshal(r["tags"], &out)
	return usecases.Tags(out)
}

func Import(ctx context.Context, sourceDSN string, target *postgres.Store, cipher *secrets.Cipher) (map[string]int, error) {
	source, err := pgx.Connect(ctx, sourceDSN)
	if err != nil {
		return nil, fmt.Errorf("open legacy database: %w", err)
	}
	defer source.Close(context.Background())
	snapshot, err := source.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer snapshot.Rollback(context.Background())
	tables := []string{"datacat_namespaces", "data_source_type", "notification_destination", "data_sources", "dashboards", "panels", "notification_channel_group", "notification_channel", "alerts", "variables", "plugins", "external_role_mappings", "datacat_users", "user_role_link"}
	data := map[string][]row{}
	for _, table := range tables {
		rows, err := snapshot.Query(ctx, "SELECT to_jsonb(t) FROM public."+pgx.Identifier{table}.Sanitize()+" AS t")
		if err != nil {
			return nil, fmt.Errorf("read legacy table %s: %w", table, err)
		}
		for rows.Next() {
			var raw []byte
			if err = rows.Scan(&raw); err != nil {
				rows.Close()
				return nil, err
			}
			var r row
			if err = json.Unmarshal(raw, &r); err != nil {
				rows.Close()
				return nil, err
			}
			data[table] = append(data[table], r)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
	}
	counts := map[string]int{}
	err = target.Within(ctx, func(store contract.Store) error {
		var n int
		if err := target.Pool().QueryRow(ctx, "SELECT count(*) FROM datacat.resources").Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			return fmt.Errorf("target schema must be empty; import into a fresh database")
		}
		write := func(ns, kind, id, name string, value any) error {
			if err := usecases.Create(ctx, store, ns, kind, id, name, value); err != nil {
				return fmt.Errorf("import %s %s: %w", kind, id, err)
			}
			counts[kind]++
			return nil
		}
		namespaces := []string{}
		for _, r := range data["datacat_namespaces"] {
			ns := r.text("id")
			namespaces = append(namespaces, ns)
			if err := write(usecases.Global, "namespace", ns, r.text("name"), entity.Namespace{ID: ns, Name: r.text("name")}); err != nil {
				return err
			}
		}
		foundDefault := false
		for _, ns := range namespaces {
			foundDefault = foundDefault || ns == defaultNamespace
		}
		if !foundDefault {
			namespaces = append(namespaces, defaultNamespace)
			if err := write(usecases.Global, "namespace", defaultNamespace, "datacat_system", entity.Namespace{ID: defaultNamespace, Name: "datacat_system"}); err != nil {
				return err
			}
		}
		types := map[string]string{}
		destinations := map[string]string{}
		for table, kind := range map[string]string{"data_source_type": "data-source-type", "notification_destination": "notification-destination"} {
			for _, r := range data[table] {
				id, name := r.text("id"), strings.ToLower(r.text("name"))
				if kind == "data-source-type" {
					types[id] = name
				} else {
					destinations[id] = name
				}
				if err := write(usecases.Global, kind, id, name, entity.NamedType{ID: r.number("id"), Name: name}); err != nil {
					return err
				}
			}
		}
		sources := map[string]entity.DataSource{}
		for _, r := range data["data_sources"] {
			id := r.text("id")
			purpose := map[int64]string{1: "Metrics", 2: "Logs", 3: "Traces"}[r.number("purpose")]
			if purpose == "" {
				purpose = r.text("purpose")
			}
			d := entity.DataSource{ID: id, Name: r.text("name"), Type: types[r.text("type_id")], TypeID: r.number("type_id"), Purpose: purpose}
			sources[id] = d
			for _, ns := range namespaces {
				copy := d
				copy.NamespaceID = ns
				encrypted, err := cipher.Seal(ns+"/data-source/"+id, r.text("connection_string"))
				if err != nil {
					return err
				}
				copy.ConnectionString = encrypted
				if err = write(ns, "data-source", id, d.Name, copy); err != nil {
					return err
				}
			}
		}
		for _, r := range data["dashboards"] {
			d := entity.Dashboard{ID: r.text("id"), NamespaceID: r.namespace(), Name: r.text("name"), Description: r.text("description"), Tags: r.tags(), CreatedAt: r.timestamp("created_at_utc"), UpdatedAt: r.timestamp("updated_at_utc")}
			if err := write(d.NamespaceID, "dashboard", d.ID, d.Name, d); err != nil {
				return err
			}
		}
		for _, r := range data["panels"] {
			d, ok := sources[r.text("data_source_id")]
			if !ok {
				return fmt.Errorf("panel references unknown data source")
			}
			d.NamespaceID = r.namespace()
			p := entity.Panel{ID: r.text("id"), NamespaceID: r.namespace(), Title: r.text("title"), TypeName: map[int64]string{1: "Graph", 2: "Table", 3: "Pie Chart"}[r.number("type_id")], Query: entity.RawQuery{Query: r.text("raw_query"), DataSource: d}, Layout: r.text("layout_configuration"), DashboardID: r.text("dashboard_id"), StyleConfiguration: r.text("style_configuration")}
			if err := write(p.NamespaceID, "panel", p.ID, "", p); err != nil {
				return err
			}
		}
		groups := map[string]entity.NotificationGroup{}
		for _, r := range data["notification_channel_group"] {
			g := entity.NotificationGroup{ID: r.text("id"), NamespaceID: r.namespace(), Name: r.text("name"), NotificationChannels: []entity.NotificationChannel{}}
			groups[g.ID] = g
			if err := write(g.NamespaceID, "notification-channel-group", g.ID, g.Name, g); err != nil {
				return err
			}
		}
		for _, r := range data["notification_channel"] {
			ch := entity.NotificationChannel{ID: r.number("id"), NamespaceID: r.namespace(), GroupID: r.text("notification_channel_group_id"), DestinationName: destinations[r.text("destination_id")], Settings: r.text("settings")}
			encrypted, err := cipher.Seal(ch.NamespaceID+"/notification-channel/"+r.text("id"), ch.Settings)
			if err != nil {
				return err
			}
			ch.Settings = encrypted
			if err := write(ch.NamespaceID, "notification-channel", r.text("id"), "", ch); err != nil {
				return err
			}
		}
		for _, r := range data["alerts"] {
			d, ok := sources[r.text("data_source_id")]
			if !ok {
				return fmt.Errorf("alert references unknown data source")
			}
			d.NamespaceID = r.namespace()
			g, ok := groups[r.text("notification_channel_group_id")]
			if !ok || g.NamespaceID != r.namespace() {
				return fmt.Errorf("alert references an invalid notification group")
			}
			status := r.text("status")
			if n := r.number("status"); n > 0 {
				status = map[int64]string{1: "Firing", 2: "Muted", 3: "Ok", 4: "Error"}[n]
			}
			a := entity.Alert{ID: r.text("id"), NamespaceID: r.namespace(), Description: r.text("description"), Template: r.text("template"), RawQuery: r.text("condition_query"), Status: status, DataSource: d, NotificationChannelGroup: g, PreviousExecutionTime: r.timestamp("previous_execution"), NextExecutionTime: r.timestamp("next_execution"), WaitTimeBeforeAlerting: entity.Duration(r.number("wait_time_before_alerting_in_ticks") * 100), RepeatInterval: entity.Duration(r.number("repeat_interval_in_ticks") * 100), Tags: r.tags()}
			if err := write(a.NamespaceID, "alert", a.ID, "", a); err != nil {
				return err
			}
		}
		for _, r := range data["variables"] {
			v := entity.Variable{ID: r.text("id"), NamespaceID: r.namespace(), Placeholder: r.text("placeholder"), Value: r.text("value")}
			name := v.Placeholder
			if id := r.text("dashboard_id"); id != "" {
				v.DashboardID = &id
				name = id + "/" + name
			}
			if err := write(v.NamespaceID, "variable", v.ID, name, v); err != nil {
				return err
			}
		}
		for _, r := range data["plugins"] {
			p := entity.Plugin{ID: r.text("id"), Name: r.text("name"), Version: r.text("version"), Author: r.text("author"), Description: r.text("description"), Settings: r.text("settings"), IsActive: false}
			if err := write(usecases.Global, "plugin", p.ID, p.Name, p); err != nil {
				return err
			}
		}
		for _, r := range data["external_role_mappings"] {
			m := entity.RoleMapping{ExternalRole: r.text("external_role"), NamespaceID: r.namespace(), RoleID: int(r.number("role_id"))}
			if err := write(usecases.Global, "role-mapping", usecases.ID(), m.ExternalRole+"/"+m.NamespaceID, m); err != nil {
				return err
			}
		}
		identities := map[string]string{}
		for _, r := range data["datacat_users"] {
			identities[r.text("id")] = r.text("identity_id")
		}
		claims := map[string][]entity.Claim{}
		for _, r := range data["user_role_link"] {
			if string(r["is_manual"]) != "true" {
				continue
			}
			identity := identities[r.text("user_id")]
			if identity == "" {
				return fmt.Errorf("manual role references unknown user")
			}
			role := map[int64]string{1: "Admin", 2: "Viewer", 3: "Editor"}[r.number("role_id")]
			if role == "" {
				return fmt.Errorf("unknown legacy role")
			}
			claims[identity] = append(claims[identity], entity.Claim{Role: role, NamespaceID: r.namespace()})
		}
		for identity, roles := range claims {
			if err := write(usecases.Global, "user-claims", identity, "", roles); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Imported integer IDs must not collide with subsequently created resources.
	_, err = target.Pool().Exec(ctx, `SELECT setval('datacat.resource_id',GREATEST(1,(SELECT COALESCE(max(id::bigint),0) FROM datacat.resources WHERE kind IN ('notification-channel','notification-destination','data-source-type'))))`)
	return counts, err
}
