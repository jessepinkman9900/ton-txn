package main

import (
	"context"
	"encoding/hex"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"log"
)

func main() {
	// setup connection
	client := liteclient.NewConnectionPool()

	configUrl := "https://ton.org/global.config.json"
	//configUrl := "https://ton-blockchain.github.io/testnet-global.config.json"
	err := client.AddConnectionsFromConfigUrl(context.Background(), configUrl)
	if err != nil {
		panic(err)
	}
	log.Printf("client connections added from config file")

	apiClient := ton.NewAPIClient(client)
	api := apiClient.WithRetry(10)
	log.Printf("api client created")

	// if all requests to be processed by single node
	// use this context
	ctx := client.StickyContext(context.Background())

	// rpc calls
	// block
	block, err := api.CurrentMasterchainInfo(ctx)
	if err != nil {
		panic(err)
	}
	log.Printf("rpc call block - file_hash:%s root_hash:%s seq_no:%d shard:%d",
		hex.EncodeToString(block.FileHash),
		hex.EncodeToString(block.RootHash),
		block.SeqNo,
		block.Shard,
	)

	// get contract events
	// dedust factory contract
	addr := address.MustParseAddr("EQBfBWT7X2BHg9tXAxzhz2aKiNTU1tpt5NsiK0uSDW_YAJ67")
	log.Print(addr.String())
	//api.GetTransaction()

	// pool contract state
	// dedust pool contract - EQDcm06RlreuMurm-yik9WbL6kI617B77OrSRF_ZjoCYFuny
	pool_addr := address.MustParseAddr("EQDcm06RlreuMurm-yik9WbL6kI617B77OrSRF_ZjoCYFuny")

	// get asssets
	// assets, err := api.RunGetMethod(ctx, block, pool_addr, "get_assets")
	// if err != nil {
	// 	log.Printf("get_assets error: %s", err.Error())
	// } else {
	// 	log.Printf("pool: %s get_assets: numerator:%d denominator:%d", pool_addr.String(), hex.EncodeToString(assets.AsTuple()[0]), hex.EncodeToString(assets.AsTuple()[1].data))
	// }

	// get fee
	fee, err := api.RunGetMethod(ctx, block, pool_addr, "get_trade_fee")
	if err != nil {
		log.Printf("get_trade_fee error: %s", err.Error())
	} else {
		log.Printf("pool: %s get_trade_fee: numerator:%d denominator:%d", pool_addr.String(), fee.AsTuple()[0], fee.AsTuple()[1])
	}

	// get reserves
	reserves, err := api.RunGetMethod(ctx, block, pool_addr, "get_reserves")
	if err != nil {
		log.Printf("get_reserves error: %s", err.Error())
	} else {
		log.Printf("pool: %s get_reserves: reserve0:%d reserve1:%d", pool_addr.String(), reserves.AsTuple()[0], reserves.AsTuple()[1])
	}
}

// todo
// 1. filter events to get pools
// 2. get pool data
// 3. build swap message data
