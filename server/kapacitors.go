package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/influxdata/chronograf"
)

type postKapacitorRequest struct {
	Name     *string `json:"name"`               // User facing name of kapacitor instance.; Required: true
	URL      *string `json:"url"`                // URL for the kapacitor backend (e.g. http://localhost:9092);/ Required: true
	Username string  `json:"username,omitempty"` // Username for authentication to kapacitor
	Password string  `json:"password,omitempty"`
}

func (p *postKapacitorRequest) Valid() error {
	if p.Name == nil || p.URL == nil {
		return fmt.Errorf("name and url required")
	}

	url, err := url.ParseRequestURI(*p.URL)
	if err != nil {
		return fmt.Errorf("invalid source URI: %v", err)
	}
	if len(url.Scheme) == 0 {
		return fmt.Errorf("Invalid URL; no URL scheme defined")
	}

	return nil
}

type kapaLinks struct {
	Proxy string `json:"proxy"` // URL location of proxy endpoint for this source
	Self  string `json:"self"`  // Self link mapping to this resource
}

type kapacitor struct {
	ID       string    `json:"id,string"`          // Unique identifier representing a kapacitor instance.
	Name     string    `json:"name"`               // User facing name of kapacitor instance.
	URL      string    `json:"url"`                // URL for the kapacitor backend (e.g. http://localhost:9092)
	Username string    `json:"username,omitempty"` // Username for authentication to kapacitor
	Password string    `json:"password,omitempty"`
	Links    kapaLinks `json:"links"` // Links are URI locations related to kapacitor
}

// NewKapacitor adds valid kapacitor store store.
func (h *Service) NewKapacitor(w http.ResponseWriter, r *http.Request) {
	srcID, err := paramID("id", r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	ctx := r.Context()
	_, err = h.SourcesStore.Get(ctx, srcID)
	if err != nil {
		notFound(w, srcID)
		return
	}

	var req postKapacitorRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		invalidJSON(w)
		return
	}
	if req.Valid() != nil {
		invalidData(w, err)
		return
	}

	srv := chronograf.Server{
		SrcID:    srcID,
		Name:     *req.Name,
		Username: req.Username,
		Password: req.Password,
		URL:      *req.URL,
	}

	if srv, err = h.ServersStore.Add(ctx, srv); err != nil {
		msg := fmt.Errorf("Error storing kapacitor %v: %v", req, err)
		unknownErrorWithMessage(w, msg)
		return
	}

	res := newKapacitor(srv)
	w.Header().Add("Location", res.Links.Self)
	encodeJSON(w, http.StatusCreated, res, h.Logger)
}

func newKapacitor(srv chronograf.Server) kapacitor {
	httpAPISrcs := "/chronograf/v1/sources"
	return kapacitor{
		ID:       strconv.Itoa(srv.ID),
		Name:     srv.Name,
		Username: srv.Username,
		Password: srv.Password,
		URL:      srv.URL,
		Links: kapaLinks{
			Self:  fmt.Sprintf("%s/%d/kapacitors/%d", httpAPISrcs, srv.SrcID, srv.ID),
			Proxy: fmt.Sprintf("%s/%d/kapacitors/%d/proxy", httpAPISrcs, srv.SrcID, srv.ID),
		},
	}
}

type kapacitors struct {
	Kapacitors []kapacitor `json:"kapacitors"`
}

// Kapacitors retrieves all kapacitors from store.
func (h *Service) Kapacitors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	mrSrvs, err := h.ServersStore.All(ctx)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Error loading kapacitors")
		return
	}

	srvs := make([]kapacitor, len(mrSrvs))
	for i, srv := range mrSrvs {
		srvs[i] = newKapacitor(srv)
	}

	res := kapacitors{
		Kapacitors: srvs,
	}

	encodeJSON(w, http.StatusOK, res, h.Logger)
}

// KapacitorsID retrieves a kapacitor with ID from store.
func (h *Service) KapacitorsID(w http.ResponseWriter, r *http.Request) {
	id, err := paramID("kid", r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	srcID, err := paramID("id", r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	ctx := r.Context()
	srv, err := h.ServersStore.Get(ctx, id)
	if err != nil || srv.SrcID != srcID {
		notFound(w, id)
		return
	}

	res := newKapacitor(srv)
	encodeJSON(w, http.StatusOK, res, h.Logger)
}

// RemoveKapacitor deletes kapacitor from store.
func (h *Service) RemoveKapacitor(w http.ResponseWriter, r *http.Request) {
	id, err := paramID("kid", r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	srcID, err := paramID("id", r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	ctx := r.Context()
	srv, err := h.ServersStore.Get(ctx, id)
	if err != nil || srv.SrcID != srcID {
		notFound(w, id)
		return
	}

	if err = h.ServersStore.Delete(ctx, srv); err != nil {
		unknownErrorWithMessage(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type patchKapacitorRequest struct {
	Name     *string `json:"name,omitempty"`     // User facing name of kapacitor instance.
	URL      *string `json:"url,omitempty"`      // URL for the kapacitor
	Username *string `json:"username,omitempty"` // Username for kapacitor auth
	Password *string `json:"password,omitempty"`
}

func (p *patchKapacitorRequest) Valid() error {
	if p.URL != nil {
		url, err := url.ParseRequestURI(*p.URL)
		if err != nil {
			return fmt.Errorf("invalid source URI: %v", err)
		}
		if len(url.Scheme) == 0 {
			return fmt.Errorf("Invalid URL; no URL scheme defined")
		}
	}
	return nil
}

// UpdateKapacitor incrementally updates a kapacitor definition in the store
func (h *Service) UpdateKapacitor(w http.ResponseWriter, r *http.Request) {
	id, err := paramID("kid", r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	srcID, err := paramID("id", r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	ctx := r.Context()
	srv, err := h.ServersStore.Get(ctx, id)
	if err != nil || srv.SrcID != srcID {
		notFound(w, id)
		return
	}

	var req patchKapacitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		invalidJSON(w)
		return
	}

	if err := req.Valid(); err != nil {
		invalidData(w, err)
		return
	}

	if req.Name != nil {
		srv.Name = *req.Name
	}
	if req.URL != nil {
		srv.URL = *req.URL
	}
	if req.Password != nil {
		srv.Password = *req.Password
	}
	if req.Username != nil {
		srv.Username = *req.Username
	}

	if err := h.ServersStore.Update(ctx, srv); err != nil {
		msg := fmt.Sprintf("Error updating kapacitor ID %d", id)
		Error(w, http.StatusInternalServerError, msg)
		return
	}

	res := newKapacitor(srv)
	encodeJSON(w, http.StatusOK, res, h.Logger)
}
