package main

import (
	"base58"
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"os"
)

const walletFile = "wallet.dat"

//定义一个wallets数据结构
type Wallets struct {
	//map[地址]钱包
	WalletMap map[string]*Wallet
}

//显示钱包的所有地址
func NewWallets() *Wallets {

	var ws Wallets
	ws.WalletMap = make(map[string]*Wallet)

	ws.LoadFile()
	return &ws
}

//创建钱包地址
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := wallet.NewAddress()

	ws.WalletMap[address] = wallet
	//把地址保存到文件中
	ws.SaveToFile()
	return address
}

//保存方法,把新建的wallet添加进去
func (ws *Wallets) SaveToFile() {
	var buffer bytes.Buffer

	//使用gob是先注册,否则会报错
	//panic: gob: type not registered for interface: elliptic.p256Curve
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, buffer.Bytes(), 0600)
	if err != nil {
		log.Panic(err)
	}
}

//读取文件方法，把所有的wallet读出来
func (ws *Wallets) LoadFile() {
	//在读取之前，要先确认文件是否在，如果不存在，直接退出
	_, err := os.Stat(walletFile)
	if os.IsNotExist(err) {
		//ws.WalletsMap = make(map[string]*Wallet)
		return
	}

	//读取内容
	content, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	//解码
	//panic: gob: type not registered for interface: elliptic.p256Curve
	gob.Register(elliptic.P256())

	decoder := gob.NewDecoder(bytes.NewReader(content))

	var wsLocal Wallets

	err = decoder.Decode(&wsLocal)
	if err != nil {
		log.Panic(err)
	}

	//ws = &wsLocal
	//对于结构来说，里面有map的，要指定赋值，不要再最外层直接赋值
	ws.WalletMap = wsLocal.WalletMap
}

func (ws *Wallets) ListAllAddresses() []string {
	var addresses []string
	//遍历钱包，将所有的key取出来返回
	for address := range ws.WalletMap {
		addresses = append(addresses, address)
	}

	return addresses
}

//从地址得到公钥哈希
func GetPubKeyFromAddress(address string) []byte {
	//1. 解码
	//2. 截取出公钥哈希：去除version（1字节），去除校验码（4字节）
	addressByte := base58.Decode(address) //25字节
	len := len(addressByte)

	pubKeyHash := addressByte[1 : len-4]

	return pubKeyHash
}
