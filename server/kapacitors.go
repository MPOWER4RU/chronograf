package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/bouk/httprouter"
	"github.com/influxdata/chronograf"
	kapa "github.com/influxdata/chronograf/kapacitor"
	"github.com/influxdata/chronograf/uuid"
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
	Rules string `json:"rules"` // Riles link for defining roles alerts for kapacitor
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
	if err := req.Valid(); err != nil {
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
			Rules: fmt.Sprintf("%s/%d/kapacitors/%d/rules", httpAPISrcs, srv.SrcID, srv.ID),
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

// KapacitorRulesPost proxies POST to kapacitor
func (h *Service) KapacitorRulesPost(w http.ResponseWriter, r *http.Request) {
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

	c := kapa.Client{
		URL:      srv.URL,
		Username: srv.Username,
		Password: srv.Password,
		Ticker:   &kapa.Alert{},
		ID:       &uuid.V4{},
	}

	var req chronograf.AlertRule
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		invalidJSON(w)
		return
	}
	// TODO: validate this data
	/*
		if err := req.Valid(); err != nil {
			invalidData(w, err)
			return
		}
	*/

	task, err := c.Create(ctx, req)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	req.ID = task.ID
	rule, err := h.AlertRulesStore.Add(ctx, srcID, id, req)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	res := alertResponse{
		AlertRule: rule,
		Links: alertLinks{
			Self:      fmt.Sprintf("/chronograf/v1/sources/%d/kapacitors/%d/rules/%s", srv.SrcID, srv.ID, req.ID),
			Kapacitor: fmt.Sprintf("/chronograf/v1/sources/%d/kapacitors/%d/proxy?path=%s", srv.SrcID, srv.ID, url.QueryEscape(task.Href)),
		},
		TICKScript: string(task.TICKScript),
	}

	w.Header().Add("Location", res.Links.Self)
	encodeJSON(w, http.StatusCreated, res, h.Logger)
}

type alertLinks struct {
	Self      string `json:"self"`
	Kapacitor string `json:"kapacitor"`
}

type alertResponse struct {
	chronograf.AlertRule
	TICKScript string     `json:"tickscript"`
	Links      alertLinks `json:"links"`
}

// KapacitorRulesPut proxies PATCH to kapacitor
func (h *Service) KapacitorRulesPut(w http.ResponseWriter, r *http.Request) {
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

	tid := httprouter.GetParamFromContext(ctx, "tid")
	c := kapa.Client{
		URL:      srv.URL,
		Username: srv.Username,
		Password: srv.Password,
		Ticker:   &kapa.Alert{},
	}
	var req chronograf.AlertRule
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		invalidJSON(w)
		return
	}
	// TODO: validate this data
	/*
		if err := req.Valid(); err != nil {
			invalidData(w, err)
			return
		}
	*/

	// Check if the rule exists and is scoped correctly
	if _, err := h.AlertRulesStore.Get(ctx, srcID, id, tid); err != nil {
		if err == chronograf.ErrAlertNotFound {
			notFound(w, id)
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	req.ID = tid
	task, err := c.Update(ctx, c.Href(tid), req)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.AlertRulesStore.Update(ctx, srcID, id, req); err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	res := alertResponse{
		AlertRule: req,
		Links: alertLinks{
			Self:      fmt.Sprintf("/chronograf/v1/sources/%d/kapacitors/%d/rules/%s", srv.SrcID, srv.ID, req.ID),
			Kapacitor: fmt.Sprintf("/chronograf/v1/sources/%d/kapacitors/%d/proxy?path=%s", srv.SrcID, srv.ID, url.QueryEscape(task.Href)),
		},
		TICKScript: string(task.TICKScript),
	}
	encodeJSON(w, http.StatusOK, res, h.Logger)
}

// KapacitorRulesGet retrieves all rules
func (h *Service) KapacitorRulesGet(w http.ResponseWriter, r *http.Request) {
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

	rules, err := h.AlertRulesStore.All(ctx, srcID, id)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	ticker := &kapa.Alert{}
	c := kapa.Client{}
	res := allAlertsResponse{
		Rules: []alertResponse{},
	}
	for _, rule := range rules {
		tickscript, err := ticker.Generate(rule)
		if err != nil {
			Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		ar := alertResponse{
			AlertRule: rule,
			Links: alertLinks{
				Self:      fmt.Sprintf("/chronograf/v1/sources/%d/kapacitors/%d/rules/%s", srv.SrcID, srv.ID, rule.ID),
				Kapacitor: fmt.Sprintf("/chronograf/v1/sources/%d/kapacitors/%d/proxy?path=%s", srv.SrcID, srv.ID, url.QueryEscape(c.Href(rule.ID))),
			},
			TICKScript: string(tickscript),
		}
		res.Rules = append(res.Rules, ar)
	}
	encodeJSON(w, http.StatusOK, res, h.Logger)
}

type allAlertsResponse struct {
	Rules []alertResponse `json:"rules"`
}

// KapacitorRulesGet retrieves specific task
func (h *Service) KapacitorRulesID(w http.ResponseWriter, r *http.Request) {
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
	tid := httprouter.GetParamFromContext(ctx, "tid")
	// Check if the rule exists within scope
	rule, err := h.AlertRulesStore.Get(ctx, srcID, id, tid)
	if err != nil {
		if err == chronograf.ErrAlertNotFound {
			notFound(w, id)
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	ticker := &kapa.Alert{}
	c := kapa.Client{}
	tickscript, err := ticker.Generate(rule)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	res := alertResponse{
		AlertRule: rule,
		Links: alertLinks{
			Self:      fmt.Sprintf("/chronograf/v1/sources/%d/kapacitors/%d/rules/%s", srv.SrcID, srv.ID, rule.ID),
			Kapacitor: fmt.Sprintf("/chronograf/v1/sources/%d/kapacitors/%d/proxy?path=%s", srv.SrcID, srv.ID, url.QueryEscape(c.Href(rule.ID))),
		},
		TICKScript: string(tickscript),
	}
	encodeJSON(w, http.StatusOK, res, h.Logger)
}

// KapacitosRulesDelete proxies DELETE to kapacitor
func (h *Service) KapacitorRulesDelete(w http.ResponseWriter, r *http.Request) {
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

	tid := httprouter.GetParamFromContext(ctx, "tid")

	// Check if the rule is linked to this server and kapacitor
	if _, err := h.AlertRulesStore.Get(ctx, srcID, id, tid); err != nil {
		if err == chronograf.ErrAlertNotFound {
			notFound(w, id)
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	c := kapa.Client{
		URL:      srv.URL,
		Username: srv.Username,
		Password: srv.Password,
	}
	if err := c.Delete(ctx, c.Href(tid)); err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.AlertRulesStore.Delete(ctx, srcID, id, chronograf.AlertRule{ID: tid}); err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
