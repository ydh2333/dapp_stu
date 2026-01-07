package main

import (
	"fmt"
	"log"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	token "github.com/ydh2333/dapp_stu/08_search_token_balance/erc20"
)

func main() {
	// 1. 连接以太坊节点
	client, err := ethclient.Dial("https://ethereum-sepolia-rpc.publicnode.com")
	if err != nil {
		log.Fatal(err)
	}
	// 2. 初始化代币合约实例
	tokenAddress := common.HexToAddress("0xf8112b83f4ABA089Acf7E8fb77c480D6778d029d")
	// 创建 ERC20 合约实例，绑定节点客户端与合约地址，后续可通过该实例调用合约方法
	instance, err := token.NewErc20(tokenAddress, client)
	if err != nil {
		log.Fatal(err)
	}
	// 3. 查询目标地址的代币余额
	address := common.HexToAddress("0x51ccc58AE0a621b78196CcE2e01920dd6E5be38b")
	bal, err := instance.BalanceOf(&bind.CallOpts{}, address)
	if err != nil {
		log.Fatal(err)
	}
	// 4. 查询代币基础信息
	// 代币名称
	name, err := instance.Name(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}
	// 代币符号
	symbol, err := instance.Symbol(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}
	// 代币小数位数
	decimals, err := instance.Decimals(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}
	// 打印原始信息
	fmt.Printf("name: %s\n", name)         // "name: Golem Network"
	fmt.Printf("symbol: %s\n", symbol)     // "symbol: GNT"
	fmt.Printf("decimals: %v\n", decimals) // "decimals: 18"
	fmt.Printf("wei: %s\n", bal)           // "wei: 74605500647408739782407023"

	// 转换 wei 为可读的代币单位（除以 10^decimals）
	fbal := new(big.Float)
	fbal.SetString(bal.String())
	value := new(big.Float).Quo(fbal, big.NewFloat(math.Pow10(int(decimals))))
	fmt.Printf("balance: %f\n", value) // "balance: 74605500.647409"
}
