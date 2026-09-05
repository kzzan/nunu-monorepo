package server

import (
	"github.com/samber/do/v2"
)

// Package registers all server-layer providers into the injector.
// Providers are lazy: only the servers invoked by the composition root are built.
// pkg/server/http must NOT be registered here: NewHTTPServer and
// httpx.NewServer provide the same *httpx.Server type, which would create
// a circular dependency.
var Package = do.Package(
	do.Lazy(NewHTTPServer),
	do.Lazy(NewJobServer),
	do.Lazy(NewTaskServer),
	do.Lazy(NewMigrateServer),
)
