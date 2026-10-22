package testutil

import (
	"context"
	"database/sql"
	"strings"
)

func Plan(ctx context.Context, db *sql.DB, query string, args ...any) ([]string, error) {
	rows, err := db.QueryContext(ctx, "EXPLAIN QUERY PLAN "+query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var steps []string
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
			return nil, err
		}
		steps = append(steps, detail)
	}
	return steps, rows.Err()
}

func Unindexed(query string, plan []string) []string {
	upper := strings.ToUpper(query)
	listing := !strings.Contains(upper, "WHERE") && !strings.Contains(upper, "JOIN")
	var steps []string
	for _, step := range plan {
		switch {
		case strings.Contains(step, "USE TEMP B-TREE"):
			steps = append(steps, step)
		case strings.HasPrefix(step, "SCAN ") && !listing:
			steps = append(steps, step)
		}
	}
	return steps
}
