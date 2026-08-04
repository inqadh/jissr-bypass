package jissrbypass

import "testing"

func TestNewClient(t *testing.T) {
	if NewClient() == nil {
		t.Fatal("NewClient returned nil")
	}
}
