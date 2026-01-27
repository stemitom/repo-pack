mod github;

pub use github::GitHubProvider;

use crate::error::RepoPackError;
use crate::url::ParsedUrl;

pub trait Provider {
    fn list_files(
        &self,
        parsed_url: &mut ParsedUrl,
        token: Option<&str>,
    ) -> impl std::future::Future<Output = Result<Vec<String>, RepoPackError>> + Send;
}
