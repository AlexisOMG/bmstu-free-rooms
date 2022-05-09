package migrations

import (
	"embed"
	"io/fs"
)

//go:embed scripts/*
var migrationsAssets embed.FS

var MigrationAssets, _ = fs.Sub(migrationsAssets, "scripts")
