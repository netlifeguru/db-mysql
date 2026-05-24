package mysql

import (
	"errors"
	"testing"
	"time"

	"github.com/netlifeguru/db"
)

func newSettingsTestPool(cfg db.Config, refs int) *Pool {
	cfg.Identifier = normalizeIdentifier(cfg.Identifier)
	key := connectionFingerprint(cfg)

	return &Pool{
		Connect:    cfg,
		Connection: nil,
		ConfigKey:  key,
		PoolKey:    key,
		Refs:       refs,
	}
}

func settingsTestConfig(identifier string) db.Config {
	return db.Config{
		Identifier: identifier,
		Host:       "localhost",
		Port:       3306,
		Database:   "app",
		Username:   "user",
		Password:   "secret",
		SSLMode:    "disable",
		TimeZone:   "UTC",
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	c := New()
	if c == nil {
		t.Fatalf("expected connection")
	}
	if c.shared == nil {
		t.Fatalf("expected shared state")
	}
	if c.shared.Pools == nil {
		t.Fatalf("expected Pools map")
	}
	if c.shared.sharedPools == nil {
		t.Fatalf("expected sharedPools map")
	}
	if c.shared.myConnections == nil {
		t.Fatalf("expected myConnections slice")
	}
	if len(c.shared.Pools) != 0 {
		t.Fatalf("expected empty Pools map, got %d", len(c.shared.Pools))
	}
	if len(c.shared.sharedPools) != 0 {
		t.Fatalf("expected empty sharedPools map, got %d", len(c.shared.sharedPools))
	}
}

func TestLoadFile(t *testing.T) {
	t.Parallel()

	c := New()
	if got := c.LoadFile(); got != "model.sql" {
		t.Fatalf("expected model.sql, got %q", got)
	}
}

func TestDriverName(t *testing.T) {
	t.Parallel()

	c := New()
	if got := c.DriverName(); got != "mysql" {
		t.Fatalf("expected mysql, got %q", got)
	}
}

func TestSettingsRejectsUnknownIdentifier(t *testing.T) {
	t.Parallel()

	c := New()

	err := c.Settings("missing")
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}
	if c.Identifier != "" {
		t.Fatalf("expected identifier to remain empty, got %q", c.Identifier)
	}
}

func TestSettingsUsesDefaultIdentifierForEmptyInput(t *testing.T) {
	t.Parallel()

	c := New()
	cfg := settingsTestConfig("default")
	c.shared.Pools["default"] = newSettingsTestPool(cfg, 1)

	err := c.Settings("   ")
	if err != nil {
		t.Fatalf("Settings returned error: %v", err)
	}
	if c.Identifier != "default" {
		t.Fatalf("expected identifier default, got %q", c.Identifier)
	}
}

func TestSettingsSwitchesIdentifier(t *testing.T) {
	t.Parallel()

	c := New()
	cfg := settingsTestConfig("main")
	c.shared.Pools["main"] = newSettingsTestPool(cfg, 1)

	err := c.Settings("main")
	if err != nil {
		t.Fatalf("Settings returned error: %v", err)
	}
	if c.Identifier != "main" {
		t.Fatalf("expected identifier main, got %q", c.Identifier)
	}
}

func TestSettingsNormalizesIdentifier(t *testing.T) {
	t.Parallel()

	c := New()
	cfg := settingsTestConfig("localhost")
	c.shared.Pools["localhost"] = newSettingsTestPool(cfg, 1)

	err := c.Settings("  http://127.0.0.1/  ")
	if err != nil {
		t.Fatalf("Settings returned error: %v", err)
	}
	if c.Identifier != "localhost" {
		t.Fatalf("expected identifier localhost, got %q", c.Identifier)
	}
}

func TestConnectionFingerprintStable(t *testing.T) {
	t.Parallel()

	cfg := settingsTestConfig("main")
	first := connectionFingerprint(cfg)
	second := connectionFingerprint(cfg)

	if first == "" {
		t.Fatalf("expected fingerprint")
	}
	if first != second {
		t.Fatalf("expected stable fingerprint, got %q and %q", first, second)
	}
}

func TestConnectionFingerprintNormalizesHost(t *testing.T) {
	t.Parallel()

	cfg1 := settingsTestConfig("main")
	cfg1.Host = "localhost"

	cfg2 := settingsTestConfig("main")
	cfg2.Host = "127.0.0.1"

	if connectionFingerprint(cfg1) != connectionFingerprint(cfg2) {
		t.Fatalf("expected localhost and 127.0.0.1 to have same fingerprint")
	}
}

func TestConnectionFingerprintIgnoresIdentifier(t *testing.T) {
	t.Parallel()

	cfg1 := settingsTestConfig("main")
	cfg2 := settingsTestConfig("readonly")

	if connectionFingerprint(cfg1) != connectionFingerprint(cfg2) {
		t.Fatalf("expected fingerprint to ignore identifier")
	}
}

func TestConnectionFingerprintChangesWhenConfigChanges(t *testing.T) {
	t.Parallel()

	base := settingsTestConfig("main")
	baseFP := connectionFingerprint(base)

	tests := []struct {
		name   string
		modify func(*db.Config)
	}{
		{"host", func(c *db.Config) { c.Host = "db.internal" }},
		{"port", func(c *db.Config) { c.Port = 3307 }},
		{"database", func(c *db.Config) { c.Database = "other" }},
		{"username", func(c *db.Config) { c.Username = "other" }},
		{"password", func(c *db.Config) { c.Password = "other" }},
		{"ssl", func(c *db.Config) { c.SSLMode = "require" }},
		{"timezone", func(c *db.Config) { c.TimeZone = "Europe/Bratislava" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := base
			tt.modify(&cfg)

			got := connectionFingerprint(cfg)
			if got == baseFP {
				t.Fatalf("expected fingerprint to change for %s", tt.name)
			}
		})
	}
}

func TestContains(t *testing.T) {
	t.Parallel()

	if !contains("mysql", []string{"postgresql", "mysql"}) {
		t.Fatalf("expected slice to contain mysql")
	}
	if contains("scylla", []string{"postgresql", "mysql"}) {
		t.Fatalf("expected slice not to contain scylla")
	}
	if contains("mysql", nil) {
		t.Fatalf("expected nil slice not to contain mysql")
	}
}

func TestTimeoutDefault(t *testing.T) {
	t.Parallel()

	c := New()
	if got := c.timeout(); got != defaultQueryTimeout {
		t.Fatalf("expected default timeout %s, got %s", defaultQueryTimeout, got)
	}
}

func TestTimeoutCustom(t *testing.T) {
	t.Parallel()

	c := New()
	c.Timeout = 10 * time.Second

	if got := c.timeout(); got != 10*time.Second {
		t.Fatalf("expected custom timeout 10s, got %s", got)
	}
}

func TestForkSharesStateAndCopiesConnectionSettings(t *testing.T) {
	t.Parallel()

	c := New()
	c.Host = "localhost"
	c.Identifier = "main"
	c.Timeout = 10 * time.Second

	cfg := settingsTestConfig("main")
	p := newSettingsTestPool(cfg, 1)
	c.shared.Pools["main"] = p

	forked := c.Fork()
	if forked == nil {
		t.Fatalf("expected forked connection")
	}

	fc, ok := forked.(*Connect)
	if !ok {
		t.Fatalf("expected *Connect, got %T", forked)
	}
	if fc == c {
		t.Fatalf("expected different Connect instance")
	}
	if fc.shared != c.shared {
		t.Fatalf("expected shared state to be reused")
	}
	if fc.Host != c.Host {
		t.Fatalf("expected host %q, got %q", c.Host, fc.Host)
	}
	if fc.Identifier != c.Identifier {
		t.Fatalf("expected identifier %q, got %q", c.Identifier, fc.Identifier)
	}
	if fc.Timeout != c.Timeout {
		t.Fatalf("expected timeout %s, got %s", c.Timeout, fc.Timeout)
	}
}

func TestForkIncrementsRefsForCurrentIdentifier(t *testing.T) {
	t.Parallel()

	c := New()
	c.Identifier = "main"

	cfg := settingsTestConfig("main")
	p := newSettingsTestPool(cfg, 1)
	c.shared.Pools["main"] = p

	_ = c.Fork()

	if p.Refs != 2 {
		t.Fatalf("expected refs 2 after fork, got %d", p.Refs)
	}
}

func TestForkWithoutIdentifierDoesNotIncrementRefs(t *testing.T) {
	t.Parallel()

	c := New()
	cfg := settingsTestConfig("main")
	p := newSettingsTestPool(cfg, 1)
	c.shared.Pools["main"] = p

	_ = c.Fork()

	if p.Refs != 1 {
		t.Fatalf("expected refs to remain 1, got %d", p.Refs)
	}
}

func TestForkUnknownIdentifierDoesNotPanic(t *testing.T) {
	t.Parallel()

	c := New()
	c.Identifier = "missing"

	forked := c.Fork()
	if forked == nil {
		t.Fatalf("expected forked connection")
	}

	fc, ok := forked.(*Connect)
	if !ok {
		t.Fatalf("expected *Connect, got %T", forked)
	}
	if fc.Identifier != "missing" {
		t.Fatalf("expected identifier missing, got %q", fc.Identifier)
	}
}
