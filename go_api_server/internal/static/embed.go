// Package static embeds the React frontend build output into the Go binary.
package static

import (
	"embed"
	"io/fs"
)

//go:embed all:web_dist
var embeddedFiles embed.FS

// FS returns a filesystem rooted at the web_dist directory.
// This strips the "web_dist" prefix so files are served from root.
func FS() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "web_dist")
}
