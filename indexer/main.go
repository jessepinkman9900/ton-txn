package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"log"
	"math"
	"math/big"
)

type PoolData struct {
	reserve0                      *big.Int         // 0 coins
	reserve1                      *big.Int         // 1 coins
	token0_wallet_addr            *address.Address // 2
	token1_wallet_addr            *address.Address // 3
	lp_fee                        *big.Int         // 4 uint8
	protocol_fee                  *big.Int         // 5 uint8
	ref_fee                       *big.Int         // 6 uint8
	protocol_fee_address          *address.Address // 7
	collected_token0_protocol_fee *big.Int         // 8 coins
	collected_token1_protocol_fee *big.Int         // 9 coins
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
	// 1 STON -> USDT
	amount_in := big.NewInt(1000000000) // 1 STON - 9 decimals
	amount_out := get_amount_out(pool_data, big.NewInt(0), amount_in, pool_data.reserve0, pool_data.reserve1)
	log.Printf("1 STON -> USDT amount_out: %d", amount_out)
	log.Printf("1 STON -> USDT amount_out scaled: %f", new(big.Float).Quo(new(big.Float).SetInt(amount_out), big.NewFloat(math.Pow10(6))))

	// 1 USDT -> STON
	amount_in = big.NewInt(1000000) // 1 USDT - 6 decimals
	amount_out = get_amount_out(pool_data, big.NewInt(0), amount_in, pool_data.reserve1, pool_data.reserve0)
	log.Printf("1 USDT -> STON amount_out: %d", amount_out)
	log.Printf("1 USDT -> STON amount_out scaled: %f", new(big.Float).Quo(new(big.Float).SetInt(amount_out), big.NewFloat(math.Pow10(9))))

	// build txn payload
	routerv1_address := address.MustParseAddr("EQB3ncyBUTjZUA5EnFKR5_EnOMI9V1tTEAAPaiU71gc4TiUt")
	op_code, _ := new(big.Int).SetString("25938561", 16)                                               // 0x25938561
	wallet_token_out_addr := address.MustParseAddr("EQBO7JIbnU1WoNlGdgFtScJrObHXkBp-FT5mAz8UagiG9KQR") // usdt
	to_addr := address.MustParseAddr("EQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAM9c")               // 0 address
	payload := cell.BeginCell().
		MustStoreUInt(op_code.Uint64(), 32).
		MustStoreAddr(wallet_token_out_addr).
		MustStoreCoins(amount_out.Uint64()).
		MustStoreAddr(to_addr).
		MustStoreBoolBit(false). // ref_address
		EndCell()

	op_code, _ = new(big.Int).SetString("7362d09c", 16) // 0x7362d09c
	jetton_amount := uint64(0)
	from_address := address.MustParseAddr("EQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAM9c") // 0 address
	body := cell.BeginCell().
		MustStoreUInt(op_code.Uint64(), 32).
		MustStoreCoins(jetton_amount).
		MustStoreAddr(from_address).
		MustStoreBoolBit(true).
		MustStoreRef(payload).
		EndCell()
	log.Printf("routerv1_address %s", routerv1_address)
	log.Printf("txn body: %+v\n", body)
	// todo: send payload on chain
}

func get_amount_out(pool_data PoolData, has_ref, amount_in, reserve_in, reserve_out *big.Int) *big.Int {
	FEE_DIVIDER := big.NewInt(10000)

	amount_in_with_fee := new(big.Int).Mul(amount_in, new(big.Int).Sub(FEE_DIVIDER, pool_data.lp_fee))
	//log.Printf("amount_in_with_fee: %d", amount_in_with_fee)
	base_out := new(big.Int).Div(new(big.Int).Mul(amount_in_with_fee, reserve_out), new(big.Int).Add(new(big.Int).Mul(reserve_in, FEE_DIVIDER), amount_in_with_fee))
	//log.Printf("base_out: %d", base_out)

	protocol_fee_out := big.NewInt(0)
	ref_fee_out := big.NewInt(0)
	if pool_data.protocol_fee.Cmp(big.NewInt(0)) > 0 {
		protocol_fee_out = new(big.Int).Div(new(big.Int).Mul(base_out, pool_data.protocol_fee), FEE_DIVIDER)
	}
	//log.Printf("protocol_fee_out: %d", protocol_fee_out)

	if has_ref.Cmp(big.NewInt(0)) == 0 && pool_data.ref_fee.Cmp(big.NewInt(0)) > 0 {
		ref_fee_out = new(big.Int).Div(new(big.Int).Mul(base_out, pool_data.ref_fee), FEE_DIVIDER)
	}
	//log.Printf("ref_fee_out: %d", ref_fee_out)

	base_out = new(big.Int).Sub(base_out, new(big.Int).Add(protocol_fee_out, ref_fee_out))
	//log.Printf("base_out: %d", base_out)
	if base_out.Cmp(big.NewInt(0)) < 0 {
		base_out = big.NewInt(0)
	}
	//log.Printf("base_out: %d", base_out)

	return base_out
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
	log.Printf("lp_fee: %d", lp_fee)

	protocol_fee := pool_data.MustInt(5)
	log.Printf("protocol_fee: %d", protocol_fee)

	ref_fee := pool_data.MustInt(6)
	log.Printf("ref_fee: %d", ref_fee)

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
