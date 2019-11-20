package main

import (
	"fmt"
)

//正向打印区块链
func (cli *CLI) PrintBlockChain() {
	cli.bc.PrintChain()
	fmt.Printf("打印区块链完成\n")
}

//反向打印区块链
func (cli *CLI) PrintBlockChainReverse() {
	bc := cli.bc
	//创建迭代器
	it := bc.NewIterator()

	//调用迭代器，返回我们的每一个区块数据
	for {
		//返回区块，左移
		block := it.Next()
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		/*
			fmt.Printf("===========================\n\n")
			fmt.Printf("版本号: %d\n", block.Version)
			fmt.Printf("前区块哈希值: %x\n", block.PrevHash)
			fmt.Printf("梅克尔根: %x\n", block.MerkelRoot)
			timeFormat := time.Unix(int64(block.TimeStamp), 0).Format("2006-01-02 15:04:05")
			fmt.Printf("时间戳: %s\n", timeFormat)
			fmt.Printf("难度值(随便写的）: %d\n", block.Difficulty)
			fmt.Printf("随机数 : %d\n", block.Nonce)
			fmt.Printf("当前区块哈希值: %x\n", block.Hash)
			//fmt.Printf("区块数据 :%s\n", block.Data)
			fmt.Printf("区块数据 :%s\n", block.Transactions[0].TXInputs[0].PubKey)
		*/
		if len(block.PrevHash) == 0 {
			fmt.Printf("区块链遍历结束！")
			break
		}
	}
}

//查找出指定地址的所有余额
func (cli *CLI) GetBalance(address string) {
	//1.验证地址是否有效
	if !IsValidAddress(address) {
		fmt.Printf("地址无效 : %s\n", address)
		return
	}
	//2.从地址反向得到公钥哈希
	pubKeyHash := GetPubKeyFromAddress(address)
	utxos := cli.bc.FindUTXOs(pubKeyHash)

	total := 0.0
	for _, utxo := range utxos {
		total += utxo.Value
	}

	fmt.Printf("\"%s\"的余额为：%f\n", address, total)
}

//转账交易
func (cli *CLI) Send(from, to string, amount float64, miner, data string) {
	//fmt.Printf("from : %s\n", from)
	//fmt.Printf("to : %s\n", to)
	//fmt.Printf("amount : %f\n", amount)
	//fmt.Printf("miner : %s\n", miner)
	//fmt.Printf("data : %s\n", data)

	//地址校验
	if !IsValidAddress(from) {
		fmt.Printf("地址无效 from: %s\n", from)
		return
	}
	if !IsValidAddress(to) {
		fmt.Printf("地址无效 to: %s\n", to)
		return
	}
	if !IsValidAddress(miner) {
		fmt.Printf("地址无效 miner: %s\n", miner)
		return
	}
	//1. 创建挖矿交易
	coinbase := NewCoinBaseTX(miner, data)
	//2. 创建一个普通交易
	tx := NewTransaction(from, to, amount, cli.bc)
	if tx == nil {
		fmt.Printf("无效的交易\n")
		return
	}
	//3. 添加到区块

	cli.bc.AddBlock([]*Transaction{coinbase, tx})
	fmt.Printf("转账成功！\n")
}

//创建钱包
func (cli *CLI) NewWallet() {
	//ws := NewWallet()
	//addr := ws.NewAddress()
	ws := NewWallets()
	address := ws.CreateWallet()
	fmt.Printf("地址：%s\n", address)
	//fmt.Printf("公钥 :%v\n", ws.PublicKey)
	//fmt.Printf("adress:=%s", addr)

}
func (cli *CLI) ListAddresses() {
	ws := NewWallets()
	addresses := ws.ListAllAddresses()
	for _, address := range addresses {
		fmt.Printf("地址：%s\n", address)
	}
}
