package api

import "net/http"

func registerTopologyRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.Handle("GET /api/topology", requireSession(http.HandlerFunc(topologyHandler(deps))))
}

func topologyHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		graph, err := deps.Runtime.Topology(r.Context())
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "failed to build topology")
			return
		}

		writeJSON(w, http.StatusOK, graph)
	}
}
