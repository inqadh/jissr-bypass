package jissrbypass

import (
	"reflect"
	"testing"
)

type recordingListener struct {
	event *Event
}

func (listener *recordingListener) OnEvent(event *Event) {
	listener.event = event
}

func TestNewClient(t *testing.T) {
	if NewClient() == nil {
		t.Fatal("NewClient returned nil")
	}
	if NewAutoClient() == nil {
		t.Fatal("NewAutoClient returned nil")
	}
	configured, err := NewAutoClientWithTimeout(250)
	if err != nil || configured == nil {
		t.Fatalf("NewAutoClientWithTimeout: client=%v err=%v", configured, err)
	}
	if _, err := NewAutoClientWithTimeout(0); err == nil {
		t.Fatal("zero native timeout accepted")
	}
}

func TestEventListenerAdapter(t *testing.T) {
	client := NewClient()
	listener := &recordingListener{}
	client.SetEventListener(listener)
	client.CloseIdleConnections()
	if listener.event == nil ||
		listener.event.GetType() != "connections.idle_closed" ||
		listener.event.GetStage() != "lifecycle" {
		t.Fatalf("unexpected event: %#v", listener.event)
	}
	client.SetEventListener(nil)
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
