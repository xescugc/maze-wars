package assets

import (
	"embed"
)

// Assets defines the embedded files
//
//go:embed css/* js/* wasm/*
var Assets embed.FS
