//go:build embeddb

package database

import _ "embed"

//go:embed mrav.sqlite
var embedded []byte
