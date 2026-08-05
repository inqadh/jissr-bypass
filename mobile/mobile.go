// Package jissrbypass exposes the deliberately small API supported by
// gomobile on both Android and Apple platforms.
package jissrbypass

import (
	"encoding/json"
	"fmt"
	"net/http"

	core "github.com/inqadh/jissr-bypass"
)

// Client wraps the shared ECH transport.
type Client struct {
	core *core.Client
}

// Response is a gomobile-compatible HTTP response.
type Response struct {
	statusCode int
	protocol   string
	body       []byte
	headers    string
}

// GetStatusCode returns the HTTP status code.
func (response *Response) GetStatusCode() int {
	if response == nil {
		return 0
	}
	return response.statusCode
}

// GetProtocol returns the negotiated HTTP protocol, for example HTTP/2.0.
func (response *Response) GetProtocol() string {
	if response == nil {
		return ""
	}
	return response.protocol
}

// GetBody returns the response body.
func (response *Response) GetBody() []byte {
	if response == nil {
		return nil
	}
	return append([]byte(nil), response.body...)
}

// GetHeadersJSON returns response headers as a JSON object of string arrays.
func (response *Response) GetHeadersJSON() string {
	if response == nil {
		return "{}"
	}
	return response.headers
}

// NewClient creates a client with the maintained Cloudflare/Google fallback
// set. Advanced native integrations should use the root Go package directly.
func NewClient() *Client {
	return &Client{core: core.NewDefaultClient()}
}

// Fetch performs an ECH-protected HTTPS request.
func (client *Client) Fetch(method, url string, body []byte) ([]byte, error) {
	if client == nil || client.core == nil {
		return nil, fmt.Errorf("client is not initialized")
	}
	return client.core.Fetch(method, url, body)
}

// Execute performs an ECH-protected HTTPS request. Request headers are a JSON
// object whose values may be a string or an array of strings.
func (client *Client) Execute(method, url string, body []byte, headersJSON string) (*Response, error) {
	if client == nil || client.core == nil {
		return nil, fmt.Errorf("client is not initialized")
	}
	headers, err := decodeHeaders(headersJSON)
	if err != nil {
		return nil, err
	}
	result, err := client.core.Execute(method, url, body, headers)
	if err != nil {
		return nil, err
	}
	encodedHeaders, err := json.Marshal(result.Headers)
	if err != nil {
		return nil, fmt.Errorf("encode response headers: %w", err)
	}
	return &Response{
		statusCode: result.StatusCode,
		protocol:   result.Protocol,
		body:       result.Body,
		headers:    string(encodedHeaders),
	}, nil
}

func decodeHeaders(encoded string) (http.Header, error) {
	headers := make(http.Header)
	if encoded == "" {
		return headers, nil
	}
	var values map[string]json.RawMessage
	if err := json.Unmarshal([]byte(encoded), &values); err != nil {
		return nil, fmt.Errorf("decode request headers: %w", err)
	}
	for name, raw := range values {
		var single string
		if err := json.Unmarshal(raw, &single); err == nil {
			headers.Add(name, single)
			continue
		}
		var multiple []string
		if err := json.Unmarshal(raw, &multiple); err != nil {
			return nil, fmt.Errorf("request header %q must be a string or string array", name)
		}
		for _, value := range multiple {
			headers.Add(name, value)
		}
	}
	return headers, nil
}

// Probe returns a JSON ProbeResult. JSON keeps the gomobile API stable while
// allowing diagnostic fields to grow compatibly.
func (client *Client) Probe(host string) (string, error) {
	if client == nil || client.core == nil {
		return "", fmt.Errorf("client is not initialized")
	}
	result := client.core.Probe(host)
	encoded, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	if !result.OK {
		return string(encoded), fmt.Errorf("%s: %s", result.Stage, result.Detail)
	}
	return string(encoded), nil
}

// VerifyCanary verifies the standard nonce, protocol, host and freshness
// checks, returning the result as JSON.
func (client *Client) VerifyCanary(url string) (string, error) {
	if client == nil || client.core == nil {
		return "", fmt.Errorf("client is not initialized")
	}
	result, err := client.core.VerifyCanary(url)
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(result)
	return string(encoded), err
}
