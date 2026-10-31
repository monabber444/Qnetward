package main

import (
	"net/http"
)

func handleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
		return
	}

	if appConnector == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"agents": []any{},
			"total":  0,
		})
		return
	}

	agents := appConnector.ListAgents(0)
	writeJSON(w, http.StatusOK, map[string]any{
		"agents": agents,
		"total":  len(agents),
	})
}
