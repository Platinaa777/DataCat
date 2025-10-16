package usecases

import "testing"

func TestQueryVariablesPreserveLegacySyntax(t *testing.T) {
	values := map[string]string{"service": "api", "region": "eu", "nested": "{.region}"}
	for _, test := range []struct{ query, want string }{
		{`up{service="{.service}"}`, `up{service="api"}`},
		{`{ .service }/${region}/{{service}}`, `api/eu/api`},
		{`{.missing}`, `{.missing}`},
		{`{.nested}`, `{.region}`},
	} {
		if got := replaceQueryVariables(test.query, values); got != test.want {
			t.Errorf("%q: got %q, want %q", test.query, got, test.want)
		}
	}
}
