package schemas

import "embed"

// Files contains every released schema so a single obsctl binary can validate
// manifests from all supported API versions without network access.
//
//go:embed v1/*.json v2/*.json
var Files embed.FS
