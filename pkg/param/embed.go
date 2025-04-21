package param

import (
	"embed"
)

//go:embed defaults/*
var Defaults embed.FS

//go:embed scaffolds/*
var Scaffolds embed.FS
