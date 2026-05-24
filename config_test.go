package mysql

import (
	"testing"
	"time"

	"github.com/netlifeguru/db"
)

func TestNormalizeIdentifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty uses default", "", "default"},
		{"spaces uses default", "   ", "default"},
		{"trim and lower", "  MAIN  ", "main"},
		{"http prefix", "http://MAIN", "main"},
		{"https prefix", "https://MAIN", "main"},
		{"trailing slash", "https://MAIN/", "main"},
		{"loopback", "127.0.0.1", "localhost"},
		{"loopback with port", "127.0.0.1:3306", "localhost:3306"},
		{"localhost kept", "localhost", "localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeIdentifier(tt.in)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestNormalizeConfigDefaults(t *testing.T) {
	t.Parallel()

	got := normalizeConfig(db.Config{})

	if got.Identifier != defaultIdentifier {
		t.Fatalf("expected identifier %q, got %q", defaultIdentifier, got.Identifier)
	}
	if got.Host != defaultHost {
		t.Fatalf("expected host %q, got %q", defaultHost, got.Host)
	}
	if got.Port != defaultPort {
		t.Fatalf("expected port %d, got %d", defaultPort, got.Port)
	}
	if got.MaxConns != defaultMaxConns {
		t.Fatalf("expected max conns %d, got %d", defaultMaxConns, got.MaxConns)
	}
	if got.MinConns != defaultMinConns {
		t.Fatalf("expected min conns %d, got %d", defaultMinConns, got.MinConns)
	}
	if got.MaxConnIdleTime != defaultMaxConnIdleTime {
		t.Fatalf("expected idle time %s, got %s", defaultMaxConnIdleTime, got.MaxConnIdleTime)
	}
	if got.MaxConnLifetime != defaultMaxConnLifetime {
		t.Fatalf("expected lifetime %s, got %s", defaultMaxConnLifetime, got.MaxConnLifetime)
	}
	if got.ConnectTimeout != defaultConnectTimeout {
		t.Fatalf("expected connect timeout %s, got %s", defaultConnectTimeout, got.ConnectTimeout)
	}
	if got.HealthCheckPeriod != defaultHealthCheckPeriod {
		t.Fatalf("expected health check period %s, got %s", defaultHealthCheckPeriod, got.HealthCheckPeriod)
	}
	if got.SSLMode != defaultSSLMode {
		t.Fatalf("expected ssl mode %q, got %q", defaultSSLMode, got.SSLMode)
	}
	if got.TimeZone != defaultTimeZone {
		t.Fatalf("expected timezone %q, got %q", defaultTimeZone, got.TimeZone)
	}
}

func TestNormalizeConfigPreservesExplicitValues(t *testing.T) {
	t.Parallel()

	cfg := db.Config{
		Identifier:        " MAIN ",
		Host:              "db.internal",
		Port:              3307,
		MaxConns:          50,
		MinConns:          10,
		MaxConnIdleTime:   time.Minute,
		MaxConnLifetime:   2 * time.Hour,
		ConnectTimeout:    3 * time.Second,
		HealthCheckPeriod: 7 * time.Second,
		SSLMode:           "require",
		TimeZone:          "UTC",
	}

	got := normalizeConfig(cfg)

	if got.Identifier != "main" {
		t.Fatalf("expected normalized identifier main, got %q", got.Identifier)
	}
	if got.Host != cfg.Host || got.Port != cfg.Port || got.MaxConns != cfg.MaxConns || got.MinConns != cfg.MinConns {
		t.Fatalf("unexpected normalized config: %#v", got)
	}
	if got.MaxConnIdleTime != cfg.MaxConnIdleTime || got.MaxConnLifetime != cfg.MaxConnLifetime || got.ConnectTimeout != cfg.ConnectTimeout || got.HealthCheckPeriod != cfg.HealthCheckPeriod {
		t.Fatalf("unexpected duration config: %#v", got)
	}
	if got.SSLMode != cfg.SSLMode || got.TimeZone != cfg.TimeZone {
		t.Fatalf("unexpected string config: %#v", got)
	}
}

func TestNormalizeConfigMinConns(t *testing.T) {
	t.Parallel()

	negative := normalizeConfig(db.Config{MaxConns: 10, MinConns: -1})
	if negative.MinConns != defaultMinConns {
		t.Fatalf("expected negative min conns to use default %d, got %d", defaultMinConns, negative.MinConns)
	}

	clamped := normalizeConfig(db.Config{MaxConns: 3, MinConns: 10})
	if clamped.MinConns != 3 {
		t.Fatalf("expected min conns to be clamped to max conns 3, got %d", clamped.MinConns)
	}
}
