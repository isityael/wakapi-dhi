package badge

import (
	"strings"
	"testing"
)

func TestRenderNormalizesNamedColor(t *testing.T) {
	svg := Render("coding", "1 hr", "green")

	if !strings.HasPrefix(string(svg), `<svg xmlns="http://www.w3.org/2000/svg"`) {
		t.Fatalf("expected SVG output, got %q", string(svg[:min(len(svg), 64)]))
	}
	if !strings.Contains(string(svg), ">coding<") || !strings.Contains(string(svg), ">1 hr<") {
		t.Fatalf("expected label and message in SVG, got %s", string(svg))
	}
	if !strings.Contains(string(svg), "#4c1") {
		t.Fatalf("expected green color scheme to render as #4c1, got %s", string(svg))
	}
}

func TestRenderPrefixesBareHexColor(t *testing.T) {
	svg := Render("label", "value", "ff00aa")

	if !strings.Contains(string(svg), "#ff00aa") {
		t.Fatalf("expected bare hex color to be prefixed, got %s", string(svg))
	}
}

func TestRenderEscapesText(t *testing.T) {
	svg := Render(`<label>`, `a&b`, "blue")

	if strings.Contains(string(svg), "<label>") || strings.Contains(string(svg), "a&b") {
		t.Fatalf("expected text content to be escaped, got %s", string(svg))
	}
	if !strings.Contains(string(svg), "&lt;label&gt;") || !strings.Contains(string(svg), "a&amp;b") {
		t.Fatalf("expected escaped text in SVG, got %s", string(svg))
	}
}
