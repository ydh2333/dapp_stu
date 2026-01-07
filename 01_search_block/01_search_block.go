package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("https://ethereum-sepolia-rpc.publicnode.com")
	if err != nil {
		log.Fatal(err)
	}

	blockNumber := big.NewInt(5671744)

	header, err := client.HeaderByNumber(context.Background(), blockNumber)
	// 区块号
	fmt.Println(header.Number.Uint64()) // 5671744
	// 区块时间戳
	fmt.Println(header.Time) // 1712798400
	// 区块难度
	fmt.Println(header.Difficulty.Uint64()) // 0
	// 区块hash
	fmt.Println(header.Hash().Hex()) // 0xae713dea1419ac72b928ebe6ba9915cd4fc1ef125a606f90f5e783c47cb1a4b5

	if err != nil {
		log.Fatal(err)
	}
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		log.Fatal(err)
	}

	// 区块号
	fmt.Println(block.Number().Uint64()) // 5671744
	// 区块时间戳
	fmt.Println(block.Time()) // 1712798400
	// 区块难度
	fmt.Println(block.Difficulty().Uint64()) // 0
	// 区块hash
	fmt.Println(block.Hash().Hex()) // 0xae713dea1419ac72b928ebe6ba9915cd4fc1ef125a606f90f5e783c47cb1a4b5
	// 交易数量
	fmt.Println(len(block.Transactions())) // 70
	count, err := client.TransactionCount(context.Background(), block.Hash())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(count) // 70
}
