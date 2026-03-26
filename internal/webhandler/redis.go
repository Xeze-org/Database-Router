package webhandler

import (
	"net/http"

	pb "db-router/proto/dbrouter"
)

func (h *Handler) handleRedisListKeys(w http.ResponseWriter, r *http.Request) {
	pattern := r.URL.Query().Get("pattern")
	if pattern == "" {
		pattern = "*"
	}
	resp, err := h.redis.ListKeys(r.Context(), &pb.ListKeysRequest{Pattern: pattern})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"keys": resp.Keys, "count": len(resp.Keys)})
}

func (h *Handler) handleRedisGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	resp, err := h.redis.GetValue(r.Context(), &pb.GetValueRequest{Key: key})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"key": resp.Key, "value": resp.Value, "ttl": resp.Ttl})
}

func (h *Handler) handleRedisSet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		TTL   int32  `json:"ttl"`
	}
	if err := h.decodeBody(r, &req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	resp, err := h.redis.SetValue(r.Context(), &pb.SetValueRequest{
		Key:   req.Key,
		Value: req.Value,
		Ttl:   req.TTL,
	})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"key": resp.Key, "value": resp.Value, "ttl": resp.Ttl})
}

func (h *Handler) handleRedisDelete(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	resp, err := h.redis.DeleteKey(r.Context(), &pb.DeleteKeyRequest{Key: key})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"key": resp.Key, "deleted": resp.Deleted})
}

func (h *Handler) handleRedisInfo(w http.ResponseWriter, r *http.Request) {
	resp, err := h.redis.Info(r.Context(), &pb.RedisInfoRequest{})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"db_size": resp.DbSize, "info": resp.Info})
}
