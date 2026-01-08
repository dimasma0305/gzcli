package uploadserver

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/dimasma0305/gzcli/internal/log"
)

const (
	maxUploadBytes = 200 << 20 // 200 MiB
)

// Options configures the upload server runtime.
type Options struct {
	Host  string
	Port  int
	Event string
}

type server struct {
	opts      Options
	templates *template.Template
}

func newServer(opts Options) (*server, error) {
	s := &server{opts: opts}

	if err := ensureTemplatePaths(); err != nil {
		return nil, fmt.Errorf("template assets unavailable: %w", err)
	}

	if err := s.loadTemplates(); err != nil {
		return nil, err
	}

	return s, nil
}

// Run starts the upload server with the provided options.
func Run(opts Options) error {
	srv, err := newServer(opts)
	if err != nil {
		return fmt.Errorf("failed to initialize upload server: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           srv.routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Info("Upload server listening on http://%s", addr)
	return httpServer.ListenAndServe()
}
