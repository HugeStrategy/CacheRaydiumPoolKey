package main

import (
	"RaydiumSync/internal/downloader"
	"RaydiumSync/internal/parser"
	"RaydiumSync/internal/redis"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var logger = logrus.New()

var redisAddr = "localhost:6379"
var redisPassword = ""
var redisDB = 0

var jsonURL = "https://api.raydium.io/v2/sdk/liquidity/mainnet.json"
var outputFile = "mainnet.json"
var programID = "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
var quoteMint = "So11111111111111111111111111111111111111112"

func init() {
	// 配置日志
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true, // 启用颜色
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 设置日志级别
	logger.SetLevel(logrus.InfoLevel)
}

var rootCmd = &cobra.Command{
	Use:   "RaydiumSync",
	Short: "Raydium CLI to cache pool keys",
	Long:  `A CLI tool for downloading, parsing, filtering, and caching Raydium liquidity pool keys to Redis.`,
	Run: func(cmd *cobra.Command, args []string) {

		// 1. 连接 Redis
		client := redis.NewRedisClient(redisAddr, redisPassword, redisDB)

		// 2. 下载 JSON 文件
		logger.Info("Downloading JSON file...")
		if err := downloader.DownloadFile(jsonURL, outputFile); err != nil {
			logger.Fatalf("Failed to download file: %v", err)
		}
		logger.Info("Download completed successfully.")

		// 3. 解析并筛选 JSON 数据
		logger.Info("Parsing and filtering JSON data...")
		pools, err := parser.ParseAndFilter(outputFile, programID, quoteMint)
		if err != nil {
			logger.Fatalf("Failed to parse or filter JSON: %v", err)
		}

		// 4. 将筛选的数据存入 Redis
		logger.Info("Processing data and storing in Redis...")
		for _, pool := range pools {
			key := pool.BaseMint
			value := fmt.Sprintf("%s,%s,%s", pool.ID, pool.BaseVault, pool.QuoteVault)
			success, err := client.SetIfNotExists(key, value)
			if success {
				if err != nil {
					logger.Warnf("Failed to set key %s in Redis: %v", key, err)
				} else {
					logger.Infof("Stored key %s in Redis", key)
				}
			} else {
				logger.Warnf("Key %s already exists in Redis", key)
			}
		}

		logger.Info("All data processed and stored in Redis.")

		// 5. 从Solana RPC监听新的AMM Pool创建并存入Redis
		// logger.Info("Listening to Solana RPC for new AMM Pool creation...")
		// subscribe.SubscribeAMMPoolCreate(wsURL)
	},
}

func main() {
	// 执行 CLI 命令
	if err := rootCmd.Execute(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}
