package webassets

import "embed"

// Dist contains the built frontend bundle.
//
//go:embed all:dist
var Dist embed.FS
