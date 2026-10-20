package testutil

import (
	"context"
	"database/sql"
	"strings"
)

// Plan is the plan that SQLite makes for a statement with the arguments given, one line for each step of it.
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

// Unindexed is the steps of a plan that no index serves: rows sorted afterwards, and a table or an index read from end to end
// to find rows. A statement with no WHERE and no JOIN lists a whole table, which no index could speed up, so its scan is let
// through, though not a sort of it.
func Unindexed(query string, plan []string) []string {
	upper := strings.ToUpper(query)
	listing := !strings.Contains(upper, "WHERE") && !strings.Contains(upper, "JOIN")
	var steps []string
	for _, step := range plan {
		switch {
		case strings.Contains(step, "USE TEMP B-TREE"):
			steps = append(steps, step)
		case strings.HasPrefix(step, "SCAN ") && !listing: // an index read from end to end finds no row by its key either
			steps = append(steps, step)
		}
	}
	return steps
}
