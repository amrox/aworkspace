package workspace

import "testing"

func TestParseRepoURL(t *testing.T) {

	tests := []struct {
		name string
		url  string
		want Repo
	}{
		{
			"https with .git",
			"https://github.com/amrox/aworkspace.git",
			Repo{Host: "github.com", Namespace: "amrox", Name: "aworkspace", Url: "https://github.com/amrox/aworkspace.git"},
		},
		{
			"https without .git",
			"https://github.com/amrox/aworkspace",
			Repo{Host: "github.com", Namespace: "amrox", Name: "aworkspace", Url: "https://github.com/amrox/aworkspace"},
		},
		{
			"https with deeper path",
			"https://gitlab.com/org/subgroup/repo.git",
			Repo{Host: "gitlab.com", Namespace: "org/subgroup", Name: "repo", Url: "https://gitlab.com/org/subgroup/repo.git"},
		},
		{
			"https with shallow path",
			"https://example.com/repo.git",
			Repo{Host: "example.com", Namespace: "", Name: "repo", Url: "https://example.com/repo.git"},
		},
		{
			"ssh with .git",
			"git@github.com:amrox/aworkspace.git",
			Repo{Host: "github.com", Namespace: "amrox", Name: "aworkspace", Url: "git@github.com:amrox/aworkspace.git"},
		},
		{
			"ssh without .git",
			"git@github.com:amrox/aworkspace",
			Repo{Host: "github.com", Namespace: "amrox", Name: "aworkspace", Url: "git@github.com:amrox/aworkspace"},
		},
		{
			"file URL",
			"file:///tmp/repo.git",
			Repo{Host: "_", Namespace: "tmp", Name: "repo", Url: "file:///tmp/repo.git"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRepoURL(tt.url)
			if err != nil {
				t.Fatalf("parseRepoURL(%q) unexpected error: %v", tt.url, err)
			}
			if got != tt.want {
				t.Errorf("parseRepoURL(%q) = %+v, want %+v", tt.url, got, tt.want)
			}
		})
	}
}

func TestBareRepoSubpath(t *testing.T) {

	tests := []struct {
		name string
		repo Repo
		subpath string
	}{
		{
			"with simple namespace",
			Repo{Host: "github.com", Namespace: "amrox", Name: "aworkspace", Url: "https://github.com/amrox/aworkspace.git"},
			"github.com/amrox/aworkspace",
		},
		{
			"with deeper deeper namespace",
			Repo{Host: "gitlab.com", Namespace: "org/subgroup", Name: "repo", Url: "https://gitlab.com/org/subgroup/repo.git"},
			"gitlab.com/org/subgroup/repo",
		},
		{
			"file URL",
			Repo{Host: "_", Namespace: "tmp", Name: "repo", Url: "file:///tmp/repo.git"},
			"_/tmp/repo",
		},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.repo.bareRepoSubpath()
			if got != tt.subpath {
				t.Errorf("bareRepoSubpath(%q) = %+v, want %+v", tt.repo, got, tt.subpath)
			}
		})
	}
}
