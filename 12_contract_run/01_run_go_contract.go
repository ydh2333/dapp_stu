package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"github.com/ydh2333/dapp_stu/10_contract_deploy/store"
)

const (
	contractAddr = "0x9F49FF297E88AD77120f0e261a76fa3A835c24Aa"
)

func main1() {
	// 步骤 1：连接以太坊节点
	client, err := ethclient.Dial("https://ethereum-sepolia-rpc.publicnode.com")
	if err != nil {
		log.Fatal(err)
	}
	// 步骤 2：实例化合约, 通过 Store 合约的 Go 绑定代码，结合合约地址和客户端实例，
	// 创建合约交互对象 storeContract，后续可通过该对象调用合约方法
	storeContract, err := store.NewStore(common.HexToAddress(contractAddr), client)
	if err != nil {
		log.Fatal(err)
	}

	// 步骤 3：加载私钥（交易签名用）
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	privateKeyStr := os.Getenv("PRIVATE_KEY")
	if privateKeyStr == "" {
		log.Fatal("PRIVATE_KEY is not set in .env file")
	}
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		log.Fatal(err)
	}

	// 步骤 4：准备合约调用参数
	// 将字符串复制到 32 字节数组中，适配合约参数类型
	var key [32]byte
	var value [32]byte
	copy(key[:], []byte("demo_save_key5"))
	copy(value[:], []byte("demo_save_value555"))

	// 步骤 5：初始化交易选项（签名 + 链 ID）
	// bind.NewKeyedTransactorWithChainID：创建交易签名器 opt（*bind.TransactOpts）；
	// 入参：私钥 + 链 ID（Sepolia 测试网链 ID 为 11155111），确保交易仅在目标链有效。
	opt, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(11155111))
	if err != nil {
		log.Fatal(err)
	}

	// 步骤 6：调用合约写入方法（发送交易）
	tx, err := storeContract.SetItem(opt, key, value)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("tx hash:", tx.Hash().Hex())

	// 等待交易被挖矿
	// 注意：只有在执行合约部署交易的情况下，合约地址（receipt）才会有值，否则为空（0x00000...）
	receipt, err := waitForReceipt(client, tx.Hash())
	if err != nil {
		log.Fatal(err)
	}
	if receipt.Status == types.ReceiptStatusSuccessful {
		fmt.Printf("交易成功！区块号：%d，消耗Gas：%d\n", receipt.BlockNumber, receipt.GasUsed)
	} else {
		fmt.Println("交易失败！")
	}

	// 步骤 7：查询合约数据（验证写入结果）
	// 构建只读调用选项（CallOpts）：不发送交易，仅查询链上数据，无 Gas 消耗
	callOpt := &bind.CallOpts{Context: context.Background()}
	// 调用合约 Items 方法（读取指定 key 的 value）
	valueInContract, err := storeContract.Items(callOpt, key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("valueInContract:", valueInContract)
	fmt.Println("is value saving in contract equals to origin value:", valueInContract == value)
}

func waitForReceipt(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	for {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			return receipt, nil
		}
		if err != ethereum.NotFound {
			return nil, err
		}
		// 等待一段时间后再次查询
		time.Sleep(1 * time.Second)
	}
}
