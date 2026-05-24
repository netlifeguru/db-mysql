package mysql

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/netlifeguru/db"
)

type Pool struct {
	Connect    db.Config
	Connection *sql.DB
	ConfigKey  string
	PoolKey    string
	Refs       int
}

type sharedPool struct {
	Connection *sql.DB
	Refs       int
}

type sharedState struct {
	Pools         map[string]*Pool
	sharedPools   map[string]*sharedPool
	poolsMu       sync.RWMutex
	myConnections []string
}

type Connect struct {
	newPoolHook func(cfg db.Config) (*sql.DB, error)
	shared      *sharedState
	Host        string
	Identifier  string
	TX          any
	err         error
	Timeout     time.Duration
	pool        *sql.DB
}

func (c *Connect) LoadFile() string {
	return "model.sql"
}

func (c *Connect) DriverName() string {
	return "mysql"
}

const defaultQueryTimeout = 5 * time.Second

func New() *Connect {
	return &Connect{
		shared: &sharedState{
			Pools:         make(map[string]*Pool),
			sharedPools:   make(map[string]*sharedPool),
			myConnections: make([]string, 0),
		},
	}
}

func (c *Connect) Settings(identifier string) error {
	identifier = normalizeIdentifier(identifier)
	if identifier == "" {
		return db.ErrEmptyIdentifier
	}

	c.shared.poolsMu.RLock()
	defer c.shared.poolsMu.RUnlock()

	if _, ok := c.shared.Pools[identifier]; !ok {
		return db.ErrNoConnection
	}

	c.Identifier = identifier
	return nil
}

func connectionFingerprint(cfg db.Config) string {
	raw := fmt.Sprintf(
		"driver=%s host=%s port=%d db=%s user=%s password=%s ssl=%s tz=%s",
		"mysql",
		strings.Replace(cfg.Host, "127.0.0.1", "localhost", 1),
		cfg.Port,
		cfg.Database,
		cfg.Username,
		cfg.Password,
		cfg.SSLMode,
		cfg.TimeZone,
	)

	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func contains(value string, slice []string) bool {
	for _, s := range slice {
		if s == value {
			return true
		}
	}
	return false
}

func (c *Connect) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return defaultQueryTimeout
}

func (c *Connect) Fork() db.Conn {
	c.shared.poolsMu.Lock()
	defer c.shared.poolsMu.Unlock()

	if c.Identifier != "" {
		if p, ok := c.shared.Pools[c.Identifier]; ok && p != nil {
			p.Refs++
		}
	}

	return &Connect{
		newPoolHook: c.newPoolHook,
		shared:      c.shared,
		Host:        c.Host,
		Identifier:  c.Identifier,
		Timeout:     c.Timeout,
	}
}
