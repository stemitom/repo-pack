use indicatif::{ProgressBar, ProgressStyle};
use std::sync::atomic::{AtomicU64, Ordering};
use std::time::Duration;

/// Progress reporter for download operations.
pub struct DownloadProgress {
    bar: ProgressBar,
    silent: bool,
    verbose: bool,
    total: u64,
    count: AtomicU64,
}

impl DownloadProgress {
    /// Creates a new progress bar with the given total file count.
    ///
    /// If `silent` is true, the progress bar is hidden but still tracks counts.
    /// If `verbose` is true, prints each file as it completes.
    pub fn new(total: u64, silent: bool, verbose: bool) -> Self {
        let bar = if silent || verbose {
            ProgressBar::hidden()
        } else {
            ProgressBar::new(total)
        };

        bar.set_style(
            ProgressStyle::with_template("Downloading [{bar:20.cyan/dim}] {pos}/{len}  {msg}")
                .expect("valid template")
                .progress_chars("██░"),
        );

        bar.enable_steady_tick(Duration::from_millis(100));

        if verbose && !silent {
            println!("Downloading {total} files...");
        }

        Self {
            bar,
            silent,
            verbose,
            total,
            count: AtomicU64::new(0),
        }
    }

    /// Increments the progress bar by one.
    pub fn inc(&self) {
        self.bar.inc(1);
        self.count.fetch_add(1, Ordering::Relaxed);
    }

    /// Updates the current file being downloaded (default mode) or prints completion (verbose).
    pub fn set_current_file(&self, path: &str) {
        if self.silent {
            return;
        }

        if self.verbose {
            let pos = self.count.load(Ordering::Relaxed) + 1;
            println!("  [{}/{}] {} ✓", pos, self.total, path);
        } else {
            let display_path = truncate_path(path, 40);
            self.bar.set_message(display_path);
        }
    }

    /// Finishes the progress bar with a completion message.
    pub fn finish(&self) {
        self.bar.finish_and_clear();
    }

    /// Abandons the progress bar (for cancellation).
    pub fn abandon(&self) {
        self.bar.abandon();
    }
}

fn truncate_path(path: &str, max_len: usize) -> String {
    if path.len() <= max_len {
        return path.to_string();
    }

    let ellipsis = "…/";
    let available = max_len - ellipsis.len();

    if let Some(pos) = path[path.len() - available..].find('/') {
        format!("{ellipsis}{}", &path[path.len() - available + pos + 1..])
    } else {
        format!("{ellipsis}{}", &path[path.len() - available..])
    }
}
