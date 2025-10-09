//nolint:revive // Struct field names match API responses
package gzapi

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/imroc/req/v3"

	"github.com/dimasma0305/gzcli/internal/gzcli/errors"
	"github.com/dimasma0305/gzcli/internal/log"
)

type Creds struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

type GZAPI struct {
	Url    string
	Creds  *Creds
	Client *req.Client
}

func Init(url string, creds *Creds) (*GZAPI, error) {
	// Validate inputs
	if creds == nil {
		return nil, errors.ErrInvalidCredentials
	}
	if url == "" {
		return nil, errors.ErrEmptyURL
	}

	url = strings.TrimRight(url, "/")
	newGz := &GZAPI{
		Client: createOptimizedClient(),
		Url:    url,
		Creds:  creds,
	}
	if err := newGz.Login(); err != nil {
		return nil, err
	}
	return newGz, nil
}

func Register(url string, creds *RegisterForm) (*GZAPI, error) {
	// Validate inputs
	if creds == nil {
		return nil, fmt.Errorf("registration form cannot be nil")
	}
	if url == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	url = strings.TrimRight(url, "/")
	newGz := &GZAPI{
		Client: createOptimizedClient(),
		Url:    url,
		Creds: &Creds{
			Username: creds.Username,
			Password: creds.Password,
		},
	}
	if err := newGz.Register(creds); err != nil {
		return nil, err
	}
	return newGz, nil
}

// createOptimizedClient creates an HTTP client with optimal performance settings
func createOptimizedClient() *req.Client {
	client := req.C().
		SetUserAgent("Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/110.0").
		SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // G402: InsecureSkipVerify needed for self-signed certs in dev/test
			MinVersion:         tls.VersionTLS12,
		}).
		SetTimeout(30 * time.Second). // Default timeout for most operations
		EnableKeepAlives()            // Enable connection keep-alive (auto-negotiates HTTP/2 for HTTPS)

	// Configure transport for optimal connection pooling
	transport := client.GetTransport()
	if transport != nil {
		transport.SetMaxIdleConns(100). // Increase connection pool
						SetIdleConnTimeout(90 * time.Second). // Keep connections alive longer
						SetMaxConnsPerHost(10)                // Max connections per host
	}

	return client
}

// requestExecutor is a function that executes an HTTP request
type requestExecutor func(*req.Request, string) (*req.Response, error)

// doRequest handles common HTTP request logic
func (cs *GZAPI) doRequest(method, url string, data any, executor requestExecutor) error {
	if cs == nil || cs.Client == nil {
		return fmt.Errorf("GZAPI client is not initialized")
	}

	// Build full URL efficiently
	var urlBuilder strings.Builder
	urlBuilder.Grow(len(cs.Url) + len(url))
	urlBuilder.WriteString(cs.Url)
	urlBuilder.WriteString(url)
	fullURL := urlBuilder.String()

	log.InfoH3("Making %s request to: %s", method, fullURL)

	// Execute the request
	resp, err := executor(cs.Client.R(), fullURL)
	if err != nil {
		log.Error("%s request failed for %s: %v", method, fullURL, err)
		return fmt.Errorf("%s request failed for %s: %w", method, fullURL, err)
	}

	// Validate status code
	if resp.StatusCode != 200 {
		log.Error("%s request returned status %d for %s: %s", method, resp.StatusCode, fullURL, resp.String())
		return fmt.Errorf("request end with %d status, %s", resp.StatusCode, resp.String())
	}

	// Unmarshal response if data pointer provided
	if data != nil {
		if len(resp.Bytes()) == 0 {
			log.DebugH3("%s response has empty body, skipping unmarshal for: %s", method, fullURL)
		} else {
			if err := resp.UnmarshalJson(&data); err != nil {
				log.Error("Failed to unmarshal JSON response from %s: %v", fullURL, err)
				return fmt.Errorf("error unmarshal json: %w, %s", err, resp.String())
			}
		}
	}

	log.InfoH3("%s request successful for: %s", method, fullURL)
	return nil
}

func (cs *GZAPI) get(url string, data any) error {
	return cs.doRequest("GET", url, data, func(r *req.Request, url string) (*req.Response, error) {
		return r.Get(url)
	})
}

func (cs *GZAPI) delete(url string, data any) error {
	return cs.doRequest("DELETE", url, data, func(r *req.Request, url string) (*req.Response, error) {
		return r.Delete(url)
	})
}

func (cs *GZAPI) post(url string, json any, data any) error {
	return cs.doRequest("POST", url, data, func(r *req.Request, url string) (*req.Response, error) {
		return r.SetBodyJsonMarshal(json).Post(url)
	})
}

func (cs *GZAPI) put(url string, json any, data any) error {
	return cs.doRequest("PUT", url, data, func(r *req.Request, url string) (*req.Response, error) {
		return r.SetBodyJsonMarshal(json).Put(url)
	})
}

func (cs *GZAPI) postMultiPart(url string, file string, data any) error {
	// Verify file exists before attempting upload
	if _, err := os.Stat(file); err != nil {
		log.Error("File does not exist: %s", file)
		return fmt.Errorf("file not found: %s", file)
	}

	// Build full URL for logging
	var urlBuilder strings.Builder
	urlBuilder.Grow(len(cs.Url) + len(url))
	urlBuilder.WriteString(cs.Url)
	urlBuilder.WriteString(url)
	fullURL := urlBuilder.String()
	log.InfoH3("Making POST multipart request to: %s with file: %s", fullURL, file)

	// Use "files" for /api/assets endpoint as per API specification
	return cs.doRequest("POST", url, data, func(r *req.Request, url string) (*req.Response, error) {
		return r.SetFile("files", file).Post(url)
	})
}

func (cs *GZAPI) putMultiPart(url string, file string, data any) error {
	// Verify file exists before attempting upload
	if _, err := os.Stat(file); err != nil {
		log.Error("File does not exist: %s", file)
		return fmt.Errorf("file not found: %s", file)
	}

	// Build full URL for logging
	var urlBuilder strings.Builder
	urlBuilder.Grow(len(cs.Url) + len(url))
	urlBuilder.WriteString(cs.Url)
	urlBuilder.WriteString(url)
	fullURL := urlBuilder.String()
	log.InfoH3("Making PUT multipart request to: %s with file: %s", fullURL, file)

	// Use "file" for PUT operations (poster/avatar uploads) as per API specification
	return cs.doRequest("PUT", url, data, func(r *req.Request, url string) (*req.Response, error) {
		return r.SetFile("file", file).Put(url)
	})
}
