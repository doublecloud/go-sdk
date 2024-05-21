package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
)

type URLRequestFunc func(serverAddr string) error

func GetToken(ctx context.Context, f URLRequestFunc, consoleURL string) (*Token, error) {
	srv, err := serve(consoleURL)
	if err != nil {
		return nil, err
	}

	err = f(srv.Addr().String())
	if err != nil {
		_ = srv.Shutdown(ctx)
		return nil, err
	}

	select {
	case <-srv.Served():
	case <-ctx.Done():
	}

	// ensure server shutdown before reading Token from the struct
	_ = srv.Shutdown(ctx)

	return srv.Token(), err
}

func (s *server) Token() *Token {
	return s.token
}

func (s *server) Served() <-chan struct{} {
	return s.served
}

type Token struct {
	IamToken  string
	ExpiresAt *timestamp.Timestamp
	Err       error
}

type server struct {
	srv        http.Server
	lsn        net.Listener
	err        chan error
	served     chan struct{}
	token      *Token
	mux        sync.Mutex
	logger     log.Logger
	consoleURL string
}

func (s *server) Addr() net.Addr {
	return s.lsn.Addr()
}

func (s *server) Serve() error {
	var err error
	s.lsn, err = listener()
	if err != nil {
		return err
	}

	s.err = make(chan error)
	s.served = make(chan struct{}, 1)
	mux := http.NewServeMux()
	mux.Handle("/", s)
	s.srv.Handler = mux
	go func() {
		s.err <- s.srv.Serve(s.lsn)
	}()

	return nil
}

func (s *server) Shutdown(ctx context.Context) error {
	err := s.srv.Shutdown(ctx)
	if err != nil {
		return err
	}
	select {
	case err = <-s.err:
		if err != http.ErrServerClosed {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.token != nil {
		return
	}

	defer redirectToConsole(w, r, s.consoleURL)

	s.token = &Token{}
	defer func() {
		close(s.served)
	}()

	flags := r.URL.Query()

	tokens, ok := flags["token"]
	if !ok || len(tokens) != 1 {
		s.token.Err = fmt.Errorf("failed to fetch token from http request")
		return
	}

	s.token.IamToken = tokens[0]

	times, ok := flags["expiresAt"]
	if !ok || len(times) != 1 {
		s.token.Err = fmt.Errorf("failed to fetch token expiration timestamp from http request")
		return
	}

	t, err := time.Parse(time.RFC3339, times[0])
	if err != nil {
		s.token.Err = fmt.Errorf("failed to parse token expiration timestamp from: %s, %v", times[0], err)
		return
	}

	s.token.ExpiresAt, err = ptypes.TimestampProto(t)
	if err != nil {
		s.token.Err = fmt.Errorf("failed to parse token expiration timestamp from: %s, %v", times[0], err)
		return
	}
}

func redirectToConsole(w http.ResponseWriter, r *http.Request, consoleURL string) {
	if consoleURL == "" {
		return
	}

	http.Redirect(w, r, consoleURL, http.StatusSeeOther)
}

func serve(consoleURL string) (*server, error) {
	srv := &server{
		consoleURL: consoleURL,
	}
	err := srv.Serve()
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func listener() (net.Listener, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			return nil, err
		}
	}
	return l, nil
}
