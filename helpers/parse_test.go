package helpers_test

import (
	"repo-pack/helpers"
	"repo-pack/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expected    model.RepoURLComponents
		expectError bool
	}{
		{
			name: "valid tree URL with simple path",
			url:  "https://github.com/owner/repo/tree/main/dir",
			expected: model.RepoURLComponents{
				Owner:      "owner",
				Repository: "repo",
				Ref:        "main",
				Dir:        "dir",
				IsFile:     false,
			},
		},
		{
			name: "branch with slash - note: branch names with slashes cannot be disambiguated from directories in URL paths",
			url:  "https://github.com/owner/repo/tree/feat/new-feature",
			expected: model.RepoURLComponents{
				Owner:      "owner",
				Repository: "repo",
				Ref:        "feat",
				Dir:        "new-feature",
				IsFile:     false,
			},
		},
		{
			name: "nested directory structure",
			url:  "https://github.com/owner/repo/tree/main/docs/guides/getting-started",
			expected: model.RepoURLComponents{
				Owner:      "owner",
				Repository: "repo",
				Ref:        "main",
				Dir:        "docs/guides/getting-started",
				IsFile:     false,
			},
		},
		{
			name: "tree url with special characters",
			url:  "https://github.com/user/proj/tree/main/docs%20%26%20resources",
			expected: model.RepoURLComponents{
				Owner:      "user",
				Repository: "proj",
				Ref:        "main",
				Dir:        "docs & resources",
				IsFile:     false,
			},
		},
		{
			name: "blob URL - single file",
			url:  "https://github.com/owner/repo/blob/main/src/file.txt",
			expected: model.RepoURLComponents{
				Owner:      "owner",
				Repository: "repo",
				Ref:        "main",
				Dir:        "src",
				FilePath:   "src/file.txt",
				IsFile:     true,
			},
		},
		{
			name: "blob URL - file in root",
			url:  "https://github.com/owner/repo/blob/main/README.md",
			expected: model.RepoURLComponents{
				Owner:      "owner",
				Repository: "repo",
				Ref:        "main",
				Dir:        ".",
				FilePath:   "README.md",
				IsFile:     true,
			},
		},
		{
			name: "blob URL - nested file",
			url:  "https://github.com/owner/repo/blob/develop/src/components/Button.tsx",
			expected: model.RepoURLComponents{
				Owner:      "owner",
				Repository: "repo",
				Ref:        "develop",
				Dir:        "src/components",
				FilePath:   "src/components/Button.tsx",
				IsFile:     true,
			},
		},
		{
			name: "blob URL with special characters",
			url:  "https://github.com/user/proj/blob/main/docs%20%26%20files/config.yaml",
			expected: model.RepoURLComponents{
				Owner:      "user",
				Repository: "proj",
				Ref:        "main",
				Dir:        "docs & files",
				FilePath:   "docs & files/config.yaml",
				IsFile:     true,
			},
		},
		{
			name: "raw.githubusercontent.com URL",
			url:  "https://raw.githubusercontent.com/owner/repo/main/src/config.json",
			expected: model.RepoURLComponents{
				Owner:      "owner",
				Repository: "repo",
				Ref:        "main",
				Dir:        "src",
				FilePath:   "src/config.json",
				IsFile:     true,
			},
		},
		{
			name: "raw URL - file in root",
			url:  "https://raw.githubusercontent.com/owner/repo/v1.0.0/package.json",
			expected: model.RepoURLComponents{
				Owner:      "owner",
				Repository: "repo",
				Ref:        "v1.0.0",
				Dir:        ".",
				FilePath:   "package.json",
				IsFile:     true,
			},
		},
		{
			name:        "unsupported host",
			url:         "https://gitlab.com/owner/repo/tree/main/dir",
			expectError: true,
		},
		{
			name:        "invalid URL structure",
			url:         "https://example.com/not-github",
			expectError: true,
		},
		{
			name: "empty directory path",
			url:  "https://github.com/owner/repo/tree/main/",
			expected: model.RepoURLComponents{
				Owner:      "owner",
				Repository: "repo",
				Ref:        "main",
				Dir:        "",
				IsFile:     false,
			},
		},
		{
			name:        "invalid github path - no tree or blob",
			url:         "https://github.com/owner/repo/main/file.txt",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components, err := helpers.ParseRepoURL(tt.url)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, components)
		})
	}
}

func TestParseRepoValidURL(t *testing.T) {
	url := "https://github.com/owner/repo/tree/main/dir"
	expected := model.RepoURLComponents{
		Owner:      "owner",
		Repository: "repo",
		Ref:        "main",
		Dir:        "dir",
		IsFile:     false,
	}

	components, err := helpers.ParseRepoURL(url)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if components != expected {
		t.Errorf("expected components: %+v, got: %+v", expected, components)
	}
}

func TestParseRepoBlobURL(t *testing.T) {
	url := "https://github.com/owner/repo/blob/main/src/file.txt"
	expected := model.RepoURLComponents{
		Owner:      "owner",
		Repository: "repo",
		Ref:        "main",
		Dir:        "src",
		FilePath:   "src/file.txt",
		IsFile:     true,
	}

	components, err := helpers.ParseRepoURL(url)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if components != expected {
		t.Errorf("expected components: %+v, got: %+v", expected, components)
	}
}

func TestParseRepoRawURL(t *testing.T) {
	url := "https://raw.githubusercontent.com/owner/repo/main/config.yaml"
	expected := model.RepoURLComponents{
		Owner:      "owner",
		Repository: "repo",
		Ref:        "main",
		Dir:        ".",
		FilePath:   "config.yaml",
		IsFile:     true,
	}

	components, err := helpers.ParseRepoURL(url)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if components != expected {
		t.Errorf("expected components: %+v, got: %+v", expected, components)
	}
}

func TestParseRepoInvalidURL(t *testing.T) {
	url := "invalid-url"
	expected := model.RepoURLComponents{}

	components, err := helpers.ParseRepoURL(url)
	if err == nil {
		t.Errorf("expected error but got: nil")
	}

	if components != expected {
		t.Errorf("expected components: %+v, got: %+v", expected, components)
	}
}

func TestParseRepoUnsupportedHost(t *testing.T) {
	url := "https://gitlab.com/owner/repo/tree/main/dir"
	expected := model.RepoURLComponents{}

	components, err := helpers.ParseRepoURL(url)
	if err == nil {
		t.Errorf("expected error but got: nil")
	}

	if components != expected {
		t.Errorf("expected components: %+v, got: %+v", expected, components)
	}
}

func BenchmarkParseRepoURL(b *testing.B) {
	url := "https://github.com/owner/repo/tree/main/docs/guides/getting-started"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = helpers.ParseRepoURL(url)
	}
}

func BenchmarkParseRepoURLParallel(b *testing.B) {
	url := "https://github.com/owner/repo/tree/main/docs/guides/getting-started"
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = helpers.ParseRepoURL(url)
		}
	})
}
