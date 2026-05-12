package templates

import (
	"embed"
)

//go:embed docker/*.yaml config/*.yaml
var FS embed.FS
