package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/netlifeguru/db"
)

func (c *Connect) CreatePool(cfg db.Config) error {
	cfg = normalizeConfig(cfg)

	identifier := cfg.Identifier
	if identifier == "" {
		return db.ErrEmptyIdentifier
	}

	c.shared.poolsMu.Lock()
	defer c.shared.poolsMu.Unlock()

	if !contains(driverName, c.shared.myConnections) {
		c.shared.myConnections = append(c.shared.myConnections, driverName)
	}

	key := connectionFingerprint(cfg)

	if existing, ok := c.shared.Pools[identifier]; ok {
		if existing.ConfigKey != key {
			return fmt.Errorf("%w: %q", db.ErrPoolIdentifierConflict, identifier)
		}

		existing.Refs++

		c.Identifier = identifier
		c.Host = cfg.Host
		return nil
	}

	var conn *sql.DB

	if existingShared, ok := c.shared.sharedPools[key]; ok {
		existingShared.Refs++
		conn = existingShared.Connection
	} else {
		created, err := c.connectWithLimits(cfg)
		if err != nil {
			return err
		}

		c.shared.sharedPools[key] = &sharedPool{
			Connection: created,
			Refs:       1,
		}

		conn = created
	}

	c.shared.Pools[identifier] = &Pool{
		Connection: conn,
		Connect:    cfg,
		ConfigKey:  key,
		PoolKey:    key,
		Refs:       1,
	}

	c.Identifier = identifier
	c.Host = cfg.Host

	return nil
}

func (c *Connect) connectWithLimits(cfg db.Config) (*sql.DB, error) {
	if cfg.MaxConns == 0 {
		cfg.MaxConns = 10
	}
	if cfg.MaxConnIdleTime == 0 {
		cfg.MaxConnIdleTime = 5 * time.Minute
	}
	if cfg.MaxConnLifetime == 0 {
		cfg.MaxConnLifetime = 1 * time.Hour
	}
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = 5 * time.Second
	}

	loc := "Local"
	if cfg.TimeZone != "" {
		loc = cfg.TimeZone
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=%s&timeout=%s&charset=utf8mb4",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		url.QueryEscape(loc),
		cfg.ConnectTimeout.String(),
	)

	switch cfg.SSLMode {
	case "", "disable":
		dsn += "&tls=false"

	case "require":
		dsn += "&tls=true"

	case "verify-ca", "verify-full":
		dsn += "&tls=true"

	case "skip-verify":
		dsn += "&tls=skip-verify"

	default:
		dsn += "&tls=" + cfg.SSLMode
	}

	d, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	d.SetMaxOpenConns(int(cfg.MaxConns))
	d.SetMaxIdleConns(int(cfg.MinConns))
	d.SetConnMaxLifetime(cfg.MaxConnLifetime)
	d.SetConnMaxIdleTime(cfg.MaxConnIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	if err := d.PingContext(ctx); err != nil {
		err := d.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return d, nil
}

func (c *Connect) Connection() (db.Config, *sql.DB) {
	c.shared.poolsMu.RLock()
	defer c.shared.poolsMu.RUnlock()
	p, ok := c.shared.Pools[c.Identifier]
	if !ok {
		return db.Config{}, nil
	}
	return p.Connect, p.Connection
}

func (c *Connect) closePoolHook(p *sql.DB) {
	err := p.Close()
	if err != nil {
		return
	}
}

func (c *Connect) Close() {
	c.shared.poolsMu.Lock()
	defer c.shared.poolsMu.Unlock()

	if c.Identifier == "" {
		return
	}

	p, ok := c.shared.Pools[c.Identifier]
	if !ok {
		c.Identifier = ""
		return
	}

	if p.Refs > 1 {
		p.Refs--
		c.Identifier = ""
		return
	}

	delete(c.shared.Pools, c.Identifier)

	if shared, ok := c.shared.sharedPools[p.PoolKey]; ok {
		if shared.Refs > 1 {
			shared.Refs--
			c.Identifier = ""
			return
		}

		if shared.Connection != nil {
			c.closePoolHook(shared.Connection)
		}

		delete(c.shared.sharedPools, p.PoolKey)
	}

	c.Identifier = ""
}
