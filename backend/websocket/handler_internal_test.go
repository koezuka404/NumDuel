package websocket

import (
	"testing"
)

func TestWriteJSONMarshalError(t *testing.T) {
	h := &Handler{}
	_, _, serverConn, _ := setupHubConnPair(t)
	if err := h.writeJSON(serverConn, map[string]any{"bad": make(chan int)}); err == nil {
		t.Fatal("expected marshal error")
	}
}
