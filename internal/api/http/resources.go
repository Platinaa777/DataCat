package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Platinaa777/DataCat/internal/api/auth"
	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/usecases"
	"github.com/go-chi/chi/v5"
)

func (s *Server) resources(r chi.Router) {
	s.route(r, "GET", "/user/get-me", func(r *http.Request) (any, error) {
		p := auth.Principal(r.Context())
		return map[string]any{"identityId": p.Subject, "claims": p.Claims, "isGlobalAdmin": p.GlobalAdmin}, nil
	})
	s.namespaces(r)
	s.dashboards(r)
	s.panels(r)
	s.variables(r)
	s.dataSources(r)
	s.notifications(r)
	s.alerts(r)
}
func (s *Server) namespaces(r chi.Router) {
	s.route(r, "POST", "/namespace", func(r *http.Request) (any, error) {
		if err := admin(r); err != nil {
			return nil, err
		}
		input, err := body[struct{ Name string }](r)
		if err != nil {
			return nil, err
		}
		return s.Service.CreateNamespace(r.Context(), input.Name)
	})
	s.route(r, "GET", "/namespace/get-available", func(r *http.Request) (any, error) {
		all, err := usecases.List[entity.Namespace](r.Context(), s.Service.Store, usecases.Global, "namespace")
		if err != nil {
			return nil, err
		}
		out := []entity.Namespace{}
		for _, ns := range all {
			if auth.Principal(r.Context()).Can(ns.ID, false) {
				out = append(out, ns)
			}
		}
		return out, nil
	})
	s.route(r, "GET", "/namespace/{name}", func(r *http.Request) (any, error) {
		ns, _, err := usecases.Read[entity.Namespace](r.Context(), s.Service.Store, usecases.Global, "namespace", param(r, "name"))
		if err != nil {
			return nil, err
		}
		if !auth.Principal(r.Context()).Can(ns.ID, false) {
			return nil, entity.ErrForbidden
		}
		dashboards, err := usecases.List[entity.Dashboard](r.Context(), s.Service.Store, ns.ID, "dashboard")
		return map[string]any{"name": ns.Name, "dashboards": dashboards}, err
	})
	s.route(r, "POST", "/external-role", func(r *http.Request) (any, error) {
		if err := admin(r); err != nil {
			return nil, err
		}
		input, err := body[entity.RoleMapping](r)
		if err != nil {
			return nil, err
		}
		if input.RoleID < 1 || input.RoleID > 3 {
			return nil, entity.ErrInvalid
		}
		if err = usecases.Required(input.ExternalRole, input.NamespaceID); err != nil {
			return nil, err
		}
		if _, _, err = usecases.Read[entity.Namespace](r.Context(), s.Service.Store, usecases.Global, "namespace", input.NamespaceID); err != nil {
			return nil, err
		}
		return nil, usecases.Create(r.Context(), s.Service.Store, usecases.Global, "role-mapping", usecases.ID(), input.ExternalRole+"/"+input.NamespaceID, input)
	})
}
func (s *Server) dashboards(r chi.Router) {
	s.route(r, "POST", "/dashboard/add", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, true)
		if err != nil {
			return nil, err
		}
		input, err := body[struct {
			Name, Description, NamespaceID string
			Tags                           []string
		}](r)
		if err != nil {
			return nil, err
		}
		if input.NamespaceID != "" && input.NamespaceID != ns {
			return nil, entity.ErrForbidden
		}
		return s.Service.CreateDashboard(r.Context(), entity.Dashboard{Name: input.Name, Description: input.Description, NamespaceID: ns, Tags: input.Tags})
	})
	s.route(r, "GET", "/dashboard/{id}", func(r *http.Request) (any, error) { return get[entity.Dashboard](s, r, "dashboard") })
	s.route(r, "GET", "/dashboard/{id}/full", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, false)
		if err != nil {
			return nil, err
		}
		d, _, err := usecases.Read[entity.Dashboard](r.Context(), s.Service.Store, ns, "dashboard", param(r, "id"))
		if err != nil {
			return nil, err
		}
		panels, err := usecases.List[entity.Panel](r.Context(), s.Service.Store, ns, "panel")
		out := []map[string]any{}
		for _, p := range panels {
			if p.DashboardID == d.ID {
				out = append(out, map[string]any{"id": p.ID, "title": p.Title, "panelType": p.TypeName, "query": p.Query.Query, "layout": p.Layout, "styleConfiguration": p.StyleConfiguration})
			}
		}
		return struct {
			entity.Dashboard
			Panels []map[string]any `json:"panels"`
		}{d, out}, err
	})
	s.route(r, "POST", "/dashboard/search", func(r *http.Request) (any, error) { return search[entity.Dashboard](s, r, "dashboard") })
	s.route(r, "PUT", "/dashboard/update/{id}", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, true)
		if err != nil {
			return nil, err
		}
		input, err := body[struct{ Name, Description string }](r)
		if err != nil {
			return nil, err
		}
		if err = usecases.Required(input.Name); err != nil {
			return nil, err
		}
		d, record, err := usecases.Read[entity.Dashboard](r.Context(), s.Service.Store, ns, "dashboard", param(r, "id"))
		if err != nil {
			return nil, err
		}
		d.Name = input.Name
		d.Description = input.Description
		d.UpdatedAt = s.Service.Now()
		record.Name = d.Name
		return nil, usecases.Update(r.Context(), s.Service.Store, record, d)
	})
	s.route(r, "DELETE", "/dashboard/remove/{id}", func(r *http.Request) (any, error) { return s.remove(r, "dashboard") })
}

type panelInput struct {
	Title                                                           string
	Type                                                            int
	RawQuery, DataSourceID, Layout, DashboardID, StyleConfiguration string
}

func (s *Server) panels(r chi.Router) {
	s.route(r, "GET", "/panel/types", func(r *http.Request) (any, error) {
		return []entity.NamedType{{ID: 1, Name: "Graph"}, {ID: 2, Name: "Table"}, {ID: 3, Name: "Pie Chart"}}, nil
	})
	s.route(r, "POST", "/panel/add", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, true)
		if err != nil {
			return nil, err
		}
		input, err := body[panelInput](r)
		if err != nil {
			return nil, err
		}
		return s.Service.CreatePanel(r.Context(), entity.Panel{NamespaceID: ns, Title: input.Title, TypeName: panelName(input.Type), Query: entity.RawQuery{Query: input.RawQuery}, Layout: input.Layout, DashboardID: input.DashboardID, StyleConfiguration: input.StyleConfiguration}, input.DataSourceID)
	})
	s.route(r, "GET", "/panel/{id}", func(r *http.Request) (any, error) { return get[entity.Panel](s, r, "panel") })
	s.route(r, "DELETE", "/panel/remove/{id}", func(r *http.Request) (any, error) { return s.remove(r, "panel") })
	s.route(r, "PUT", "/panel/update/{id}", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, true)
		if err != nil {
			return nil, err
		}
		input, err := body[panelInput](r)
		if err != nil {
			return nil, err
		}
		if err = usecases.Required(input.Title, input.RawQuery, input.Layout, panelName(input.Type)); err != nil {
			return nil, err
		}
		return nil, s.Service.UpdatePanel(r.Context(), ns, param(r, "id"), input.Title, panelName(input.Type), input.RawQuery, input.DataSourceID, input.Layout, input.StyleConfiguration)
	})
}
func (s *Server) variables(r chi.Router) {
	s.route(r, "POST", "/variable/add", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, true)
		if err != nil {
			return nil, err
		}
		input, err := body[struct {
			Placeholder, Value, NamespaceID string
			DashboardID                     *string
		}](r)
		if err != nil {
			return nil, err
		}
		if input.NamespaceID != "" && input.NamespaceID != ns {
			return nil, entity.ErrForbidden
		}
		return s.Service.CreateVariable(r.Context(), entity.Variable{Placeholder: input.Placeholder, Value: input.Value, NamespaceID: ns, DashboardID: input.DashboardID})
	})
	s.route(r, "GET", "/variable/{id}", func(r *http.Request) (any, error) { return get[entity.Variable](s, r, "variable") })
	for _, field := range []string{"dashboard", "namespace"} {
		field := field
		s.route(r, "GET", "/variables/"+field+"/{id}", func(r *http.Request) (any, error) {
			ns, err := s.scope(r, false)
			if err != nil {
				return nil, err
			}
			if field == "namespace" && param(r, "id") != ns {
				return nil, entity.ErrForbidden
			}
			all, err := usecases.List[entity.Variable](r.Context(), s.Service.Store, ns, "variable")
			out := []entity.Variable{}
			for _, v := range all {
				if field == "namespace" || (v.DashboardID != nil && *v.DashboardID == param(r, "id")) {
					out = append(out, v)
				}
			}
			return out, err
		})
	}
	s.route(r, "DELETE", "/variable/remove/{id}", func(r *http.Request) (any, error) { return s.remove(r, "variable") })
	s.route(r, "PUT", "/variable/update/{id}", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, true)
		if err != nil {
			return nil, err
		}
		input, err := body[struct{ Placeholder, Value string }](r)
		if err != nil {
			return nil, err
		}
		if err = usecases.Required(input.Placeholder, input.Value); err != nil {
			return nil, err
		}
		v, rec, err := usecases.Read[entity.Variable](r.Context(), s.Service.Store, ns, "variable", param(r, "id"))
		if err != nil {
			return nil, err
		}
		v.Placeholder = input.Placeholder
		v.Value = input.Value
		rec.Name = v.Placeholder
		if v.DashboardID != nil {
			rec.Name = *v.DashboardID + "/" + v.Placeholder
		}
		return nil, usecases.Update(r.Context(), s.Service.Store, rec, v)
	})
}
func (s *Server) namedTypes(r chi.Router, kind string) {
	s.route(r, "POST", "/"+kind+"/add", func(r *http.Request) (any, error) {
		if err := admin(r); err != nil {
			return nil, err
		}
		input, err := body[struct{ Name string }](r)
		if err != nil {
			return nil, err
		}
		if err = usecases.Required(input.Name); err != nil {
			return nil, err
		}
		id, err := s.Service.Store.NextID(r.Context())
		if err != nil {
			return nil, err
		}
		return id, usecases.Create(r.Context(), s.Service.Store, usecases.Global, kind, strconv.FormatInt(id, 10), input.Name, entity.NamedType{ID: id, Name: input.Name})
	})
	s.route(r, "GET", "/"+kind+"/get-all", func(r *http.Request) (any, error) {
		return usecases.List[entity.NamedType](r.Context(), s.Service.Store, usecases.Global, kind)
	})
	s.route(r, "GET", "/"+kind+"/{name}", func(r *http.Request) (any, error) {
		v, _, err := usecases.Read[entity.NamedType](r.Context(), s.Service.Store, usecases.Global, kind, param(r, "name"))
		return v, err
	})
	path := "/" + kind + "/remove/{name}"
	if kind == "data-source-type" {
		path = "/datasource/remove/{name}"
	}
	s.route(r, "DELETE", path, func(r *http.Request) (any, error) {
		if err := admin(r); err != nil {
			return nil, err
		}
		_, rec, err := usecases.Read[entity.NamedType](r.Context(), s.Service.Store, usecases.Global, kind, param(r, "name"))
		if err != nil {
			return nil, err
		}
		return nil, s.Service.Delete(r.Context(), usecases.Global, kind, rec.ID)
	})
}
func (s *Server) dataSources(r chi.Router) {
	s.namedTypes(r, "data-source-type")
	s.route(r, "POST", "/data-source/add", func(r *http.Request) (any, error) {
		ns, err := s.namespaceAdmin(r)
		if err != nil {
			return nil, err
		}
		input, err := body[struct {
			UniqueName, DataSourceType, ConnectionString string
			Purpose                                      json.RawMessage
		}](r)
		if err != nil {
			return nil, err
		}
		if err = s.validateConnection(input.ConnectionString); err != nil {
			return nil, err
		}
		var purposeName string
		if json.Unmarshal(input.Purpose, &purposeName) != nil {
			var purposeNumber int
			if json.Unmarshal(input.Purpose, &purposeNumber) != nil {
				return nil, entity.ErrInvalid
			}
			purposeName = map[int]string{1: "Metrics", 2: "Logs", 3: "Traces"}[purposeNumber]
		}
		purpose := map[string]string{"metrics": "Metrics", "logs": "Logs", "traces": "Traces"}[strings.ToLower(purposeName)]
		return s.Service.CreateDataSource(r.Context(), ns, entity.DataSource{Name: input.UniqueName, Type: input.DataSourceType, ConnectionString: input.ConnectionString, Purpose: purpose})
	})
	s.route(r, "GET", "/data-source/{id}", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, false)
		if err != nil {
			return nil, err
		}
		d, err := s.Service.DataSource(r.Context(), ns, param(r, "id"))
		if err != nil {
			return nil, err
		}
		if _, err = s.namespaceAdmin(r); err != nil {
			d.ConnectionString = ""
		}
		return d, nil
	})
	s.route(r, "POST", "/data-source/search", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, false)
		if err != nil {
			return nil, err
		}
		input, err := body[entity.Search](r)
		if err != nil {
			return nil, err
		}
		page, size, err := pagination(r)
		if err != nil {
			return nil, err
		}
		result, err := resourcePage[entity.DataSource](s, r, ns, "data-source", input, page, size)
		if err != nil {
			return nil, err
		}
		for i := range result.Items {
			result.Items[i].ConnectionString = ""
		}
		return result, nil
	})
	s.route(r, "PUT", "/data-source/update-connection-string/{id}", func(r *http.Request) (any, error) {
		ns, err := s.namespaceAdmin(r)
		if err != nil {
			return nil, err
		}
		input, err := body[struct{ ConnectionString string }](r)
		if err != nil {
			return nil, err
		}
		if err = s.validateConnection(input.ConnectionString); err != nil {
			return nil, err
		}
		d, rec, err := usecases.Read[entity.DataSource](r.Context(), s.Service.Store, ns, "data-source", param(r, "id"))
		if err != nil {
			return nil, err
		}
		d.ConnectionString, err = s.Service.Cipher.Seal(ns+"/data-source/"+d.ID, input.ConnectionString)
		if err != nil {
			return nil, err
		}
		return nil, usecases.Update(r.Context(), s.Service.Store, rec, d)
	})
	s.route(r, "DELETE", "/data-source/remove/{id}", func(r *http.Request) (any, error) {
		if _, err := s.namespaceAdmin(r); err != nil {
			return nil, err
		}
		return s.remove(r, "data-source")
	})
}
func (s *Server) notifications(r chi.Router) {
	s.namedTypes(r, "notification-destination")
	s.route(r, "POST", "/notification-channel-group/add", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, true)
		if err != nil {
			return nil, err
		}
		input, err := body[struct{ Name string }](r)
		if err != nil {
			return nil, err
		}
		if err = usecases.Required(input.Name); err != nil {
			return nil, err
		}
		id := usecases.ID()
		return id, usecases.Create(r.Context(), s.Service.Store, ns, "notification-channel-group", id, input.Name, entity.NotificationGroup{ID: id, Name: input.Name, NamespaceID: ns, NotificationChannels: []entity.NotificationChannel{}})
	})
	s.route(r, "GET", "/notification-channel-group/get-all", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, false)
		if err != nil {
			return nil, err
		}
		groups, err := s.Service.Groups(r.Context(), ns)
		if _, permissionErr := s.namespaceAdmin(r); permissionErr != nil {
			for i := range groups {
				for j := range groups[i].NotificationChannels {
					groups[i].NotificationChannels[j].Settings = "{}"
				}
			}
		}
		return groups, err
	})
	s.route(r, "GET", "/notification-channel-group/{id}", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, false)
		if err != nil {
			return nil, err
		}
		group, err := s.Service.Group(r.Context(), ns, param(r, "id"))
		if _, permissionErr := s.namespaceAdmin(r); permissionErr != nil {
			for i := range group.NotificationChannels {
				group.NotificationChannels[i].Settings = "{}"
			}
		}
		return group, err
	})
	s.route(r, "POST", "/notification-channel/add", func(r *http.Request) (any, error) {
		ns, err := s.namespaceAdmin(r)
		if err != nil {
			return nil, err
		}
		input, err := body[struct{ NotificationChannelGroupName, DestinationName, Settings string }](r)
		if err != nil {
			return nil, err
		}
		if err = s.validateNotification(input.DestinationName, input.Settings); err != nil {
			return nil, err
		}
		return s.Service.CreateChannel(r.Context(), ns, input.NotificationChannelGroupName, input.DestinationName, input.Settings)
	})
	s.route(r, "GET", "/notification-channel/{id}", func(r *http.Request) (any, error) {
		if _, err := s.namespaceAdmin(r); err != nil {
			return nil, err
		}
		ns, err := s.scope(r, false)
		if err != nil {
			return nil, err
		}
		return s.Service.Channel(r.Context(), ns, param(r, "id"))
	})
	s.route(r, "PUT", "/notification-channel/{id}", func(r *http.Request) (any, error) {
		ns, err := s.namespaceAdmin(r)
		if err != nil {
			return nil, err
		}
		input, err := body[struct{ DestinationName, Settings string }](r)
		if err != nil {
			return nil, err
		}
		if err = s.validateNotification(input.DestinationName, input.Settings); err != nil {
			return nil, err
		}
		ch, rec, err := usecases.Read[entity.NotificationChannel](r.Context(), s.Service.Store, ns, "notification-channel", param(r, "id"))
		if err != nil {
			return nil, err
		}
		if _, _, err = usecases.Read[entity.NamedType](r.Context(), s.Service.Store, usecases.Global, "notification-destination", input.DestinationName); err != nil {
			return nil, err
		}
		ch.DestinationName = input.DestinationName
		ch.Settings, err = s.Service.Cipher.Seal(ns+"/notification-channel/"+param(r, "id"), input.Settings)
		if err != nil {
			return nil, err
		}
		return nil, usecases.Update(r.Context(), s.Service.Store, rec, ch)
	})
	s.route(r, "DELETE", "/notification-channel/remove/{id}", func(r *http.Request) (any, error) {
		if _, err := s.namespaceAdmin(r); err != nil {
			return nil, err
		}
		return s.remove(r, "notification-channel")
	})
}
