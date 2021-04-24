package app

import (
	"context"
	"fmt"
	"github.com/dustin/go-humanize"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

const (
	// port is the default HTTP port.
	port = 80
)

// Server represents a collection of functions for starting and running an RPC server.
type Server struct {
	Server  *http.Server
	Now     func() time.Time
	addrC   chan *net.TCPAddr
	tcpAddr *net.TCPAddr
}

// Run will start the gRPC server and listen for requests.
func (s *Server) Run(_ context.Context) error {
	addr := s.Server.Addr
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.addrC <- lis.Addr().(*net.TCPAddr)

	if s.Server.Handler == nil {
		return fmt.Errorf("http server needs a handler")
	}

	s.Server.Handler = wrapperHandler(s.Now, s.Server.Handler)
	log.Printf("serving http on http://%s", s.addr().String())
	return s.Server.Serve(lis)
}

// Halt will attempt to gracefully shut down the server.
func (s *Server) Halt(ctx context.Context) error {
	log.Printf("stopping serving http on http://%s...", s.addr().String())
	return s.Server.Shutdown(ctx)
}

// Addr will block until you have received an address for your server.
func (s *Server) addr() *net.TCPAddr {
	if s.tcpAddr != nil {
		return s.tcpAddr
	}
	t := time.NewTimer(5 * time.Second)
	select {
	case addr := <-s.addrC:
		s.tcpAddr = addr
	case <-t.C:
		return &net.TCPAddr{}
	}
	return s.tcpAddr
}

// NewServer sets up a new HTTP server.
func NewServer(server *http.Server) *Server {
	if server == nil {
		server = &defaultHTTPServer
	}
	if server.WriteTimeout == 0 {
		server.WriteTimeout = defaultHTTPServer.WriteTimeout
	}
	if server.ReadTimeout == 0 {
		server.ReadTimeout = defaultHTTPServer.ReadTimeout
	}
	if server.IdleTimeout == 0 {
		server.IdleTimeout = defaultHTTPServer.IdleTimeout
	}
	if server.ReadHeaderTimeout == 0 {
		server.ReadHeaderTimeout = defaultHTTPServer.ReadHeaderTimeout
	}
	if server.Addr == "" {
		var addr string
		if addr = os.Getenv("HTTP_ADDR"); addr == "" {
			addr = net.JoinHostPort("0.0.0.0", strconv.Itoa(port))
		}
		server.Addr = addr
	}
	return &Server{
		Server: server,
		Now:    time.Now,
		addrC:  make(chan *net.TCPAddr, 1),
	}
}

// defaultHTTPServer represents the default configuration for the http server
var defaultHTTPServer = http.Server{
	WriteTimeout:      5 * time.Second,
	ReadTimeout:       5 * time.Second,
	IdleTimeout:       5 * time.Second,
	ReadHeaderTimeout: 1 * time.Second,
}

// wrapperHandler returns the wrapper handler for the http server.
func wrapperHandler(now func() time.Time, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/healthz":
			healthHandler(now)(w, r)
		default:
			next.ServeHTTP(w, r)
		}
	}
}

// healthHandler responds with service health.
func healthHandler(now func() time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := now()

		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		latency := time.Since(start).Nanoseconds() / (1 * 1000 * 1000) // Milliseconds
		res := &response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Data: &data{
				Type: "health",
				Content: healthResponse{
					Latency:       fmt.Sprintf("%d ms", latency),
					HeapInUse:     humanize.Bytes(mem.HeapInuse),
					HeapAlloc:     humanize.Bytes(mem.HeapAlloc),
					StackInUse:    humanize.Bytes(mem.StackInuse),
					NumGoRoutines: runtime.NumGoroutine(),
				},
			},
		}
		res.writeTo(w)
	}
}

// healthResponse contains information about the service health.
type healthResponse struct {
	Latency       string `json:"latency"`
	StackInUse    string `json:"stack_in_use"`
	HeapInUse     string `json:"heap_in_use"`
	HeapAlloc     string `json:"heap_alloc"`
	NumGoRoutines int    `json:"num_go_routines"`
}
