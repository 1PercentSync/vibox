package static

import "embed"

// StaticFiles embeds the frontend build output.
// This uses Go 1.16+ embed feature to include static files in the binary.
// The frontend must be built first using: cd frontend && npm run build
//
// Directory structure:
//   dist/
//   ├── index.html
//   └── assets/
//       ├── *.js
//       └── *.css
//
//go:embed dist
var StaticFiles embed.FS
