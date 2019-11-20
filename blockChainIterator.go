package main

import (
	"bolt"
	"github.com/labstack/gommon/log"
)

//定义一个迭代器结构
type BlockChainIterator struct {
	db *bolt.DB //数据库
	//定义一个游标,用于不断的索引
	currentHashPointer []byte
}

//创建一个迭代器,区块链来调用
func (bc *BlockChain) NewIterator() *BlockChainIterator {
	return &BlockChainIterator{
		bc.db,
		//最初指向最后一个区块,随着next()的调用不断变化
		bc.tail,
	}
}

//迭代器是属于区块链的:即迭代器由区块链调用
//Next方式是属于迭代器的:即Next()由迭代器来调用
//1. 返回当前的区块
//2. 指针前移
func (it *BlockChainIterator) Next() *Block {
	var block Block
	it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			log.Panic("迭代器遍历时bucket不应该为空，请检查!")
		}
		blockTmp := bucket.Get(it.currentHashPointer)

		//解码动作
		block = Deserialize(blockTmp)
		//游标哈希左移
		it.currentHashPointer = block.PrevHash
		return nil
	})
	return &block
}
