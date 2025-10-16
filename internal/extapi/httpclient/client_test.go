package httpclient

import "testing"

func TestOutboundAllowlist(t *testing.T) {
	c := New([]string{"prometheus:9090"})
	for _, raw := range []string{"http://169.254.169.254/latest/meta-data", "file:///etc/passwd", "http://user:pass@prometheus:9090", "http://prometheus:9091", "http://prometheus:9090.evil.test"} {
		if _, err := c.Validate(raw); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	if _, err := c.Validate("http://prometheus:9090/api/v1/query"); err != nil {
		t.Fatal(err)
	}
}
