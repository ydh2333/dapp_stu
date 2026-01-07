package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// 步骤1：连接以太坊节点（Infura提供的Sepolia测试网节点）
	client, err := ethclient.Dial("https://sepolia.infura.io/v3/48838061f7a544dfb74776ac7a0680bb")
	if err != nil {
		log.Fatal(err)
	}
	// 步骤2：加载发送方私钥（需替换为实际有效私钥）
	privateKey, err := crypto.HexToECDSA("d71a701e75b49c9a337ac20bacf15ccf62b92b86fee94cb6f5bc0240453f4f64")
	if err != nil {
		log.Fatal(err)
	}

	// 步骤3：从私钥推导发送方地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	// 步骤4：获取发送方账户的下一个nonce（防止交易重放）
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	// 步骤5：定义交易参数
	value := big.NewInt(1000000000000000000) // 转账金额：1 ETH（以wei为单位，1 ETH = 10^18 wei）
	gasLimit := uint64(21000)                // 燃气上限：普通ETH转账固定21000 gas
	// 根据'x'个先前块来获得平均燃气价格，燃气价格总是根据市场需求和用户愿意支付的价格而波动
	// 因此对燃气价格进行硬编码并不理想（gasPrice := big.NewInt(30000000000)）
	// 建议燃气价格（动态获取，适配市场）
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	toAddress := common.HexToAddress("0xF9B6FF30D67e802690C94edD7B4CFFCfdF6A4deF") // 接收方地址
	var data []byte                                                                // 转账数据：普通ETH转账无需附加数据，设为空

	// 步骤6：构造未签名交易
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	// 步骤7：获取链ID并签名交易（EIP155标准，防止跨链重放）
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// 步骤8：发送签名后的交易到区块链
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
	// 步骤9：输出交易哈希（可在区块链浏览器查询交易状态）
	fmt.Printf("tx sent: %s\n", signedTx.Hash().Hex())
}
