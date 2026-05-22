package workspace

import (
	"bytes"
	"testing"
)

func TestPrefixWriter(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		writes []string
		want   string
	}{
		{
			name:   "single line no newline",
			prefix: "[git] ",
			writes: []string{"hello"},
			want:   "[git] hello",
		},
		{
			name:   "single line with newline",
			prefix: "[git] ",
			writes: []string{"hello\n"},
			want:   "[git] hello\n",
		},
		{
			name:   "two lines",
			prefix: ">> ",
			writes: []string{"line1\nline2\n"},
			want:   ">> line1\n>> line2\n",
		},
		{
			name:   "multiple writes",
			prefix: "| ",
			writes: []string{"hello\n", "world\n"},
			want:   "| hello\n| world\n",
		},
		{
			name:   "newline split across writes",
			prefix: ": ",
			writes: []string{"ab", "cd\nef"},
			want:   ": abcd\n: ef",
		},
		{
			name:   "empty write",
			prefix: "[git] ",
			writes: []string{""},
			want:   "",
		},
		{
			name:   "only newline",
			prefix: ">> ",
			writes: []string{"\n"},
			want:   ">> \n",
		},
		{
			name:   "multiple newlines",
			prefix: "| ",
			writes: []string{"\n\n"},
			want:   "| \n| \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			pw := &prefixWriter{
				prefix: []byte(tt.prefix),
				w:      &buf,
			}
			for _, s := range tt.writes {
				_, err := pw.Write([]byte(s))
				if err != nil {
					t.Fatalf("Write() error: %v", err)
				}
			}
			got := buf.String()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
