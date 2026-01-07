package main

import (
	"crypto/ecdsa"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

func main() {
	// 1、生成私钥
	// 基于 secp256k1 椭圆曲线生成随机的 ECDSA 私钥，这是以太坊私钥的核心生成逻辑
	// 私钥是 256 位（32 字节）的随机数，是以太坊账户的核心凭证
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(" 256 位（32 字节）的随机数:", privateKey)
	fmt.Println("--------------------------------------------------------")

	// 2、私钥转十六进制字符串
	privateKeyBytes := crypto.FromECDSA(privateKey)
	fmt.Println("转十六进制字符串:", hexutil.Encode(privateKeyBytes))
	fmt.Println("--------------------------------------------------------")

	// 3、从私钥生成公钥
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey) // 类型断言确保公钥是 *ecdsa.PublicKey 类型（防止类型错误）
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	// 将公钥结构体转为原始字节数组
	// 以太坊公钥前缀为 04，总长度 65 字节：1 字节前缀 + 32 字节 x 坐标 + 32 字节 y 坐标
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	fmt.Println("from pubKey:", hexutil.Encode(publicKeyBytes)[4:]) // 去掉'0x04'
	fmt.Println("--------------------------------------------------------")
	// 4、从公钥计算以太坊地址
	// 方式一：以太坊官方提供的计算方法
	// crypto.PubkeyToAddress(...)：以太坊官方封装的地址计算方法，
	// 内部逻辑是「公钥字节数组（去掉 04 前缀）→ Keccak256 哈希 → 截取最后 20 字节 → 加 0x 前缀」
	// .Hex()：将地址转为带 0x 前缀的十六进制字符串（以太坊地址标准格式）
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Println(address)
	fmt.Println("--------------------------------------------------------")
	// 方式二：自己实现计算方法
	// 创建 Keccak256 哈希器（以太坊专用的 SHA3 变种，非标准 NIST SHA3）
	hash := sha3.NewLegacyKeccak256()
	hash.Write(publicKeyBytes[1:]) // 写入去掉 04 前缀的公钥字节数组（64 字节）
	fmt.Println("full:", hexutil.Encode(hash.Sum(nil)[:]))
	fmt.Println(hexutil.Encode(hash.Sum(nil)[12:])) // 原长32位，截去12位，保留后20位
}
