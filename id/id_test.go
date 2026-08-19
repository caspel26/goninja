package id

import "testing"

func TestNewUUID(t *testing.T) {
	a := NewUUID()
	b := NewUUID()

	if len(a) != 36 {
		t.Errorf("NewUUID() length = %d, want 36 (canonical UUID string)", len(a))
	}
	if a == b {
		t.Error("NewUUID() returned the same value twice in a row, want random v4 UUIDs")
	}
}
