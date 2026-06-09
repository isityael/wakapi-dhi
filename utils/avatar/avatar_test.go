package avatar

import (
	"strings"
	"testing"
)

func TestMakeReturnsDeterministicSVG(t *testing.T) {
	first := Make("user@example.com")
	second := Make("user@example.com")

	if first != second {
		t.Fatal("expected avatar SVG to be deterministic for the same seed")
	}
	if !strings.HasPrefix(first, `<svg xmlns="http://www.w3.org/2000/svg"`) {
		t.Fatalf("expected SVG document, got %q", first[:min(len(first), 64)])
	}
	if !strings.Contains(first, "<rect") || !strings.Contains(first, "<circle") {
		t.Fatalf("expected geometric avatar SVG, got %s", first)
	}
}

func TestMakeVariesBySeed(t *testing.T) {
	first := Make("alice")
	second := Make("bob")

	if first == second {
		t.Fatal("expected different seeds to render different avatars")
	}
}
