package ui

import "embed"

// DistFS holds the embedded production build of the UI.
//
//go:embed dist
var DistFS embed.FS
