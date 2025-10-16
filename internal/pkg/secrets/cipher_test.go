package secrets

import (
	"strings"
	"testing"
)

func TestCipherBindsScopeAndRejectsTampering(t *testing.T) {
	c, err := New([]byte(strings.Repeat("k", 32)))
	if err != nil {
		t.Fatal(err)
	}
	a, err := c.Seal("tenant-a/secret/token", "private-token")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := c.Seal("tenant-a/secret/token", "private-token")
	if a == b {
		t.Fatal("nonce was reused")
	}
	if value, err := c.Open("tenant-a/secret/token", a); err != nil || value != "private-token" {
		t.Fatalf("decrypt: %s %v", value, err)
	}
	if _, err = c.Open("tenant-b/secret/token", a); err == nil {
		t.Fatal("cross-tenant decryption succeeded")
	}
	if _, err = c.Open("tenant-a/secret/token", a[:len(a)-3]+"AAA"); err == nil {
		t.Fatal("tampered ciphertext accepted")
	}
}
