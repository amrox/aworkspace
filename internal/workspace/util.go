package workspace

import (
	"os"
	"path/filepath"
	"strings"
)

func CanonicalPath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs) // resolves symlinks + case; needs the path to exist
}

func isUnderDir(dir, path string) (bool, error) {
	d, err := CanonicalPath(dir)
	if err != nil {
		return false, err
	}
	p, err := CanonicalPath(path)
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(d, p)
	if err != nil {
		return false, nil // different volumes → not under
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)), nil
}
