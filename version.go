// Package dscrd exposes build metadata shared by the CLI and release tooling.
package dscrd

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var rawVersion string

// Version is the current dscrd version, sourced from the VERSION file.
var Version = strings.TrimSpace(rawVersion)
