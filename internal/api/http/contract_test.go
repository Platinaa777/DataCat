package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/Platinaa777/DataCat/internal/api/auth"
	"github.com/go-chi/chi/v5"
)

func TestLegacyRoutesRemainRegistered(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "legacy-api.json"))
	if err != nil {
		t.Fatal(err)
	}
	var routes []struct{ Method, Path, Status string }
	if err = json.Unmarshal(b, &routes); err != nil {
		t.Fatal(err)
	}
	parameter := regexp.MustCompile(`\{[^}]+\}`)
	canonical := func(path string) string { return parameter.ReplaceAllString(path, "{}") }
	s := &Server{Auth: &auth.Auth{}}
	registered := map[string]bool{}
	if err = chi.Walk(s.Handler().(chi.Routes), func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		registered[method+" "+canonical(route)] = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for _, route := range routes {
		if route.Status == "inactive-in-legacy" {
			continue
		}
		if !registered[route.Method+" "+canonical(route.Path)] {
			t.Errorf("missing legacy route: %s %s", route.Method, route.Path)
		}
	}
}
