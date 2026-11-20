// Package migrations embeds the goose migration files so the migrate command
// and its tests do not depend on the process working directory.
package migrations

import "embed"

// FS contains every goose migration.
//
//go:embed *.sql
var FS embed.FS
