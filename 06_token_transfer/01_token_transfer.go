package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main1() {
	client, err := ethclient.Dial("https://ethereum-sepolia-rpc.publicnode.com")
	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := crypto.HexToECDSA("d71a701e75b49c9a337ac20bacf15ccf62b92b86fee94cb6f5bc0240453f4f64")
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	value := big.NewInt(0) // in wei (0 eth)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 接收方地址
	toAddress := common.HexToAddress("0xF9B6FF30D67e802690C94edD7B4CFFCfdF6A4deF")
	// erc20代币合约地址
	tokenAddress := common.HexToAddress("0xf8112b83f4ABA089Acf7E8fb77c480D6778d029d")

	// 1. 计算 transfer 函数的方法 ID（前4字节 Keccak256 哈希）
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb
	// 2. 对接收方地址、转账金额做 32 字节左填充（ABI 编码要求）
	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAddress)) // 0x0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d
	amount := new(big.Int)
	amount.SetString("10000000000000000000", 10) // 10 tokens
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000
	var data []byte
	// 3. 拼接最终的交易数据
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	// 估算 Gas 限额
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: fromAddress,
		To:   &tokenAddress,
		Data: data,
	})
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	fmt.Println(gasLimit) // 23256

	//构建未签名交易
	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)
	// 获取链 ID（Sepolia 测试网）
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	// 用私钥签名交易（EIP155 签名规则，防止跨链重放）
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("tx sent: %s\n", signedTx.Hash().Hex()) // tx sent: 0xa56316b637a94c4cc0331c73ef26389d6c097506d581073f927275e7a6ece0bc
}
