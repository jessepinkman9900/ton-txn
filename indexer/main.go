package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"log"
	"math/big"
)

type PoolData struct {
	reserve0                      *big.Int         // 0
	reserve1                      *big.Int         // 1
	token0_wallet_addr            *address.Address // 2
	token1_wallet_addr            *address.Address // 3
	lp_fee                        *big.Int         // 4 uint8
	protocol_fee                  *big.Int         // 5 uint8
	ref_fee                       *big.Int         // 6 uint8
	protocol_fee_address          *address.Address // 7
	collected_token0_protocol_fee *big.Int         // 8
	collected_token1_protocol_fee *big.Int         // 9
}

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

	// get create pool msg

	// get pool data
	pool_address := address.MustParseAddr("EQBWjPASSjsgibEv3fGUCwSwFyUxLVFaywZzNmuBXPFOFfOG")
	pool_data, err := get_pool_data(ctx, api, block, pool_address)
	if err != nil {
		log.Printf("ERROR fetching pool_data ", err.Error())
	}
	log.Printf("pool_data", pool_data)

	// calculate swap

	// build txn payload

}

func get_pool_data(ctx context.Context, api ton.APIClientWrapped, block *ton.BlockIDExt, pool_address *address.Address) (PoolData, error) {
	pool_data, err := api.RunGetMethod(ctx, block, pool_address, "get_pool_data")
	if err != nil {
		log.Printf("ERROR get_pool_data", err.Error())
		return PoolData{}, fmt.Errorf("unableto fetch data from pool")
	}
	log.Printf("SUCCESS get_pool_data")

	reserve0 := pool_data.MustInt(0)
	log.Printf("reserve0: %d", reserve0)

	reserve1 := pool_data.MustInt(1)
	log.Printf("reserve1: %d", reserve1)

	token0_wallet_address := pool_data.MustSlice(2).MustLoadAddr()
	log.Printf("token0_wallet_address: %s", token0_wallet_address)

	token1_wallet_address := pool_data.MustSlice(3).MustLoadAddr()
	log.Printf("token1_wallet_address: %s", token1_wallet_address)

	lp_fee := pool_data.MustInt(4)
	log.Printf("lp_fee: '%d", lp_fee)

	protocol_fee := pool_data.MustInt(5)
	log.Printf("protocol_fee: '%d", protocol_fee)

	ref_fee := pool_data.MustInt(6)
	log.Printf("ref_fee: '%d", ref_fee)

	protocol_fee_address := pool_data.MustSlice(7).MustLoadAddr()
	log.Printf("protocol_fee_address: %s", protocol_fee_address)

	collected_token0_protocol_fee := pool_data.MustInt(0)
	log.Printf("collected_token0_protocol_fee: %d", collected_token0_protocol_fee)

	collected_token1_protocol_fee := pool_data.MustInt(0)
	log.Printf("collected_token1_protocol_fee: %d", collected_token1_protocol_fee)

	pool_data_struct := PoolData{
		reserve0:                      reserve0,
		reserve1:                      reserve1,
		token0_wallet_addr:            token0_wallet_address,
		token1_wallet_addr:            token1_wallet_address,
		lp_fee:                        lp_fee,
		protocol_fee:                  protocol_fee,
		ref_fee:                       ref_fee,
		protocol_fee_address:          protocol_fee_address,
		collected_token0_protocol_fee: collected_token0_protocol_fee,
		collected_token1_protocol_fee: collected_token1_protocol_fee,
	}

	return pool_data_struct, nil
}

// todo
// 1. filter events to get pools
// 2. get pool data
// 3. build swap message data
