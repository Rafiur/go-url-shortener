// Package web embeds the static frontend into the binary so a single artifact
// serves both the API and the UI.
package web

import _ "embed"

//go:embed index.html
var Index []byte
