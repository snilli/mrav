//go:build embeddb

package database

import _ "embed"

// embedded is the database built into the program, for one file that needs nothing beside it. The build copies
// data/mrav.sqlite here first (see build-windows.sh); the copy is not in git.
//
//go:embed mrav.sqlite
var embedded []byte
