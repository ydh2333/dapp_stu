package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("https://ethereum-sepolia-rpc.publicnode.com")
	if err != nil {
		log.Fatal(err)
	}

	account := common.HexToAddress("0x51ccc58AE0a621b78196CcE2e01920dd6E5be38b")

	// 1. 查询最新区块的账户余额
	balance, err := client.BalanceAt(context.Background(), account, nil) // nil：表示查询最新区块的余额
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(balance)
	// ----------------------------------------------------------------------------
	// 2. 查询指定历史区块的余额
	blockNumber := big.NewInt(9996975)
	balanceAt, err := client.BalanceAt(context.Background(), account, blockNumber)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(balanceAt)
	// 将 wei 转换为 ETH 单位
	fbalance := new(big.Float)
	fbalance.SetString(balanceAt.String())
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
	fmt.Println(ethValue)

	// 3. 查询待处理交易的余额（Pending 余额）
	pendingBalance, err := client.PendingBalanceAt(context.Background(), account)
	fmt.Println(pendingBalance)
}
