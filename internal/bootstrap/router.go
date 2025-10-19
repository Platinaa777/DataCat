package bootstrap

import (
	"log/slog"

	"github.com/Platinaa777/DataCat/internal/api/auth"
	httpapi "github.com/Platinaa777/DataCat/internal/api/http"
	"github.com/Platinaa777/DataCat/internal/api/http/alert"
	"github.com/Platinaa777/DataCat/internal/api/http/catalog"
	"github.com/Platinaa777/DataCat/internal/api/http/dashboard"
	"github.com/Platinaa777/DataCat/internal/api/http/datasource"
	"github.com/Platinaa777/DataCat/internal/api/http/namespace"
	"github.com/Platinaa777/DataCat/internal/api/http/notification"
	"github.com/Platinaa777/DataCat/internal/api/http/panel"
	"github.com/Platinaa777/DataCat/internal/api/http/plugin"
	"github.com/Platinaa777/DataCat/internal/api/http/query"
	"github.com/Platinaa777/DataCat/internal/api/http/secret"
	"github.com/Platinaa777/DataCat/internal/api/http/transport"
	"github.com/Platinaa777/DataCat/internal/api/http/user"
	"github.com/Platinaa777/DataCat/internal/api/http/variable"
	"github.com/Platinaa777/DataCat/internal/config"
	"github.com/Platinaa777/DataCat/internal/extapi/telemetry"
	"github.com/Platinaa777/DataCat/internal/pkg/cache"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

func NewRouter(service *usecases.Service, identity *auth.Auth, provider *telemetry.Provider, cfg config.Config, log *slog.Logger, resultCache cache.Cache) *httpapi.Server {
	access := &transport.Access{Service: service}
	responses := &transport.Responder{Log: log}
	return &httpapi.Server{Auth: identity, Config: cfg, Access: access, Responses: responses, Features: []httpapi.Feature{
		user.New(responses),
		namespace.New(service, responses),
		dashboard.New(service, access, responses),
		panel.New(service, access, responses),
		variable.New(service, access, responses),
		datasource.New(service, access, responses, provider),
		notification.New(service, access, responses, provider),
		alert.New(service, access, responses),
		plugin.New(service, access, responses, provider),
		query.New(service, access, responses, provider, resultCache),
		secret.New(service, access, responses),
		catalog.New(service, responses, "data-source-type"),
		catalog.New(service, responses, "notification-destination"),
	}}
}
