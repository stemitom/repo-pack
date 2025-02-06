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
			name: "valid URL with simple path",
			url:  "https://github.com/owner/repo/tree/main/dir",
			expected: model.RepoURLComponents{
				Owner:      "owner",
				Repository: "repo",
				Ref:        "main",
				Dir:        "dir",
			},
		},
		{
			name: "branch with slash",
			url:  "https://github.com/owner/repo/tree/feat/new-feature/path/to/dir",
			expected: model.RepoURLComponents{
				Owner:      "owner",
				Repository: "repo",
				Ref:        "feat/new-feature",
				Dir:        "path/to/dir",
			},
		},
		{
			name: "url with special characters",
			url:  "https://github.com/user/proj/tree/main/docs%20%26%20resources",
			expected: model.RepoURLComponents{
				Owner:      "user",
				Repository: "proj",
				Ref:        "main",
				Dir:        "docs & resources",
			},
		},
		{
			name:        "invalid URL format",
			url:         "https://github.com/owner/repo/blob/main/file.txt",
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
			},
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
	expectedErr := "invalid URL format: invalid-url"

	components, err := helpers.ParseRepoURL(url)
	if err == nil {
		t.Errorf("expected error: %s, got: nil", expectedErr)
	} else if err.Error() != expectedErr {
		t.Errorf("expected error: %s, got: %v", expectedErr, err)
	}

	if components != expected {
		t.Errorf("expected components: %+v, got: %+v", expected, components)
	}
}

func TestParseRepoInvalidURLFormat(t *testing.T) {
	url := "https://github.com/owner/repo/blob/main/file.txt"
	expected := model.RepoURLComponents{}
	expectedErr := "invalid URL format: https://github.com/owner/repo/blob/main/file.txt"

	components, err := helpers.ParseRepoURL(url)
	if err == nil {
		t.Errorf("expected error: %s, got: nil", expectedErr)
	} else if err.Error() != expectedErr {
		t.Errorf("expected error: %s, got: %v", expectedErr, err)
	}

	if components != expected {
		t.Errorf("expected components: %+v, got: %+v", expected, components)
	}
}
