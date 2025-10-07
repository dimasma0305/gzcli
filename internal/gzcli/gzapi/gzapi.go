//nolint:revive // Struct field names match API responses
package gzapi

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/imroc/req/v3"

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
		return nil, fmt.Errorf("credentials cannot be nil")
	}
	if url == "" {
		return nil, fmt.Errorf("URL cannot be empty")
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

//nolint:dupl // HTTP methods have intentional similarity for clarity
func (cs *GZAPI) get(url string, data any) error {
	if cs == nil || cs.Client == nil {
		return fmt.Errorf("GZAPI client is not initialized")
	}
	// Use string builder for efficient URL construction
	var urlBuilder strings.Builder
	urlBuilder.Grow(len(cs.Url) + len(url))
	urlBuilder.WriteString(cs.Url)
	urlBuilder.WriteString(url)
	fullURL := urlBuilder.String()
	log.InfoH3("Making GET request to: %s", fullURL)

	req, err := cs.Client.R().Get(fullURL)
	if err != nil {
		log.Error("GET request failed for %s: %v", fullURL, err)
		return fmt.Errorf("GET request failed for %s: %w", fullURL, err)
	}

	if req.StatusCode != 200 {
		log.Error("GET request returned status %d for %s: %s", req.StatusCode, fullURL, req.String())
		return fmt.Errorf("request end with %d status, %s", req.StatusCode, req.String())
	}

	if data != nil {
		if err := req.UnmarshalJson(&data); err != nil {
			log.Error("Failed to unmarshal JSON response from %s: %v", fullURL, err)
			return fmt.Errorf("error unmarshal json: %w, %s", err, req.String())
		}
	}

	log.InfoH3("GET request successful for: %s", fullURL)
	return nil
}

//nolint:dupl // HTTP methods have intentional similarity for clarity
func (cs *GZAPI) delete(url string, data any) error {
	if cs == nil || cs.Client == nil {
		return fmt.Errorf("GZAPI client is not initialized")
	}
	var urlBuilder strings.Builder
	urlBuilder.Grow(len(cs.Url) + len(url))
	urlBuilder.WriteString(cs.Url)
	urlBuilder.WriteString(url)
	fullURL := urlBuilder.String()
	log.InfoH3("Making DELETE request to: %s", fullURL)

	req, err := cs.Client.R().Delete(fullURL)
	if err != nil {
		log.Error("DELETE request failed for %s: %v", fullURL, err)
		return fmt.Errorf("DELETE request failed for %s: %w", fullURL, err)
	}

	if req.StatusCode != 200 {
		log.Error("DELETE request returned status %d for %s: %s", req.StatusCode, fullURL, req.String())
		return fmt.Errorf("request end with %d status, %s", req.StatusCode, req.String())
	}

	if data != nil {
		if err := req.UnmarshalJson(&data); err != nil {
			log.Error("Failed to unmarshal JSON response from %s: %v", fullURL, err)
			return fmt.Errorf("error unmarshal json: %w, %s", err, req.String())
		}
	}

	log.InfoH3("DELETE request successful for: %s", fullURL)
	return nil
}

//nolint:dupl // HTTP methods have intentional similarity for clarity
func (cs *GZAPI) post(url string, json any, data any) error {
	if cs == nil || cs.Client == nil {
		return fmt.Errorf("GZAPI client is not initialized")
	}
	var urlBuilder strings.Builder
	urlBuilder.Grow(len(cs.Url) + len(url))
	urlBuilder.WriteString(cs.Url)
	urlBuilder.WriteString(url)
	fullURL := urlBuilder.String()
	log.InfoH3("Making POST request to: %s", fullURL)

	req, err := cs.Client.R().SetBodyJsonMarshal(json).Post(fullURL)
	if err != nil {
		log.Error("POST request failed for %s: %v", fullURL, err)
		return fmt.Errorf("POST request failed for %s: %w", fullURL, err)
	}

	if req.StatusCode != 200 {
		log.Error("POST request returned status %d for %s: %s", req.StatusCode, fullURL, req.String())
		return fmt.Errorf("request end with %d status, %s", req.StatusCode, req.String())
	}

	if data != nil {
		if err := req.UnmarshalJson(&data); err != nil {
			log.Error("Failed to unmarshal JSON response from %s: %v", fullURL, err)
			return fmt.Errorf("error unmarshal json: %w, %s", err, req.String())
		}
	}

	log.InfoH3("POST request successful for: %s", fullURL)
	return nil
}

//nolint:dupl // Multipart methods have intentional similarity for clarity
func (cs *GZAPI) postMultiPart(url string, file string, data any) error {
	if cs == nil || cs.Client == nil {
		return fmt.Errorf("GZAPI client is not initialized")
	}
	var urlBuilder strings.Builder
	urlBuilder.Grow(len(cs.Url) + len(url))
	urlBuilder.WriteString(cs.Url)
	urlBuilder.WriteString(url)
	fullURL := urlBuilder.String()
	log.InfoH3("Making POST multipart request to: %s with file: %s", fullURL, file)

	// Verify file exists before attempting upload
	if _, err := os.Stat(file); err != nil {
		log.Error("File does not exist: %s", file)
		return fmt.Errorf("file not found: %s", file)
	}

	// Use "files" for /api/assets endpoint as per API specification
	req, err := cs.Client.R().SetFile("files", file).Post(fullURL)
	if err != nil {
		log.Error("POST multipart request failed for %s: %v", fullURL, err)
		return fmt.Errorf("POST multipart request failed for %s: %w", fullURL, err)
	}

	if req.StatusCode != 200 {
		log.Error("POST multipart request returned status %d for %s: %s", req.StatusCode, fullURL, req.String())
		return fmt.Errorf("request end with %d status, %s", req.StatusCode, req.String())
	}

	if data != nil {
		if err := req.UnmarshalJson(&data); err != nil {
			log.Error("Failed to unmarshal JSON response from %s: %v", fullURL, err)
			return fmt.Errorf("error unmarshal json: %w, %s", err, req.String())
		}
	}

	log.InfoH3("POST multipart request successful for: %s", fullURL)
	return nil
}

//nolint:dupl // Multipart methods have intentional similarity for clarity
func (cs *GZAPI) putMultiPart(url string, file string, data any) error {
	if cs == nil || cs.Client == nil {
		return fmt.Errorf("GZAPI client is not initialized")
	}
	var urlBuilder strings.Builder
	urlBuilder.Grow(len(cs.Url) + len(url))
	urlBuilder.WriteString(cs.Url)
	urlBuilder.WriteString(url)
	fullURL := urlBuilder.String()
	log.InfoH3("Making PUT multipart request to: %s with file: %s", fullURL, file)

	// Verify file exists before attempting upload
	if _, err := os.Stat(file); err != nil {
		log.Error("File does not exist: %s", file)
		return fmt.Errorf("file not found: %s", file)
	}

	// Use "file" for PUT operations (poster/avatar uploads) as per API specification
	req, err := cs.Client.R().SetFile("file", file).Put(fullURL)
	if err != nil {
		log.Error("PUT multipart request failed for %s: %v", fullURL, err)
		return fmt.Errorf("PUT multipart request failed for %s: %w", fullURL, err)
	}

	if req.StatusCode != 200 {
		log.Error("PUT multipart request returned status %d for %s: %s", req.StatusCode, fullURL, req.String())
		return fmt.Errorf("request end with %d status, %s", req.StatusCode, req.String())
	}

	if data != nil {
		if err := req.UnmarshalJson(&data); err != nil {
			log.Error("Failed to unmarshal JSON response from %s: %v", fullURL, err)
			return fmt.Errorf("error unmarshal json: %w, %s", err, req.String())
		}
	}

	log.InfoH3("PUT multipart request successful for: %s", fullURL)
	return nil
}

//nolint:dupl // HTTP methods have intentional similarity for clarity
func (cs *GZAPI) put(url string, json any, data any) error {
	if cs == nil || cs.Client == nil {
		return fmt.Errorf("GZAPI client is not initialized")
	}
	var urlBuilder strings.Builder
	urlBuilder.Grow(len(cs.Url) + len(url))
	urlBuilder.WriteString(cs.Url)
	urlBuilder.WriteString(url)
	fullURL := urlBuilder.String()
	log.InfoH3("Making PUT request to: %s", fullURL)

	req, err := cs.Client.R().SetBodyJsonMarshal(json).Put(fullURL)
	if err != nil {
		log.Error("PUT request failed for %s: %v", fullURL, err)
		return fmt.Errorf("PUT request failed for %s: %w", fullURL, err)
	}

	if req.StatusCode != 200 {
		log.Error("PUT request returned status %d for %s: %s", req.StatusCode, fullURL, req.String())
		return fmt.Errorf("request end with %d status, %s", req.StatusCode, req.String())
	}

	if data != nil {
		if err := req.UnmarshalJson(&data); err != nil {
			log.Error("Failed to unmarshal JSON response from %s: %v", fullURL, err)
			return fmt.Errorf("error unmarshal json: %w, %s", err, req.String())
		}
	}

	log.InfoH3("PUT request successful for: %s", fullURL)
	return nil
}
