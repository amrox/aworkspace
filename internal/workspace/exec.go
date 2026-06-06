package workspace

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type prefixWriter struct {
	prefix      []byte
	w           io.Writer
	wrotePrefix bool
}

func (pw *prefixWriter) Write(p []byte) (int, error) {

	lastIdx := 0
	for idx, b := range p {
		if !pw.wrotePrefix {
			n, err := pw.w.Write(pw.prefix)
			if err != nil {
				return n, err
			}
			pw.wrotePrefix = true
		}

		if b == '\n' {
			n, err := pw.w.Write(p[lastIdx : idx+1])
			if err != nil {
				return n, err
			}
			pw.wrotePrefix = false
			lastIdx = idx + 1
		}
	}

	if lastIdx < len(p) {
		n, err := pw.w.Write(p[lastIdx:])
		if err != nil {
			return n, err
		}
	}

	return len(p), nil
}

func execCommand(path string, args ...string) (string, error) {

	var buf bytes.Buffer
	var w io.Writer = &buf

	exe := filepath.Base(path)

	if CurLogLevel >= LogLevelVerbose {
		pw := &prefixWriter{
			prefix: []byte("[" + exe + "] "),
			w:      os.Stderr,
		}
		w = io.MultiWriter(&buf, pw)
	}

	LogVerbose("Executing command: %q\n", append([]string{path}, args...))

	cmd := exec.Command(path, args...)
	cmd.Stdout = w
	cmd.Stderr = w
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("subcommand failed: %q: %w", buf.String(), err)
	}
	return buf.String(), nil
}

