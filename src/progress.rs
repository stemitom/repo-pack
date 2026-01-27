use indicatif::{ProgressBar, ProgressStyle};
use std::time::Duration;

/// Progress reporter for download operations.
pub struct DownloadProgress {
    bar: ProgressBar,
    silent: bool,
}

impl DownloadProgress {
    /// Creates a new progress bar with the given total file count.
    ///
    /// If `silent` is true, the progress bar is hidden but still tracks counts.
    pub fn new(total: u64, silent: bool) -> Self {
        let bar = if silent {
            ProgressBar::hidden()
        } else {
            ProgressBar::new(total)
        };

        bar.set_style(
            ProgressStyle::with_template(
                "{spinner:.green} {msg} [{bar:40.cyan/dim}] {pos}/{len} ({per_sec})",
            )
            .expect("valid template")
            .progress_chars("━━─"),
        );
        bar.set_message("Downloading");
        bar.enable_steady_tick(Duration::from_millis(100));

        Self { bar, silent }
    }

    /// Increments the progress bar by one.
    pub fn inc(&self) {
        self.bar.inc(1);
    }

    /// Updates the current file being downloaded.
    pub fn set_current_file(&self, path: &str) {
        if !self.silent {
            self.bar.set_message(format!("Downloading {path}"));
        }
    }

    /// Finishes the progress bar with a completion message.
    pub fn finish(&self) {
        self.bar.set_message("Downloaded");
        self.bar.finish();
    }

    /// Abandons the progress bar (for cancellation).
    pub fn abandon(&self) {
        self.bar.abandon();
    }
}
