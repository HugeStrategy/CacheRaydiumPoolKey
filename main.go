package main

import (
	"RaydiumSync/internal/downloader"
	"RaydiumSync/internal/log"
	"RaydiumSync/internal/parser"
	"RaydiumSync/internal/redis"
	"RaydiumSync/internal/subscribe"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var redisAddr = "localhost:6379"
var redisPassword = ""
var redisDB = 0

var jsonURL = "https://api.raydium.io/v2/sdk/liquidity/mainnet.json"
var outputFile = "mainnet.json"
var programID = "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
var solAddress = "So11111111111111111111111111111111111111112"

var rootCmd = &cobra.Command{
	Use:   "RaydiumSync",
	Short: "Raydium CLI to cache pool keys",
	Long:  `A CLI tool for downloading, parsing, filtering, and caching Raydium liquidity pool keys to Redis.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. 连接 Redis
		client := redis.NewRedisClient(redisAddr, redisPassword, redisDB)

		// 2. 下载 JSON 文件
		log.Logger.Info("Downloading JSON file...")
		if err := downloader.DownloadFile(jsonURL, outputFile); err != nil {
			log.Logger.Fatalf("Failed to download file: %v", err)
		}
		log.Logger.Info("Download completed successfully.")

		// 3. 解析并筛选 JSON 数据
		log.Logger.Info("Parsing and filtering JSON data...")
		pools, err := parser.ParseAndFilter(outputFile, programID, solAddress)
		if err != nil {
			log.Logger.Fatalf("Failed to parse or filter JSON: %v", err)
		}

		// 4. 将筛选的数据存入 Redis
		log.Logger.Infof("Processing %d data and storing in Redis...", len(pools))
		batchData := make(map[string]string)
		ScammerPoolCount := 0
		for _, pool := range pools {
			key := pool.QuoteMint
			value := fmt.Sprintf("%s,%s,%s", pool.ID, pool.BaseVault, pool.QuoteVault)
			if batchData[key] != "" {
				log.Logger.Warnf("Overwrite duplicate key: %v", key)
				ScammerPoolCount++
			}
			batchData[key] = value
		}
		log.Logger.Infof("All pool count: %v", len(pools))
		log.Logger.Infof("Scammer pool count: %v", ScammerPoolCount)
		log.Logger.Infof("effective pool count: %v", len(batchData))
		// 批量写入 Redis
		if err := client.BatchSet(batchData); err != nil {
			log.Logger.Fatalf("Failed to store data in Redis: %v", err)
		}
		log.Logger.Info("All data processed and stored in Redis.")

		//5. 从YellowStone GRPC监听新的AMM Pool创建并存入Redis
		log.Logger.Info("Listening to Solana RPC for new AMM Pool creation...")
		subscribe.SubscribeAMMPoolCreate("solana-yellowstone-grpc.publicnode.com:443", *client)
	},
}

func main() {
	// 执行 CLI 命令
	if err := rootCmd.Execute(); err != nil {
		log.Logger.Error(err)
		os.Exit(1)
	}
}
