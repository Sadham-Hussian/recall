package migrations

import "embed"

// FS embeds all SQL migration files so migrations can run from any working directory.
//
//go:embed *.sql
var FS embed.FS
