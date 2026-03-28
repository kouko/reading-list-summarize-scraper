// Package embed provides embedded static assets for the rlss CLI.
// Go's embed directive requires paths relative to the source file,
// so this package lives alongside the embedded files.
package embed

import _ "embed"

//go:embed defuddle.min.js
var DefuddleJS string

//go:embed extension/manifest.json
var ExtensionManifest []byte

//go:embed extension/background.js
var ExtensionBackground []byte
