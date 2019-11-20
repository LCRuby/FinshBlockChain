package main

import (
	"fmt"
	"os"
	"strconv"
)

//这是一个用来接收命令行参数并且控制区块链操作的文件
//CLI的成员是区块链
type CLI struct {
	bc *BlockChain
}

//反引号可以多行写,usge相当于一个命令使用说明
const Usage = `
	printChain               "正向打印区块链"
	printChainR 			"反向打印区块链"
	getBalcance  --address ADDRESS "获取地址的余额"
    send FROM TO AMOUNT MINER DATA "由FROM转AMOUNT给TO，由MINER挖矿，同时写入DATA"
	newWallet "创建一个钱包"
    listAddresses "列举所有的钱包地址"
`

//接受参数的动作，我们放到一个函数中

func (cli *CLI) Run() {

	//./block printChain
	//./block addBlock --data "HelloWorld"
	//1. 得到所有的命令
	args := os.Args
	if len(args) < 2 {
		fmt.Printf(Usage)
		return
	}

	//2. 分析命令
	cmd := args[1]
	switch cmd {

	case "printChain":
		fmt.Printf("打印区块\n")
		cli.PrintBlockChain()
	case "printChainR":
		fmt.Printf("反向打印区块链\n")
		cli.PrintBlockChainReverse()
	case "getBalance":
		fmt.Printf("h获取余额\n")
		if len(args) == 4 && args[2] == "--address" {
			adress := args[3]
			cli.GetBalance(adress)
		}
	case "send":
		fmt.Printf("转账开始...\n")
		if len(args) != 7 {
			fmt.Printf("参数个数错误，请检查！\n")
			fmt.Printf(Usage)
			return
		}
		//./block send FROM TO AMOUNT MINER DATA "由FROM转AMOUNT给TO，由MINER挖矿，同时写入DATA"
		from := args[2]
		to := args[3]
		amount, _ := strconv.ParseFloat(args[4], 64) //知识点，请注意
		miner := args[5]
		data := args[6]
		cli.Send(from, to, amount, miner, data)
	case "newWallet":
		fmt.Printf("创建新的钱包....\n")
		cli.NewWallet()
	case "listAddresses":
		fmt.Printf("列举所有地址...\n")
		cli.ListAddresses()
	default:
		fmt.Printf("无效的命令，请检查!\n")
		fmt.Printf(Usage)
	}
}
