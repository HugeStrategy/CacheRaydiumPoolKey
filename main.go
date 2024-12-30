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
	"strconv"
	"time"
)

var (
	redisAddr     = "localhost:6379"
	redisPassword = ""
	redisDB       = 0
	jsonURL       = "https://api.raydium.io/v2/sdk/liquidity/mainnet.json"
	outputFile    = "mainnet.json"
	programID     = "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
	solAddress    = "So11111111111111111111111111111111111111112"
	interval      = 30 * time.Minute // default interval
)

var rootCmd = &cobra.Command{
	Use:   "RaydiumSync",
	Short: "Raydium CLI to cache pool keys",
	Long:  `A CLI tool for downloading, parsing, filtering, and caching Raydium liquidity pool keys to Redis.`,
}

var scheduledCmd = &cobra.Command{
	Use:   "scheduled [interval]",
	Short: "Scheduled synchronization mode",
	Long:  `Run the Raydium synchronization in scheduled mode with a specified interval.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := redis.NewRedisClient(redisAddr, redisPassword, redisDB)
		SyncWithRaydiumJsonFile(client)
		log.Logger.Info("Scheduled mode completed successfully.")
		if len(args) > 0 {
			intVal, err := strconv.Atoi(args[0])
			if err != nil {
				log.Logger.Errorf("Invalid interval: %v", err)
				return
			}
			interval = time.Minute * time.Duration(intVal)
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			client := redis.NewRedisClient(redisAddr, redisPassword, redisDB)
			SyncWithRaydiumJsonFile(client)
			log.Logger.Info("Scheduled mode completed successfully.")
		}
	},
}

var instantCmd = &cobra.Command{
	Use:   "instant",
	Short: "Instant synchronization mode",
	Long:  `Execute the Raydium synchronization steps instantly once.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := redis.NewRedisClient(redisAddr, redisPassword, redisDB)
		SyncWithRaydiumJsonFile(client)
		log.Logger.Info("Instant mode completed successfully.")
	},
}

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor mode",
	Long:  `Only connect to Redis and listen for new AMM Pool creations.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := redis.NewRedisClient(redisAddr, redisPassword, redisDB)
		log.Logger.Info("Listening to Solana RPC for new AMM Pool creation...")
		for {
			subscribe.SubscribeAMMPoolCreate("solana-yellowstone-grpc.publicnode.com:443", *client)
			log.Logger.Error("GRPC connection failed. Retrying...")
		}
	},
}

func SyncWithRaydiumJsonFile(client *redis.RedisClient) {
	log.Logger.Info("Downloading JSON file...")
	if err := downloader.DownloadFile(jsonURL, outputFile); err != nil {
		log.Logger.Fatalf("Failed to download file: %v", err)
	}
	log.Logger.Info("Download completed successfully.")

	log.Logger.Info("Parsing and filtering JSON data...")
	pools, err := parser.ParseAndFilter(outputFile, programID, solAddress)
	if err != nil {
		log.Logger.Fatalf("Failed to parse or filter JSON: %v", err)
	}

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
	if err := client.BatchSet(batchData); err != nil {
		log.Logger.Fatalf("Failed to store data in Redis: %v", err)
	}
	log.Logger.Info("All data processed and stored in Redis.")
}

func main() {
	rootCmd.AddCommand(scheduledCmd)
	rootCmd.AddCommand(instantCmd)
	rootCmd.AddCommand(monitorCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Logger.Error(err)
		os.Exit(1)
	}
}
