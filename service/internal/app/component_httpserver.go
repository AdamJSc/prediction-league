package app

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"prediction-league/service/internal/domain"
	"time"
)

// httpServer represents a collection of functions for starting and running a web server
type httpServer struct {
	srv     *http.Server
	addrC   chan *net.TCPAddr
	tcpAddr *net.TCPAddr
	l       domain.Logger
}

// Run will start the server and listen for requests
func (h *httpServer) Run(_ context.Context) error {
	addr := h.srv.Addr
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	h.addrC <- lis.Addr().(*net.TCPAddr)
	h.l.Infof("serving http on http://%s", h.addr().String())
	return h.srv.Serve(lis)
}

// Halt will attempt to gracefully shutdown the server
func (h *httpServer) Halt(ctx context.Context) error {
	h.l.Info("halting http server...")
	if err := h.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("cannot shutdown http server: %w", err)
	}
	return nil
}

// Addr will block until you have received an address for your server.
func (h *httpServer) addr() *net.TCPAddr {
	if h.tcpAddr != nil {
		return h.tcpAddr
	}
	t := time.NewTimer(5 * time.Second)
	select {
	case addr := <-h.addrC:
		h.tcpAddr = addr
	case <-t.C:
		return &net.TCPAddr{}
	}
	return h.tcpAddr
}

// NewHTTPServer returns a new HTTP server
func NewHTTPServer(cnt *container) (*httpServer, error) {
	if cnt == nil {
		return nil, fmt.Errorf("container: %w", domain.ErrIsNil)
	}
	if cnt.config == nil {
		return nil, fmt.Errorf("config: %w", domain.ErrIsNil)
	}
	if cnt.logger == nil {
		return nil, fmt.Errorf("logger: %w", domain.ErrIsNil)
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cnt.config.ServicePort),
		Handler:           newRouter(cnt),
		WriteTimeout:      5 * time.Second,
		ReadTimeout:       5 * time.Second,
		IdleTimeout:       5 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
	}

	addrC := make(chan *net.TCPAddr, 1)

	return &httpServer{srv, addrC, nil, cnt.logger}, nil
}

// newRouter instantiates a router for the HTTP server
func newRouter(cnt *container) *mux.Router {
	rtr := mux.NewRouter()

	// api endpoints
	api := rtr.PathPrefix("/api").Subrouter()

	api.HandleFunc("/season/{season_id}", retrieveSeasonHandler(cnt)).Methods(http.MethodGet)
	api.HandleFunc("/season/{season_id}/entry", createEntryHandler(cnt)).Methods(http.MethodPost)
	api.HandleFunc("/season/{season_id}/leaderboard/{round_number:[0-9]+}", retrieveLeaderBoardHandler(cnt)).Methods(http.MethodGet)

	api.HandleFunc("/entry/{entry_id}/prediction", createEntryPredictionHandler(cnt)).Methods(http.MethodPost)
	api.HandleFunc("/entry/{entry_id}/prediction", retrieveLatestEntryPredictionHandler(cnt)).Methods(http.MethodGet)
	api.HandleFunc("/entry/{entry_id}/scored/{round_number:[0-9]+}", retrieveLatestScoredEntryPrediction(cnt)).Methods(http.MethodGet)
	api.HandleFunc("/entry/{entry_id}/payment", updateEntryPaymentDetailsHandler(cnt)).Methods(http.MethodPatch)

	// requires basic auth
	api.HandleFunc("/entry/{entry_id}/approve", approveEntryByIDHandler(cnt)).Methods(http.MethodPatch)
	api.HandleFunc("/entry/{entry_id}/generate-login", generateExtendedMagicLoginTokenHandler(cnt)).Methods(http.MethodPost)

	// serve static assets
	assets := http.Dir("./resources/dist")
	rtr.PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(assets)))

	// frontend endpoints
	rtr.HandleFunc("/", frontendIndexHandler(cnt)).Methods(http.MethodGet)
	rtr.HandleFunc("/leaderboard", frontendLeaderBoardHandler(cnt)).Methods(http.MethodGet)
	rtr.HandleFunc("/faq", frontendFAQHandler(cnt)).Methods(http.MethodGet)
	rtr.HandleFunc("/join", frontendJoinHandler(cnt)).Methods(http.MethodGet)
	rtr.HandleFunc("/prediction", frontendPredictionHandler(cnt)).Methods(http.MethodGet)

	rtr.HandleFunc("/login", frontendGenerateMagicLoginHandler(cnt)).Methods(http.MethodPost)
	rtr.HandleFunc("/login/failed", frontendMagicLoginFailedHandler(cnt)).Methods(http.MethodGet)
	rtr.HandleFunc("/login/{magic_token_id}", frontendRedeemMagicLoginHandler(cnt)).Methods(http.MethodGet)

	return rtr
}
