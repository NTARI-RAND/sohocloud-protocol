package httpjson

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/NTARI-RAND/sohocloud-protocol/coordinator"
	"github.com/NTARI-RAND/sohocloud-protocol/identity"
)

// Handler adapts a coordinator.Coordinator to an http.Handler speaking the
// reference HTTP+JSON transport.
//
// SPIFFE identity binding is expected to be applied by middleware in front of
// this handler — SoHoLINK already enforces the canonical /node/<id> binding
// (identity.BindsTo) for these operations at its HTTP layer. This reference
// handler wires routes and JSON only; it is not the transport-layer
// authenticator, and it deliberately does not re-implement one.
func Handler(c coordinator.Coordinator) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/v0/listing", post(c.SubmitListing))
	mux.Handle("/v0/heartbeat", post(c.Heartbeat))
	mux.Handle("/v0/decline", post(c.Decline))
	mux.Handle("/v0/report", post(c.ReportJob))

	mux.HandleFunc("/v0/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		id := identity.NodeID(r.URL.Query().Get("node_id"))
		if id == "" {
			http.Error(w, "missing node_id", http.StatusBadRequest)
			return
		}
		out, err := c.PollJobs(r.Context(), id)
		if err != nil {
			fail(w, err)
			return
		}
		writeJSON(w, out)
	})

	mux.HandleFunc("/v0/fees", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		out, err := c.Fees(r.Context())
		if err != nil {
			fail(w, err)
			return
		}
		writeJSON(w, out)
	})

	return mux
}

// post builds a POST handler that decodes a JSON body of type T and forwards it
// to call. It compiles against any coordinator method of the shape
// func(context.Context, T) error, so the concrete message type is inferred and
// need not be named here.
func post[T any](call func(context.Context, T) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		var v T
		if !decode(w, r, &v) {
			return
		}
		if err := call(r.Context(), v); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
