// src/modules/redis_client.rs
use anyhow::Result;
use redis::{AsyncCommands, Client};

pub struct RedisClient {
    client: Client,
}

impl RedisClient {
    pub fn new(addr: &str, password: &str, db: i64) -> Result<Self> {
        let url = if password.is_empty() {
            format!("redis://{}/{}?", addr, db)
        } else {
            format!("redis://:{}@{}/{}", password, addr, db)
        };

        let client = Client::open(url)?;
        Ok(RedisClient { client })
    }

    pub async fn set_if_not_exists(&self, key: &str, value: &str) -> Result<bool> {
        let mut conn = self.client.get_async_connection().await?;
        let result: bool = conn.set_nx(key, value).await?;
        Ok(result)
    }

    pub async fn get(&self, key: &str) -> Result<String> {
        let mut conn = self.client.get_async_connection().await?;
        let value: String = conn.get(key).await?;
        Ok(value)
    }
}
