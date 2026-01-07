package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	// 1. 加载.env文件（核心：读取配置）
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// 2. 从.env读取配置（逐个解析）
	// RPC节点地址
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		log.Fatal("RPC_URL is not set in .env file")
	}
	// 发送方私钥
	privateKeyStr := os.Getenv("PRIVATE_KEY")
	if privateKeyStr == "" {
		log.Fatal("PRIVATE_KEY is not set in .env file")
	}
	// 接收方地址
	toAddressStr := os.Getenv("TO_ADDRESS")
	if toAddressStr == "" {
		log.Fatal("TO_ADDRESS is not set in .env file")
	}
	// 代币合约地址
	tokenAddressStr := os.Getenv("TOKEN_CONTRACT_ADDRESS")
	if tokenAddressStr == "" {
		log.Fatal("TOKEN_CONTRACT_ADDRESS is not set in .env file")
	}
	// 转账金额
	transferAmountStr := os.Getenv("TRANSFER_AMOUNT")
	if transferAmountStr == "" {
		log.Fatal("TRANSFER_AMOUNT is not set in .env file")
	}

	// 3. 连接以太坊节点
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatal(err)
	}

	// 4. 解析私钥
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		log.Fatal(err)
	}

	// 5. 获取发送方地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 6. 获取交易Nonce（未确认交易计数）
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	// 7. 初始化基础参数
	value := big.NewInt(0) // in wei (0 eth)
	// <------------------------------------------------------------------------------------
	// 8. 获取EIP-1559动态手续费参数
	// a. GasTipCap（优先费）：建议的最大小费（给矿工）
	gasTipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		log.Fatalf("failed to get gas tip cap: %v", err)
	}
	fmt.Printf("suggested gas tip cap: %s wei\n", gasTipCap.String())

	// b. GasFeeCap（最大手续费）：baseFee + tip 的上限（baseFee由链上计算）
	// 这里简单取：GasTipCap * 2（也可以通过区块头的baseFee动态计算）
	gasFeeCap := new(big.Int).Mul(gasTipCap, big.NewInt(2))
	fmt.Printf("gas fee cap (tip*2): %s wei\n", gasFeeCap.String())
	// ------------------------------------------------------------------------------------>

	// 9. 解析关键地址
	toAddress := common.HexToAddress(toAddressStr)       // 接收方地址
	tokenAddress := common.HexToAddress(tokenAddressStr) // erc20代币合约地址

	// <------------------------------------------------------------------------------------
	// 10. 使用官方ABI包自动编码transfer交易数据
	// 10.1 定义ERC20的transfer方法ABI（JSON格式，描述函数名和参数类型）
	// 格式说明：name=函数名，type=function，inputs=参数列表（name=参数名，type=参数类型）
	erc20ABIJson := `[
		{
			"name": "transfer",
			"type": "function",
			"inputs": [
				{"name": "to", "type": "address"},
				{"name": "value", "type": "uint256"}
			]
		}
	]`
	// 10.2 解析ABI字符串为ABI对象
	erc20ABI, err := abi.JSON(strings.NewReader(erc20ABIJson))
	if err != nil {
		fmt.Println(err) // 打印ABI解析错误（如JSON格式错误、参数类型写错）
		log.Fatal("Failed to parse ERC20 ABI")
	}

	// 10.3 解析转账金额（确保格式正确）
	amount := new(big.Int)
	amount, ok = amount.SetString(transferAmountStr, 10) // 第二个参数10表示十进制
	if !ok {
		fmt.Printf("Invalid transfer amount: %s (must be a positive integer)\n", transferAmountStr)
		log.Fatal("Failed to parse transfer amount")
	}

	// 10.4 自动编码交易数据：Pack(函数名, 参数1, 参数2, ...)
	// 作用：自动生成methodID + 32字节左填充的参数，无需手动拼接！
	data, err := erc20ABI.Pack(
		"transfer", // 要调用的函数名（必须与ABI中定义的name一致）
		toAddress,  // 第一个参数：接收方地址（类型需与ABI中"to"的type=address匹配）
		amount,     // 第二个参数：转账金额（类型需与ABI中"value"的type=uint256匹配）
	)
	if err != nil {
		fmt.Println(err) // 新增：打印参数编码错误（如参数类型不匹配、地址格式错误）
		log.Fatal("Failed to pack transfer parameters (ABI encoding)")
	}
	fmt.Printf("Auto-generated transfer data (hex): %s\n", common.Bytes2Hex(data)) // 打印编码后的交易数据
	// ------------------------------------------------------------------------------------>
	// 估算 Gas 限额
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: fromAddress,
		To:   &tokenAddress,
		Data: data,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(gasLimit) // 23256

	// 获取链 ID（Sepolia 测试网）
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	// <------------------------------------------------------------------------------------
	// 构建EIP-1559动态手续费交易
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,       // 链ID
		Nonce:     nonce,         // 交易Nonce
		GasTipCap: gasTipCap,     // 优先费（小费）
		GasFeeCap: gasFeeCap,     // 最大手续费（baseFee + tip）
		Gas:       gasLimit,      // Gas限额
		To:        &tokenAddress, // 代币合约地址
		Value:     value,         // 转账ETH金额（ERC20转账为0）
		Data:      data,          // 交易数据（transfer方法+参数）
	})

	// 用私钥签名交易（EIP155 签名规则，防止跨链重放）
	signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	// ------------------------------------------------------------------------------------>
	// 发送交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("tx sent: %s\n", signedTx.Hash().Hex()) // tx sent: 0xa56316b637a94c4cc0331c73ef26389d6c097506d581073f927275e7a6ece0bc
}
