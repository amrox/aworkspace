package workspace

import (
	"reflect"
	"testing"
)

func TestParseWorktreeList(t *testing.T) {

	t.Run("bare, prunable, and live records", func(t *testing.T) {
		out := `worktree /repos/dotfiles
bare

worktree /ws/hi/code/dotfiles
HEAD fbb7487ee3d69ee86cfe75e83febca1858f6d3b4
branch refs/heads/ws/hi
prunable gitdir file points to non-existent location

worktree /ws/test1/code/dotfiles
HEAD 019a06151d0e0fdcc8870dacb967623493400712
branch refs/heads/ws/test1
`
		got := parseWorktreeList(out)
		want := []Worktree{
			{Path: "/repos/dotfiles", Bare: true},
			{Path: "/ws/hi/code/dotfiles", HEAD: "fbb7487ee3d69ee86cfe75e83febca1858f6d3b4", Branch: "ws/hi", Prunable: true},
			{Path: "/ws/test1/code/dotfiles", HEAD: "019a06151d0e0fdcc8870dacb967623493400712", Branch: "ws/test1"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("parseWorktreeList:\n got %+v\nwant %+v", got, want)
		}
	})

	t.Run("detached worktree has empty branch", func(t *testing.T) {
		out := `worktree /ws/x/code/repo
HEAD abc123
detached
`
		got := parseWorktreeList(out)
		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}
		if got[0].Branch != "" {
			t.Errorf("Branch = %q, want empty for detached", got[0].Branch)
		}
	})

	t.Run("empty input yields no worktrees", func(t *testing.T) {
		if got := parseWorktreeList(""); len(got) != 0 {
			t.Fatalf("len = %d, want 0", len(got))
		}
	})
}
