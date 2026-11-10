package migrations

import "embed"

// Files contains the SQL migrations embedded in the backend binary.
//
//go:embed *.sql
var Files embed.FS
