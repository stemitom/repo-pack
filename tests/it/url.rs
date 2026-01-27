use repo_pack::{ParsedUrl, RepoPackError};

#[test]
fn parse_standard_url() {
    let url = "https://github.com/owner/repo/tree/main/src/lib";
    let parsed = ParsedUrl::parse(url).unwrap();
    assert_eq!(parsed.owner, "owner");
    assert_eq!(parsed.repo, "repo");
    assert_eq!(parsed.git_ref(), "main");
    assert_eq!(parsed.dir, "src/lib");
}

#[test]
fn parse_url_with_branch_name() {
    let url = "https://github.com/astral-sh/uv/tree/main/crates/uv-fs/src";
    let parsed = ParsedUrl::parse(url).unwrap();
    assert_eq!(parsed.owner, "astral-sh");
    assert_eq!(parsed.repo, "uv");
    assert_eq!(parsed.git_ref(), "main");
    assert_eq!(parsed.dir, "crates/uv-fs/src");
}

#[test]
fn parse_url_with_tag_ref() {
    let url = "https://github.com/owner/repo/tree/v1.2.3/docs";
    let parsed = ParsedUrl::parse(url).unwrap();
    assert_eq!(parsed.git_ref(), "v1.2.3");
    assert_eq!(parsed.dir, "docs");
}

#[test]
fn parse_url_with_commit_sha() {
    let url = "https://github.com/owner/repo/tree/abc123def/src";
    let parsed = ParsedUrl::parse(url).unwrap();
    assert_eq!(parsed.git_ref(), "abc123def");
}

#[test]
fn parse_url_with_deep_path() {
    let url = "https://github.com/owner/repo/tree/main/a/b/c/d/e/f";
    let parsed = ParsedUrl::parse(url).unwrap();
    assert_eq!(parsed.dir, "a/b/c/d/e/f");
}

#[test]
fn parse_url_with_single_dir() {
    let url = "https://github.com/owner/repo/tree/main/src";
    let parsed = ParsedUrl::parse(url).unwrap();
    assert_eq!(parsed.dir, "src");
}

#[test]
fn parse_blob_url_fails() {
    let url = "https://github.com/owner/repo/blob/main/file.rs";
    let result = ParsedUrl::parse(url);
    assert!(matches!(result, Err(RepoPackError::InvalidUrl { .. })));
}

#[test]
fn parse_invalid_url_fails() {
    let url = "not-a-valid-url";
    let result = ParsedUrl::parse(url);
    assert!(matches!(result, Err(RepoPackError::InvalidUrl { .. })));
}

#[test]
fn parse_non_github_url_parses_path() {
    // URL parser only validates path structure, not host
    let url = "https://gitlab.com/owner/repo/tree/main/src";
    let parsed = ParsedUrl::parse(url).unwrap();
    assert_eq!(parsed.owner, "owner");
    assert_eq!(parsed.dir, "src");
}

#[test]
fn parse_url_without_tree_uses_default_branch() {
    let url = "https://github.com/owner/repo/src/lib";
    let parsed = ParsedUrl::parse(url).unwrap();
    assert_eq!(parsed.owner, "owner");
    assert_eq!(parsed.repo, "repo");
    assert!(parsed.needs_default_branch());
    assert_eq!(parsed.dir, "src/lib");
}

#[test]
fn parse_repo_root_url() {
    let url = "https://github.com/owner/repo";
    let parsed = ParsedUrl::parse(url).unwrap();
    assert_eq!(parsed.owner, "owner");
    assert_eq!(parsed.repo, "repo");
    assert!(parsed.needs_default_branch());
    assert_eq!(parsed.dir, "");
}

#[test]
fn parse_url_with_encoded_chars() {
    let url = "https://github.com/owner/repo/tree/main/path%20with%20spaces";
    let parsed = ParsedUrl::parse(url).unwrap();
    assert_eq!(parsed.dir, "path%20with%20spaces");
}
