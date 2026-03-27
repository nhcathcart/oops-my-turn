package schemata

import "embed"

//go:embed migrations
var MigrationsFolder embed.FS

const RootFolder = "migrations"
