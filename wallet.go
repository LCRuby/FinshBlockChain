package main

import (
	"base58"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/ripemd160"
)

type Wallet struct {
	Private *ecdsa.PrivateKey
	//约定，这里的PubKey不存储原始的公钥，而是存储X和Y拼接的字符串，在校验端重新拆分（参考r,s传递）
	/*
		type PublicKey struct {
			elliptic.Curve
			X, Y *big.Int
		}
	*/
	PublicKey []byte
}

//创建钱包
func NewWallet() *Wallet {
	//创建地址看着流程图做
	//创建一个椭圆曲线
	curve := elliptic.P256()
	//生成私钥
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic()
	}
	//生成公钥
	pubKeyOrig := privateKey.PublicKey

	//拼接 x,y
	pubKey := append(pubKeyOrig.X.Bytes(), pubKeyOrig.Y.Bytes()...)
	return &Wallet{Private: privateKey, PublicKey: pubKey}
}

//生成地址
func (w *Wallet) NewAddress() string {
	pubKey := w.PublicKey
	rip160HashValue := HashPubKey(pubKey)
	version := byte(00)
	//拼接version
	payload := append([]byte{version}, rip160HashValue...)

	//checksum
	checkCode := CheckSum(payload)

	//25字节数据
	payload = append(payload, checkCode...)

	//go语言有一个库，叫做btcd,这个是go语言实现的比特币全节点源码
	address := base58.Encode(payload)

	return address
}

//将公约转化成公钥哈希
func HashPubKey(data []byte) []byte {
	hash := sha256.Sum256(data)
	//编码
	rip160hasher := ripemd160.New()
	_, err := rip160hasher.Write(hash[:])
	if err != nil {
		log.Panic(err)
	}

	//返回rip160的哈希结果
	rip160HashValue := rip160hasher.Sum(nil)
	return rip160HashValue
}

func CheckSum(data []byte) []byte {
	//两次sha256
	hash1 := sha256.Sum256(data)
	hash2 := sha256.Sum256(hash1[:])

	//前4字节校验码
	checkCode := hash2[:4]
	return checkCode
}

//地址校验
func IsValidAddress(address string) bool {
	//1. 解码
	addressByte := base58.Decode(address)

	if len(addressByte) < 4 {
		return false
	}

	//2. 取数据
	payload := addressByte[:len(addressByte)-4]
	checksum1 := addressByte[len(addressByte)-4:]

	//3. 做checksum函数
	checksum2 := CheckSum(payload)

	//fmt.Printf("checksum1 : %x\n", checksum1)
	//fmt.Printf("checksum2 : %x\n", checksum2)

	//4. 比较
	return bytes.Equal(checksum1, checksum2)
}
