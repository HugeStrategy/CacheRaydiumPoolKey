// src/modules/downloader.rs
use anyhow::Result;
use indicatif::{ProgressBar, ProgressStyle};
use tokio::io::AsyncWriteExt;

pub async fn download_file(url: &str, filepath: &str) -> Result<()> {
    let client = reqwest::Client::new();
    let response = client.get(url).send().await?;
    let total_size = response.content_length().unwrap_or(0);

    let pb = ProgressBar::new(total_size);
    pb.set_style(ProgressStyle::default_bar()
        .template("{spinner:.green} [{elapsed_precise}] [{bar:40.cyan/blue}] {bytes}/{total_bytes} ({eta})")?
        .progress_chars("#>-"));

    let mut file = tokio::fs::File::create(filepath).await?;
    let mut downloaded: u64 = 0;

    let content = response.bytes().await?;
    file.write_all(&content).await?;
    downloaded = content.len() as u64;
    pb.set_position(downloaded);

    pb.finish_with_message("Download completed");
    Ok(())
}
