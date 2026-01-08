package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// 连接以太坊 Sepolia 测试网的 WebSocket 节点（WSS 协议）
	client, err := ethclient.Dial("wss://ethereum-sepolia-rpc.publicnode.com")
	if err != nil {
		log.Fatal(err)
	}

	// 创建一个新的通道，用于接收最新的区块头
	headers := make(chan *types.Header)
	// SubscribeNewHead 方法，接收刚创建的区块头通道，该方法将返回一个订阅对象
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:
			fmt.Println("header:", header.Hash().Hex())
			fmt.Println("header:", header.Number.Uint64())
			fmt.Println("header:", header.Time)
			fmt.Println("header:", header.Nonce)

			block, err := client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("block:", block.Hash().Hex())
			fmt.Println("block:", block.Number().Uint64())
			fmt.Println("block:", block.Time())
			fmt.Println("block:", block.Nonce())
			fmt.Println("block:", len(block.Transactions()))

			fmt.Println("-----------------------------------------------------------")
		}
	}
}
