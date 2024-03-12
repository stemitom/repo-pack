package helpers_test

import (
	"repo-pack/helpers"
	"repo-pack/model"
	"testing"
)

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
