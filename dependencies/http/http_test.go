package http

import "testing"

func TestNewDefaultClient(t *testing.T) {
	c := NewDefaultClient()
	if c == nil {
		t.Fail()
	}
}
