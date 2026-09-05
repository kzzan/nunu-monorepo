package docs

import _ "embed"

// SwaggerJSON is served directly because gin-swagger currently uses the v1
// registry while generated docs are registered through swag/v2.
//
//go:embed swagger.json
var SwaggerJSON []byte
