package avatar

import (
	"fmt"
	"hash/fnv"
	"strings"
)

var palette = []string{
	"#2563eb",
	"#0891b2",
	"#059669",
	"#7c3aed",
	"#db2777",
	"#ea580c",
	"#4f46e5",
	"#0d9488",
}

// Make returns a deterministic geometric SVG avatar for the supplied seed.
func Make(seed string) string {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(strings.TrimSpace(strings.ToLower(seed))))
	sum := hash.Sum64()

	bg := palette[int(sum%uint64(len(palette)))]
	accent := palette[int((sum>>8)%uint64(len(palette)))]
	light := palette[int((sum>>16)%uint64(len(palette)))]

	cx := 24 + int((sum>>24)%17)
	cy := 24 + int((sum>>32)%17)
	r := 10 + int((sum>>40)%11)

	return fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" role="img"><rect width="64" height="64" fill="%s"/><circle cx="%d" cy="%d" r="%d" fill="%s" opacity="0.78"/><circle cx="%d" cy="%d" r="%d" fill="%s" opacity="0.42"/><rect x="8" y="46" width="48" height="10" rx="5" fill="%s" opacity="0.54"/></svg>`,
		bg,
		cx,
		cy,
		r,
		light,
		64-cx,
		64-cy,
		max(6, r-4),
		accent,
		light,
	)
}
