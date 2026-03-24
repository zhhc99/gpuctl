package ipc

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
)

type Handler interface {
	HandleLoad() LoadResponse
	HandleList(ListRequest) ListResponse
	HandleTuneGet(TuneGetRequest) TuneGetResponse
	HandleTuneSet(TuneSetRequest) TuneSetResponse
	HandleTuneReset(TuneResetRequest) TuneResetResponse
	HandleVersion() VersionResponse
}

type Server struct {
	l    net.Listener
	http *http.Server
	once sync.Once
}

func NewServer(h Handler) (*Server, error) {
	l, err := listen()
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	route := func(path string, fn func(w http.ResponseWriter, r *http.Request)) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fn(w, r)
		})
	}

	badRequest := func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(struct {
			Err string `json:"error"`
		}{Err: "invalid request body"})
	}

	route("/load", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(h.HandleLoad())
	})

	route("/version", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(h.HandleVersion())
	})

	route("/list", func(w http.ResponseWriter, r *http.Request) {
		var req ListRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			badRequest(w)
			return
		}
		_ = json.NewEncoder(w).Encode(h.HandleList(req))
	})

	route("/tune/get", func(w http.ResponseWriter, r *http.Request) {
		var req TuneGetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			badRequest(w)
			return
		}
		_ = json.NewEncoder(w).Encode(h.HandleTuneGet(req))
	})

	route("/tune/set", func(w http.ResponseWriter, r *http.Request) {
		var req TuneSetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			badRequest(w)
			return
		}
		_ = json.NewEncoder(w).Encode(h.HandleTuneSet(req))
	})

	route("/tune/reset", func(w http.ResponseWriter, r *http.Request) {
		var req TuneResetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			badRequest(w)
			return
		}
		_ = json.NewEncoder(w).Encode(h.HandleTuneReset(req))
	})

	return &Server{
		l:    l,
		http: &http.Server{Handler: mux},
	}, nil
}

func (s *Server) Serve() { _ = s.http.Serve(s.l) }

func (s *Server) Close() {
	s.once.Do(func() { _ = s.http.Close() })
}
