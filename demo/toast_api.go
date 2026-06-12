package demo

import (
	"net/http"
	"strconv"

	"github.com/markopolo123/baud-ui/baud"
)

// toastDemos fixes the demo copy per tone so each round-trip produces
// distinct, assertable content (tabs_api.go precedent).
var toastDemos = map[string]baud.ToastProps{
	"ok":   {Tone: "ok", Title: "Deployed ingest-gw v2.14.1", Body: "rollout complete on 12 pods"},
	"err":  {Tone: "err", Title: "Connection lost", Body: "eu-west-1 etcd quorum unreachable"},
	"warn": {Tone: "warn", Title: "p99 latency above SLO", Body: "ingest-gw 412ms for 5m"},
	"info": {Tone: "info", Title: "Maintenance window", Body: "eu-west-1 rotates certs at 02:00"},
}

// handleToast serves GET /demo/toast?tone=…[&ms=…] — the htmx OOB toast
// push. The sheet's trigger buttons use hx-swap="none": the response's
// only payload is the hx-swap-oob fragment htmx appends into the
// page-level #toasts region. ms overrides the ~4s auto-dismiss interval
// (the e2e suite uses a short one to assert a real auto-dismiss without
// multi-second waits).
func handleToast(w http.ResponseWriter, r *http.Request) {
	p, ok := toastDemos[r.URL.Query().Get("tone")]
	if !ok {
		http.Error(w, "unknown tone", http.StatusBadRequest)
		return
	}
	if msStr := r.URL.Query().Get("ms"); msStr != "" {
		ms, err := strconv.Atoi(msStr)
		if err != nil || ms < 1 {
			http.Error(w, "bad ms", http.StatusBadRequest)
			return
		}
		p.DismissMs = ms
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	baud.ToastOOB(p).Render(r.Context(), w)
}
