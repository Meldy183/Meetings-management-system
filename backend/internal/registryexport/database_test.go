package registryexport

import (
	"strings"
	"testing"
	"time"
)

func TestReadOnlyConfig(t *testing.T) {
	cfg, err := readOnlyConfig("postgres://example:secret@localhost/registry?default_transaction_read_only=off&statement_timeout=0&lock_timeout=0&idle_in_transaction_session_timeout=0")
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{
		"application_name":                    "meetings-registry-export",
		"default_transaction_read_only":       "on",
		"statement_timeout":                   "30000",
		"lock_timeout":                        "2000",
		"idle_in_transaction_session_timeout": "30000",
	} {
		if got := cfg.RuntimeParams[name]; got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	if cfg.ConnectTimeout != 10*time.Second {
		t.Errorf("unexpected connection timeout: %v", cfg.ConnectTimeout)
	}
}

func TestReadOnlyConfigRejectsMissingAndInvalidDSN(t *testing.T) {
	for _, dsn := range []string{"", "postgres://example:private-password@localhost:invalid/registry"} {
		_, err := readOnlyConfig(dsn)
		if err == nil {
			t.Fatalf("expected error for invalid connection string")
		}
		if strings.Contains(err.Error(), "private-password") {
			t.Fatal("error must not disclose the database password")
		}
	}
}
