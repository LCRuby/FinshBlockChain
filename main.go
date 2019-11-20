package main

func main() {
	bc := NewBlockChain("1DCm5Z3WhKVMqVTPFrBqhjDbjyF3ahtpgE")
	cli := CLI{bc}
	cli.Run()
	/*
		bc.AddBlock("503是第一个区块！")
		bc.AddBlock("未来的国家执掌者")

		//创建一个迭代器
		it := bc.NewIterator()
		//调用迭代器,返回每一个区块的数据

		for {
			block := it.Next()
			fmt.Printf("=====================\n")
			fmt.Printf("prevhash:=%x\n", block.PrevHash)
			fmt.Printf("hash:=%x\n", block.Hash)
			fmt.Printf("data:=%s\n", block.Data)
			if len(block.PrevHash) == 0 {
				fmt.Printf("区块遍历结束!!!\n")
				break
			}

		}
	*/

}
