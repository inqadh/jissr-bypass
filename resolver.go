package jissrbypass

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

type wireResolution struct {
	ip       string
	ech      []byte
	resolver string
}

func resolveECH(ctx context.Context, host string, config Config) (Resolution, error) {
	wire, wireErr := resolveWire(ctx, host, config.WireResolvers, config.UserAgent)
	if wireErr == nil {
		return Resolution{IP: wire.ip, ECH: wire.ech, Resolver: wire.resolver}, nil
	}
	jsonResolution, jsonErr := resolveJSON(ctx, host, config.JSONResolvers, config.UserAgent)
	if jsonErr == nil {
		return Resolution{IP: jsonResolution.ip, ECH: jsonResolution.ech, Resolver: jsonResolution.resolver}, nil
	}
	return Resolution{}, fmt.Errorf("wire DoH: %v; JSON DoH: %w", wireErr, jsonErr)
}

func resolveWire(ctx context.Context, host string, endpoints []string, userAgent string) (wireResolution, error) {
	raceContext, cancel := context.WithCancel(ctx)
	defer cancel()
	type result struct {
		resolution wireResolution
		err        error
	}
	results := make(chan result, len(endpoints))
	client := &http.Client{Timeout: 6 * time.Second}
	for _, endpoint := range endpoints {
		go func(endpoint string) {
			resolution, err := resolveWireEndpoint(raceContext, client, host, endpoint, userAgent)
			results <- result{resolution: resolution, err: err}
		}(endpoint)
	}
	errorsFound := make([]string, 0, len(endpoints))
	for range endpoints {
		result := <-results
		if result.err == nil {
			cancel()
			return result.resolution, nil
		}
		errorsFound = append(errorsFound, result.err.Error())
	}
	return wireResolution{}, fmt.Errorf("%s", strings.Join(errorsFound, "; "))
}

func resolveWireEndpoint(ctx context.Context, client *http.Client, host, endpoint, userAgent string) (wireResolution, error) {
	type queryResult struct {
		message *dnsmessage.Message
		err     error
	}
	addressResult := make(chan queryResult, 1)
	httpsResult := make(chan queryResult, 1)
	go func() {
		message, err := wireQuery(ctx, client, endpoint, host, dnsmessage.TypeA, userAgent)
		addressResult <- queryResult{message: message, err: err}
	}()
	go func() {
		message, err := wireQuery(ctx, client, endpoint, host, dnsmessage.TypeHTTPS, userAgent)
		httpsResult <- queryResult{message: message, err: err}
	}()
	address := <-addressResult
	https := <-httpsResult
	if address.err != nil {
		return wireResolution{}, fmt.Errorf("%s A: %w", endpoint, address.err)
	}
	if https.err != nil {
		return wireResolution{}, fmt.Errorf("%s HTTPS: %w", endpoint, https.err)
	}

	var ip string
	for _, answer := range address.message.Answers {
		if resource, ok := answer.Body.(*dnsmessage.AResource); ok {
			ip = net.IP(resource.A[:]).String()
			break
		}
	}
	if ip == "" {
		return wireResolution{}, fmt.Errorf("%s: no A record", endpoint)
	}
	for _, answer := range https.message.Answers {
		resource, ok := answer.Body.(*dnsmessage.HTTPSResource)
		if !ok {
			continue
		}
		ech, ok := resource.GetParam(dnsmessage.SVCParamECH)
		if ok && validECHConfigList(ech) {
			return wireResolution{ip: ip, ech: append([]byte(nil), ech...), resolver: endpoint}, nil
		}
	}
	return wireResolution{}, fmt.Errorf("%s: no valid ECH parameter", endpoint)
}

func wireQuery(ctx context.Context, client *http.Client, endpoint, host string, queryType dnsmessage.Type, userAgent string) (*dnsmessage.Message, error) {
	name, err := dnsmessage.NewName(strings.TrimSuffix(host, ".") + ".")
	if err != nil {
		return nil, err
	}
	query := dnsmessage.Message{
		Header: dnsmessage.Header{ID: randomDNSID(), RecursionDesired: true},
		Questions: []dnsmessage.Question{{
			Name: name, Type: queryType, Class: dnsmessage.ClassINET,
		}},
	}
	body, err := query.Pack()
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/dns-message")
	request.Header.Set("Content-Type", "application/dns-message")
	request.Header.Set("User-Agent", userAgent)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 64<<10))
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d", response.StatusCode)
	}
	if !strings.Contains(strings.ToLower(response.Header.Get("Content-Type")), "application/dns-message") {
		return nil, fmt.Errorf("unexpected content type %q", response.Header.Get("Content-Type"))
	}
	var message dnsmessage.Message
	if err := message.Unpack(responseBody); err != nil {
		return nil, err
	}
	if !message.Header.Response || message.Header.RCode != dnsmessage.RCodeSuccess {
		return nil, fmt.Errorf("dns status %s", message.Header.RCode)
	}
	if message.Header.ID != query.Header.ID {
		return nil, fmt.Errorf("DNS transaction ID mismatch")
	}
	if len(message.Questions) != 1 ||
		message.Questions[0].Name.String() != name.String() ||
		message.Questions[0].Type != queryType ||
		message.Questions[0].Class != dnsmessage.ClassINET {
		return nil, fmt.Errorf("DNS question mismatch")
	}
	return &message, nil
}

func randomDNSID() uint16 {
	var value [2]byte
	if _, err := rand.Read(value[:]); err != nil {
		return 0
	}
	return binary.BigEndian.Uint16(value[:])
}

type jsonDoHResponse struct {
	Status int `json:"Status"`
	Answer []struct {
		Type int    `json:"type"`
		Data string `json:"data"`
	} `json:"Answer"`
}

func resolveJSON(ctx context.Context, host string, endpoints []string, userAgent string) (wireResolution, error) {
	type result struct {
		resolution wireResolution
		err        error
	}
	results := make(chan result, len(endpoints))
	client := &http.Client{Timeout: 6 * time.Second}
	for _, endpoint := range endpoints {
		go func(endpoint string) {
			address, err := jsonQuery(ctx, client, endpoint, host, "A", userAgent)
			if err != nil {
				results <- result{err: fmt.Errorf("%s A: %w", endpoint, err)}
				return
			}
			https, err := jsonQuery(ctx, client, endpoint, host, "HTTPS", userAgent)
			if err != nil {
				results <- result{err: fmt.Errorf("%s HTTPS: %w", endpoint, err)}
				return
			}
			var ip string
			for _, answer := range address.Answer {
				if answer.Type == 1 && net.ParseIP(answer.Data) != nil {
					ip = answer.Data
					break
				}
			}
			for _, answer := range https.Answer {
				if answer.Type != 65 {
					continue
				}
				ech, err := extractECH(answer.Data)
				if err == nil && validECHConfigList(ech) && ip != "" {
					results <- result{resolution: wireResolution{ip: ip, ech: ech, resolver: endpoint}}
					return
				}
			}
			results <- result{err: fmt.Errorf("%s: incomplete A/HTTPS answer", endpoint)}
		}(endpoint)
	}
	errorsFound := make([]string, 0, len(endpoints))
	for range endpoints {
		result := <-results
		if result.err == nil {
			return result.resolution, nil
		}
		errorsFound = append(errorsFound, result.err.Error())
	}
	return wireResolution{}, fmt.Errorf("%s", strings.Join(errorsFound, "; "))
}

func jsonQuery(ctx context.Context, client *http.Client, endpoint, host, queryType, userAgent string) (*jsonDoHResponse, error) {
	requestURL := fmt.Sprintf("%s?name=%s&type=%s", endpoint, host, queryType)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/dns-json")
	request.Header.Set("User-Agent", userAgent)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 64<<10))
	if err != nil {
		return nil, err
	}
	contentType := strings.ToLower(response.Header.Get("Content-Type"))
	if response.StatusCode != http.StatusOK || (!strings.Contains(contentType, "json") && !strings.Contains(contentType, "dns")) {
		return nil, fmt.Errorf("http %d content-type %q", response.StatusCode, contentType)
	}
	var result jsonDoHResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if result.Status != 0 {
		return nil, fmt.Errorf("dns status %d", result.Status)
	}
	return &result, nil
}

func extractECH(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, `\#`) {
		raw, err := parseRawRData(value)
		if err != nil {
			return nil, err
		}
		ech := extractECHFromSVCB(raw)
		if ech == nil {
			return nil, fmt.Errorf("ECH parameter absent")
		}
		return ech, nil
	}
	for _, field := range strings.Fields(value) {
		key, encoded, ok := strings.Cut(field, "=")
		if !ok || (strings.ToLower(key) != "ech" && strings.ToLower(key) != "echconfig") {
			continue
		}
		return base64.StdEncoding.DecodeString(strings.Trim(encoded, `"`))
	}
	return nil, fmt.Errorf("ECH parameter absent")
}

func validECHConfigList(ech []byte) bool {
	return len(ech) >= 3 && int(binary.BigEndian.Uint16(ech[:2])) == len(ech)-2
}

func parseRawRData(value string) ([]byte, error) {
	parts := strings.Fields(value)
	if len(parts) < 3 || parts[0] != `\#` {
		return nil, fmt.Errorf("malformed raw RDATA")
	}
	declaredLength, err := strconv.Atoi(parts[1])
	if err != nil || declaredLength < 0 || declaredLength > 64<<10 {
		return nil, fmt.Errorf("invalid raw RDATA length")
	}
	var result []byte
	for _, part := range parts[2:] {
		decoded, err := hex.DecodeString(part)
		if err != nil {
			return nil, err
		}
		result = append(result, decoded...)
	}
	if len(result) != declaredLength {
		return nil, fmt.Errorf("raw RDATA length mismatch")
	}
	return result, nil
}

func extractECHFromSVCB(data []byte) []byte {
	if len(data) < 3 {
		return nil
	}
	data = data[2:]
	for len(data) > 0 {
		length := int(data[0])
		data = data[1:]
		if length == 0 {
			break
		}
		if length > len(data) {
			return nil
		}
		data = data[length:]
	}
	for len(data) >= 4 {
		key := binary.BigEndian.Uint16(data[:2])
		length := int(binary.BigEndian.Uint16(data[2:4]))
		data = data[4:]
		if length > len(data) {
			return nil
		}
		if key == 5 {
			return append([]byte(nil), data[:length]...)
		}
		data = data[length:]
	}
	return nil
}
