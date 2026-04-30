package api

import "net/http"

type Dependencies struct {
	TokenPlaintext string
}

func NewRouter(deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	registerHealthRoutes(mux)
	registerSessionRoutes(mux, deps)
	registerDiskRoutes(mux)
	registerEventRoutes(mux)
	registerTopologyRoutes(mux)
	registerSettingsRoutes(mux)
	registerStaticRoutes(mux)

	return mux
}
