// Package orm is the persistence module: the Ent client, its schema, and the shapes kept in JSON columns.
package orm

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"

	"orm/ent"
)

// Client is the Ent client every adapter queries through.
type Client struct {
	*ent.Client
}

const pragmas = "_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"

// Open opens the SQLite file at path, creating it if it is missing, with foreign keys on. It never
// creates a table: call client.Schema.Create. The path goes into a file: URI, so it must not hold
// "?" or "#".
func Open(ctx context.Context, path string) (*Client, error) {
	return open(ctx, "file:"+path+"?"+pragmas, 0)
}

var memoryDBs atomic.Int64

// OpenMemory opens a private in-memory database, for tests.
func OpenMemory(ctx context.Context) (*Client, error) {
	dsn := fmt.Sprintf("file:mrav-memory-%d?mode=memory&cache=shared&%s", memoryDBs.Add(1), pragmas)
	return open(ctx, dsn, 1)
}

func open(ctx context.Context, dsn string, maxConns int) (*Client, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if maxConns > 0 {
		db.SetMaxOpenConns(maxConns) // an in-memory database lives and dies with its connections
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	return &Client{Client: ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, db)))}, nil
}

// Close closes the Ent driver, which closes the *sql.DB.
func (c *Client) Close() error {
	return c.Client.Close()
}
