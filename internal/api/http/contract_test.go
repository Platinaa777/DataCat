package httpapi_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Platinaa777/DataCat/internal/api/auth"
	"github.com/Platinaa777/DataCat/internal/bootstrap"
	"github.com/Platinaa777/DataCat/internal/config"
	"github.com/go-chi/chi/v5"
)

func TestOpenAPIRoutesMatchRouter(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "api", "openapi.json"))
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err = json.Unmarshal(b, &document); err != nil {
		t.Fatal(err)
	}
	parameter := regexp.MustCompile(`\{[^}]+\}`)
	canonical := func(path string) string { return parameter.ReplaceAllString(path, "{}") }
	s := bootstrap.NewRouter(nil, &auth.Auth{}, nil, config.Config{}, slog.Default())
	registered := map[string]bool{}
	if err = chi.Walk(s.Handler().(chi.Routes), func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		registered[method+" "+canonical(route)] = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for path, methods := range document.Paths {
		for method := range methods {
			key := strings.ToUpper(method) + " " + canonical(path)
			if !registered[key] {
				t.Errorf("OpenAPI operation has no route: %s", key)
			}
			delete(registered, key)
		}
	}
	for route := range registered {
		if strings.Contains(route, " /api/v1/") {
			t.Errorf("route has no OpenAPI operation: %s", route)
		}
	}
}
