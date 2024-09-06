package deploy

import (
	"embed"
)

//go:embed common kind frontend
var DeployFiles embed.FS
