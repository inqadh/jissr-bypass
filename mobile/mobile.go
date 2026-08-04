// Package jissrbypass exposes the deliberately small API supported by
// gomobile on both Android and Apple platforms.
package jissrbypass

import (
	"encoding/json"
	"fmt"

	core "github.com/jissr-org/jissr-bypass"
)

// Client wraps the shared ECH transport.
type Client struct {
	core *core.Client
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
