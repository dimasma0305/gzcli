package utils

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"mime/multipart"
	"net/url"
	"strings"

	"github.com/lqqyt2423/go-mitmproxy/proxy"
)

// FormData represents parsed form data from an HTTP request
type FormData struct {
	Files  []*multipart.FileHeader
	Values map[string][]string
}

// ParseBody parses the request body based on content type
func ParseBody(f *proxy.Flow) (interface{}, error) {
	contentType := f.Request.Header.Get("content-type")
	var data interface{}
	var err error
	data, err = parseJSON(f)
	if err != nil {
		data, err = parseMultipartFormData(f)
		if err != nil {
			data, err = parseFormURLEncoded(f)
			if err != nil {
				data, err = parseXML(f)
				if err != nil {
					data, err = parseTextPlain(f)
					if err != nil {
						err = fmt.Errorf("the body didn't have a match parser %s", contentType)
					}
				}
			}
		}
	}
	if data != nil {
		err = nil
	}
	return data, err
}

func parseJSON(f *proxy.Flow) (interface{}, error) {
	var data interface{}
	if err := json.Unmarshal(f.Request.Body, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func parseXML(f *proxy.Flow) (interface{}, error) {
	var data interface{}
	if err := xml.Unmarshal(f.Request.Body, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func parseFormURLEncoded(f *proxy.Flow) (interface{}, error) {
	values, err := url.ParseQuery(string(f.Request.Body))
	if err != nil {
		return nil, err
	}
	return values, nil
}

func parseMultipartFormData(f *proxy.Flow) (interface{}, error) {
	boundary := strings.ReplaceAll(f.Request.Header.Get("Content-Type"), "multipart/form-data; boundary=", "")
	reader := multipart.NewReader(strings.NewReader(string(f.Request.Body)), boundary)
	form, err := reader.ReadForm(100000000000000)
	if err != nil {
		return nil, err
	}
	result := &FormData{
		Files:  make([]*multipart.FileHeader, 0),
		Values: make(map[string][]string),
	}

	// Append files
	for _, files := range form.File {
		result.Files = append(result.Files, files...)
	}

	// Append values
	for key, values := range form.Value {
		result.Values[key] = values
	}
	return result, nil
}

//nolint:unparam // error return kept for interface consistency
func parseTextPlain(f *proxy.Flow) (string, error) {
	return string(f.Request.Body), nil
}
