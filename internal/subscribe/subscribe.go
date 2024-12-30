package subscribe

import (
	"RaydiumSync/internal/log"
	"RaydiumSync/internal/redis"
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/keepalive"
	"time"

	"github.com/blocto/solana-go-sdk/common"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const ammProgramID = "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
const solAddress = "So11111111111111111111111111111111111111112"

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

type Pool struct {
	CA         string
	PoolID     string
	BaseVault  string
	QuoteVault string
}

func SubscribeAMMPoolCreate(grpcAddress string, redisClient redis.RedisClient) error {
	// Start GRPC connection
	conn := grpcConnect(grpcAddress)
	defer conn.Close()

	// Create Geyser client
	grpcClient := pb.NewGeyserClient(conn)

	// Set up the transactions subscription
	subscription := pb.SubscribeRequest{}

	// Subscribe Raydium Fee Account
	subscription.Transactions = make(map[string]*pb.SubscribeRequestFilterTransactions)
	subscription.Transactions["transactions_sub"] = &pb.SubscribeRequestFilterTransactions{
		AccountExclude: []string{},
		AccountInclude: []string{},
		AccountRequired: []string{
			"7YttLkHDoNj9wyDur5pM1ejNaAvT9X4eqaYcHQqtj2G5",
		},
	}
	subscriptionJson, err := json.Marshal(&subscription)
	if err != nil {
		log.Logger.Printf("Failed to marshal subscription request: %v", subscriptionJson)
	}
	log.Logger.Printf("Subscription request: %s", string(subscriptionJson))

	//Send the subscription request
	ctx := context.Background()
	stream, err := grpcClient.Subscribe(ctx)
	if err != nil {
		log.Logger.Fatalf("Failed to subscribe: %v", err)
	}
	err = stream.Send(&subscription)
	if err != nil {
		log.Logger.Fatalf("Failed to send subscription: %v", err)
	}

	// Receive and process updates
	for {
		update, err := stream.Recv()
		if err != nil {
			log.Logger.Errorf("Failed to receive update: %v", err)
			return err
		}

		// Monitor Raydium Pool Create
		if transactionUpdate, ok := update.UpdateOneof.(*pb.SubscribeUpdate_Transaction); ok {
			pool := processTransaction(transactionUpdate)
			err = redisClient.SetKeyValue(pool.CA, fmt.Sprintf("%s,%s,%s", pool.PoolID, pool.BaseVault, pool.QuoteVault))
			if err != nil {
				log.Logger.Errorf("Failed to write pool info to Redis: %v", err)
			} else {
				log.Logger.Infof("Write New Raydium Pool Successfully. CA: %s Pool ID: %s BaseVault: %s QuoteVault: %s\n", pool.CA, pool.PoolID, pool.BaseVault, pool.QuoteVault)
			}
		}
	}
}

func processTransaction(transactionUpdate *pb.SubscribeUpdate_Transaction) Pool {
	var pool Pool

	txInfo := transactionUpdate.Transaction
	msg := txInfo.Transaction.Transaction.Message

	for _, ix := range msg.Instructions {
		programID := common.PublicKeyFromBytes(msg.AccountKeys[ix.ProgramIdIndex]).String()
		if programID == ammProgramID && len(ix.Data) > 0 && ix.Data[0] == 1 {
			if common.PublicKeyFromBytes(msg.AccountKeys[ix.Accounts[8]]).String() == solAddress {
				pool.CA = common.PublicKeyFromBytes(msg.AccountKeys[ix.Accounts[9]]).String()
			} else if common.PublicKeyFromBytes(msg.AccountKeys[ix.Accounts[9]]).String() == solAddress {
				pool.CA = common.PublicKeyFromBytes(msg.AccountKeys[ix.Accounts[8]]).String()
			} else {
				log.Logger.Info("No CA found")
			}
			pool.PoolID = common.PublicKeyFromBytes(msg.AccountKeys[ix.Accounts[4]]).String()
			pool.BaseVault = common.PublicKeyFromBytes(msg.AccountKeys[ix.Accounts[10]]).String()
			pool.QuoteVault = common.PublicKeyFromBytes(msg.AccountKeys[ix.Accounts[11]]).String()
			break
		}
	}

	return pool
}

func grpcConnect(address string) *grpc.ClientConn {
	var opts []grpc.DialOption
	pool, _ := x509.SystemCertPool()
	creds := credentials.NewClientTLSFromCert(pool, "")
	opts = append(opts, grpc.WithTransportCredentials(creds))

	opts = append(opts, grpc.WithKeepaliveParams(kacp))

	log.Logger.Println("Starting grpc client, connecting to", address)
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Logger.Fatalf("fail to dial: %v", err)
	}

	return conn
}
