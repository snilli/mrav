//go:build !embeddb

package database

// embedded is the database built into the program: none, unless it was built with the embeddb tag.
var embedded []byte
