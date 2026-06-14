package utils

import (
	"regexp"
	"strings"
)

var wildcardEscaper = strings.NewReplacer(
	`\\`, `\\`,
	`.`, `\.`,
	`+`, `\+`,
	`(`, `\(`,
	`)`, `\)`,
	`|`, `\|`,
	`^`, `\^`,
	`$`, `\$`,
	`{`, `\{`,
	`}`, `\}`,
	`[`, `\[`,
	`]`, `\]`,
)

// WildcardMatch implements the limited "*" and "?" matching used by Wakapi's
// alias and import-host patterns while avoiding an extra third-party matcher.
func WildcardMatch(pattern, value string) bool {
	regex := wildcardEscaper.Replace(pattern)
	regex = strings.ReplaceAll(regex, "*", ".*")
	regex = strings.ReplaceAll(regex, "?", ".")
	return regexp.MustCompile("^" + regex + "$").MatchString(value)
}
