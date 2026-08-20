package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/example/multitenant-search/internal/collection"
	"github.com/example/multitenant-search/internal/document"
	"github.com/example/multitenant-search/internal/platform"
	"github.com/example/multitenant-search/internal/query"
	"github.com/example/multitenant-search/internal/snapshot"
	"github.com/example/multitenant-search/internal/tenant"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	store       *platform.Store
	tenants     *tenant.Service
	collections *collection.Service
	docs        *document.Service
	snaps       *snapshot.Service
	logger      *platform.Logger
	started     time.Time
}

func NewServer(store *platform.Store, l *platform.Logger) *Server {
	return &Server{store: store, tenants: tenant.NewService(store, platform.RealClock{}), collections: collection.NewService(store, platform.RealClock{}), docs: document.NewService(store, platform.RealClock{}), snaps: snapshot.New(store, "./data/snapshots"), logger: l, started: time.Now()}
}
func (s *Server) Routes() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/healthz", s.health)
	m.HandleFunc("/readyz", s.ready)
	m.HandleFunc("/metrics", s.metrics)
	m.HandleFunc("/v1/tenants", s.tenantsHandler)
	m.HandleFunc("/v1/collections", s.collectionsHandler)
	m.HandleFunc("/v1/documents", s.documentsHandler)
	m.HandleFunc("/v1/search", s.searchHandler)
	m.HandleFunc("/v1/snapshots", s.snapshotHandler)
	return requestMiddleware(bodyLimitMiddleware(m))
}
func requestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := platform.NewID("req")
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), "request_id", id)
		defer func() {
			if x := recover(); x != nil {
				http.Error(w, `{"error":"internal"}`, 500)
			}
		}()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func bodyLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4<<20)
		next.ServeHTTP(w, r)
	})
}
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	write(w, 200, map[string]any{"status": "ok", "uptime": time.Since(s.started).String()})
}
func (s *Server) ready(w http.ResponseWriter, _ *http.Request) {
	write(w, 200, map[string]string{"status": "ready"})
}
func (s *Server) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "search_uptime_seconds %f\n", time.Since(s.started).Seconds())
}
func (s *Server) tenantsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		write(w, 200, s.store.ListTenants(r.Context()))
		return
	}
	var in struct {
		Name  string `json:"name"`
		Quota int    `json:"quota"`
	}
	if !decode(r, &in, w) {
		return
	}
	t, e := s.tenants.Create(r.Context(), in.Name, in.Quota)
	if e != nil {
		errorJSON(w, e)
		return
	}
	write(w, 201, t)
}
func (s *Server) collectionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tenantID := r.URL.Query().Get("tenant_id")
		write(w, 200, s.store.ListCollections(r.Context(), tenantID))
		return
	}
	var in struct {
		TenantID string                           `json:"tenant_id"`
		Name     string                           `json:"name"`
		Mappings map[string]platform.FieldMapping `json:"mappings"`
		Shards   int                              `json:"shards"`
		Replicas int                              `json:"replicas"`
	}
	if !decode(r, &in, w) {
		return
	}
	c, e := s.collections.Create(r.Context(), in.TenantID, in.Name, in.Mappings, in.Shards, in.Replicas)
	if e != nil {
		errorJSON(w, e)
		return
	}
	write(w, 201, c)
}
func (s *Server) documentsHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if r.Method == http.MethodDelete && len(parts) > 2 {
		if e := s.docs.Delete(r.Context(), parts[2]); e != nil {
			errorJSON(w, e)
			return
		}
		write(w, 200, map[string]string{"status": "deleted"})
		return
	}
	if r.Method == http.MethodGet {
		write(w, 200, s.docs.List(r.Context(), r.URL.Query().Get("tenant_id"), r.URL.Query().Get("collection_id")))
		return
	}
	var d platform.Document
	if !decode(r, &d, w) {
		return
	}
	if d.ID == "" {
		d.ID = platform.NewID("doc")
	}
	if e := s.docs.Put(r.Context(), &d); e != nil {
		errorJSON(w, e)
		return
	}
	write(w, 201, d)
}
func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	var req query.Request
	if !decode(r, &req, w) {
		return
	}
	col := r.URL.Query().Get("collection_id")
	eng := query.New(s.docs.EnsureIndex(col), 10000)
	res, e := eng.Search(r.Context(), req)
	if e != nil {
		errorJSON(w, e)
		return
	}
	write(w, 200, res)
}
func (s *Server) snapshotHandler(w http.ResponseWriter, r *http.Request) {
	p, e := s.snaps.Create(r.Context())
	if e != nil {
		errorJSON(w, e)
		return
	}
	write(w, 202, map[string]string{"path": p})
}
func decode(r *http.Request, v any, w http.ResponseWriter) bool {
	b, e := io.ReadAll(r.Body)
	if e == nil {
		e = json.Unmarshal(b, v)
	}
	if e != nil {
		errorJSON(w, fmt.Errorf("decode body: %w", e))
		return false
	}
	return true
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func errorJSON(w http.ResponseWriter, e error) {
	status := 400
	if strings.Contains(e.Error(), "not found") {
		status = 404
	}
	if strings.Contains(e.Error(), "quota") {
		status = 429
	}
	if strings.Contains(e.Error(), "conflict") {
		status = 409
	}
	write(w, status, map[string]any{"error": e.Error()})
}

var _ = slog.Default
