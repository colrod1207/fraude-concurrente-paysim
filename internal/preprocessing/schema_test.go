package preprocessing

import "testing"

func TestHeader_MatchesExpectedColumnCountAndOrder(t *testing.T) {
	h := Header()
	if len(h) != 16 {
		t.Fatalf("expected 16 columns, got %d: %v", len(h), h)
	}
	if h[0] != "step" || h[len(h)-1] != "hour" {
		t.Errorf("expected header to start with %q and end with %q, got %v", "step", "hour", h)
	}
}
