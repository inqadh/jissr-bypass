package jissrbypass

import (
	"reflect"
	"testing"
)

func TestNewClient(t *testing.T) {
	if NewClient() == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestDecodeHeaders(t *testing.T) {
	headers, err := decodeHeaders(`{"Authorization":"Bearer test","X-Value":["one","two"]}`)
	if err != nil {
		t.Fatal(err)
	}
	if headers.Get("Authorization") != "Bearer test" ||
		!reflect.DeepEqual(headers.Values("X-Value"), []string{"one", "two"}) {
		t.Fatalf("unexpected headers: %#v", headers)
	}
}

func TestDecodeHeadersRejectsNonStringValue(t *testing.T) {
	if _, err := decodeHeaders(`{"X-Value":123}`); err == nil {
		t.Fatal("numeric header value was accepted")
	}
}
