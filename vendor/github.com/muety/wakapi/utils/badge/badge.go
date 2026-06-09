package badge

import (
	"fmt"
	"html"
	"strings"
	"unicode/utf8"
)

var colorScheme = map[string]string{
	"brightgreen":   "#4c1",
	"green":         "#4c1",
	"yellowgreen":   "#97ca00",
	"yellow":        "#dfb317",
	"orange":        "#fe7d37",
	"red":           "#e05d44",
	"blue":          "#007ec6",
	"grey":          "#555",
	"gray":          "#555",
	"lightgrey":     "#9f9f9f",
	"lightgray":     "#9f9f9f",
	"success":       "#4c1",
	"important":     "#fe7d37",
	"critical":      "#e05d44",
	"informational": "#007ec6",
	"inactive":      "#9f9f9f",
}

// Render returns a compact shields-style SVG badge.
func Render(label, message, color string) []byte {
	label = strings.TrimSpace(label)
	message = strings.TrimSpace(message)
	color = normalizeColor(color)

	labelWidth := textWidth(label)
	messageWidth := textWidth(message)
	totalWidth := labelWidth + messageWidth
	labelX := labelWidth / 2
	messageX := labelWidth + messageWidth/2

	return []byte(fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s: %s"><linearGradient id="s" x2="0" y2="100%%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="%d" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="%d" height="20" fill="#555"/><rect x="%d" width="%d" height="20" fill="%s"/><rect width="%d" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="11"><text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text><text x="%d" y="14">%s</text><text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text><text x="%d" y="14">%s</text></g></svg>`,
		totalWidth,
		html.EscapeString(label),
		html.EscapeString(message),
		totalWidth,
		labelWidth,
		labelWidth,
		messageWidth,
		color,
		totalWidth,
		labelX,
		html.EscapeString(label),
		labelX,
		html.EscapeString(label),
		messageX,
		html.EscapeString(message),
		messageX,
		html.EscapeString(message),
	))
}

func normalizeColor(color string) string {
	color = strings.TrimSpace(strings.ToLower(color))
	if color == "" {
		return colorScheme["blue"]
	}
	if mapped, ok := colorScheme[color]; ok {
		return mapped
	}
	if strings.HasPrefix(color, "#") {
		return color
	}
	return "#" + color
}

func textWidth(text string) int {
	return utf8.RuneCountInString(text)*7 + 10
}
