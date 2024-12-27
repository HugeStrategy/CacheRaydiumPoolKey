use anyhow::Result;
use bytes::BytesMut;
use futures::Stream;
use serde::{Deserialize, Serialize};
use tokio::fs::File;
use tokio::io::{BufReader};
use tokio_stream::StreamExt;
use tokio_util::codec::{Decoder, FramedRead};

const BUFFER_SIZE: usize = 16 * 1024; // 16KB buffer

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

struct JsonDecoder {
    buffer: BytesMut,
}

impl JsonDecoder {
    fn new() -> Self {
        Self {
            buffer: BytesMut::with_capacity(BUFFER_SIZE),
        }
    }
}

impl Decoder for JsonDecoder {
    type Item = Pool;
    type Error = anyhow::Error;

    fn decode(&mut self, src: &mut BytesMut) -> Result<Option<Self::Item>, Self::Error> {
        if src.is_empty() {
            return Ok(None);
        }

        match serde_json::from_slice(src) {
            Ok(pool) => {
                src.clear();
                Ok(Some(pool))
            }
            Err(_) => Ok(None),
        }
    }
}

pub async fn parse_and_filter_stream(
    filepath: &str,
    program_id: &str,
    quote_mint: &str,
) -> Result<impl Stream<Item = Result<Pool>>> {
    let file = File::open(filepath).await?;
    let reader = BufReader::with_capacity(BUFFER_SIZE, file);

    let stream = FramedRead::new(reader, JsonDecoder::new())
        .filter(move |pool_result| {
            let pool = match pool_result {
                Ok(pool) => pool,
                Err(_) => return false,
            };
            pool.program_id == program_id && pool.quote_mint == quote_mint
        });

    Ok(stream)
}

// 为了兼容性，保留同步版本
pub fn parse_and_filter(filepath: &str, program_id: &str, quote_mint: &str) -> Result<Vec<Pool>> {
    let file = std::fs::File::open(filepath)?;
    let data: RootData = serde_json::from_reader(file)?;

    let mut filtered_pools = Vec::new();

    // Process official pools with pre-allocated capacity
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