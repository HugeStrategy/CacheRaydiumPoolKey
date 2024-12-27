// src/modules/parser.rs
use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::fs::File;

#[derive(Debug, Serialize, Deserialize)]
pub struct Pool {
    pub id: String,
    #[serde(rename = "baseMint")]
    pub base_mint: String,
    #[serde(rename = "baseVault")]
    pub base_vault: String,
    #[serde(rename = "quoteVault")]
    pub quote_vault: String,
    #[serde(rename = "programId")]
    pub program_id: String,
    #[serde(rename = "quoteMint")]
    pub quote_mint: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct RootData {
    official: Vec<Pool>,
    #[serde(rename = "unOfficial")]
    unofficial: Vec<Pool>,
}

pub fn parse_and_filter(filepath: &str, program_id: &str, quote_mint: &str) -> Result<Vec<Pool>> {
    let file = File::open(filepath)?;
    let data: RootData = serde_json::from_reader(file)?;

    let mut filtered_pools = Vec::new();

    // Process official pools
    filtered_pools.extend(
        data.official
            .into_iter()
            .filter(|pool| pool.program_id == program_id && pool.quote_mint == quote_mint),
    );

    // Process unofficial pools
    filtered_pools.extend(
        data.unofficial
            .into_iter()
            .filter(|pool| pool.program_id == program_id && pool.quote_mint == quote_mint),
    );

    Ok(filtered_pools)
}
