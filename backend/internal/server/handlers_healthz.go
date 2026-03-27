package server

import (
	"encoding/json"
	"net/http"
)

type HealthCheckResponse struct {
	Version string `json:"version"`
	Status  string `json:"status"`
}

func HandleHealthCheck(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := HealthCheckResponse{
			Version: s.Version(),
			Status:  "OK",
		}

		b, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}
}
