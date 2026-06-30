package workspace

import "testing"

// TestDefaultBranchName documents branch-naming behavior from the docs
// (README.md "Branch Naming" and ROADMAP.md "Design Notes").
//
// Workspace names are validated to be branch-safe at creation time
// (see TestCreateWorkspaceNameValidation), so defaultBranchName does NOT
// transform the name — it is a passthrough that only applies the optional
// prefix:
//
//   - default (no prefix): branch name == workspace name, verbatim (case preserved)
//   - an optional configurable prefix is prepended when set
func TestDefaultBranchName(t *testing.T) {
	tests := []struct {
		name   string
		wsPath string
		prefix string
		want   string
	}{
		{
			name:   "no prefix: branch name equals workspace name",
			wsPath: "/home/you/Workspaces/my-feature",
			prefix: "",
			want:   "my-feature",
		},
		{
			name:   "case is preserved (names are validated, not lowercased)",
			wsPath: "/home/you/Workspaces/MyFeature",
			prefix: "",
			want:   "MyFeature",
		},
		{
			name:   "optional prefix is applied when configured",
			wsPath: "/home/you/Workspaces/JIRA-1234",
			prefix: "ws/",
			want:   "ws/JIRA-1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := Workspace{Path: tt.wsPath}
			config := Config{BranchPrefix: tt.prefix}

			got := defaultBranchName(ws, config)
			if got != tt.want {
				t.Errorf("defaultBranchName(%q, prefix %q) = %q, want %q",
					tt.wsPath, tt.prefix, got, tt.want)
			}
		})
	}
}
