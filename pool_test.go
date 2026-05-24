package mysql

import (
	"errors"
	"testing"

	"github.com/netlifeguru/db"
)

func newPoolTestConnect() *Connect {
	return &Connect{
		shared: &sharedState{
			Pools:         make(map[string]*Pool),
			sharedPools:   make(map[string]*sharedPool),
			myConnections: make([]string, 0),
		},
	}
}

func mysqlTestConfig(identifier string) db.Config {
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

func seedPool(c *Connect, cfg db.Config, refs int, sharedRefs int) *Pool {
	cfg = normalizeConfig(cfg)
	key := connectionFingerprint(cfg)

	p := &Pool{
		Connection: nil,
		Connect:    cfg,
		ConfigKey:  key,
		PoolKey:    key,
		Refs:       refs,
	}

	c.shared.Pools[cfg.Identifier] = p
	c.shared.sharedPools[key] = &sharedPool{
		Connection: nil,
		Refs:       sharedRefs,
	}
	c.Identifier = cfg.Identifier
	c.Host = cfg.Host

	return p
}

func TestCreatePoolSameIdentifierDifferentConfigReturnsConflictBeforeConnect(t *testing.T) {
	t.Parallel()

	c := newPoolTestConnect()

	cfg1 := mysqlTestConfig("main")
	cfg2 := mysqlTestConfig("main")
	cfg2.Database = "other"

	seedPool(c, cfg1, 1, 1)

	err := c.CreatePool(cfg2)
	if !errors.Is(err, db.ErrPoolIdentifierConflict) {
		t.Fatalf("expected ErrPoolIdentifierConflict, got %v", err)
	}

	p := c.shared.Pools["main"]
	if p == nil {
		t.Fatalf("expected original pool to remain")
	}
	if p.Connect.Database != "app" {
		t.Fatalf("expected original database app, got %q", p.Connect.Database)
	}
	if p.Refs != 1 {
		t.Fatalf("expected refs to remain 1, got %d", p.Refs)
	}
	if len(c.shared.sharedPools) != 1 {
		t.Fatalf("expected shared pool count to remain 1, got %d", len(c.shared.sharedPools))
	}
}

func TestCreatePoolSameIdentifierSameConfigIncrementsRefsWithoutConnect(t *testing.T) {
	t.Parallel()

	c := newPoolTestConnect()

	cfg := mysqlTestConfig("main")
	seedPool(c, cfg, 1, 1)

	err := c.CreatePool(cfg)
	if err != nil {
		t.Fatalf("CreatePool returned error: %v", err)
	}

	p := c.shared.Pools["main"]
	if p == nil {
		t.Fatalf("expected pool main")
	}
	if p.Refs != 2 {
		t.Fatalf("expected pool refs 2, got %d", p.Refs)
	}

	shared := c.shared.sharedPools[p.PoolKey]
	if shared == nil {
		t.Fatalf("expected shared pool")
	}
	if shared.Refs != 1 {
		t.Fatalf("expected shared refs to remain 1, got %d", shared.Refs)
	}
	if c.Identifier != "main" {
		t.Fatalf("expected current identifier main, got %q", c.Identifier)
	}
	if c.Host != "localhost" {
		t.Fatalf("expected host localhost, got %q", c.Host)
	}
	if !contains(driverName, c.shared.myConnections) {
		t.Fatalf("expected myConnections to contain %q", driverName)
	}
}

func TestCreatePoolNormalizesExistingIdentifier(t *testing.T) {
	t.Parallel()

	c := newPoolTestConnect()

	cfg := mysqlTestConfig("localhost:3306")
	seedPool(c, cfg, 1, 1)

	incoming := mysqlTestConfig("http://127.0.0.1:3306/")
	err := c.CreatePool(incoming)
	if err != nil {
		t.Fatalf("CreatePool returned error: %v", err)
	}

	if c.Identifier != "localhost:3306" {
		t.Fatalf("expected normalized identifier localhost:3306, got %q", c.Identifier)
	}
	if c.shared.Pools["localhost:3306"].Refs != 2 {
		t.Fatalf("expected refs 2, got %d", c.shared.Pools["localhost:3306"].Refs)
	}
}

func TestConnectionReturnsCurrentPool(t *testing.T) {
	t.Parallel()

	c := newPoolTestConnect()

	cfg := mysqlTestConfig("main")
	seedPool(c, cfg, 1, 1)

	gotCfg, gotConn := c.Connection()
	if gotConn != nil {
		t.Fatalf("expected nil sql.DB in unit test, got %#v", gotConn)
	}
	if gotCfg.Identifier != "main" {
		t.Fatalf("expected identifier main, got %q", gotCfg.Identifier)
	}
	if gotCfg.Database != "app" {
		t.Fatalf("expected database app, got %q", gotCfg.Database)
	}
}

func TestConnectionUnknownIdentifierReturnsZeroConfigAndNilConnection(t *testing.T) {
	t.Parallel()

	c := newPoolTestConnect()
	c.Identifier = "missing"

	gotCfg, gotConn := c.Connection()
	if gotConn != nil {
		t.Fatalf("expected nil connection, got %#v", gotConn)
	}
	if gotCfg.Identifier != "" || gotCfg.Host != "" || gotCfg.Port != 0 {
		t.Fatalf("expected zero config, got %#v", gotCfg)
	}
}

func TestCloseEmptyIdentifierDoesNothing(t *testing.T) {
	t.Parallel()

	c := newPoolTestConnect()
	c.Close()

	if len(c.shared.Pools) != 0 {
		t.Fatalf("expected no pools, got %d", len(c.shared.Pools))
	}
	if len(c.shared.sharedPools) != 0 {
		t.Fatalf("expected no shared pools, got %d", len(c.shared.sharedPools))
	}
}

func TestCloseUnknownIdentifierClearsIdentifier(t *testing.T) {
	t.Parallel()

	c := newPoolTestConnect()
	c.Identifier = "missing"

	c.Close()

	if c.Identifier != "" {
		t.Fatalf("expected identifier to be cleared, got %q", c.Identifier)
	}
}

func TestCloseDecrementsLogicalPoolRefs(t *testing.T) {
	t.Parallel()

	c := newPoolTestConnect()
	cfg := mysqlTestConfig("main")
	seedPool(c, cfg, 2, 1)

	c.Close()

	if c.Identifier != "" {
		t.Fatalf("expected identifier to be cleared, got %q", c.Identifier)
	}

	p := c.shared.Pools["main"]
	if p == nil {
		t.Fatalf("expected logical pool to remain")
	}
	if p.Refs != 1 {
		t.Fatalf("expected pool refs 1, got %d", p.Refs)
	}
	if len(c.shared.sharedPools) != 1 {
		t.Fatalf("expected shared pool to remain, got %d", len(c.shared.sharedPools))
	}
}

func TestCloseRemovesLastLogicalPoolAndSharedPool(t *testing.T) {
	t.Parallel()

	c := newPoolTestConnect()
	cfg := mysqlTestConfig("main")
	seedPool(c, cfg, 1, 1)

	c.Close()

	if c.Identifier != "" {
		t.Fatalf("expected identifier to be cleared, got %q", c.Identifier)
	}
	if len(c.shared.Pools) != 0 {
		t.Fatalf("expected no logical pools, got %d", len(c.shared.Pools))
	}
	if len(c.shared.sharedPools) != 0 {
		t.Fatalf("expected no shared pools, got %d", len(c.shared.sharedPools))
	}
}

func TestCloseSharedPoolWithMultipleLogicalPools(t *testing.T) {
	t.Parallel()

	c := newPoolTestConnect()

	cfg1 := mysqlTestConfig("main")
	cfg2 := normalizeConfig(mysqlTestConfig("readonly"))

	p1 := seedPool(c, cfg1, 1, 2)
	p2 := &Pool{
		Connection: nil,
		Connect:    cfg2,
		ConfigKey:  p1.ConfigKey,
		PoolKey:    p1.PoolKey,
		Refs:       1,
	}

	c.shared.Pools["readonly"] = p2
	c.Identifier = "main"

	c.Close()

	if c.Identifier != "" {
		t.Fatalf("expected identifier to be cleared, got %q", c.Identifier)
	}
	if _, ok := c.shared.Pools["main"]; ok {
		t.Fatalf("expected main pool to be removed")
	}
	if _, ok := c.shared.Pools["readonly"]; !ok {
		t.Fatalf("expected readonly pool to remain")
	}
	if len(c.shared.sharedPools) != 1 {
		t.Fatalf("expected shared pool to remain, got %d", len(c.shared.sharedPools))
	}

	shared := c.shared.sharedPools[p1.PoolKey]
	if shared == nil {
		t.Fatalf("expected shared pool")
	}
	if shared.Refs != 1 {
		t.Fatalf("expected shared refs 1, got %d", shared.Refs)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	t.Parallel()

	c := newPoolTestConnect()
	cfg := mysqlTestConfig("main")
	seedPool(c, cfg, 1, 1)

	c.Close()
	c.Close()

	if c.Identifier != "" {
		t.Fatalf("expected identifier to stay empty, got %q", c.Identifier)
	}
	if len(c.shared.Pools) != 0 {
		t.Fatalf("expected no pools after repeated close, got %d", len(c.shared.Pools))
	}
	if len(c.shared.sharedPools) != 0 {
		t.Fatalf("expected no shared pools after repeated close, got %d", len(c.shared.sharedPools))
	}
}

func TestClosePoolMissingSharedPoolDoesNotPanic(t *testing.T) {
	t.Parallel()

	c := newPoolTestConnect()
	cfg := mysqlTestConfig("main")
	p := seedPool(c, cfg, 1, 1)
	delete(c.shared.sharedPools, p.PoolKey)

	c.Close()

	if c.Identifier != "" {
		t.Fatalf("expected identifier to be cleared, got %q", c.Identifier)
	}
	if len(c.shared.Pools) != 0 {
		t.Fatalf("expected logical pool to be removed, got %d", len(c.shared.Pools))
	}
	if len(c.shared.sharedPools) != 0 {
		t.Fatalf("expected no shared pools, got %d", len(c.shared.sharedPools))
	}
}

func TestCloseNilSharedConnectionDoesNotPanic(t *testing.T) {
	t.Parallel()

	c := newPoolTestConnect()
	cfg := mysqlTestConfig("main")
	seedPool(c, cfg, 1, 1)

	c.Close()

	if len(c.shared.Pools) != 0 {
		t.Fatalf("expected no logical pools, got %d", len(c.shared.Pools))
	}
	if len(c.shared.sharedPools) != 0 {
		t.Fatalf("expected no shared pools, got %d", len(c.shared.sharedPools))
	}
}
