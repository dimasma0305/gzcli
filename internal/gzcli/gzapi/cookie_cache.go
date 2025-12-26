package gzapi

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"
	"gopkg.in/yaml.v2"

	"github.com/dimasma0305/gzcli/internal/log"
)

type cookieStore struct {
	path    string
	baseURL *url.URL
	mu      sync.Mutex
}

type storedCookie struct {
	Name     string        `yaml:"name"`
	Value    string        `yaml:"value"`
	Path     string        `yaml:"path"`
	Domain   string        `yaml:"domain"`
	Expires  time.Time     `yaml:"expires"`
	Secure   bool          `yaml:"secure"`
	HTTPOnly bool          `yaml:"httpOnly"`
	SameSite http.SameSite `yaml:"sameSite"`
}

type storedCookiesFile struct {
	SavedAt time.Time      `yaml:"savedAt"`
	URL     string         `yaml:"url"`
	Cookies []storedCookie `yaml:"cookies"`
}

func newCookieStore(rawURL string, username string) (*cookieStore, error) {
	parsed, err := normalizeBaseURL(rawURL)
	if err != nil {
		return nil, err
	}

	path, err := cookieStorePath(parsed, username)
	if err != nil {
		return nil, err
	}

	return &cookieStore{
		path:    path,
		baseURL: parsed,
	}, nil
}

func (s *cookieStore) load() (*cookiejar.Jar, bool, error) {
	jar := s.newJar()
	if jar == nil {
		return nil, false, fmt.Errorf("failed to initialize cookie jar")
	}

	buf, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return jar, false, nil
		}
		return jar, false, fmt.Errorf("read cookie cache: %w", err)
	}

	var file storedCookiesFile
	if err := yaml.Unmarshal(buf, &file); err != nil {
		return jar, false, fmt.Errorf("parse cookie cache: %w", err)
	}

	cookies := make([]*http.Cookie, 0, len(file.Cookies))
	now := time.Now()
	for _, c := range file.Cookies {
		httpCookie := c.toHTTPCookie()
		if httpCookie.Expires.IsZero() || httpCookie.Expires.After(now) {
			cookies = append(cookies, httpCookie)
		}
	}

	if len(cookies) == 0 {
		return jar, false, nil
	}

	jar.SetCookies(s.baseURL, cookies)
	return jar, true, nil
}

func (s *cookieStore) save(jar *cookiejar.Jar) error {
	if s == nil || jar == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cookies := jar.Cookies(s.baseURL)
	filtered := make([]storedCookie, 0, len(cookies))
	now := time.Now()
	for _, c := range cookies {
		if c.Expires.IsZero() || c.Expires.After(now) {
			filtered = append(filtered, storedCookieFromHTTPCookie(c))
		}
	}

	if len(filtered) == 0 {
		if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove empty cookie cache: %w", err)
		}
		return nil
	}

	payload := storedCookiesFile{
		SavedAt: time.Now(),
		URL:     s.baseURL.String(),
		Cookies: filtered,
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0750); err != nil {
		return fmt.Errorf("create cookie cache dir: %w", err)
	}

	data, err := yaml.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode cookies: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0600); err != nil {
		return fmt.Errorf("write cookie cache: %w", err)
	}

	return nil
}

func (s *cookieStore) newJar() *cookiejar.Jar {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Error("Failed to create cookie jar: %v", err)
		return nil
	}
	return jar
}

func storedCookieFromHTTPCookie(c *http.Cookie) storedCookie {
	return storedCookie{
		Name:     c.Name,
		Value:    c.Value,
		Path:     c.Path,
		Domain:   c.Domain,
		Expires:  c.Expires,
		Secure:   c.Secure,
		HTTPOnly: c.HttpOnly,
		SameSite: c.SameSite,
	}
}

func (c storedCookie) toHTTPCookie() *http.Cookie {
	return &http.Cookie{
		Name:     c.Name,
		Value:    c.Value,
		Path:     c.Path,
		Domain:   c.Domain,
		Expires:  c.Expires,
		Secure:   c.Secure,
		HttpOnly: c.HTTPOnly,
		SameSite: c.SameSite,
	}
}

func normalizeBaseURL(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL for cookie cache: %w", err)
	}

	parsed.Path = "/"
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed, nil
}

func cookieStorePath(baseURL *url.URL, username string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("determine working directory: %w", err)
	}

	name := baseURL.Host
	if name == "" {
		name = "default"
	}
	if scheme := baseURL.Scheme; scheme != "" {
		name = scheme + "-" + name
	}

	// Sanitize username and append if present
	if username != "" {
		safeUsername := strings.ReplaceAll(username, string(filepath.Separator), "_")
		safeUsername = strings.ReplaceAll(safeUsername, ":", "-")
		name = name + "-" + safeUsername
	}

	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, string(filepath.Separator), "_")

	return filepath.Join(cwd, ".gzcli", "cache", "cookies", name+".yaml"), nil
}
