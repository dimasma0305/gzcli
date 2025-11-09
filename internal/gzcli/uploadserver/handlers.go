// Package uploadserver hosts the HTTP handlers and templates for the challenge upload server.
package uploadserver

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/log"
)

//go:embed assets/*
var assetsFS embed.FS

const (
	templateHomeFile = "home.gohtml"
	templateHome     = "home"
)

type viewData struct {
	Title       string
	Events      []string
	Categories  []string
	Templates   []templateInfo
	SuccessMsg  string
	ErrorMsg    string
	DefaultHost string
	DefaultPort int
	MaxUpload   string
	MaxExtract  string
	MaxEntry    string
}

func (s *server) loadTemplates() error {
	tmpl, err := template.ParseFS(assetsFS, filepath.Join("assets", templateHomeFile))
	if err != nil {
		return err
	}

	s.templates = tmpl
	return nil
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.handleHome)
	mux.HandleFunc("/upload", s.handleUpload)
	mux.HandleFunc("/templates/", s.handleTemplateDownload)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return mux
}

func (s *server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := s.baseViewData()
	if err := s.templates.ExecuteTemplate(w, templateHome, data); err != nil {
		log.Error("Template render error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := s.baseViewData()

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		data.ErrorMsg = friendlyError(err)
		s.renderWithStatus(w, data, http.StatusBadRequest)
		return
	}

	event := strings.TrimSpace(r.FormValue("event"))
	category := strings.TrimSpace(r.FormValue("category"))

	file, header, err := r.FormFile("challenge")
	if err != nil {
		data.ErrorMsg = "challenge ZIP is required"
		s.renderWithStatus(w, data, http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	if err := s.processUpload(r.Context(), event, category, file, header.Filename); err != nil {
		data.ErrorMsg = err.Error()
		s.renderWithStatus(w, data, http.StatusBadRequest)
		return
	}

	data.SuccessMsg = "Challenge uploaded successfully."
	s.renderWithStatus(w, data, http.StatusOK)
}

func (s *server) handleTemplateDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/templates/")
	slug = strings.TrimSuffix(slug, ".zip")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	template, ok := getTemplateBySlug(slug)
	if !ok {
		http.NotFound(w, r)
		return
	}

	buf := &bytes.Buffer{}
	if err := writeTemplateArchive(buf, template); err != nil {
		log.Error("Failed generating template archive %s: %v", slug, err)
		http.Error(w, "failed to build template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+slug+`.zip"`)
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Error("Failed writing template archive %s: %v", slug, err)
	}
}

func (s *server) baseViewData() viewData {
	events, err := config.ListEvents()
	if err != nil {
		log.Error("Failed to list events: %v", err)
		events = []string{}
	}

	return viewData{
		Title:       "GZCLI Challenge Upload Server",
		Events:      events,
		Categories:  config.CHALLENGE_CATEGORY,
		Templates:   listTemplateInfo(),
		DefaultHost: s.opts.Host,
		DefaultPort: s.opts.Port,
		MaxUpload:   formatBytes(uint64(maxUploadBytes)),
		MaxExtract:  formatBytes(maxExtractedBytes),
		MaxEntry:    formatBytes(maxEntryBytes),
	}
}

func (s *server) renderWithStatus(w http.ResponseWriter, data viewData, status int) {
	w.WriteHeader(status)
	if err := s.templates.ExecuteTemplate(w, templateHome, data); err != nil {
		log.Error("Template render error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func friendlyError(err error) string {
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		return "uploaded file exceeds size limit"
	}
	return "invalid upload payload"
}

func formatBytes(limit uint64) string {
	const mib = 1 << 20
	if limit <= 0 {
		return "unknown"
	}
	if limit%mib == 0 {
		return fmt.Sprintf("%d MiB", limit/mib)
	}
	value := float64(limit) / float64(mib)
	return fmt.Sprintf("%.1f MiB", value)
}
