package api

import "net/http"

type Dependencies struct{}

func NewRouter(deps Dependencies) http.Handler {
	_ = deps

	mux := http.NewServeMux()
	registerHealthRoutes(mux)

	return mux
}
