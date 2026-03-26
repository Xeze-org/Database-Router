package webhandler

import (
	"net/http"

	pb "db-router/proto/dbrouter"
)

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp, err := h.health.Check(ctx, &pb.HealthCheckRequest{})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}

	format := func(s *pb.ConnectionStatus) map[string]interface{} {
		if s == nil {
			return map[string]interface{}{"status": "disabled"}
		}
		m := map[string]interface{}{"status": s.Status}
		if s.Host != "" {
			m["host"] = s.Host
		}
		if s.Database != "" {
			m["database"] = s.Database
		}
		if s.Port != "" {
			m["port"] = s.Port
		}
		if s.Error != "" {
			m["error"] = s.Error
		}
		return m
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"postgres": format(resp.Postgres),
		"mongo":    format(resp.Mongo),
		"redis":    format(resp.Redis),
	})
}
