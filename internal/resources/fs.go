package resources

import "embed"

//go:embed tiles sprites fonts levels music
var EmbeddedFS embed.FS
