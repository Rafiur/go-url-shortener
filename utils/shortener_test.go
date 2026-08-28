package utils

import (
	"regexp"
	"testing"
)

// Short codes end up in URLs, so every character must survive a path segment
// untouched.
var urlSafe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func TestGenerateShortCodeLength(t *testing.T) {
	for _, n := range []int{4, 7, 12} {
		code := GenerateShortCode(n)

		if len(code) != n {
			t.Errorf("GenerateShortCode(%d) = %q, want length %d, got %d", n, code, n, len(code))
		}
		if !urlSafe.MatchString(code) {
			t.Errorf("GenerateShortCode(%d) = %q, contains characters unsafe for a URL path", n, code)
		}
	}
}

func TestGenerateShortCodeIsRandom(t *testing.T) {
	const draws = 500

	seen := make(map[string]bool, draws)
	for i := 0; i < draws; i++ {
		code := GenerateShortCode(7)
		if seen[code] {
			t.Fatalf("GenerateShortCode produced duplicate %q within %d draws", code, draws)
		}
		seen[code] = true
	}
}
