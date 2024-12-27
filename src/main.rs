// src/main.rs
mod modules {
    pub mod downloader;
    pub mod parser;
    pub mod redis_client;
    pub mod subscribe;
}

use anyhow::Result;
use clap::Parser;
use env_logger::Env;
use log::{error, info};
use modules::{downloader, parser, redis_client::RedisClient};
use futures::StreamExt;

#[derive(Parser)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(long, default_value = "localhost:6379")]
    redis_addr: String,

    #[arg(long, default_value = "")]
    redis_password: String,

    #[arg(long, default_value_t = 0)]
    redis_db: i64,

    #[arg(long, default_value = "https://api.raydium.io/v2/sdk/liquidity/mainnet.json")]
    json_url: String,

    #[arg(long, default_value = "mainnet.json")]
    output_file: String,

    #[arg(long, default_value = "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8")]
    program_id: String,

    #[arg(long, default_value = "So11111111111111111111111111111111111111112")]
    quote_mint: String,
}

#[tokio::main]
async fn main() -> Result<()> {
    env_logger::Builder::from_env(Env::default().default_filter_or("info")).init();

    let cli = Cli::parse();

    // 1. Connect to Redis
    let client = RedisClient::new(&cli.redis_addr, &cli.redis_password, cli.redis_db)?;

    // 2. Download JSON file
    info!("Downloading JSON file...");
    downloader::download_file(&cli.json_url, &cli.output_file).await?;
    info!("Download completed successfully.");

    // 3. Parse and filter JSON data using stream
    info!("Parsing and filtering JSON data...");
    let mut pool_stream = parser::parse_and_filter_stream(
        &cli.output_file,
        &cli.program_id,
        &cli.quote_mint,
    ).await?;

    // 4. Process stream and store in Redis
    info!("Processing data and storing in Redis...");
    while let Some(pool_result) = pool_stream.next().await {
        match pool_result {
            Ok(pool) => {
                let key = pool.base_mint;
                let value = format!("{},{},{}", pool.id, pool.base_vault, pool.quote_vault);

                match client.set_if_not_exists(&key, &value).await {
                    Ok(true) => info!("Stored key {} in Redis", key),
                    Ok(false) => info!("Key {} already exists in Redis", key),
                    Err(e) => error!("Failed to set key {} in Redis: {}", key, e),
                }
            }
            Err(e) => error!("Error processing pool: {}", e),
        }
    }

    info!("All data processed and stored in Redis.");
    Ok(())
}
