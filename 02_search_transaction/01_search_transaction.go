package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("https://sepolia.infura.io/v3/48838061f7a544dfb74776ac7a0680bb")
	if err != nil {
		log.Fatal(err)
	}
	// 1、连接以太坊节点，获取链 ID（用于交易签名验证）
	chainID, err := client.ChainID(context.Background())
	fmt.Println("chainID:", chainID)
	if err != nil {
		log.Fatal(err)
	}
	// 2、指定区块号（5671744），通过 BlockByNumber 获取该区块的完整数据
	blockNumber := big.NewInt(5671744)
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		log.Fatal(err)
	}
	// 3、遍历区块内的交易（仅取第一条，break 终止循环），打印交易核心字段
	for _, tx := range block.Transactions() {
		// 交易hash
		fmt.Println(tx.Hash().Hex()) // 0x20294a03e8766e9aeab58327fc4112756017c6c28f6f99c7722f4a29075601c5
		// 转账金额
		fmt.Println(tx.Value().String()) // 100000000000000000
		// gas限制
		fmt.Println(tx.Gas()) // 21000
		// gas价格
		fmt.Println(tx.GasPrice().Uint64()) // 100000000000
		// 交易随机数
		fmt.Println(tx.Nonce()) // 245132
		// 交易数据，空表示普通转账，非空为合约交互数据
		fmt.Println(tx.Data()) // []
		// 交易接收地址
		fmt.Println(tx.To().Hex()) // 0x8F9aFd209339088Ced7Bc0f57Fe08566ADda3587
		// 交易发送地址，通过 types.Sender 结合链 ID 解析签名获取
		if sender, err := types.Sender(types.NewEIP155Signer(chainID), tx); err == nil {
			fmt.Println("sender", sender.Hex()) // 0x2CdA41645F2dBffB852a605E92B185501801FC28
		} else {
			log.Fatal(err)
		}
		// 交易收据（签名），包含交易状态（receipt.Status，1 表示成功）、日志（receipt.Logs）
		receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(receipt.Status) // 1
		fmt.Println(receipt.Logs)   // []
		break
	}
	fmt.Println("-------------------------------------------------")
	// 1、指定区块哈希，通过 TransactionCount 获取该区块内的交易总数
	blockHash := common.HexToHash("0xae713dea1419ac72b928ebe6ba9915cd4fc1ef125a606f90f5e783c47cb1a4b5")
	count, err := client.TransactionCount(context.Background(), blockHash)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("count:", count)
	// 2、按交易索引遍历（仅取第 1 条），通过 TransactionInBlock 获取指定索引的交易，打印交易哈希
	for idx := uint(0); idx < count; idx++ {
		tx, err := client.TransactionInBlock(context.Background(), blockHash, idx)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(tx.Hash().Hex()) // 0x20294a03e8766e9aeab58327fc4112756017c6c28f6f99c7722f4a29075601c5
		break
	}
	fmt.Println("-------------------------------------------------")
	// 1、指定交易哈希，通过 TransactionByHash 查询交易
	txHash := common.HexToHash("0x20294a03e8766e9aeab58327fc4112756017c6c28f6f99c7722f4a29075601c5")
	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		log.Fatal(err)
	}
	// isPending：标识交易是否处于 “待确认” 状态（false，表示已上链）
	fmt.Println(isPending)
	// 交易哈希（验证查询结果的准确性）
	fmt.Println(tx.Hash().Hex()) // 0x20294a03e8766e9aeab58327fc4112756017c6c28f6f99c7722f4a29075601c5.Println(isPending)       // false
}
