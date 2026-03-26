package webhandler

import (
	"net/http"

	pb "db-router/proto/dbrouter"

	"google.golang.org/protobuf/types/known/structpb"
)

func (h *Handler) handleMongoListDatabases(w http.ResponseWriter, r *http.Request) {
	resp, err := h.mongo.ListDatabases(r.Context(), &pb.ListMongoDatabasesRequest{})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"databases": resp.Databases})
}

func (h *Handler) handleMongoListCollections(w http.ResponseWriter, r *http.Request) {
	db := r.URL.Query().Get("db")
	resp, err := h.mongo.ListCollections(r.Context(), &pb.ListCollectionsRequest{Database: db})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"database": db, "collections": resp.Collections})
}

func (h *Handler) handleMongoFind(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	resp, err := h.mongo.FindDocuments(r.Context(), &pb.FindDocumentsRequest{
		Database:   q.Get("db"),
		Collection: q.Get("col"),
	})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"database":   resp.Database,
		"collection": resp.Collection,
		"documents":  mongoDocsToMaps(resp.Documents),
		"count":      resp.Count,
	})
}

func (h *Handler) handleMongoInsert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Database   string                 `json:"database"`
		Collection string                 `json:"collection"`
		Document   map[string]interface{} `json:"document"`
	}
	if err := h.decodeBody(r, &req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	pbDoc, err := structpb.NewStruct(req.Document)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	resp, err := h.mongo.InsertDocument(r.Context(), &pb.InsertDocumentRequest{
		Database:   req.Database,
		Collection: req.Collection,
		Document:   pbDoc,
	})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"database":    resp.Database,
		"collection":  resp.Collection,
		"inserted_id": resp.InsertedId,
	})
}

func (h *Handler) handleMongoUpdate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Database   string                 `json:"database"`
		Collection string                 `json:"collection"`
		ID         string                 `json:"id"`
		Update     map[string]interface{} `json:"update"`
	}
	if err := h.decodeBody(r, &req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	pbUpdate, err := structpb.NewStruct(req.Update)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	resp, err := h.mongo.UpdateDocument(r.Context(), &pb.UpdateDocumentRequest{
		Database:   req.Database,
		Collection: req.Collection,
		Id:         req.ID,
		Update:     pbUpdate,
	})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"database":       resp.Database,
		"collection":     resp.Collection,
		"matched_count":  resp.MatchedCount,
		"modified_count": resp.ModifiedCount,
	})
}

func (h *Handler) handleMongoDelete(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	resp, err := h.mongo.DeleteDocument(r.Context(), &pb.DeleteDocumentRequest{
		Database:   q.Get("db"),
		Collection: q.Get("col"),
		Id:         q.Get("id"),
	})
	if err != nil {
		h.writeGRPCError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"database":      resp.Database,
		"collection":    resp.Collection,
		"deleted_count": resp.DeletedCount,
	})
}

func mongoDocsToMaps(docs []*pb.MongoDocument) []map[string]interface{} {
	out := make([]map[string]interface{}, len(docs))
	for i, doc := range docs {
		m := make(map[string]interface{}, len(doc.Fields))
		for k, v := range doc.Fields {
			m[k] = v.AsInterface()
		}
		out[i] = m
	}
	return out
}
