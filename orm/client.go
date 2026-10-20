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

// Client is the Ent client every adapter queries through. DB is the connection under it, for the query that Ent's
// builder does not say well.
type Client struct {
	*ent.Client
	DB *sql.DB
}

// traced tells a function of every statement that Ent sends, before it goes.
type traced struct {
	dialect.Driver
	trace func(query string, args []any)
}

func (t traced) Exec(ctx context.Context, query string, args, v any) error {
	t.trace(query, argList(args))
	return t.Driver.Exec(ctx, query, args, v)
}

func (t traced) Query(ctx context.Context, query string, args, v any) error {
	t.trace(query, argList(args))
	return t.Driver.Query(ctx, query, args, v)
}

func argList(args any) []any {
	list, _ := args.([]any)
	return list
}

const pragmas = "_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"

// Open opens the SQLite file at path, creating it if it is missing, with foreign keys on. It never
// creates a table: call client.Schema.Create. The path goes into a file: URI, so it must not hold
// "?" or "#".
func Open(ctx context.Context, path string) (*Client, error) {
	return open(ctx, "file:"+path+"?"+pragmas, 0, nil)
}

// OpenTraced is Open with trace told of every statement that Ent sends, with its arguments: for a test that wants to see
// what the repositories ask the database. The repositories send nothing but through Ent.
func OpenTraced(ctx context.Context, path string, trace func(query string, args []any)) (*Client, error) {
	return open(ctx, "file:"+path+"?"+pragmas, 0, trace)
}

var memoryDBs atomic.Int64

// OpenMemory opens a private in-memory database, for tests.
func OpenMemory(ctx context.Context) (*Client, error) {
	dsn := fmt.Sprintf("file:mrav-memory-%d?mode=memory&cache=shared&%s", memoryDBs.Add(1), pragmas)
	return open(ctx, dsn, 1, nil)
}

func open(ctx context.Context, dsn string, maxConns int, trace func(query string, args []any)) (*Client, error) {
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
	var drv dialect.Driver = entsql.OpenDB(dialect.SQLite, db)
	if trace != nil {
		drv = traced{Driver: drv, trace: trace}
	}
	return &Client{Client: ent.NewClient(ent.Driver(drv)), DB: db}, nil
}

// Close closes the Ent driver, which closes the *sql.DB.
func (c *Client) Close() error {
	return c.Client.Close()
}
