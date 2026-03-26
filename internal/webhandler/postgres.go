package webhandler

import (
	"net/http"
	"strconv"

	pb "db-router/proto/dbrouter"

	"google.golang.org/protobuf/types/known/structpb"
)

func (h *Handler) handlePGListDatabases(w http.ResponseWriter, r *http.Request) {
	resp, err := h.pg.ListDatabases(r.Context(), &pb.ListDatabasesRequest{})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"databases": resp.Databases})
}

func (h *Handler) handlePGCreateDatabase(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := h.decodeBody(r, &req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	resp, err := h.pg.CreateDatabase(r.Context(), &pb.CreateDatabaseRequest{Name: req.Name})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"name":    resp.Name,
		"message": resp.Message,
	})
}

func (h *Handler) handlePGListTables(w http.ResponseWriter, r *http.Request) {
	db := r.URL.Query().Get("db")
	resp, err := h.pg.ListTables(r.Context(), &pb.ListTablesRequest{Database: db})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"database": db, "tables": resp.Tables})
}

func (h *Handler) handlePGQuery(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query    string `json:"query"`
		Database string `json:"database"`
	}
	if err := h.decodeBody(r, &req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	resp, err := h.pg.ExecuteQuery(r.Context(), &pb.ExecuteQueryRequest{
		Query:    req.Query,
		Database: req.Database,
	})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}

	if len(resp.Rows) > 0 {
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"columns": resp.Columns,
			"rows":    protoRowsToMaps(resp.Rows),
			"count":   resp.Count,
		})
	} else {
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"rows_affected": resp.RowsAffected,
			"message":       resp.Message,
		})
	}
}

func (h *Handler) handlePGSelect(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit == 0 {
		limit = 100
	}

	resp, err := h.pg.SelectData(r.Context(), &pb.SelectDataRequest{
		Database: q.Get("db"),
		Table:    q.Get("table"),
		Limit:    int32(limit),
	})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"database": resp.Database,
		"table":    resp.Table,
		"data":     protoRowsToMaps(resp.Data),
		"count":    resp.Count,
	})
}

func (h *Handler) handlePGInsert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Database string                 `json:"database"`
		Table    string                 `json:"table"`
		Data     map[string]interface{} `json:"data"`
	}
	if err := h.decodeBody(r, &req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	pbData, err := toStructpbValueMap(req.Data)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	resp, err := h.pg.InsertData(r.Context(), &pb.InsertDataRequest{
		Database: req.Database,
		Table:    req.Table,
		Data:     pbData,
	})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"database":    resp.Database,
		"table":       resp.Table,
		"inserted_id": resp.InsertedId,
		"message":     resp.Message,
	})
}

func (h *Handler) handlePGUpdate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Database string                 `json:"database"`
		Table    string                 `json:"table"`
		ID       string                 `json:"id"`
		Data     map[string]interface{} `json:"data"`
	}
	if err := h.decodeBody(r, &req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	pbData, err := toStructpbValueMap(req.Data)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	resp, err := h.pg.UpdateData(r.Context(), &pb.UpdateDataRequest{
		Database: req.Database,
		Table:    req.Table,
		Id:       req.ID,
		Data:     pbData,
	})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"database":      resp.Database,
		"table":         resp.Table,
		"id":            resp.Id,
		"rows_affected": resp.RowsAffected,
	})
}

func (h *Handler) handlePGDelete(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	resp, err := h.pg.DeleteData(r.Context(), &pb.DeleteDataRequest{
		Database: q.Get("db"),
		Table:    q.Get("table"),
		Id:       q.Get("id"),
	})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"database":      resp.Database,
		"table":         resp.Table,
		"id":            resp.Id,
		"rows_affected": resp.RowsAffected,
	})
}

// ── proto → JSON helpers ──────────────────────────────────────────────────────

func protoRowsToMaps(rows []*pb.QueryResultRow) []map[string]interface{} {
	out := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		m := make(map[string]interface{}, len(row.Fields))
		for k, v := range row.Fields {
			m[k] = v.AsInterface()
		}
		out[i] = m
	}
	return out
}

func toStructpbValueMap(src map[string]interface{}) (map[string]*structpb.Value, error) {
	m := make(map[string]*structpb.Value, len(src))
	for k, v := range src {
		pv, err := structpb.NewValue(v)
		if err != nil {
			return nil, err
		}
		m[k] = pv
	}
	return m, nil
}
