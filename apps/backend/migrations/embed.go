// Package migrations embeds the SQL migration files so they ship inside the
// compiled binary and can be applied without any external tooling.
package migrations

import "embed"

// Files holds every numbered *.sql migration, applied in lexical order.
//
//go:embed *.sql
var Files embed.FS
