package entity

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDurationWireFormat(t *testing.T) {
	for _, tc := range []struct {
		wire string
		want time.Duration
	}{{"00:00:00", 0}, {"01:02:03.1234567", time.Hour + 2*time.Minute + 3*time.Second + 123456700}, {"2.03:04:05", 51*time.Hour + 4*time.Minute + 5*time.Second}, {"30s", 30 * time.Second}} {
		d, err := ParseDuration(tc.wire)
		if err != nil || time.Duration(d) != tc.want {
			t.Fatalf("%s: %v %v", tc.wire, d, err)
		}
		b, err := json.Marshal(d)
		if err != nil {
			t.Fatal(err)
		}
		var roundtrip Duration
		if err = json.Unmarshal(b, &roundtrip); err != nil || roundtrip != d {
			t.Fatalf("roundtrip: %s %v", b, err)
		}
	}
	for _, bad := range []string{"-1s", "00:70:00", "00:00:NaN", "00:00:Inf", "1.2.3", "garbage", "999999999999:00:00"} {
		if _, err := ParseDuration(bad); err == nil {
			t.Errorf("accepted %q", bad)
		}
	}
}
func TestNamespaceRoles(t *testing.T) {
	p := Principal{Claims: []Claim{{Role: "Viewer", NamespaceID: "a"}, {Role: "Editor", NamespaceID: "b"}}}
	if !p.Can("a", false) || p.Can("a", true) || !p.Can("b", true) || p.Can("c", false) {
		t.Fatal("namespace role boundary failed")
	}
}
