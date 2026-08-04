package font

import (
	"os"
	"strings"
	"testing"
)

// TestRootIsWasmSafe closes the SPECS invariant: the root is pure identity and
// crosses to WASM whole, so nothing here may pull the runtime with it. It
// scans the root's non-test .go files and fails on any import, any occurrence
// of os./embed/[]byte/map[, or a build directive in either of its two forms.
// Test files are exempt: they never enter a binary. The day the root needs an
// import, changing this test is the explicit decision that authorizes it.
func TestRootIsWasmSafe(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(src), "\n") {
			if strings.Contains(line, "import") {
				t.Errorf("%s: any import in the root is forbidden: %q", name, line)
			}
			for _, tok := range []string{"os.", "embed", "[]byte", "map["} {
				if strings.Contains(line, tok) {
					t.Errorf("%s: %q is forbidden in the root: %q", name, tok, line)
				}
			}
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//go:build") || strings.HasPrefix(trimmed, "// +build") {
				t.Errorf("%s: build directives are forbidden in the root: %q", name, line)
			}
		}
	}
}
