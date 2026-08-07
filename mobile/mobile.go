// Package jissrbypass exposes the deliberately small API supported by
// gomobile on both Android and Apple platforms.
package jissrbypass

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	core "github.com/inqadh/jissr-bypass"
)

// Client wraps the shared ECH transport.
type Client struct {
	core *core.Client
}

// EventListener receives major transport decisions. Implementations must
// return promptly and must not call the same Client from OnEvent.
type EventListener interface {
	OnEvent(event *Event)
}

// Event is a gomobile-compatible transport event.
type Event struct {
	eventType          string
	host               string
	route              string
	stage              string
	reason             string
	detail             string
	resolver           string
	cached             bool
	safeToFallback     bool
	likelyInterference bool
}

func (event *Event) GetType() string {
	if event == nil {
		return ""
	}
	return event.eventType
}

func (event *Event) GetHost() string {
	if event == nil {
		return ""
	}
	return event.host
}

func (event *Event) GetRoute() string {
	if event == nil {
		return ""
	}
	return event.route
}

func (event *Event) GetStage() string {
	if event == nil {
		return ""
	}
	return event.stage
}

func (event *Event) GetReason() string {
	if event == nil {
		return ""
	}
	return event.reason
}

func (event *Event) GetDetail() string {
	if event == nil {
		return ""
	}
	return event.detail
}

func (event *Event) GetResolver() string {
	if event == nil {
		return ""
	}
	return event.resolver
}

func (event *Event) GetCached() bool {
	return event != nil && event.cached
}

func (event *Event) GetSafeToFallback() bool {
	return event != nil && event.safeToFallback
}

func (event *Event) GetLikelyInterference() bool {
	return event != nil && event.likelyInterference
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

// NewAutoClient creates a client that remembers whether each host should use
// the ordinary native transport or Jissr Bypass.
func NewAutoClient() *Client {
	return &Client{core: core.NewAutoClient()}
}

// NewAutoClientWithTimeout creates an auto client with an application-defined
// total native route-attempt budget in milliseconds.
func NewAutoClientWithTimeout(timeoutMillis int64) (*Client, error) {
	if timeoutMillis <= 0 {
		return nil, fmt.Errorf("native attempt timeout must be greater than zero")
	}
	if timeoutMillis > int64(time.Duration(1<<63-1)/time.Millisecond) {
		return nil, fmt.Errorf("native attempt timeout is too large")
	}
	config := core.DefaultConfig()
	config.NativeFirst = true
	config.NativeAttemptTimeout = time.Duration(timeoutMillis) * time.Millisecond
	client, err := core.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Client{core: client}, nil
}

// SetEventListener replaces the listener. Passing nil disables events.
func (client *Client) SetEventListener(listener EventListener) {
	if client == nil || client.core == nil {
		return
	}
	if listener == nil {
		client.core.SetEventHandler(nil)
		return
	}
	client.core.SetEventHandler(func(event core.Event) {
		listener.OnEvent(&Event{
			eventType:          event.Type,
			host:               event.Host,
			route:              event.Route,
			stage:              event.Stage,
			reason:             event.Reason,
			detail:             event.Detail,
			resolver:           event.Resolver,
			cached:             event.Cached,
			safeToFallback:     event.SafeToFallback,
			likelyInterference: event.LikelyInterference,
		})
	})
}

// ResetNetworkState clears cached route and resolution decisions after the
// device changes networks.
func (client *Client) ResetNetworkState() {
	if client == nil || client.core == nil {
		return
	}
	client.core.ResetNetworkState()
}

// CloseIdleConnections releases pooled native and ECH connections.
func (client *Client) CloseIdleConnections() {
	if client == nil || client.core == nil {
		return
	}
	client.core.CloseIdleConnections()
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
