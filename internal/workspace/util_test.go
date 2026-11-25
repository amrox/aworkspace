package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsUnderDir(t *testing.T) {

	// EvalSymlinks requires the paths to exist, so create them.
	base := t.TempDir()
	foo := filepath.Join(base, "foo")
	foobar := filepath.Join(base, "foobar") // shares the "foo" prefix on purpose
	child := filepath.Join(foo, "a", "b")
	for _, d := range []string{foo, foobar, child} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	cases := []struct {
		name string
		dir  string
		path string
		want bool
	}{
		{"child is under dir", foo, child, true},
		{"dir equal to itself is under", foo, foo, true},
		{"prefix sibling is not under", foo, foobar, false}, // /foo vs /foobar
		{"parent is not under", foo, base, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := isUnderDir(tc.dir, tc.path)
			if err != nil {
				t.Fatalf("isUnderDir: %v", err)
			}
			if got != tc.want {
				t.Errorf("isUnderDir(%q, %q) = %v, want %v", tc.dir, tc.path, got, tc.want)
			}
		})
	}

	t.Run("resolves symlinks before comparing", func(t *testing.T) {
		link := filepath.Join(base, "link")
		if err := os.Symlink(foo, link); err != nil {
			t.Fatalf("symlink: %v", err)
		}
		// link/a resolves to foo/a, which is under foo.
		got, err := isUnderDir(foo, filepath.Join(link, "a"))
		if err != nil {
			t.Fatalf("isUnderDir: %v", err)
		}
		if !got {
			t.Errorf("symlinked path should resolve to under dir")
		}
	})
}
