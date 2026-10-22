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

type Client struct {
	*ent.Client
	DB *sql.DB
}

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

func Open(ctx context.Context, path string) (*Client, error) {
	return open(ctx, "file:"+path+"?"+pragmas, 0, nil)
}

func OpenTraced(ctx context.Context, path string, trace func(query string, args []any)) (*Client, error) {
	return open(ctx, "file:"+path+"?"+pragmas, 0, trace)
}

var memoryDBs atomic.Int64

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
		db.SetMaxOpenConns(maxConns)
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

func (c *Client) Close() error {
	return c.Client.Close()
}
